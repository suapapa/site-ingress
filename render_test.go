package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/suapapa/site-ingress/internal/ingress"
)

const testShell = `<!doctype html><html lang="en"><head>
<meta name="description" content="shell description" />
<link rel="canonical" href="https://homin.dev/" />
<meta property="og:title" content="HOMIN-DEV" />
<meta property="og:description" content="shell description" />
<meta property="og:url" content="https://homin.dev/" />
<meta name="twitter:title" content="HOMIN-DEV" />
<meta name="twitter:description" content="shell description" />
<title>HOMIN-DEV</title></head><body><div id="app"></div></body></html>`

func testLinks() ingress.Links {
	return ingress.Links{
		"/": {
			{Name: "/resume", Link: "https://homin.dev/resume/", Desc: "📋 Resume", SiteMap: true},
			{Name: "/hidden", Link: "https://example.com/hidden", Desc: "ShouldNotAppear", Hide: true},
		},
		"/social": {
			{Name: "/yt", Link: "https://www.youtube.com/@hominlog", Desc: "🍿 YouTube"},
			{Name: `/x"><script>alert(1)</script>`, Link: `https://example.com/?a="b"`, Desc: "<b>nope</b>"},
		},
	}
}

func TestPrerenderPages(t *testing.T) {
	pages, err := prerenderPages([]byte(testShell), testLinks(), "https://homin.dev")
	if err != nil {
		t.Fatalf("prerenderPages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2 (%v)", len(pages), keysOf(pages))
	}

	root, ok := pages["/"]
	if !ok {
		t.Fatalf(`no page for "/" (got %v) — the root prefix must survive tidyPath`, keysOf(pages))
	}

	// "/" keeps the hand-written head from index.html.
	if !strings.Contains(root, "<title>HOMIN-DEV</title>") {
		t.Errorf(`root title was overwritten`)
	}
	if !strings.Contains(root, `content="shell description"`) {
		t.Errorf("root description was overwritten")
	}
	// ...but still gets the injected navigation and the crawler-visible markup.
	for _, want := range []string{
		`href="https://homin.dev/resume/"`,
		`class="link-card"`,
		`class="site-header"`,
		`class="site-footer"`,
		`application/ld+json`,
		`"@type":"WebSite"`,
		`"@type":"ItemList"`,
		`"sameAs":[`,
		`rel="alternate" type="text/markdown" href="https://homin.dev/llms.txt"`,
		`<span>📋 Resume</span> <span class="link-sub">/resume</span>`,
	} {
		if !strings.Contains(root, want) {
			t.Errorf("root page missing %q", want)
		}
	}
	if strings.Contains(root, "ShouldNotAppear") {
		t.Errorf("hidden link leaked into the prerendered page")
	}
	if strings.Contains(root, `class="back-link"`) {
		t.Errorf("root page should not render back-link")
	}

	social := pages["/social"]
	if social == "" {
		t.Fatalf(`no page for "/social"`)
	}
	for _, want := range []string{
		"<title>HOMIN-DEV — Social</title>",
		`<link rel="canonical" href="https://homin.dev/social"/>`,
		`content="https://homin.dev/social"`,
		`class="back-link"`,
		"YouTube",
	} {
		if !strings.Contains(social, want) {
			t.Errorf("social page missing %q", want)
		}
	}
	if desc := pageDescription("/social", testLinks()["/social"]); !strings.Contains(desc, "YouTube") {
		t.Errorf("pageDescription(/social) = %q", desc)
	}

	// Escaping: nothing from the config may be emitted as raw markup.
	if strings.Contains(social, "<script>alert(1)</script>") || strings.Contains(social, "<b>nope</b>") {
		t.Errorf("config values were not escaped:\n%s", social)
	}
	if !strings.Contains(social, "&lt;b&gt;nope&lt;/b&gt;") {
		t.Errorf("expected escaped label in social page")
	}
}

func TestPrerenderPages_withoutApp(t *testing.T) {
	if _, err := prerenderPages([]byte("<html><body>no app here</body></html>"), testLinks(), "https://homin.dev"); err == nil {
		t.Fatal("expected an error when the shell has no #app element")
	}
}

