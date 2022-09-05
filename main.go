package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

const (
	SSL_CERT_FILE = "/etc/letsencrypt/live/homin.dev/fullchain.pem"
	SSL_KEY_FILE  = "/etc/letsencrypt/live/homin.dev/privkey.pem"
)

var (
	httpPort, httpsPort int
	linksConf           string
)

func main() {
	log.Println("homin.dev ingress start")
	defer log.Println("homin.dev ingress stop")

	flag.IntVar(&httpPort, "http", 80, "set http port")
	flag.IntVar(&httpsPort, "https", 443, "set https port")
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "links")
	flag.Parse()

	http.HandleFunc("/ingress", ingressHandler)
	http.HandleFunc("/404", notfoundHandler)
	http.HandleFunc("/support", supportHandler)

	http.HandleFunc("/", redirectHadler)
	http.HandleFunc("/img/", imgHandler)

	http.Handle("/.well-known/acme-challenge/", NewAcmeChallenge("/tmp/letsencrypt/"))

	// start HTTPServer
	go func() {
		log.Printf("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	go startHTTPSServer()

	exitCh := make(chan any)
	<-exitCh
}
