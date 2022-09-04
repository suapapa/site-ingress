package main

import (
	_ "embed"
	"log"
	"net/http"
	"path"
	"text/template"
)

var (
	tmpl *template.Template
)

var (
	//go:embed asset/iamfine.jpg
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

func imgHandler(w http.ResponseWriter, r *http.Request) {
	basePath := path.Base(r.URL.Path)
	log.Println("basePath: ", basePath)

	imgs := map[string][]byte{
		"iamfine": iamfineImg,
	}

	w.Header().Set("Content-Type", "image/jpg")
	w.Write(imgs[basePath])
}
