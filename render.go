package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
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

// prerenderPages renders one document per configured link prefix (and /about when About is set).
func prerenderPages(shell []byte, links ingress.Links, site string) (map[string]string, error) {
	base, err := html.Parse(bytes.NewReader(shell))
	if err != nil {
		return nil, fmt.Errorf("parse shell: %w", err)
	}

	llmsTxt, markdownPages = buildMarkdownPages(links, site)

	out := make(map[string]string, len(links)+2)
	for _, prefix := range sortedPrefixes(links) {
		path := tidyPath(prefix)
		routeLinks := effectiveRouteLinks(links, path, siteAbout)
		doc := cloneNode(base)

		// "/" keeps the hand-written title/description from index.html: they are better
		// than anything derived from a link list.
		if path != "/" {
			title := pageTitle(path)
			desc := pageDescription(path, routeLinks)
			if path == aboutPath && !siteAbout.IsEmpty() {
				title = aboutPageTitle(siteAbout)
				desc = aboutPageDescription(siteAbout)
			}
			setTitle(doc, title)
			setMetaContent(doc, "name", "description", desc)
			setMetaContent(doc, "property", "og:title", title)
			setMetaContent(doc, "property", "og:description", desc)
			setMetaContent(doc, "name", "twitter:title", title)
			setMetaContent(doc, "name", "twitter:description", desc)
		}
		setMetaContent(doc, "property", "og:url", absURL(site, path))
		setCanonical(doc, absURL(site, path))
		setAlternateMarkdown(doc, absURL(site, llmsTxtPath))
		addStructuredData(doc, site, prefix, routeLinks, links)
		if err := setAppContent(doc, path, routeLinks, siteAbout); err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := html.Render(&buf, doc); err != nil {
			return nil, fmt.Errorf("render %s: %w", path, err)
		}
		out[path] = buf.String()
	}

	if !siteAbout.IsEmpty() {
		if _, exists := out[aboutPath]; !exists {
			doc := cloneNode(base)
			title := aboutPageTitle(siteAbout)
			desc := aboutPageDescription(siteAbout)
			setTitle(doc, title)
			setMetaContent(doc, "name", "description", desc)
			setMetaContent(doc, "property", "og:title", title)
			setMetaContent(doc, "property", "og:description", desc)
			setMetaContent(doc, "name", "twitter:title", title)
			setMetaContent(doc, "name", "twitter:description", desc)
			setMetaContent(doc, "property", "og:url", absURL(site, aboutPath))
			setCanonical(doc, absURL(site, aboutPath))
			setAlternateMarkdown(doc, absURL(site, llmsTxtPath))

			rootLinks := effectiveRouteLinks(links, "/", siteAbout)
			addStructuredData(doc, site, "/", rootLinks, links)
			if err := setAppContent(doc, aboutPath, rootLinks, siteAbout); err != nil {
				return nil, err
			}

			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				return nil, fmt.Errorf("render %s: %w", aboutPath, err)
			}
			out[aboutPath] = buf.String()
		}
	}

	if _, exists := out[wipPath]; !exists {
		doc := cloneNode(base)
		title := wipPageTitle()
		desc := wipPageDescription()
		setTitle(doc, title)
		setMetaContent(doc, "name", "description", desc)
		setMetaContent(doc, "property", "og:title", title)
		setMetaContent(doc, "property", "og:description", desc)
		setMetaContent(doc, "name", "twitter:title", title)
		setMetaContent(doc, "name", "twitter:description", desc)
		setMetaContent(doc, "property", "og:url", absURL(site, wipPath))
		setCanonical(doc, absURL(site, wipPath))
		setAlternateMarkdown(doc, absURL(site, llmsTxtPath))

		rootLinks := effectiveRouteLinks(links, "/", siteAbout)
		addStructuredData(doc, site, "/", rootLinks, links)
		if err := setAppContent(doc, wipPath, rootLinks, siteAbout); err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := html.Render(&buf, doc); err != nil {
			return nil, fmt.Errorf("render %s: %w", wipPath, err)
		}
		out[wipPath] = buf.String()
	}

	return out, nil
}

