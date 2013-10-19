package gopmda

import (
	"bytes"
	"encoding/xml"
	"github.com/kokardy/saxlike"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	URL_UPDATE_FROM_1_WEEK_AGO  = URL(`http://www.info.pmda.go.jp/downfiles/ph/1week.html`)
	URL_UPDATE_FROM_2_WEEKS_AGO = URL(`http://www.info.pmda.go.jp/downfiles/ph/2week.html`)
	URL_UPDATE_FROM_1_MONTH_AGO = URL(`http://www.info.pmda.go.jp/downfiles/ph/tenpulist.html`)
	REGEXP_UPDATE_DRUG_URL      = regexp.MustCompile(`/go/pack/\d{7}[A-Z]\d{4}_\d_\d{2}/`)
	REGEXP_DELETE_DRUG_COMMENT  = regexp.MustCompile(`\d{7}[A-Z]\d{4}_\d_\d{2}`)
	PATH_UPDATE_DATE            = "update.xml"
)

type DeleteList []string

func (dl DeleteList) Delete() {
	for _, dirname := range dl {
		dirname = strings.Trim(dirname, " ")
		path := filepath.Join(SAVE_PATH, dirname)
		new_path := filepath.Join(SAVE_PATH, "deleted", dirname)
		err := os.Rename(path, new_path)
		if err != nil {
			log.Println("rename err:", err)
		} else {
			log.Prinln("delete:", dirname)
		}
	}
}

type UpdateList []DrugURL

func (ul UpdateList) Update() {
	for _, drug_url := range ul {
		if err := drug_url.Download(); err != ERR_ALREADY_EXISTS {
			time.Sleep(2 * time.Second)
		}
	}
	now := time.Now()
	if b, err := xml.Marshal(now); err != nil {
		log.Panicln("'Now' cannot be marshaled")
	} else {
		if w, err := os.OpenFile(PATH_UPDATE_DATE, os.O_CREATE|os.O_RDWR, 0666); err != nil {
			log.Panicln("cannot open ", PATH_UPDATE_DATE)
		} else {
			_, err = w.Write(b)
			if err != nil {
				log.Panicln("cannot write to ", PATH_UPDATE_DATE)
			}
		}
	}

}

func DeleteAndUpdate(i int) (deleteList DeleteList, updateList UpdateList, err error) {
	r := UpdateReader(i)
	deleteList, updateList, err = UpdateParse(r)
	return
}

func UpdateReader(i int) io.Reader {
	var update_url URL
	switch i {
	case 0:
		update_url = URL_UPDATE_FROM_1_WEEK_AGO
	case 1:
		update_url = URL_UPDATE_FROM_2_WEEKS_AGO
	default:
		update_url = URL_UPDATE_FROM_1_MONTH_AGO
	}

	buf := bytes.NewBufferString("")
	io.Copy(buf, update_url.GetAndDecode())
	b := buf.Bytes()

	return bytes.NewReader((bytes.Replace(b, []byte("</BODY>"), []byte(""), -1)))
}

type UpdateHandler struct {
	saxlike.VoidHandler
	inH2       bool
	isDelete   bool
	isUpdate   bool
	DeleteList []string
	UpdateList []DrugURL
}

func NewUpdateHandler() (uh *UpdateHandler) {
	uh = new(UpdateHandler)
	uh.inH2 = false
	uh.isDelete = false
	uh.isUpdate = false
	uh.DeleteList = make(DeleteList, 0, 0)
	uh.UpdateList = make(UpdateList, 0, 0)
	return
}

func (uh *UpdateHandler) StartElement(e xml.StartElement) {
	if "h2" == strings.ToLower(e.Name.Local) {
		uh.inH2 = true
	}
	if "a" == strings.ToLower(e.Name.Local) {
		if uh.isUpdate {
			for _, attr := range e.Attr {
				value := REGEXP_UPDATE_DRUG_URL.FindString(attr.Value)
				if "href" == strings.ToLower(attr.Name.Local) && value != "" {
					url := URL_ROOT + value
					uh.UpdateList = append(uh.UpdateList, DrugURL(url))
				}
			}
		}
	}
}

func (uh *UpdateHandler) EndElement(e xml.EndElement) {
	if "h2" == strings.ToLower(e.Name.Local) {
		uh.inH2 = false
	}
}

func (uh *UpdateHandler) CharData(c xml.CharData) {
	if uh.inH2 && "削除分" == string(c) {
		uh.isDelete = true
		uh.isUpdate = false
	} else if uh.inH2 && "掲載分" == string(c) {
		uh.isDelete = false
		uh.isUpdate = true
	}
}

func (uh *UpdateHandler) Comment(c xml.Comment) {
	value := strings.Trim(string(REGEXP_DELETE_DRUG_COMMENT.Find(c)), " ")
	if uh.isDelete && value != "" {
		uh.DeleteList = append(uh.DeleteList, string(c))
	}
}

func UpdateParse(r io.Reader) (deleteDrugs DeleteList, updateDrugs []DrugURL, err error) {
	uh := NewUpdateHandler()
	p := saxlike.NewParser(r, uh)
	p.SetHTMLMode()
	err = p.Parse()
	if err != nil {
		return
	}
	deleteDrugs = uh.DeleteList
	updateDrugs = uh.UpdateList
	return
}
