package gopmda

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	REGEXP_PHARMACEUTICAL_REFERENCE = regexp.MustCompile(`href="(/downfiles/ph/PDF/.*?pdf)"`)
	REGEXP_INTERVIEW_FORM           = regexp.MustCompile(`href="(/go/interview/.*?)"`)
	REGEXP_SGML                     = regexp.MustCompile(`href="(/downfiles/ph/SGM/.*?)"`)
	ERR_SCAN_LINK                   = errors.New("scan_link failed.")
	ERR_ALREADY_EXISTS              = errors.New("already exists")
)

func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		log.Panicln("unexpected err:", err)
	}
	return false
}

func scan_link(line string) (filetype string, url URL, err error) {
	exps := []*regexp.Regexp{
		REGEXP_PHARMACEUTICAL_REFERENCE,
		REGEXP_INTERVIEW_FORM,
		REGEXP_SGML,
	}
	filetypes := []string{
		"PR",
		"IF",
		"SGML",
	}
	for i, exp := range exps {
		matches := exp.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		} else {
			filetype = filetypes[i]
			url = URL(URL_ROOT + matches[1])
			return
		}
	}
	err = ERR_SCAN_LINK
	return
}

type DrugURL URL

func NewDrugURL(url URL) DrugURL {
	return DrugURL(url)
}

func (url DrugURL) foot() *bufio.Reader {
	q := map[string]string{"view": "foot"}
	return URL(url).Query(q)
}
func (url DrugURL) footURL() URL {
	u := URL(url + "?view=foot")
	return u
}
func (url DrugURL) tocURL() URL {
	u := URL(url + "?view=toc")
	return u
}
func (url DrugURL) bodyURL() URL {
	u := URL(url + "?view=body")
	return u
}

func (url DrugURL) Title() (title string) {
	reader := URL(url).Get()
	title_getter := NewTitleGetter()
	title_getter.Parse(reader)
	title = title_getter.Title
	return
}

func titlize(pre_title string) (title string) {
	title = strings.Trim(pre_title, "\r\n＊/*※　　")
	return
}
func (url DrugURL) Titles() (titles []string) {
	title := url.Title()
	titles = make([]string, 0, 1)
	for _, pre_title := range strings.Split(title, "／") {
		titles = append(titles, titlize(pre_title))
	}
	drugs_log, err := os.OpenFile(DRUG_LOG, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("drug path file open error: %s", err)
	}
	defer drugs_log.Close()

	for _, title := range titles {
		drugs_log.WriteString(title + "\t" + string(url.DrugPath()) + "\n")
	}
	return
}

func (url DrugURL) Links() (links map[string]URL) {
	links = make(map[string]URL)
	r := url.foot()
	var current_line string
	links["toc"] = url.tocURL()
	links["body"] = url.bodyURL()
	//links["foot"] = url.footURL()
	//footはSGMLとPDFのリンク先が一致しないのでいらない
	//というか自前で作らないとダメ
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			break
		}
		current_line += string(line)
		if isPrefix {
			continue
		}
		filetype, link, err := scan_link(current_line)
		current_line = ""
		if err != nil {
			continue
		}
		links[filetype] = link
	}
	return
}

func (durl DrugURL) IsFinished() bool {
	drug_path := durl.DrugPath()
	pr_path := filepath.Join(drug_path, "PR.pdf")
	sgml_path := filepath.Join(drug_path, "SGML.zip")
	path_list := []string{
		drug_path,
		pr_path,
		sgml_path,
	}
	for _, path := range path_list {
		if !Exists(path) {
			return false
		}
	}
	return true
}

func (durl DrugURL) Download() (err error) {
	log.Println("download:", durl)
	drug_path := durl.DrugPath()
	if durl.IsFinished() {
		log.Println("already exists", durl)
		return ERR_ALREADY_EXISTS
	}
	urls := durl.Links()
	for filetype, url := range urls {
		var path string
		switch filetype {
		case "PR":
			path = filepath.Join(drug_path, "PR.pdf")
		case "IF":
			path = filepath.Join(drug_path, "IF.pdf")
		case "SGML":
			path = filepath.Join(drug_path, "SGML.zip")
		case "foot":
			path = filepath.Join(drug_path, "foot.html")
		case "toc":
			path = filepath.Join(drug_path, "toc.html")
		case "body":
			path = filepath.Join(drug_path, "body.html")
		default:
			log.Panicf("unknown filetype '%s'", filetype)

		}
		br := url.GetRaw()
		dir := filepath.Dir(path)
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			log.Panicf("%v \ncannot create a directory: %s", err, dir)
		}
		fw, err := os.Create(path)
		if err != nil {
			log.Panicf("%v \ncannot create a file: %s", err, path)

		}
		defer fw.Close()
		_, err = io.Copy(fw, br)
		if err != nil {
			log.Panicf("%v \ncannot copy a file: %s", err, path)
		}
	}
	return
}
func (durl DrugURL) DrugPath() string {
	pathes := strings.Split(string(durl), "/")
	for i := len(pathes) - 1; i >= 0; i-- {
		if len(pathes[i]) > 0 {
			return filepath.Join(SAVE_PATH, pathes[i])
		}
	}
	panic(fmt.Sprintf("cannot get path from url: %s", string(durl)))
}

func DrugURLGenerator() (ch chan DrugURL) {
	ch = make(chan DrugURL, 500)
	yakkou_list := get_yakkou_codes()
	go func() {
		for _, yakkou := range yakkou_list {
			log.Println("yakkou code", yakkou)
			urls := yakkou.DrugUrls()
			for _, url := range urls {
				ch <- url
			}
		}
		close(ch)
	}()
	return ch
}

type DrugURLHistory map[DrugURL]bool

func NewDrugURLHistory() DrugURLHistory {
	history := make(DrugURLHistory)
	for url := range DrugURLGenerator() {
		history[url] = false
	}
	return history
}

func LoadDrugURLHistory(fpath string) (his DrugURLHistory, err error) {
	var pre_his map[string]bool
	his = make(DrugURLHistory)
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &pre_his)
	if err != nil {
		panic(err)
	}
	for key, value := range pre_his {
		his[DrugURL(key)] = value
	}
	return
}

func (his DrugURLHistory) Save(fpath string) (err error) {
	b, err := json.Marshal(his)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(fpath, b, 0666)
	log.Printf("save history:%s\n", fpath)
	return
}
