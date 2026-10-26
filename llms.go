package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/suapapa/site-ingress/internal/ingress"
)

const (
	siteSummary     = "Homin Lee's personal projects and social links."
	markdownMime    = "text/markdown; charset=utf-8"
	llmsTxtPath     = "/llms.txt"
	llmsFullTxtPath = "/llms-full.txt"
)

// markdownPages maps a normalized route path ("/", "/social", …) to its Markdown
// representation for Accept: text/markdown content negotiation and *.md routes.
var markdownPages map[string]string

// llmsTxt caches the full /llms.txt document built at startup.
var llmsTxt string

// buildMarkdownPages prebuilds the site-wide /llms.txt document and per-route Markdown
// views from the configured links.
func buildMarkdownPages(links ingress.Links, site string) (string, map[string]string) {
	full := renderLLMsTxt(links, site)
	byRoute := make(map[string]string, len(links)+3)

	for _, prefix := range orderedPrefixes(links) {
		path := tidyPath(prefix)
		if path == "/" {
			byRoute[path] = full
			continue
		}
		routeLinks := effectiveRouteLinks(links, path, siteAbout)
		byRoute[path] = renderRouteMarkdown(path, routeLinks, site)
	}
	if _, ok := byRoute["/"]; !ok {
		byRoute["/"] = full
	}
	if !siteAbout.IsEmpty() {
		if _, ok := byRoute[aboutPath]; !ok {
			byRoute[aboutPath] = renderAboutMarkdown(siteAbout, site)
		}
	}
	if _, ok := byRoute[wipPath]; !ok {
		byRoute[wipPath] = renderWipMarkdown(site)
	}

	return full, byRoute
}

// renderLLMsTxt renders the site-wide index following the https://llmstxt.org/ format.
func renderLLMsTxt(links ingress.Links, site string) string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(siteName)
	b.WriteString("\n\n> ")
	b.WriteString(siteSummary)
	b.WriteString("\n\n")
	fmt.Fprintf(
		&b,
		"Author: Homin Lee (%s)\nCanonical URL: %s\nShort links on %s/<name> redirect (HTTP 307) to the corresponding destination.\n",
		absURL(site, "/resume/"),
		absURL(site, "/"),
		strings.TrimSuffix(site, "/"),
	)

	if !siteAbout.IsEmpty() {
		b.WriteString("\n## About (`")
		b.WriteString(aboutPath)
		b.WriteString("`)\n\n")
		name := strings.TrimSpace(siteAbout.Name)
		title := strings.TrimSpace(siteAbout.Title)
		if name != "" && title != "" {
			b.WriteString("**")
			b.WriteString(name)
			b.WriteString("** — ")
			b.WriteString(title)
			b.WriteString("\n\n")
		} else if name != "" {
			b.WriteString("**")
			b.WriteString(name)
			b.WriteString("**\n\n")
		} else if title != "" {
			b.WriteString("**")
			b.WriteString(title)
			b.WriteString("**\n\n")
		}
		if desc := strings.TrimSpace(siteAbout.Description); desc != "" {
			b.WriteString(desc)
			b.WriteByte('\n')
		}
	}

	for _, prefix := range orderedPrefixes(links) {
		path := tidyPath(prefix)
		routeLinks := effectiveRouteLinks(links, path, siteAbout)
		items := formatMarkdownLinks(prefix, routeLinks, site)
		if len(items) == 0 {
			continue
		}

		b.WriteString("\n## ")
		b.WriteString(sectionHeading(path))
		b.WriteString("\n\n")
		for _, item := range items {
			b.WriteString(item)
			b.WriteByte('\n')
		}
	}

	writeMachineEndpoints(&b, site)
	return b.String()
}

func renderAboutMarkdown(about ingress.About, site string) string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(aboutPageTitle(about))
	b.WriteString("\n\n")
	name := strings.TrimSpace(about.Name)
	title := strings.TrimSpace(about.Title)
	if name != "" && title != "" {
		fmt.Fprintf(&b, "> **%s** — %s\n\n", name, title)
	} else if name != "" {
		fmt.Fprintf(&b, "> **%s**\n\n", name)
	} else if title != "" {
		fmt.Fprintf(&b, "> %s\n\n", title)
	}
	fmt.Fprintf(&b, "Canonical URL: %s\n", absURL(site, aboutPath))
	if hero := strings.TrimSpace(about.Hero); hero != "" {
		fmt.Fprintf(&b, "Hero Image: %s\n", hero)
	}
	if desc := strings.TrimSpace(about.Description); desc != "" {
		b.WriteString("\n")
		b.WriteString(desc)
		b.WriteByte('\n')
	}
	fmt.Fprintf(
		&b,
		"\n## Discovery\n\n- [Full Site Index (llms.txt)](%s): Complete Markdown list of all routes and destinations on %s.\n",
		absURL(site, llmsTxtPath),
		siteName,
	)

	return b.String()
}

func renderWipMarkdown(site string) string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(wipPageTitle())
	b.WriteString("\n\n> ")
	b.WriteString(wipPageDescription())
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Canonical URL: %s\n", absURL(site, wipPath))
	fmt.Fprintf(&b, "Image: %s\n\n", absURL(site, "/wip.webp"))
	b.WriteString("현재 작업 중인 페이지입니다.\n")
	fmt.Fprintf(
		&b,
		"\n## Discovery\n\n- [Full Site Index (llms.txt)](%s): Complete Markdown list of all routes and destinations on %s.\n",
		absURL(site, llmsTxtPath),
		siteName,
	)

	return b.String()
}

