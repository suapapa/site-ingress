package main

import (
	_ "embed"
	"net/http"

	"github.com/suapapa/site-ingress/ingress"
)

func supportHandler(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// _, span := tracer.Start(ctx, "support-handler")
	// defer span.End()

	c := &PageContent{
		Title: "๐ธ ํ์ ๐ธ",
		Img:   "https://homin.dev/asset/image/flex_01_360.jpg",
		Msg:   "์ฌ์ดํธ๋ฅผ ํ์ํด ์ฃผ์ธ์",
		Links: []*ingress.Link{
			{
				Name: "REDBUBBLE",
				Link: "https://www.redbubble.com/people/suapapa/shop?asc=u",
				Desc: "๐ ๊ตณ์ฆ์พ ๐",
			},
			{
				Name: "Buy Me a coffee",
				Link: "https://www.buymeacoffee.com/homin",
				Desc: "โ๏ธ ์ถฉ์ ํด์ฃผ๊ธฐ",
			},
			{
				Name: "Paypal",
				Link: "https://www.paypal.com/paypalme/suapapa",
				Desc: "ํ์ดํ ์ก๊ธ",
			},
			{
				Name: "KakaoPay QR",
				Link: "https://homin.dev/asset/image/kakaopay.jpg",
				Desc: "์นด์นด์คํ์ด QR <- ๋ฐ์คํฌํ์์๋ ์ฌ๊ธฐ๋ก",
			},
			{
				Name: "KakaoPay",
				Link: "https://qr.kakaopay.com/281006011000002416281797",
				Desc: "์นด์นด์คํ์ด ์คํ <- ๋ชจ๋ฐ์ผ์์๋ ์ฌ๊ธฐ๋ก",
			},
			{
				Name: "/ingress",
				Link: "https://homin.dev/",
				Desc: "๋๋ฌธ์ผ๋ก ์ด๋",
			},
		},
	}

	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Errorf("fail on support handler: %v", err)
	}
}