func aboutPageTitle(about ingress.About) string {
	return pageTitle(aboutPath)
}

func wipPageTitle() string {
	return siteName + " — Work In Progress"
}

func wipPageDescription() string {
	return "Work In Progress — Gopher debugging: log.Print(\"h.e.r.e\")"
}

func aboutPageDescription(about ingress.About) string {
	name := cleanText(about.Name)
	title := cleanText(about.Title)
	desc := cleanText(about.Description)

	var header string
	switch {
	case name != "" && title != "":
		header = name + " — " + title
	case name != "":
		header = name
	case title != "":
		header = title
	}

	switch {
	case header != "" && desc != "":
		return truncate(header+" — "+desc, maxMetaDescription)
	case header != "":
		return truncate(header, maxMetaDescription)
	case desc != "":
		return truncate(desc, maxMetaDescription)
	default:
		return siteName + " — About"
	}
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
// the no-JS view is the real navigation, not an empty div. Whitespace nodes between flex
// items do not affect CSS layout but keep HTML-to-text scrapers from concatenating words.
func setAppContent(doc *html.Node, path string, links []*ingress.Link, about ingress.About) error {
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
	app.AppendChild(text("\n"))

	main := &html.Node{Type: html.ElementNode, Data: "main",
		Attr: []html.Attribute{{Key: "class", Val: "center-content"}, {Key: "aria-label", Val: "Site links"}}}
	app.AppendChild(main)
	app.AppendChild(text("\n"))

	if path != "/" && path != aboutPath && path != wipPath {
		back := &html.Node{
			Type: html.ElementNode,
			Data: "a",
			Attr: []html.Attribute{
				{Key: "class", Val: "back-link"},
				{Key: "href", Val: "/"},
			},
		}
		arrow := &html.Node{
			Type: html.ElementNode,
			Data: "span",
			Attr: []html.Attribute{
				{Key: "class", Val: "back-arrow"},
				{Key: "aria-hidden", Val: "true"},
			},
		}
		arrow.AppendChild(text("←"))
		label := &html.Node{Type: html.ElementNode, Data: "span"}
		label.AppendChild(text("Home"))

		back.AppendChild(arrow)
		back.AppendChild(text(" "))
		back.AppendChild(label)
		main.AppendChild(back)
		main.AppendChild(text("\n"))
	}

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
		card.AppendChild(text(" "))
		card.AppendChild(sub)
		nav.AppendChild(card)
		nav.AppendChild(text("\n"))
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

	if !about.IsEmpty() {
		app.AppendChild(text("\n"))
		app.AppendChild(buildAboutOverlayNode(about, path == aboutPath))
	}

	app.AppendChild(text("\n"))
	app.AppendChild(buildWipOverlayNode(path == wipPath))

	return nil
}

func buildAboutOverlayNode(about ingress.About, open bool) *html.Node {
	overlayAttrs := []html.Attribute{
		{Key: "class", Val: "about-overlay"},
		{Key: "id", Val: "about-overlay"},
	}
	if !open {
		overlayAttrs = append(overlayAttrs, html.Attribute{Key: "hidden", Val: ""})
	}
	overlay := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: overlayAttrs,
	}

	modal := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "about-modal"},
			{Key: "role", Val: "dialog"},
			{Key: "aria-modal", Val: "true"},
			{Key: "aria-labelledby", Val: "about-modal-title"},
		},
	}
	overlay.AppendChild(modal)

	closeBtn := &html.Node{
		Type: html.ElementNode,
		Data: "a",
		Attr: []html.Attribute{
			{Key: "class", Val: "about-close"},
			{Key: "href", Val: "/"},
			{Key: "aria-label", Val: "Close about dialog"},
		},
	}
	closeBtn.AppendChild(text("×"))
	modal.AppendChild(closeBtn)

	heroURL := strings.TrimSpace(about.Hero)
	name := strings.TrimSpace(about.Name)
	title := strings.TrimSpace(about.Title)

	if heroURL != "" && isSafeMarkdownURL(heroURL) {
		heroImg := &html.Node{
			Type: html.ElementNode,
			Data: "img",
			Attr: []html.Attribute{
				{Key: "class", Val: "about-hero"},
				{Key: "src", Val: heroURL},
				{Key: "alt", Val: firstNonEmpty(name, title, siteName)},
			},
		}
		modal.AppendChild(heroImg)
	}

	if name != "" {
		nameNode := &html.Node{
			Type: html.ElementNode,
			Data: "h2",
			Attr: []html.Attribute{
				{Key: "class", Val: "about-name"},
				{Key: "id", Val: "about-modal-title"},
			},
		}
		nameNode.AppendChild(text(name))
		modal.AppendChild(nameNode)

		if title != "" {
			titleNode := &html.Node{
				Type: html.ElementNode,
				Data: "p",
				Attr: []html.Attribute{
					{Key: "class", Val: "about-title"},
				},
			}
			titleNode.AppendChild(text(title))
			modal.AppendChild(titleNode)
		}
	} else {
		titleNode := &html.Node{
			Type: html.ElementNode,
			Data: "h2",
			Attr: []html.Attribute{
				{Key: "class", Val: "about-title"},
				{Key: "id", Val: "about-modal-title"},
			},
		}
		titleNode.AppendChild(text(firstNonEmpty(title, "About")))
		modal.AppendChild(titleNode)
	}

	if strings.TrimSpace(about.Description) != "" {
		descNode := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{Key: "class", Val: "about-description"},
			},
		}
		renderMarkdownNodes(descNode, about.Description)
		modal.AppendChild(descNode)
	}

	return overlay
}

