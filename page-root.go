package main

import (
	"net/http"

	"go.opentelemetry.io/otel"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer(tracerName).Start(r.Context(), "ingress-root-page")
	defer span.End()

	links, err := getLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}

	c := &PageContent{
		Title:     "🔥 대문 🔥",
		Img:       "https://homin.dev/asset/image/dungeon_01_360.jpg",
		Msg:       randLine(),
		Links:     links,
		LastWords: "<a href=\"https://homin.dev/blog/post/20220908_homin-dev_with_k8s/\">사이트 소개</a>",
	}

	err = tmplPage.Execute(w, c)
	if err != nil {
		log.Errorf("fail on root handler: %v", err)
	}
}
