package gopmda

import (
	"log"
	"testing"
)

func TestUpdate(t *testing.T) {
	log.Println("update test start")

	deleteList, updateList, err := DeleteAndUpdate(3)
	if err != nil {
		t.Fatal("err:", err)
	}
	/*
		for _, d := range deleteList {
			log.Println("delete:", d)
		}
		for _, u := range updateList {
			log.Println("update:", u)
		}*/
	log.Println("deleteList:", len(deleteList))
	log.Println("UpdateList:", len(updateList))
}