func buildWipOverlayNode(open bool) *html.Node {
	overlayAttrs := []html.Attribute{
		{Key: "class", Val: "wip-overlay"},
		{Key: "id", Val: "wip-overlay"},
	}
	if !open {
		overlayAttrs = append(overlayAttrs, html.Attribute{Key: "hidden", Val: ""})
	}
	overlay := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: overlayAttrs,
	}

	modal := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "wip-modal"},
			{Key: "role", Val: "dialog"},
			{Key: "aria-modal", Val: "true"},
			{Key: "aria-labelledby", Val: "wip-modal-title"},
		},
	}
	overlay.AppendChild(modal)

	closeBtn := &html.Node{
		Type: html.ElementNode,
		Data: "a",
		Attr: []html.Attribute{
			{Key: "class", Val: "wip-close"},
			{Key: "href", Val: "/"},
			{Key: "aria-label", Val: "Close dialog"},
		},
	}
	closeBtn.AppendChild(text("×"))
	modal.AppendChild(closeBtn)

	titleNode := &html.Node{
		Type: html.ElementNode,
		Data: "h2",
		Attr: []html.Attribute{
			{Key: "class", Val: "visually-hidden"},
			{Key: "id", Val: "wip-modal-title"},
		},
	}
	titleNode.AppendChild(text("Work In Progress"))
	modal.AppendChild(titleNode)

	frame := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "wip-frame"},
		},
	}
	img := &html.Node{
		Type: html.ElementNode,
		Data: "img",
		Attr: []html.Attribute{
			{Key: "class", Val: "wip-img"},
			{Key: "src", Val: "/wip.webp"},
			{Key: "alt", Val: wipPageDescription()},
			{Key: "width", Val: "711"},
			{Key: "height", Val: "711"},
		},
	}
	frame.AppendChild(img)
	modal.AppendChild(frame)

	caption := &html.Node{
		Type: html.ElementNode,
		Data: "p",
		Attr: []html.Attribute{
			{Key: "class", Val: "wip-caption"},
		},
	}
	caption.AppendChild(text("Work in Progress"))
	modal.AppendChild(caption)

	return overlay
}

