package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/suapapa/site-ingress/ingress"
)

const (
	SSL_CERT_FILE = "/etc/letsencrypt/live/homin.dev/fullchain.pem"
	SSL_KEY_FILE  = "/etc/letsencrypt/live/homin.dev/privkey.pem"

	// otplEP = "simplest-collector.default.svc.cluster.local:4317"
)

var (
	programName = "site-ingress"
	programVer  = "dev"

	urlPrefix string
	httpPort  int
	linksConf string
	debug     bool

	links []*ingress.Link
)

func main() {
	log.WithField("alert", "telegram").Infof("homin.dev ingress start")
	defer func() {
		log.WithField("alert", "telegram").Infof("homin.dev ingress stop")
	}()

	flag.StringVar(&urlPrefix, "p", "/ingress", "set url prefix")
	flag.IntVar(&httpPort, "http", 8080, "set http port")
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "links")
	flag.BoolVar(&debug, "d", false, "print debug logs")
	flag.Parse()

	// ctx := context.Background()
	// tp := initTracerProvider(ctx, otplEP)
	// defer func() {
	// 	if err := tp.Shutdown(ctx); err != nil {
	// 		log.Errorf("Error shutting down tracer provider: %v", err)
	// 	}
	// }()

	// // mp := initMeterProvider(ctx, otplEP)
	// // defer func() {
	// // 	if err := mp.Shutdown(ctx); err != nil {
	// // 		log.Errorf("Error shutting down meter provider: %v", err)
	// // 	}
	// // }()

	// tracer = tp.Tracer(programName)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	// otel.SetTracerProvider(tp)

	var err error
	if links, err = getLinks(linksConf); err != nil {
		log.Fatalf("fail to read links conf: %v", err)
		os.Exit(-1)
	}

	if urlPrefix[0] != '/' {
		urlPrefix = "/" + urlPrefix
	}

	if urlPrefix != "/" {
		http.HandleFunc(urlPrefix+"/support", supportHandler)
		http.HandleFunc(urlPrefix+"/404", notfoundHandler)
		http.HandleFunc(urlPrefix, rootHandler)
	}
	http.HandleFunc("/404", notfoundHandler)
	http.HandleFunc("/support", supportHandler)
	http.HandleFunc("/", rootHandler)

	// start HTTPServer
	go func() {
		log.Infof("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	ctx := context.Background()
	// watch lins conf file's chage
	go func(ctx context.Context) {
		fsW, err := fsnotify.NewWatcher()
		if err != nil {
			log.Errorf("fail to create fsnotify watcher: %v", err)
			return
		}

		if err := fsW.Add(linksConf); err != nil {
			log.Errorf("fail to add file to fsnotify watcher: %v", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-fsW.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Infof("links conf file changed: %s", event.Name)
					if links, err = getLinks(linksConf); err != nil {
						log.Errorf("fail to read links conf: %v", err)
					}
				}
			case err := <-fsW.Errors:
				log.Errorf("fsnotify error: %v", err)
			}
		}
	}(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
