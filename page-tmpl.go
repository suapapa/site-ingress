package main

import (
	_ "embed"
	"text/template"

	"github.com/suapapa/site-ingress/ingress"
)

var (
	//go:embed tmpl/page.tmpl
	tmplPageStr string
	tmplPage    *template.Template
)

func init() {
	var err error
	tmplPage, err = template.New("page").Parse(tmplPageStr)
	if err != nil {
		panic(err)
	}
}

type PageContent struct {
	Title     string
	Img       string
	Msg       string
	Links     []*ingress.Link
	LastWords string
}