func renderMarkdownNodes(parent *html.Node, md string) {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			i++
			continue
		}

		// Fenced code block
		if strings.HasPrefix(trimmed, "```") {
			i++
			var codeLines []string
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				codeLines = append(codeLines, lines[i])
				i++
			}
			if i < len(lines) {
				i++
			}
			pre := &html.Node{Type: html.ElementNode, Data: "pre"}
			code := &html.Node{Type: html.ElementNode, Data: "code"}
			code.AppendChild(text(strings.Join(codeLines, "\n")))
			pre.AppendChild(code)
			parent.AppendChild(pre)
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			parent.AppendChild(&html.Node{Type: html.ElementNode, Data: "hr"})
			i++
			continue
		}

		// Heading
		if level, content, ok := parseMarkdownHeading(trimmed); ok {
			h := &html.Node{Type: html.ElementNode, Data: fmt.Sprintf("h%d", level)}
			appendInlineMarkdown(h, content)
			parent.AppendChild(h)
			i++
			continue
		}

		// Blockquote
		if strings.HasPrefix(trimmed, ">") {
			var quoteLines []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(t, ">") {
					break
				}
				quoteLines = append(quoteLines, strings.TrimSpace(strings.TrimPrefix(t, ">")))
				i++
			}
			bq := &html.Node{Type: html.ElementNode, Data: "blockquote"}
			p := &html.Node{Type: html.ElementNode, Data: "p"}
			appendParagraphLines(p, quoteLines)
			bq.AppendChild(p)
			parent.AppendChild(bq)
			continue
		}

		// Unordered list
		if item, ok := parseUnorderedListItem(trimmed); ok {
			ul := &html.Node{Type: html.ElementNode, Data: "ul"}
			li := &html.Node{Type: html.ElementNode, Data: "li"}
			appendInlineMarkdown(li, item)
			ul.AppendChild(li)
			i++
			for i < len(lines) {
				nextItem, nextOK := parseUnorderedListItem(strings.TrimSpace(lines[i]))
				if !nextOK {
					break
				}
				nextLI := &html.Node{Type: html.ElementNode, Data: "li"}
				appendInlineMarkdown(nextLI, nextItem)
				ul.AppendChild(nextLI)
				i++
			}
			parent.AppendChild(ul)
			continue
		}

		// Ordered list
		if item, ok := parseOrderedListItem(trimmed); ok {
			ol := &html.Node{Type: html.ElementNode, Data: "ol"}
			li := &html.Node{Type: html.ElementNode, Data: "li"}
			appendInlineMarkdown(li, item)
			ol.AppendChild(li)
			i++
			for i < len(lines) {
				nextItem, nextOK := parseOrderedListItem(strings.TrimSpace(lines[i]))
				if !nextOK {
					break
				}
				nextLI := &html.Node{Type: html.ElementNode, Data: "li"}
				appendInlineMarkdown(nextLI, nextItem)
				ol.AppendChild(nextLI)
				i++
			}
			parent.AppendChild(ol)
			continue
		}

		// Paragraph
		var paraLines []string
		for i < len(lines) {
			t := strings.TrimSpace(lines[i])
			if t == "" ||
				strings.HasPrefix(t, "```") ||
				t == "---" || t == "***" || t == "___" ||
				strings.HasPrefix(t, ">") {
				break
			}
			if _, _, ok := parseMarkdownHeading(t); ok {
				break
			}
			if _, ok := parseUnorderedListItem(t); ok {
				break
			}
			if _, ok := parseOrderedListItem(t); ok {
				break
			}
			paraLines = append(paraLines, t)
			i++
		}
		if len(paraLines) > 0 {
			p := &html.Node{Type: html.ElementNode, Data: "p"}
			appendParagraphLines(p, paraLines)
			parent.AppendChild(p)
		}
	}
}

func appendParagraphLines(p *html.Node, lines []string) {
	for idx, l := range lines {
		if idx > 0 {
			p.AppendChild(&html.Node{Type: html.ElementNode, Data: "br"})
		}
		appendInlineMarkdown(p, l)
	}
}

