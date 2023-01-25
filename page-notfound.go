package main

import (
	"net/http"

	"github.com/suapapa/site-ingress/ingress"
)

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// _, span := tracer.Start(ctx, "notfound-handler")
	// defer span.End()
	log.Warn("not found handler called")

	c := &PageContent{
		Title: "🚧 404 🚧",
		Img:   "https://homin.dev/asset/image/panic_01_360.jpg",
		Msg:   "이 산이 아닌갑다",
		Links: []*ingress.Link{
			{
				Name: "ingress",
				Link: "/",
				Desc: "대문으로 이동",
			},
		},
		LastWords: "<a href=\"/support\">대가없는 🥩 환영합니다</a>",
	}

	w.WriteHeader(http.StatusNotFound)
	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Errorf("fail on not found handler: %v", err)
	}
}
