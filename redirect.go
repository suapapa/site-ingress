package main

import (
	"log"
	"net/http"
	"path"
)

var (
	redirects = map[string]string{
		// work-pool test. FQDN: <service>.<namespace>.svc.cluster.local
		"website":   "http://website.default.svc.cluster.local:8080",
		"resume":    "https://suapapa.github.io/resume/",
		"blog":      "http://suapapa.github.io/blog/",
		"github":    "https://github.com/suapapa",
		"youtube":   "https://www.youtube.com/c/HominLee",
		"instagram": "https://www.instagram.com/homin1227/",
	}
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	basePath := path.Base(r.URL.Path)
	log.Printf("hit basePath")

	site, ok := redirects[basePath]
	if !ok {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return
	}

	if basePath == "website" {
		log.Printf("reverse-proxy: %s", site)
		serveReverseProxy(site, w, r)
		return
	}

	http.Redirect(w, r, site, http.StatusMovedPermanently)
}