func parseMarkdownHeading(line string) (int, string, bool) {
	level := 0
	for level < len(line) && level < 6 && line[level] == '#' {
		level++
	}
	if level == 0 || level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(line[level+1:]), true
}

func parseUnorderedListItem(line string) (string, bool) {
	for _, prefix := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(line[len(prefix):]), true
		}
	}
	return "", false
}

func parseOrderedListItem(line string) (string, bool) {
	dot := strings.Index(line, ". ")
	if dot <= 0 {
		return "", false
	}
	for i := 0; i < dot; i++ {
		if line[i] < '0' || line[i] > '9' {
			return "", false
		}
	}
	return strings.TrimSpace(line[dot+2:]), true
}

func appendInlineMarkdown(parent *html.Node, s string) {
	for len(s) > 0 {
		kind, start, end, a, b := findEarliestInlineToken(s)
		if kind == "" {
			parent.AppendChild(text(s))
			return
		}
		if start > 0 {
			parent.AppendChild(text(s[:start]))
		}
		switch kind {
		case "code":
			code := &html.Node{Type: html.ElementNode, Data: "code"}
			code.AppendChild(text(a))
			parent.AppendChild(code)
		case "img":
			if isSafeMarkdownURL(b) {
				img := &html.Node{
					Type: html.ElementNode,
					Data: "img",
					Attr: []html.Attribute{
						{Key: "src", Val: strings.TrimSpace(b)},
						{Key: "alt", Val: a},
					},
				}
				parent.AppendChild(img)
			} else {
				parent.AppendChild(text(a))
			}
		case "link":
			href := strings.TrimSpace(b)
			if isSafeMarkdownURL(href) {
				attrs := []html.Attribute{{Key: "href", Val: href}}
				if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
					attrs = append(attrs,
						html.Attribute{Key: "target", Val: "_blank"},
						html.Attribute{Key: "rel", Val: "noopener noreferrer"},
					)
				}
				link := &html.Node{Type: html.ElementNode, Data: "a", Attr: attrs}
				appendInlineMarkdown(link, a)
				parent.AppendChild(link)
			} else {
				appendInlineMarkdown(parent, a)
			}
		case "bold":
			strong := &html.Node{Type: html.ElementNode, Data: "strong"}
			appendInlineMarkdown(strong, a)
			parent.AppendChild(strong)
		case "italic":
			em := &html.Node{Type: html.ElementNode, Data: "em"}
			appendInlineMarkdown(em, a)
			parent.AppendChild(em)
		}
		s = s[end:]
	}
}

func findEarliestInlineToken(s string) (kind string, start, end int, a, b string) {
	bestStart := len(s) + 1

	consider := func(k string, st, en int, argA, argB string) {
		if st >= 0 && st < bestStart {
			bestStart = st
			kind, start, end, a, b = k, st, en, argA, argB
		}
	}

	// Inline code: `...`
	if st := strings.IndexByte(s, '`'); st >= 0 {
		if cl := strings.IndexByte(s[st+1:], '`'); cl > 0 {
			consider("code", st, st+1+cl+1, s[st+1:st+1+cl], "")
		}
	}

	// Image: ![alt](url)
	if st := strings.Index(s, "!["); st >= 0 {
		if mid := strings.Index(s[st+2:], "]("); mid >= 0 {
			urlStart := st + 2 + mid + 2
			if cl := strings.IndexByte(s[urlStart:], ')'); cl >= 0 {
				consider("img", st, urlStart+cl+1, s[st+2:st+2+mid], s[urlStart:urlStart+cl])
			}
		}
	}

	// Link: [label](url)
	for searchFrom := 0; searchFrom < len(s); {
		rel := strings.IndexByte(s[searchFrom:], '[')
		if rel < 0 {
			break
		}
		st := searchFrom + rel
		if st > 0 && s[st-1] == '!' {
			searchFrom = st + 1
			continue
		}
		if mid := strings.Index(s[st+1:], "]("); mid >= 0 {
			urlStart := st + 1 + mid + 2
			if cl := strings.IndexByte(s[urlStart:], ')'); cl >= 0 {
				consider("link", st, urlStart+cl+1, s[st+1:st+1+mid], s[urlStart:urlStart+cl])
				break
			}
		}
		searchFrom = st + 1
	}

	// Bold: **...** or __...__
	for _, delim := range []string{"**", "__"} {
		if st := strings.Index(s, delim); st >= 0 {
			if cl := strings.Index(s[st+2:], delim); cl > 0 {
				consider("bold", st, st+2+cl+2, s[st+2:st+2+cl], "")
			}
		}
	}

	// Italic: *...* or _..._
	for _, delim := range []byte{'*', '_'} {
		for searchFrom := 0; searchFrom < len(s); {
			rel := strings.IndexByte(s[searchFrom:], delim)
			if rel < 0 {
				break
			}
			st := searchFrom + rel
			if st+1 < len(s) && s[st+1] == delim {
				searchFrom = st + 2
				continue
			}
			if cl := strings.IndexByte(s[st+1:], delim); cl > 0 {
				consider("italic", st, st+1+cl+1, s[st+1:st+1+cl], "")
				break
			}
			searchFrom = st + 1
		}
	}

	return kind, start, end, a, b
}

