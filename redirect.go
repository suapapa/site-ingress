package main

import (
	"log"
	"net/http"
	"path"
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	basePath := path.Base(r.URL.Path)
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}

	log.Printf("hit basePath, %s", basePath)
	if basePath == "/" {
		http.Redirect(w, r, "/ingress", http.StatusMovedPermanently)
		return
	}

	if basePath == "favicon.ico" {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(favicon)
		return
	}

	// redirect for external sites
	link, ok := redirects[basePath]
	if !ok {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.ReverseProxy {
		serveReverseProxy(link.Link, w, r)
		return
	}

	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}
