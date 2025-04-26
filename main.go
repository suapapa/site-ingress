package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

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

	router := gin.Default()

	if urlPrefix != "/" {
		router.GET(urlPrefix+"/support", gin.WrapF(supportHandler))
		router.GET(urlPrefix+"/404", gin.WrapF(notfoundHandler))
		router.GET(urlPrefix, gin.WrapF(rootHandler))
		router.GET(urlPrefix+"/:path", redirectHandler)
	}
	router.GET("/404", gin.WrapF(notfoundHandler))
	router.GET("/support", gin.WrapF(supportHandler))
	router.GET("/", gin.WrapF(rootHandler))
	router.GET("/:path", redirectHandler)

	// start HTTPServer
	go func() {
		log.Infof("listening http on :%d", httpPort)
		if err := router.Run(fmt.Sprintf(":%d", httpPort)); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func redirectHandler(c *gin.Context) {
	dest := c.Param("path")
	if dest == "" {
		if urlPrefix != "" {
			c.Redirect(http.StatusTemporaryRedirect, urlPrefix)
		} else {
			c.Redirect(http.StatusTemporaryRedirect, "/")
		}
		return
	}

	for _, link := range links {
		if link.Name == dest {
			c.Redirect(http.StatusTemporaryRedirect, link.Link)
			log.Printf("redirect %s -> %s", dest, link.Link)
			return
		}
	}

	c.Redirect(http.StatusTemporaryRedirect, urlPrefix+"/404")
}
