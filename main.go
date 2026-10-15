package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sort"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/snabb/sitemap"

	"github.com/suapapa/site-ingress/internal/ingress"
)

var (
	programName = "site-ingress"
	programVer  = "dev"

	httpPort  int
	linksConf string
	siteURL   = "https://homin.dev" // absolute base for relative links in sitemap.xml
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
	flag.StringVar(&siteURL, "u", siteURL, "site URL used to absolutize relative links in sitemap.xml")
	flag.BoolVar(&debug, "d", false, "print debug logs")
	flag.Parse()

	site, err := ingress.LoadSiteFromFile(linksConf)
	if err != nil {
		log.Fatalf("fail to read links conf: %v", err)
	}
	siteLinks = site.Links
	says = site.Says
	redirectByPath = buildRedirectIndex(site.Links)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.SetTrustedProxies([]string{
		"10.0.0.0/8",     // Kubernetes cluster network
		"172.16.0.0/12",  // Kubernetes cluster network
		"192.168.0.0/16", // Kubernetes cluster network
		"127.0.0.1",      // Localhost
	})

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
	router.StaticFile("/ads.txt", "asset/ads.txt")
	router.StaticFile("/robots.txt", "asset/robots.txt")
	router.GET("/sitemap.xml", sitemapHandler)
	router.GET("/:path", redirectHandler)

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

	for _, prefix := range sortedPrefixes(siteLinks) {
		for _, link := range siteLinks[prefix] {
			if link == nil || !link.SiteMap {
				continue
			}
			loc, ok := sitemapLoc(prefix, link.Link)
			if !ok {
				log.WithField("link", link.Link).Warn("skip sitemap entry: not a crawlable URL")
				continue
			}
			sm.Add(&sitemap.URL{Loc: loc})
		}
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Status(http.StatusOK)
	sm.WriteTo(c.Writer)
}

// sitemapLoc resolves a links.yaml entry into the absolute <loc> value required by the
// sitemap spec. Absolute links are emitted verbatim: running them through path.Join (as
// this used to do) collapses the "//" after the scheme and strips the trailing slash,
// turning https://homin.dev/resume/ into the invalid /https:/homin.dev/resume.
// Relative links are resolved against siteURL plus the enclosing prefix.
func sitemapLoc(prefix, link string) (string, bool) {
	u, err := url.Parse(link)
	if err != nil {
		return "", false
	}

	if u.IsAbs() {
		if u.Scheme != "http" && u.Scheme != "https" {
			return "", false // mailto:, tel:, javascript: … have no place in a sitemap
		}
		return u.String(), true
	}

	base, err := url.Parse(siteURL)
	if err != nil || base.Host == "" {
		return "", false
	}
	base.Path = path.Join("/", prefix, u.Path)
	base.RawQuery = u.RawQuery
	return base.String(), true
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

	for _, prefix := range sortedPrefixes(links) {
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

// sortedPrefixes returns the link prefixes in a deterministic order so that generated
// output (redirect index, sitemap.xml) does not depend on Go's random map iteration.
func sortedPrefixes(links ingress.Links) []string {
	prefixes := make([]string, 0, len(links))
	for p := range links {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)
	return prefixes
}
