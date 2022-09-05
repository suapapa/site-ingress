package main

import (
	_ "embed"
	"log"
	"net/http"
	"text/template"
)

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("tmpl/nothound.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, struct{ Logo, SupportLink string }{
		Logo:        "/img/iamfine",
		SupportLink: "https://www.paypal.com/paypalme/suapapa",
	})
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
