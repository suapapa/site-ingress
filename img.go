package main

import (
	_ "embed"
	"net/http"
	"path"
)

var (
	//go:embed asset/iamfine.jpg
	iamfineImg []byte

	//go:embed asset/kakaopay.jpg
	kakaopayImg []byte
)

func imgHandler(w http.ResponseWriter, r *http.Request) {
	basePath := path.Base(r.URL.Path)
	// log.Println("basePath: ", basePath)

	imgs := map[string][]byte{
		"iamfine":  iamfineImg,
		"kakaopay": kakaopayImg,
	}

	data, ok := imgs[basePath]
	if !ok {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return
	}

	w.Header().Set("Content-Type", "image/jpg")
	w.Write(data)
}
