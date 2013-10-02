package gopmda

import (
	"bufio"
	"log"
	"os"
	"testing"
)

var (
	_history_path = "history.json"
	_save_path    = "saved.txt"
	_drugs_log    = "drugpath.txt"
)

func TestDrugURLs(t *testing.T) {
	var _history DrugURLHistory
	log.Println("CodeTest start")
	var err error
	if _history, err = LoadDrugURLHistory(_history_path); err == nil {
		log.Println("drug url history loaded")
	} else {
		_history = NewDrugURLHistory()
		err = _history.Save(_history_path)
		if err != nil {
			t.Fatal(err)
		}
		log.Println("drug url history created")
	}
}

func TestDrugDownload(t *testing.T) {
	//DownloadAll()
}
