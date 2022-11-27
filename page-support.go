package main

import (
	_ "embed"
	"net/http"

	"github.com/suapapa/site-ingress/ingress"
)

func supportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "support-handler")
	defer span.End()
	// trace.SpanFromContext(ctx)

	c := &PageContent{
		Title: "💸 후원 💸",
		Img:   "https://homin.dev/asset/image/flex_01_360.jpg",
		Msg:   "사이트를 후원해 주세요",
		Links: []*ingress.Link{
			{
				Name: "REDBUBBLE",
				Link: "https://www.redbubble.com/people/suapapa/shop?asc=u",
				Desc: "🎁 굳즈샾 🎁",
			},
			{
				Name: "Buy Me a coffee",
				Link: "https://www.buymeacoffee.com/homin",
				Desc: "☕️ 충전해주기",
			},
			{
				Name: "Paypal",
				Link: "https://www.paypal.com/paypalme/suapapa",
				Desc: "페이팔 송금",
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
				Name: "/ingress",
				Link: "https://homin.dev/",
				Desc: "대문으로 이동",
			},
		},
	}

	err := tmplPage.Execute(w, c)
	if err != nil {
		log.Errorf("fail on support handler: %v", err)
	}
}
