package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sort"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/snabb/sitemap"

	"github.com/suapapa/site-ingress/internal/ingress"
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

	siteLinks      ingress.Links
	redirectByPath map[string]string // tidyPath(link.Name) -> target URL (first match by sorted prefix)
)

func main() {
	log.WithField("alert", "telegram").Infof("homin.dev ingress start")
	defer func() {
		log.WithField("alert", "telegram").Infof("homin.dev ingress stop")
	}()

	flag.IntVar(&httpPort, "http", 8080, "set http port")
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "links")
	flag.BoolVar(&debug, "d", false, "print debug logs")
	flag.Parse()

	var site *ingress.Site
	var err error
	if site, err = ingress.LoadSiteFromFile(linksConf); err != nil {
		log.Fatalf("fail to read links conf: %v", err)
		os.Exit(-1)
	}

	if site == nil {
		log.Fatalf("fail to read links conf: %v", err)
		os.Exit(-1)
	}
	siteLinks = site.Links
	says = site.Says
	redirectByPath = buildRedirectIndex(site.Links)

	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add essential middleware for production
	router.Use(gin.Recovery()) // Panic recovery middleware

	// Configure trusted proxies for Kubernetes network ranges
	router.SetTrustedProxies([]string{
		"10.0.0.0/8",     // Kubernetes cluster network
		"172.16.0.0/12",  // Kubernetes cluster network
		"192.168.0.0/16", // Kubernetes cluster network
		"127.0.0.1",      // Localhost
	})

	// API
	router.GET("/api/links", func(c *gin.Context) {
		showHides := c.Query("show_hides") == "true"
		prefix := tidyPath(c.Query("prefix"))

		list := siteLinks[prefix]
		resp := make([]*ingress.Link, 0, len(list))
		for _, l := range list {
			if !l.Hide || showHides {
				resp = append(resp, l)
			}
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/api/fish", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetRandomSay())
	})
	router.Static("/assets", "/assets")
	router.GET("/sitemap.xml", sitemapHandler)
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
	dest := tidyPath(c.Param("path"))

	if debug {
		log.Infof("redirect handler: path=%s dest=%s", c.Request.URL.Path, dest)
	}

	target, ok := redirectByPath[dest]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if debug {
		log.Infof("redirect %s -> %s", dest, target)
	}
	c.Redirect(http.StatusTemporaryRedirect, target)
}

func sitemapHandler(c *gin.Context) {
	sm := sitemap.New()

	for prefix, ls := range siteLinks {
		for _, link := range ls {
			if link.SiteMap {
				sm.Add(&sitemap.URL{Loc: path.Join(prefix, link.Link)})
			}
		}
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Status(http.StatusOK)
	sm.WriteTo(c.Writer)
}

func tidyPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	return p
}

// buildRedirectIndex maps normalized link names to redirect targets. Prefix keys are sorted so
// the first matching link order matches a stable YAML-like ordering (Go map iteration is not).
func buildRedirectIndex(links ingress.Links) map[string]string {
	n := 0
	for _, ls := range links {
		n += len(ls)
	}
	out := make(map[string]string, n)

	prefixes := make([]string, 0, len(links))
	for p := range links {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	for _, prefix := range prefixes {
		for _, link := range links[prefix] {
			if link == nil {
				continue
			}
			key := tidyPath(link.Name)
			if _, exists := out[key]; !exists {
				out[key] = link.Link
			}
		}
	}
	return out
}
