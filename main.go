package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	SSL_CERT_FILE = "/etc/letsencrypt/live/homin.dev/fullchain.pem"
	SSL_KEY_FILE  = "/etc/letsencrypt/live/homin.dev/privkey.pem"
)

var (
	httpPort, httpsPort int
)

func main() {
	log.Println("homin.dev ingress start")
	defer log.Println("homin.dev ingress stop")

	flag.IntVar(&httpPort, "http", 80, "set http port")
	flag.IntVar(&httpsPort, "https", 443, "set https port")
	flag.Parse()

	http.HandleFunc("/", redirectHadler)
	http.HandleFunc("/img/", imgHandler)
	http.HandleFunc("/404", notfoundHandler)
	http.Handle("/.well-known/acme-challenge/", NewAcmeChallenge("/tmp/letsencrypt/"))

	// start HTTPServer
	go func() {
		log.Printf("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	startHTTPSServer()

	exitCh := make(chan any)
	<-exitCh
}

func startHTTPSServer() {
	// TODO: compare checksum of last cert
	if filesExist(SSL_CERT_FILE, SSL_KEY_FILE) {
		go func() {
			log.Printf("listening https on :%d", httpsPort)
			if err := http.ListenAndServeTLS(fmt.Sprintf(":%d", httpsPort), SSL_CERT_FILE, SSL_KEY_FILE, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}
