package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const (
	SSL_CERT_FILE = "/etc/letsencrypt/live/homin.dev/fullchain.pem"
	SSL_KEY_FILE  = "/etc/letsencrypt/live/homin.dev/privkey.pem"
)

var (
	//go:embed asset/favicon.ico
	favicon []byte

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
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(favicon)
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

	go func() {
		out, err := exec.Command("/create_ssl_cert.sh").Output()
		if err != nil {
			log.Printf("ERR: %v", err)
			notifyErrToTelegram(err)
			return
		}
		log.Printf("INFO: %s", out)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
