package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/suapapa/site-ingress/internal/ingress"
)

// Server-side rendering of the SPA routes.
//
// The frontend is a WebGL client-rendered app: without JavaScript the document is just
// <div id="app"></div>. Google renders that, but crawlers that do not execute JS —
// notably Naver's Yeti — index an empty page. So the Go backend renders the same markup
// the frontend would build, straight from the same site config the JSON API serves, and
// hands it over inside #app. The SPA replaces it on hydration (main.js calls
// app.replaceChildren() first), so nothing is duplicated.
//
// Pages are rendered once at boot: the config is read once at startup too, so a config
// change already requires a backend restart.

const (
	maxMetaDescription = 160
	siteName           = "HOMIN-DEV"
)

// prerenderShell is the Vite-built index.html used as the template.
var prerenderShell []byte

// pages maps a route path ("/", "/social", …) to the fully server-rendered document.
var pages map[string]string

// prerenderPages renders one document per configured link prefix.
func prerenderPages(shell []byte, links ingress.Links, site string) (map[string]string, error) {
	base, err := html.Parse(bytes.NewReader(shell))
	if err != nil {
		return nil, fmt.Errorf("parse shell: %w", err)
	}

	out := make(map[string]string, len(links)+1)
	for _, prefix := range sortedPrefixes(links) {
		path := tidyPath(prefix)
		doc := cloneNode(base)

		// "/" keeps the hand-written title/description from index.html: they are better
		// than anything derived from a link list.
		if path != "/" {
			title := pageTitle(path)
			desc := pageDescription(path, links[prefix])
			setTitle(doc, title)
			setMetaContent(doc, "name", "description", desc)
			setMetaContent(doc, "property", "og:title", title)
			setMetaContent(doc, "property", "og:description", desc)
			setMetaContent(doc, "name", "twitter:title", title)
			setMetaContent(doc, "name", "twitter:description", desc)
		}
		setMetaContent(doc, "property", "og:url", absURL(site, path))
		setCanonical(doc, absURL(site, path))
		addStructuredData(doc, site)
		if err := setAppContent(doc, links[prefix]); err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := html.Render(&buf, doc); err != nil {
			return nil, fmt.Errorf("render %s: %w", path, err)
		}
		out[path] = buf.String()
	}
	return out, nil
}

func pageTitle(path string) string {
	seg := strings.Trim(path, "/")
	if seg == "" {
		return siteName
	}
	words := strings.Split(seg, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return siteName + " — " + strings.Join(words, " ")
}

func pageDescription(path string, links []*ingress.Link) string {
	labels := visibleLabels(links)
	if len(labels) == 0 {
		return siteName + " — Homin Lee's personal ingress."
	}
	desc := fmt.Sprintf("%s %s: %s", siteName, strings.Trim(path, "/"), strings.Join(labels, ", "))
	return truncate(desc, maxMetaDescription)
}

// visibleLabels returns the labels a visitor actually sees, in config order. Hidden
// links are dropped exactly like the JSON API drops them.
func visibleLabels(links []*ingress.Link) []string {
	out := make([]string, 0, len(links))
	for _, l := range links {
		if l == nil || l.Hide {
			continue
		}
		label := l.Desc
		if strings.TrimSpace(label) == "" {
			label = l.Name
		}
		if strings.TrimSpace(label) != "" {
			out = append(out, strings.Join(strings.Fields(label), " "))
		}
	}
	return out
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n-1])) + "…"
}

func absURL(site, path string) string {
	return strings.TrimSuffix(site, "/") + path
}

