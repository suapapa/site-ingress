package main

import (
	"log"
	"net/http"
	"os"
)

const (
	SSL_CERT_FILE = "cert/homin-dev.crt"
	SSL_KEY_FILE  = "cert/homin-dev.key"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Homin.dev ingress"))
	})
	http.Handle("/.well-known/acme-challenge/",
		http.FileServer(http.FileSystem(http.Dir("/tmp/letsencrypt/"))),
	)

	go func() {
		log.Println("starting http on :80")
		if err := http.ListenAndServe(":80", nil); err != nil {
			log.Fatal(err)
		}
	}()

	if filesExist(SSL_CERT_FILE, SSL_KEY_FILE) {
		go func() {
			log.Println("starting http on :443")
			if err := http.ListenAndServeTLS(":443", SSL_CERT_FILE, SSL_KEY_FILE, nil); err != nil {
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
