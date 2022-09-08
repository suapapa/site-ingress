package main

import (
	"log"
	"net/http"
	"strings"
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path //"/blog/favicon.ico"
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}

	// log.Printf("hit basePath, %s", urlPath)
	if urlPath == "/" {
		http.Redirect(w, r, "/ingress", http.StatusMovedPermanently)
		return
	}

	if urlPath == "/favicon.ico" {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(favicon)
		return
	}

	subDomain, subPath := genFakeSubdomain(urlPath)

	// redirect for external sites
	link, ok := redirects[subDomain]
	if !ok {
		log.Printf("hit %s from %s", urlPath, r.RemoteAddr)
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RP {
		log.Printf("rp: %s => '%s' (sd=%s)", urlPath, link.RPLink+subPath, subDomain)
		// TODO: cache proxy handlers?
		serveReverseProxy(
			link.RPLink+subPath,
			w, r,
		)
		return
	}

	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}

func genFakeSubdomain(urlPath string) (string, string) {
	urlPath = strings.Trim(urlPath, "/")
	paths := strings.Split(urlPath, "/")

	var subDomain, subPath string
	subDomain = paths[0]
	if len(paths[1:]) == 0 {
		subPath = ""
	} else {
		subPath = "/" + strings.Join(paths[1:], "/")
	}

	return subDomain, subPath
}
