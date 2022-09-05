package main

import (
	"log"
	"net/http"
)

func ingressHandler(w http.ResponseWriter, r *http.Request) {
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}

	c := &PageContent{
		Title:     "🚧 Ingress 🚧",
		Img:       "https://homin.dev/img/iamfine",
		Links:     links,
		LastWords: "<a href=\"/support\">대가없는 🥩 환영합니다</a>",
	}

	err = tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
