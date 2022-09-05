package main

import (
	"log"
	"net/http"
)

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	c := &PageContent{
		Title: "🚧 404 🚧",
		Img:   "https://homin.dev/img/iamfine",
		Msg:   "공사가 마무리되기 전에, 다른 곳들을 둘러보세요",
		Links: []*PageLink{
			{
				Title: "/Github",
				Link:  "/github",
				Desc:  "hobby projects",
			},
			{
				Title: "/Youtube",
				Link:  "/youtube",
				Desc:  "personal videos",
			},
			{
				Title: "/Resume",
				Link:  "/resume",
				Desc:  "online resume",
			},
		},
		LastWords: "<a href=\"/support\">대가없는 🥩 환영합니다</a>",
	}

	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
