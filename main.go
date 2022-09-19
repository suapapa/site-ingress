package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	SSL_CERT_FILE = "/etc/letsencrypt/live/homin.dev/fullchain.pem"
	SSL_KEY_FILE  = "/etc/letsencrypt/live/homin.dev/privkey.pem"
)

var (
	//go:embed asset/favicon.ico
	//go:embed asset/ads.txt
	//go:embed asset/sitemap.xml
	efs embed.FS

	httpPort, httpsPort int
	linksConf           string
)

func main() {
	log.Println("homin.dev ingress start")
	defer log.Println("homin.dev ingress stop")

	notifyErrToTelegram(errors.New("start homin.dev ingress"))

	flag.IntVar(&httpPort, "http", 80, "set http port")
	flag.IntVar(&httpsPort, "https", 443, "set https port")
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "links")
	flag.Parse()

	http.HandleFunc("/ingress", ingressHandler)
	http.HandleFunc("/404", notfoundHandler)
	http.HandleFunc("/support", supportHandler)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		if b, err := efs.ReadFile("asset/favicon.ico"); err != nil {
			log.Printf("fail to read asset for %s, %v", r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "image/x-icon")
			w.Write(b)
		}
	})
	http.HandleFunc("/ads.txt", func(w http.ResponseWriter, r *http.Request) {
		if b, err := efs.ReadFile("asset/ads.txt"); err != nil {
			log.Printf("fail to read asset for %s, %v", r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.Write(b)
		}
	})
	http.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		if b, err := efs.ReadFile("asset/sitemap.xml"); err != nil {
			log.Printf("fail to read asset for %s, %v", r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text/xml")
			w.Write(b)
		}
	})

	http.HandleFunc("/", redirectHadler)

	http.Handle("/.well-known/acme-challenge/", NewAcmeChallenge("/tmp/letsencrypt/"))

	// start HTTPServer
	go func() {
		log.Printf("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	go startHTTPSServer()
	go startPortFoward()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