func TestPageDescription_truncates(t *testing.T) {
	long := make([]*ingress.Link, 0, 40)
	for i := 0; i < 40; i++ {
		long = append(long, &ingress.Link{Name: "/x", Desc: "a fairly long destination label"})
	}
	got := pageDescription("/social", long)
	if n := len([]rune(got)); n > maxMetaDescription {
		t.Errorf("description is %d runes, want <= %d: %q", n, maxMetaDescription, got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncated description should end with an ellipsis: %q", got)
	}
}

// TestHTMLRoutes covers the HTTP surface: the prerendered pages, the gophers-only escape
// hatch, short links and 404s.
func TestHTMLRoutes(t *testing.T) {
	origPages, origShell, origRedirects := pages, prerenderShell, redirectByPath
	origLLMs, origMDPages := llmsTxt, markdownPages
	t.Cleanup(func() {
		pages, prerenderShell, redirectByPath = origPages, origShell, origRedirects
		llmsTxt, markdownPages = origLLMs, origMDPages
	})

	rendered, err := prerenderPages([]byte(testShell), testLinks(), "https://homin.dev")
	if err != nil {
		t.Fatalf("prerenderPages: %v", err)
	}
	pages, prerenderShell = rendered, []byte(testShell)
	redirectByPath = buildRedirectIndex(ingress.Links{
		"/": {{Name: "/gh", Link: "https://github.com/suapapa"}},
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET(llmsTxtPath, llmsTxtHandler)
	r.GET(llmsFullTxtPath, llmsTxtHandler)
	r.GET("/", rootHandler)
	r.GET("/:path", pathHandler)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
		wantNot    string
	}{
		{"/", http.StatusOK, `class="link-card"`, ""},
		{"/social", http.StatusOK, "YouTube", ""},
		// ?show_gophers_only=true must return the bare shell (no chrome).
		{"/?show_gophers_only=true", http.StatusOK, `<div id="app"></div>`, "site-header"},
		{"/nonsense", http.StatusNotFound, "", ""},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if w.Code != tt.wantStatus {
			t.Errorf("GET %s = %d, want %d", tt.path, w.Code, tt.wantStatus)
			continue
		}
		body := w.Body.String()
		if tt.wantBody != "" && !strings.Contains(body, tt.wantBody) {
			t.Errorf("GET %s body missing %q", tt.path, tt.wantBody)
		}
		if tt.wantNot != "" && strings.Contains(body, tt.wantNot) {
			t.Errorf("GET %s body should not contain %q", tt.path, tt.wantNot)
		}
	}

	// Short links keep working through the same handler.
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/gh", nil))
	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("GET /gh = %d, want 307", w.Code)
	}
	if got := w.Header().Get("Location"); got != "https://github.com/suapapa" {
		t.Errorf("GET /gh Location = %q", got)
	}
}

func TestLLMsTxtAndMarkdownNegotiation(t *testing.T) {
	origPages, origShell := pages, prerenderShell
	origLLMs, origMDPages := llmsTxt, markdownPages
	t.Cleanup(func() {
		pages, prerenderShell = origPages, origShell
		llmsTxt, markdownPages = origLLMs, origMDPages
	})

	links := ingress.Links{
		"/": {
			{Name: "/resume", Link: "https://homin.dev/resume/", Desc: "📋 Resume"},
			{Name: "/hidden", Link: "https://example.com/hidden", Desc: "ShouldNotAppear", Hide: true},
		},
		"/support": {
			{Name: "Github Sponsor", Link: "https://github.com/sponsors/suapapa", Desc: "🐟🥮 붕어빵 지원"},
		},
	}

	rendered, err := prerenderPages([]byte(testShell), links, "https://homin.dev")
	if err != nil {
		t.Fatalf("prerenderPages: %v", err)
	}
	pages, prerenderShell = rendered, []byte(testShell)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET(llmsTxtPath, llmsTxtHandler)
	r.GET(llmsFullTxtPath, llmsTxtHandler)
	r.GET("/", rootHandler)
	r.GET("/:path", pathHandler)

	// 1. /llms.txt and /llms-full.txt serve the full site index in Markdown.
	for _, endpoint := range []string{"/llms.txt", "/llms-full.txt", "/index.md"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, endpoint, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200", endpoint, w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
			t.Errorf("GET %s Content-Type = %q, want text/markdown", endpoint, ct)
		}
		body := w.Body.String()
		for _, want := range []string{
			"# HOMIN-DEV",
			"## Main (`/`)",
			"- [📋 Resume](https://homin.dev/resume/): Short link `/resume`",
			"## Support (`/support`)",
			"- [Github Sponsor](https://github.com/sponsors/suapapa): 🐟🥮 붕어빵 지원",
			"## Machine-Readable Endpoints",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s missing %q\n%s", endpoint, want, body)
			}
		}
		if strings.Contains(body, "ShouldNotAppear") {
			t.Errorf("GET %s leaked hidden link:\n%s", endpoint, body)
		}
	}

	// 2. Accept: text/markdown on "/" returns the full site markdown.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/markdown, text/html;q=0.8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET / (Accept: text/markdown) = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
		t.Errorf("GET / (Accept: text/markdown) Content-Type = %q", ct)
	}
	if !strings.Contains(w.Body.String(), "## Support (`/support`)") {
		t.Errorf("GET / (Accept: text/markdown) should include all sections:\n%s", w.Body.String())
	}

	// 3. Sub-route via Accept: text/markdown or .md suffix returns route-specific markdown.
	for _, subReq := range []*http.Request{
		func() *http.Request {
			r := httptest.NewRequest(http.MethodGet, "/support", nil)
			r.Header.Set("Accept", "text/markdown")
			return r
		}(),
		httptest.NewRequest(http.MethodGet, "/support.md", nil),
		httptest.NewRequest(http.MethodGet, "/support?format=md", nil),
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, subReq)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200", subReq.URL.String(), w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
			t.Errorf("GET %s Content-Type = %q", subReq.URL.String(), ct)
		}
		body := w.Body.String()
		for _, want := range []string{
			"# HOMIN-DEV — Support",
			"- [Github Sponsor](https://github.com/sponsors/suapapa): 🐟🥮 붕어빵 지원",
			"https://homin.dev/llms.txt",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s missing %q\n%s", subReq.URL.String(), want, body)
			}
		}
	}

	// 4. Browser Accept header (preferring text/html) still receives HTML + Link header.
	htmlReq := httptest.NewRequest(http.MethodGet, "/", nil)
	htmlReq.Header.Set("Accept", "text/html,application/xhtml+xml,text/markdown;q=0.5,*/*;q=0.8")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, htmlReq)
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("browser Accept got Content-Type = %q, want text/html", ct)
	}
	if linkHdr := w.Header().Get("Link"); !strings.Contains(linkHdr, `rel="alternate"`) || !strings.Contains(linkHdr, "/llms.txt") {
		t.Errorf("HTML response Link header = %q, want alternate /llms.txt", linkHdr)
	}
	if varyHdr := w.Header().Get("Vary"); varyHdr != "Accept" {
		t.Errorf("HTML response Vary header = %q, want Accept", varyHdr)
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
