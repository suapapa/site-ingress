package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/snabb/sitemap"

	"github.com/suapapa/site-ingress/internal/ingress"
)

const (
	aboutPath        = "/about"
	aboutButtonLabel = "\U0001f471\U0001f3fb\u200d\u2642\ufe0f About"
	wipPath          = "/wip"
)

var (
	programName = "site-ingress"
	programVer  = "dev"

	httpPort  int
	linksConf string
	staticDir string
	siteURL   = "https://homin.dev" // absolute base for relative links in sitemap.xml
	debug     bool

	siteAbout      ingress.About
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
	flag.StringVar(&staticDir, "static", "static", "directory holding the Vite-built index.html used as the prerender template")
	flag.BoolVar(&debug, "d", false, "print debug logs")
	flag.Parse()

	site, err := ingress.LoadSiteFromFile(linksConf)
	if err != nil {
		log.Fatalf("fail to read links conf: %v", err)
	}
	siteAbout = site.About
	siteLinks = site.Links
	says = site.Says
	redirectByPath = buildRedirectIndex(site.Links)
	llmsTxt, markdownPages = buildMarkdownPages(site.Links, siteURL)
	loadPrerenderShell()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.SetTrustedProxies([]string{
		"10.0.0.0/8",     // Kubernetes cluster network
		"172.16.0.0/12",  // Kubernetes cluster network
		"192.168.0.0/16", // Kubernetes cluster network
		"127.0.0.1",      // Localhost
	})

	router.GET("/api/links", linksHandler)
	router.GET("/api/about", aboutHandler)
	router.GET("/api/fish", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetRandomSay())
	})
	router.StaticFile("/ads.txt", "asset/ads.txt")
	router.StaticFile("/robots.txt", "asset/robots.txt")
	router.StaticFile("/wip.webp", "asset/wip.webp")
	router.GET("/sitemap.xml", sitemapHandler)
	router.HEAD("/sitemap.xml", sitemapHandler)
	router.GET(llmsTxtPath, llmsTxtHandler)
	router.HEAD(llmsTxtPath, llmsTxtHandler)
	router.GET(llmsFullTxtPath, llmsTxtHandler)
	router.HEAD(llmsFullTxtPath, llmsTxtHandler)
	router.GET("/", rootHandler)
	router.HEAD("/", rootHandler)
	router.GET("/:path", pathHandler)
	router.HEAD("/:path", pathHandler)

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

// loadPrerenderShell reads the Vite-built index.html and prerenders one document per
// configured route. A missing shell is not fatal — nginx then keeps serving the plain
// client-rendered shell, which is what happened before prerendering existed.
func loadPrerenderShell() {
	for _, p := range []string{
		filepath.Join(staticDir, "index.html"),
		filepath.Join("frontend", "dist", "index.html"), // local `go run .`
	} {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		rendered, err := prerenderPages(b, siteLinks, siteURL)
		if err != nil {
			log.WithError(err).Warn("prerender failed: HTML routes fall back to the client-rendered shell")
			return
		}
		prerenderShell = b
		pages = rendered
		log.WithField("static", p).Infof("prerendered %d SPA route(s)", len(pages))
		return
	}
	log.WithField("static", staticDir).Warn("prerender shell not found: HTML routes fall back to nginx")
}

func linksHandler(c *gin.Context) {
	showHides := c.Query("show_hides") == "true"
	prefix := tidyPath(c.Query("prefix"))

	list := effectiveRouteLinks(siteLinks, prefix, siteAbout)
	resp := make([]*ingress.Link, 0, len(list))
	for _, l := range list {
		if l == nil {
			continue
		}
		if !l.Hide || showHides {
			resp = append(resp, l)
		}
	}
	c.JSON(http.StatusOK, resp)
}

func aboutHandler(c *gin.Context) {
	if siteAbout.IsEmpty() {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, siteAbout)
}

// effectiveRouteLinks returns the configured links for prefix, prepending the "👱🏻‍♂️ About"
// button on the root route when About is configured.
func effectiveRouteLinks(links ingress.Links, prefix string, about ingress.About) []*ingress.Link {
	norm := tidyPath(prefix)
	var list []*ingress.Link
	if norm == "/" {
		if l, ok := links["/"]; ok {
			list = l
		} else if l, ok := links[""]; ok {
			list = l
		}
	} else {
		if l, ok := links[norm]; ok {
			list = l
		} else if l, ok := links[strings.TrimPrefix(norm, "/")]; ok {
			list = l
		}
	}

	if norm != "/" || about.IsEmpty() {
		return list
	}

	for _, l := range list {
		if l != nil && (tidyPath(l.Name) == aboutPath || tidyPath(l.Link) == aboutPath) {
			return list
		}
	}

	out := make([]*ingress.Link, 0, len(list)+1)
	out = append(out, &ingress.Link{
		Name: aboutPath,
		Link: aboutPath,
		Desc: aboutButtonLabel,
	})
	out = append(out, list...)
	return out
}

func rootHandler(c *gin.Context) {
	servePrerendered(c, "/")
}

// pathHandler serves a prerendered SPA route when the path is a configured prefix, and
// otherwise keeps the original short-link redirect behaviour.
func pathHandler(c *gin.Context) {
	dest := tidyPath(c.Param("path"))

	if dest == llmsTxtPath || dest == llmsFullTxtPath {
		llmsTxtHandler(c)
		return
	}

	if strings.HasSuffix(dest, ".md") {
		mdPath := strings.TrimSuffix(dest, ".md")
		if mdPath == "/index" {
			mdPath = "/"
		}
		if md, ok := markdownPages[mdPath]; ok {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Data(http.StatusOK, markdownMime, []byte(md))
			return
		}
	}

	if _, ok := pages[dest]; ok {
		servePrerendered(c, dest)
		return
	}
	if wantsMarkdown(c) {
		if md, ok := markdownPages[dest]; ok {
			c.Header("Vary", "Accept")
			c.Header("X-Content-Type-Options", "nosniff")
			c.Data(http.StatusOK, markdownMime, []byte(md))
			return
		}
	}

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

func servePrerendered(c *gin.Context, path string) {
	c.Header("Vary", "Accept")
	c.Header("Cache-Control", "no-cache")
	c.Header("Link", fmt.Sprintf("<%s>; rel=\"alternate\"; type=\"text/markdown\"", absURL(siteURL, llmsTxtPath)))

	// ?show_gophers_only=true asks for the WebGL scene without any chrome, so the bare
	// shell (no injected links) is the right answer there.
	if c.Query("show_gophers_only") == "true" && len(prerenderShell) > 0 {
		c.Data(http.StatusOK, "text/html; charset=utf-8", prerenderShell)
		return
	}

	if wantsMarkdown(c) {
		if md, ok := markdownPages[path]; ok {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Data(http.StatusOK, markdownMime, []byte(md))
			return
		}
	}

	page, ok := pages[path]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
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
	return resolvePublicURL(siteURL, prefix, link)
}

func resolvePublicURL(site, prefix, link string) (string, bool) {
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

	base, err := url.Parse(site)
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

	// Guard len(p) > 1: stripping the trailing slash of "/" would turn the root path into
	// the empty string, so "/" and "" would stop resolving to the same prefix.
	if len(p) > 1 && p[len(p)-1] == '/' {
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
