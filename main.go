package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/snabb/sitemap"

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

	flag.StringVar(&urlPrefix, "p", "", "set url prefix") // /ingress
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

	if len(urlPrefix) > 0 && urlPrefix[0] != '/' {
		urlPrefix = "/" + urlPrefix
	} else if urlPrefix == "" {
		urlPrefix = "/"
	}

	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add essential middleware for production
	router.Use(gin.Recovery()) // Panic recovery middleware

	// Configure trusted proxies for Kubernetes network ranges
	// These are the default CIDR ranges used in Kubernetes clusters
	router.SetTrustedProxies([]string{
		"10.0.0.0/8",     // Kubernetes cluster network
		"172.16.0.0/12",  // Kubernetes cluster network
		"192.168.0.0/16", // Kubernetes cluster network
		"127.0.0.1",      // Localhost
	})

	if urlPrefix != "/" {
		router.GET(urlPrefix+"/support", gin.WrapF(supportHandler))
		router.GET(urlPrefix+"/404", gin.WrapF(notfoundHandler))
		router.GET(urlPrefix, gin.WrapF(rootHandler))
		router.GET(urlPrefix+"/sitemap.xml", sitemapHandler)
		router.GET(urlPrefix+"/:path", redirectHandler)
	}
	router.GET("/404", gin.WrapF(notfoundHandler))
	router.GET("/support", gin.WrapF(supportHandler))
	router.GET("/sitemap.xml", sitemapHandler)
	router.GET("/:path", redirectHandler)
	router.GET("/", gin.WrapF(rootHandler))

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

	if dest[0] == '/' {
		dest = dest[1:]
	}

	staticAssets := map[string]string{
		"ads.txt":    "asset/ads.txt",
		"robots.txt": "asset/robots.txt",
	}

	if asset, ok := staticAssets[dest]; ok {
		c.File(asset)
		return
	}

	for _, link := range links {
		if link == nil {
			continue
		}

		if link.Name[0] == '/' {
			link.Name = link.Name[1:]
		}

		if link.Name == dest {
			// if link.RPLink != "" {
			// 	urlPath := c.Request.URL.Path
			// 	log.Printf("reverse proxy %s -> %s", urlPath, link.RPLink)
			// 	serveReverseProxy(c.Request.Context(), c.Writer, c.Request, link.RPLink, urlPath)
			// 	// serveReverseProxy(c.Request.Context(), c.Writer, c.Request, link.Link, link.RPLink)
			// } else {
			log.Printf("redirect %s -> %s", dest, link.Link)
			c.Redirect(http.StatusTemporaryRedirect, link.Link)
			// }
			return
		}
	}

	c.Redirect(http.StatusTemporaryRedirect, urlPrefix+"/404")
}

func sitemapHandler(c *gin.Context) {
	sm := sitemap.New()

	for _, link := range links {
		if link.SiteMap {
			sm.Add(&sitemap.URL{Loc: link.Link})
		}
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Status(http.StatusOK)
	sm.WriteTo(c.Writer)
}
