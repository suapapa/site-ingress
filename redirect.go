package main

import (
	"log"
	"net/http"
	"path"
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	basePath := path.Base(r.URL.Path)
	log.Printf("hit basePath, %s", basePath)
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}

	link, ok := redirects[basePath]
	if !ok {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return
	}

	if link.ReverseProxy {
		serveReverseProxy(link.Link, w, r)
		return
	}

	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}