// renderRouteMarkdown renders a single route prefix as a standalone Markdown document.
func renderRouteMarkdown(path string, links []*ingress.Link, site string) string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(pageTitle(path))
	b.WriteString("\n\n> ")
	b.WriteString(pageDescription(path, links))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Canonical URL: %s\n", absURL(site, path))

	items := formatMarkdownLinks(path, links, site)
	if len(items) > 0 {
		b.WriteString("\n## Links\n\n")
		for _, item := range items {
			b.WriteString(item)
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(
		&b,
		"\n## Discovery\n\n- [Full Site Index (llms.txt)](%s): Complete Markdown list of all routes and destinations on %s.\n",
		absURL(site, llmsTxtPath),
		siteName,
	)

	return b.String()
}

func formatMarkdownLinks(prefix string, links []*ingress.Link, site string) []string {
	items := make([]string, 0, len(links))
	for _, l := range links {
		if l == nil || l.Hide {
			continue
		}
		target, ok := resolvePublicURL(site, prefix, l.Link)
		if !ok {
			continue
		}

		name := cleanText(l.Name)
		desc := cleanText(l.Desc)

		switch {
		case strings.HasPrefix(name, "/"):
			label := firstNonEmpty(desc, name, "Link")
			shortPath := tidyPath(name)
			if desc != "" && desc != name {
				items = append(items, fmt.Sprintf("- [%s](%s): Short link `%s`", label, target, shortPath))
			} else {
				items = append(items, fmt.Sprintf("- [%s](%s)", label, target))
			}
		case name != "" && desc != "" && name != desc:
			items = append(items, fmt.Sprintf("- [%s](%s): %s", name, target, desc))
		default:
			label := firstNonEmpty(desc, name, "Link")
			items = append(items, fmt.Sprintf("- [%s](%s)", label, target))
		}
	}
	return items
}

func writeMachineEndpoints(b *strings.Builder, site string) {
	b.WriteString("\n## Machine-Readable Endpoints\n\n")
	fmt.Fprintf(b, "- [LLMs Index](%s): Complete Markdown index of all public links across all routes.\n", absURL(site, llmsTxtPath))
	fmt.Fprintf(b, "- [Links JSON API](%s): Public links in JSON (pass `?prefix=<route>` for sub-routes).\n", absURL(site, "/api/links"))
	fmt.Fprintf(b, "- [Quotes JSON API](%s): Random movie quote in JSON.\n", absURL(site, "/api/fish"))
	fmt.Fprintf(b, "- [Sitemap](%s): XML sitemap.\n", absURL(site, "/sitemap.xml"))
}

func sectionHeading(path string) string {
	if path == "/" {
		return "Main (`/`)"
	}
	seg := strings.Trim(path, "/")
	words := strings.Split(seg, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return fmt.Sprintf("%s (`%s`)", strings.Join(words, " "), path)
}

// orderedPrefixes returns the configured prefixes with the root ("/" or "") first and
// the remaining prefixes in deterministic lexical order.
func orderedPrefixes(links ingress.Links) []string {
	sorted := sortedPrefixes(links)
	out := make([]string, 0, len(sorted))
	for _, p := range sorted {
		if tidyPath(p) == "/" {
			out = append(out, p)
		}
	}
	for _, p := range sorted {
		if tidyPath(p) != "/" {
			out = append(out, p)
		}
	}
	return out
}

func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func llmsTxtHandler(c *gin.Context) {
	doc := llmsTxt
	if doc == "" {
		doc = renderLLMsTxt(siteLinks, siteURL)
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, markdownMime, []byte(doc))
}

// wantsMarkdown reports whether the request explicitly asks for a Markdown representation
// via ?format=md|markdown or an Accept header that prefers text/markdown over text/html.
func wantsMarkdown(c *gin.Context) bool {
	switch strings.ToLower(strings.TrimSpace(c.Query("format"))) {
	case "md", "markdown":
		return true
	}

	accept := c.GetHeader("Accept")
	if accept == "" {
		return false
	}

	var mdQ, htmlQ float64
	for _, part := range strings.Split(accept, ",") {
		mediaType, q := parseMediaRange(part)
		switch mediaType {
		case "text/markdown", "text/x-markdown":
			if q > mdQ {
				mdQ = q
			}
		case "text/html", "application/xhtml+xml":
			if q > htmlQ {
				htmlQ = q
			}
		}
	}

	return mdQ > 0 && mdQ >= htmlQ
}

func parseMediaRange(s string) (string, float64) {
	segments := strings.Split(strings.TrimSpace(s), ";")
	mediaType := strings.ToLower(strings.TrimSpace(segments[0]))
	q := 1.0

	for _, param := range segments[1:] {
		kv := strings.SplitN(strings.TrimSpace(param), "=", 2)
		if len(kv) != 2 || strings.ToLower(strings.TrimSpace(kv[0])) != "q" {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
		if err != nil || parsed < 0 {
			return mediaType, 0
		}
		if parsed > 1 {
			parsed = 1
		}
		q = parsed
	}

	return mediaType, q
}
