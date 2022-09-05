package main

import (
	"log"
	"net/http"
)

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
	}

	c := &PageContent{
		Title:     "🚧 404 🚧",
		Img:       "https://homin.dev/img/iamfine",
		Msg:       "공사가 마무리되기 전에, 다른 곳들을 둘러보세요",
		Links:     links,
		LastWords: "<a href=\"/support\">대가없는 🥩 환영합니다</a>",
	}

	err = tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