// setAppContent replaces #app's children with the markup main.js builds at runtime, so
// the no-JS view is the real navigation, not an empty div.
func setAppContent(doc *html.Node, links []*ingress.Link) error {
	app := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && attrOf(n, "id") == "app"
	})
	if app == nil {
		return fmt.Errorf("shell has no #app element")
	}
	app.FirstChild, app.LastChild = nil, nil

	header := &html.Node{Type: html.ElementNode, Data: "a",
		Attr: []html.Attribute{{Key: "class", Val: "site-header"}, {Key: "href", Val: "/"}}}
	header.AppendChild(text(siteName))
	app.AppendChild(header)

	main := &html.Node{Type: html.ElementNode, Data: "main",
		Attr: []html.Attribute{{Key: "class", Val: "center-content"}, {Key: "aria-label", Val: "Site links"}}}
	app.AppendChild(main)

	nav := &html.Node{Type: html.ElementNode, Data: "nav",
		Attr: []html.Attribute{{Key: "class", Val: "links-container"}, {Key: "aria-label", Val: "Destinations"}}}
	main.AppendChild(nav)

	for _, l := range links {
		if l == nil || l.Hide {
			continue
		}
		card := &html.Node{Type: html.ElementNode, Data: "a",
			Attr: []html.Attribute{{Key: "class", Val: "link-card"}, {Key: "href", Val: l.Link}}}

		label := &html.Node{Type: html.ElementNode, Data: "span"}
		label.AppendChild(text(firstNonEmpty(l.Desc, l.Name, "Link")))

		sub := &html.Node{Type: html.ElementNode, Data: "span",
			Attr: []html.Attribute{{Key: "class", Val: "link-sub"}}}
		sub.AppendChild(text(l.Name))

		card.AppendChild(label)
		card.AppendChild(sub)
		nav.AppendChild(card)
	}

	// Mirrors buildChrome() so the styled layout holds up without JavaScript.
	status := &html.Node{Type: html.ElementNode, Data: "p",
		Attr: []html.Attribute{{Key: "class", Val: "links-status"}, {Key: "hidden", Val: ""}}}
	main.AppendChild(status)

	main.AppendChild(&html.Node{Type: html.ElementNode, Data: "blockquote",
		Attr: []html.Attribute{{Key: "class", Val: "fish-container"}}})

	footer := &html.Node{Type: html.ElementNode, Data: "footer",
		Attr: []html.Attribute{{Key: "class", Val: "site-footer"}}}
	footer.AppendChild(text("© Homin Lee <i@homin.dev> All rights reserved."))
	footer.AppendChild(&html.Node{Type: html.ElementNode, Data: "br"})
	footer.AppendChild(text("The Go gopher was designed by Renee French."))
	app.AppendChild(footer)

	return nil
}

func addStructuredData(doc *html.Node, site string) {
	head := findNode(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "head" })
	if head == nil {
		return
	}
	ld := map[string]any{
		"@context":   "https://schema.org",
		"@type":      "WebSite",
		"name":       siteName,
		"url":        absURL(site, "/"),
		"inLanguage": "en",
		"author": map[string]any{
			"@type": "Person",
			"name":  "Homin Lee",
			"url":   absURL(site, "/resume/"),
		},
	}
	raw, err := json.Marshal(ld)
	if err != nil {
		return
	}
	script := &html.Node{Type: html.ElementNode, Data: "script",
		Attr: []html.Attribute{{Key: "type", Val: "application/ld+json"}}}
	script.AppendChild(text(string(raw)))
	head.AppendChild(script)
}

func setTitle(doc *html.Node, title string) {
	n := findNode(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "title" })
	if n == nil {
		return
	}
	n.FirstChild, n.LastChild = nil, nil
	n.AppendChild(text(title))
}

func setMetaContent(doc *html.Node, key, value, content string) {
	n := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && attrOf(n, key) == value
	})
	if n == nil {
		return
	}
	setAttr(n, "content", content)
}

func setCanonical(doc *html.Node, url string) {
	n := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "link" && attrOf(n, "rel") == "canonical"
	})
	if n == nil {
		return
	}
	setAttr(n, "href", url)
}

func text(s string) *html.Node { return &html.Node{Type: html.TextNode, Data: s} }

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func findNode(n *html.Node, match func(*html.Node) bool) *html.Node {
	if match(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findNode(c, match); r != nil {
			return r
		}
	}
	return nil
}

func attrOf(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, val string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func cloneNode(n *html.Node) *html.Node {
	c := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     append([]html.Attribute(nil), n.Attr...),
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		c.AppendChild(cloneNode(ch))
	}
	return c
}
