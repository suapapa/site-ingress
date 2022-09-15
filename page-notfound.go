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
		Img:   "https://homin.dev/asset/image/iamfine.jpg",
		Msg:   "이 산이 아닌갑다",
		Links: []*Link{
			{
				Name: "ingress",
				Link: "/ingress",
				Desc: "대문으로 이동",
			},
		},
		LastWords: "<a href=\"/support\">대가없는 🥩 환영합니다</a>",
	}

	err = tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
