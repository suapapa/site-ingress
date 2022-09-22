package main

import (
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
	urlPrefix string
	httpPort  int
	linksConf string
)

func main() {
	log.Println("homin.dev ingress start")
	defer func() {
		log.Println("homin.dev ingress stop")
	}()

	flag.StringVar(&urlPrefix, "p", "/ingress", "set url prefix")
	flag.IntVar(&httpPort, "http", 8080, "set http port")
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "links")
	flag.Parse()

	if urlPrefix[0] != '/' {
		urlPrefix = "/" + urlPrefix
	}

	http.HandleFunc(urlPrefix, rootHandler)
	http.HandleFunc(urlPrefix+"/support", supportHandler)
	http.HandleFunc(urlPrefix+"/404", notfoundHandler)
	http.HandleFunc("/404", notfoundHandler)
	http.HandleFunc("/", rootHandler)
	// start HTTPServer
	go func() {
		log.Printf("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
