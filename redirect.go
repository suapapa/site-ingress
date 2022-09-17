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

	if urlPath == "/" || urlPath == "" {
		http.Redirect(w, r, "/ingress", http.StatusMovedPermanently)
		return
	}

	if urlPath == "/favicon.ico" {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(favicon)
		return
	}

	// use first depth of path to sub-domain
	subDomain, _ := getSubdomain2(urlPath)

	// redirect for external sites
	link, ok := redirects[subDomain]
	if !ok {
		log.Printf("404 %s from %s", urlPath, r.RemoteAddr)
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RP {
		// log.Printf("rp: %s => '%s'", urlPath, path.Join(link.RPLink, subDomain, subPath))
		// TODO: cache proxy handlers?
		serveReverseProxy(
			link.RPLink,
			w, r,
		)
		return
	}

	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}

func getSubdomain(urlPath string) (string, string) {
	if urlPath == "" {
		return "", ""
	}

	urlPath = strings.Trim(urlPath, "/")
	paths := strings.Split(urlPath, "/")

	var subDomain, subPath string
	subDomain = paths[0]
	if len(paths[1:]) == 0 {
		subPath = "/"
	} else {
		subPath = "/" + strings.Join(paths[1:], "/")
	}

	return subDomain, subPath
}

func getSubdomain2(urlPath string) (string, string) {
	if len(urlPath) == 0 {
		return "", ""
	}

	if urlPath[0] == '/' {
		urlPath = urlPath[1:]
	}

	i := strings.Index(urlPath, "/")
	if i < 0 {
		return urlPath, "/"
	}

	return urlPath[:i], urlPath[i:]
}
