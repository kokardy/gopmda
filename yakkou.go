package gopmda

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

const ()

var (
	YAKKOU_CODE_REGEXP = regexp.MustCompile(`<option value="(\d{3})">`)
	DRUG_URL_REGEXP    = regexp.MustCompile(`href="(/go/pack/.+?)/" target="_new">(.*?)</A>`)
	ERROR_NO_MATCH     = errors.New("no match")
)

func scan_yakkou_code(line string) (code YAKKOU, err error) {
	match := YAKKOU_CODE_REGEXP.FindStringSubmatch(line)
	if len(match) < 2 {
		err = ERROR_NO_MATCH
	} else {
		code = YAKKOU(match[1])
	}
	return
}

func scan_url(line string) (url, name string, err error) {
	match := DRUG_URL_REGEXP.FindStringSubmatch(line)
	if len(match) < 3 {
		err = ERROR_NO_MATCH
		return
	}
	url, name = match[1], match[2]

	return
}

func get_yakkou_codes() (codes []YAKKOU) {
	codes = make([]YAKKOU, 0, 100)
	r := get(URL_PSEARCH_KENSAKU)
	var preline string
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			break
		}
		preline += string(line)
		if isPrefix {
			continue
		} else {
			code, err := scan_yakkou_code(preline)
			if err == nil {
				codes = append(codes, code)
			}
			preline = ""
		}

	}
	//log.Println(codes)
	return
}

type YAKKOU string

func (code YAKKOU) DrugUrls() (urls []DrugURL) {
	urls = make([]DrugURL, 0, 100)
	search_result_url := fmt.Sprintf(URL_PSEARCH+`PackinsSearch`+`?effect=%s&count=1000`, code)
	res, err := http.Get(search_result_url)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(DECODER.NewReader(res.Body))
	var preline string
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			break
		}
		preline += string(line)
		if isPrefix {
			continue
		} else {
			url, _, err := scan_url(preline)
			if err == nil {
				urls = append(urls, NewDrugURL(URL(URL_ROOT+url)))
				//log.Println(url, name)
			}
			preline = ""
		}

	}
	//log.Println(urls)
	return
}
