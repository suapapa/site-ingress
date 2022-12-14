package main

import (
	"net/http"

	"github.com/suapapa/site-ingress/ingress"
)

func notfoundHandler(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// _, span := tracer.Start(ctx, "notfound-handler")
	// defer span.End()

	c := &PageContent{
		Title: "π§ 404 π§",
		Img:   "https://homin.dev/asset/image/panic_01_360.jpg",
		Msg:   "μ΄ μ°μ΄ μλκ°λ€",
		Links: []*ingress.Link{
			{
				Name: "ingress",
				Link: "/",
				Desc: "λλ¬ΈμΌλ‘ μ΄λ",
			},
		},
		LastWords: "<a href=\"/support\">λκ°μλ π₯© νμν©λλ€</a>",
	}

	w.WriteHeader(http.StatusNotFound)
	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Errorf("fail on not found handler: %v", err)
	}
}