func isSafeMarkdownURL(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "/") && !strings.HasPrefix(trimmed, "//") {
		return true
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "mailto":
		return true
	default:
		return false
	}
}

func addStructuredData(doc *html.Node, site, prefix string, routeLinks []*ingress.Link, allLinks ingress.Links) {
	head := findNode(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "head" })
	if head == nil {
		return
	}
	author := map[string]any{
		"@type": "Person",
		"name":  "Homin Lee",
		"url":   absURL(site, "/resume/"),
	}
	if sameAs := collectSameAs(site, allLinks); len(sameAs) > 0 {
		author["sameAs"] = sameAs
	}

	ld := map[string]any{
		"@context":   "https://schema.org",
		"@type":      "WebSite",
		"name":       siteName,
		"url":        absURL(site, "/"),
		"inLanguage": "en",
		"author":     author,
	}
	if items := buildItemListElements(site, prefix, routeLinks); len(items) > 0 {
		ld["mainEntity"] = map[string]any{
			"@type":           "ItemList",
			"itemListElement": items,
		}
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

func collectSameAs(site string, allLinks ingress.Links) []string {
	seen := map[string]bool{
		absURL(site, "/"): true,
	}
	out := make([]string, 0)
	for _, prefix := range orderedPrefixes(allLinks) {
		for _, l := range allLinks[prefix] {
			if l == nil || l.Hide {
				continue
			}
			target, ok := resolvePublicURL(site, prefix, l.Link)
			if !ok || seen[target] {
				continue
			}
			seen[target] = true
			out = append(out, target)
		}
	}
	return out
}

func buildItemListElements(site, prefix string, links []*ingress.Link) []map[string]any {
	items := make([]map[string]any, 0, len(links))
	for _, l := range links {
		if l == nil || l.Hide {
			continue
		}
		target, ok := resolvePublicURL(site, prefix, l.Link)
		if !ok {
			continue
		}
		item := map[string]any{
			"@type":    "ListItem",
			"position": len(items) + 1,
			"name":     cleanText(firstNonEmpty(l.Desc, l.Name, "Link")),
			"url":      target,
		}
		items = append(items, item)
	}
	return items
}

func setAlternateMarkdown(doc *html.Node, url string) {
	n := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "link" &&
			attrOf(n, "rel") == "alternate" &&
			attrOf(n, "type") == "text/markdown"
	})
	if n != nil {
		setAttr(n, "href", url)
		return
	}
	head := findNode(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "head" })
	if head == nil {
		return
	}
	head.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "link",
		Attr: []html.Attribute{
			{Key: "rel", Val: "alternate"},
			{Key: "type", Val: "text/markdown"},
			{Key: "href", Val: url},
			{Key: "title", Val: "LLMs.txt"},
		},
	})
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
