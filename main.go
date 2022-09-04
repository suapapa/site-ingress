package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	SSL_CERT_FILE = "cert/homin-dev.crt"
	SSL_KEY_FILE  = "cert/homin-dev.key"
)

var (
	httpPort, httpsPort int
)

func main() {
	flag.IntVar(&httpPort, "http", 80, "set http port")
	flag.IntVar(&httpsPort, "https", 443, "set https port")
	flag.Parse()

	http.HandleFunc("/", notfoundHandler)
	http.HandleFunc("/img/iamfine", imgIamfineHandler)
	http.Handle("/.well-known/acme-challenge/",
		http.FileServer(http.FileSystem(http.Dir("/tmp/letsencrypt/"))),
	)

	go func() {
		log.Printf("starting http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	if filesExist(SSL_CERT_FILE, SSL_KEY_FILE) {
		go func() {
			log.Printf("starting http on :%d", httpsPort)
			if err := http.ListenAndServeTLS(fmt.Sprintf(":%d", httpsPort), SSL_CERT_FILE, SSL_KEY_FILE, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}

	exitCh := make(chan any)
	<-exitCh
}

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}
