package gopmda

import (
	"bufio"
	"bytes"
	"code.google.com/p/mahonia"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	SAVE_LOG_FILE       = "save.log"
	SAVE_PATH           = "save"
	URL_ROOT            = "http://www.info.pmda.go.jp"
	URL_PSEARCH         = URL_ROOT + "/psearch/"
	URL_PSEARCH_KENSAKU = URL_PSEARCH + "html/menu_tenpu_kensaku.html"
	HISTORY_PATH        = "history.json"
	SAVED_LIST_PATH     = "tmp_saved.txt"
	DRUG_LOG            = "drugpath.txt"
)

var (
	w, _    = os.Open(SAVE_LOG_FILE)
	SAVELOG = log.New(w, "", 0)
	DECODER = mahonia.NewDecoder("euc-jp")
)

type URL string

func (url URL) Get() *bufio.Reader {
	return get(string(url))
}

func (url URL) GetAndDecode() io.Reader {
	return getanddecode(string(url))
}

func (url URL) GetRaw() io.Reader {
	return getraw(string(url))
}

func (url URL) Query(m map[string]string) *bufio.Reader {
	buf := bytes.NewBufferString(string(url) + "?")
	for k, v := range m {
		buf.WriteString(fmt.Sprintf("%s=%s", k, v))
	}
	qurl := buf.String()
	return get(qurl)
}

func get(url string) *bufio.Reader {
	pre_r := getraw(url)
	r := bufio.NewReader(DECODER.NewReader(pre_r))
	return r
}

func getanddecode(url string) io.Reader {
	pre_r := getraw(url)
	r := DECODER.NewReader(pre_r)
	return r
}

func getraw(url string) io.Reader {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	r := bytes.NewBufferString("")
	io.Copy(r, res.Body)
	return r
}

func LoadSaveData(path string) []DrugURL {
	durl_list := make([]DrugURL, 0, 0)
	r, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("tmp_saved_file doesn't exist")
		} else {
			log.Panicln("err:", err)
		}
	}
	defer r.Close()
	br := bufio.NewReader(r)
	preline := ""
	for {
		line, isPrefix, err := br.ReadLine()
		if err != nil {
			break
		}
		preline += string(line)
		if !isPrefix {
			durl_list = append(durl_list, DrugURL(preline))
			preline = ""
		}
	}
	return durl_list
}

func DownloadAll() {
	log.Println("download start")
	_history, _ := LoadDrugURLHistory(HISTORY_PATH)
	durls := LoadSaveData(SAVED_LIST_PATH)
	for _, durl := range durls {
		_history[durl] = true
	}

	save, err := os.OpenFile(SAVED_LIST_PATH, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("save file open error:", err)

	}
	defer save.Close()

	drugs_log, err := os.OpenFile(DRUG_LOG, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("drug path file open error:", err)
	}
	defer drugs_log.Close()

	for durl, finished := range _history {
		log.Println(durl, finished)
		if !finished {
			titles := durl.Titles()
			log.Printf("title: %v", titles)
			links := durl.Links()
			if len(links) > 0 {
				log.Printf("len: %d ", len(links))
				for key, link := range links {
					log.Printf(`%s:"%v"`, key, string(link))
				}
				if err := durl.Download(); err == ERR_ALREADY_EXISTS {
					for _, title := range titles {
						drugs_log.WriteString(title + "\t" + string(durl.DrugPath()) + "\n")
					}
				} else {
					_, err := save.WriteString(string(durl) + "\n")
					if err != nil {
						panic(err)
					}
					for _, title := range titles {
						drugs_log.WriteString(title + "\t" + string(durl.DrugPath()) + "\n")
					}
					time.Sleep(1 * time.Second)
				}
			}
		}
	}
}
