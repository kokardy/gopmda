package gopmda

import (
	"encoding/xml"
	"github.com/kokardy/saxlike"
	"io"
	"strings"
)

func NewTitleGetter() *TitleGetterHandler {
	return &TitleGetterHandler{&saxlike.VoidHandler{}, "", false, false}
}

type TitleGetterHandler struct {
	*saxlike.VoidHandler
	Title   string
	mode    bool
	scanned bool
}

func (getter *TitleGetterHandler) StartElement(e xml.StartElement) {
	if strings.ToLower(e.Name.Local) == "title" {
		getter.mode = true
	}
}

func (getter *TitleGetterHandler) EndElement(e xml.EndElement) {
	if strings.ToLower(e.Name.Local) == "title" {
		getter.mode = false
		getter.scanned = true
	}
}

func (getter *TitleGetterHandler) CharData(char xml.CharData) {
	if getter.mode && !getter.scanned {
		getter.Title = string([]byte(char))
	}
}

func (getter *TitleGetterHandler) Parse(input io.Reader) {
	parser := saxlike.NewParser(input, getter)
	parser.SetHTMLMode()
	parser.Parse()
}
