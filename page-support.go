package main

import (
	_ "embed"
	"log"
	"net/http"
)

func supportHandler(w http.ResponseWriter, r *http.Request) {
	c := &PageContent{
		Title: "💸 후원 💸",
		Img:   "https://homin.dev/asset/image/gb.jpg",
		Msg:   "사이트를 후원해 주세요",
		Links: []*Link{
			{
				Name: "Buy Me a coffee",
				Link: "https://www.buymeacoffee.com/homin",
				Desc: "☕️ 충전해주기",
			},
			{
				Name: "KakaoPay QR",
				Link: "https://homin.dev/asset/image/kakaopay.jpg",
				Desc: "카카오페이 QR <- 데스크탑에서는 여기로",
			},
			{
				Name: "KakaoPay",
				Link: "https://qr.kakaopay.com/281006011000002416281797",
				Desc: "카카오페이 실행 <- 모바일에서는 여기로",
			},
			{
				Name: "Paypal",
				Link: "https://www.paypal.com/paypalme/suapapa",
				Desc: "페이팔 송금",
			},
			{
				Name: "/ingress",
				Link: "https://homin.dev/",
				Desc: "대문으로 이동",
			},
		},
	}

	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
}
