package main

import (
	"github.com/kokardy/gopmda"
	"log"
)

func main() {
	gopmda.DownloadAll()
	deleteList, UpdateList, err := gopmda.DeleteAndUpdate(1)
  if err != nil{
    log.Panicln("err:", err)
  }
	deleteList.Delete()
	UpdateList.Update()
}
