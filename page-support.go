package main

import (
	_ "embed"
	"log"
	"net/http"
)

func supportHandler(w http.ResponseWriter, r *http.Request) {
	c := &PageContent{
		Title: "🥩 후원 🥩",
		Img:   "/img/iamfine",
		Msg:   "사이트 유지 비용을 후원해 주세요",
		Links: []*PageLink{
			{
				Title: "KakaoPay",
				Link:  "/img/kakaopay",
				Desc:  "카카오페이 QR코드",
			},
			{
				Title: "Paypal",
				Link:  "https://www.paypal.com/paypalme/suapapa",
				Desc:  "페이팔 송금",
			},
		},
	}

	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
