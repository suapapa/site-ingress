package main

import (
	"log"
	"net/http"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	switch urlPath {
	case urlPrefix, urlPrefix + "ingress":
		err := updateLinks()
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}

		c := &PageContent{
			Title:     "🔥 대문 🔥",
			Img:       "https://homin.dev/asset/image/ingress.jpg",
			Msg:       "어디로 가야하죠 아죠씨",
			Links:     links,
			LastWords: "<a href=\"https://homin.dev/blog/post/20220908_homin-dev_with_k8s/\">사이트 소개</a>",
		}

		err = tmplPage.Execute(w, c)
		if err != nil {
			log.Printf("ERR: %v", err)
		}
	case urlPrefix + "support":
		supportHandler(w, r)
	default:
		notfoundHandler(w, r)
	}
}
