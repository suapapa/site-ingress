package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// log.Printf("serveRP: %s", target)
	// parse the url
	url, err := url.Parse(target)
	if err != nil {
		log.Printf("ERR!: %v", err)
	}

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
