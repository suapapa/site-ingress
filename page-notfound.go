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
		Title: "🚧 404 🚧",
		Img:   "/img/iamfine",
		Msg:   "이 산이 아닌갑다",
		Links: []*Link{
			{
				Name: "대문",
				Link: "/ingress",
				Desc: "다른 페이지들로 이동",
			},
		},
		LastWords: "<a href=\"/support\">대가없는 🥩 환영합니다</a>",
	}

	err = tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
