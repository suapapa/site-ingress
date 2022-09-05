package main

import (
	_ "embed"
	"log"
	"net/http"
	"text/template"
)

func supportHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	tmpl, err := template.ParseFiles("tmpl/support.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, struct{ Logo, SupportLink string }{
		Logo: "/img/iamfine",
	})
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
