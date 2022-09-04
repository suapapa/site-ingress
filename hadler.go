package main

import (
	_ "embed"
	"log"
	"net/http"
	"text/template"
)

var (
	tmpl *template.Template
)

var (
	//go:embed asset/iamfine.png
	iamfineImg []byte

	//go:embed asset/notfound.tmpl
	tmpl404Html string
)

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if tmpl == nil {
		tmpl, err = template.New("404").Parse(tmpl404Html)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tmpl.Execute(w, struct{ Logo, SupportLink string }{
		Logo:        "/img/iamfine",
		SupportLink: "https://www.paypal.com/paypalme/suapapa",
	})
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}

func imgIamfineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(iamfineImg)
}
