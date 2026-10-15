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
	} {
		if !strings.Contains(root, want) {
			t.Errorf("root page missing %q", want)
		}
	}
	if strings.Contains(root, "ShouldNotAppear") {
		t.Errorf("hidden link leaked into the prerendered page")
	}

	social := pages["/social"]
	if social == "" {
		t.Fatalf(`no page for "/social"`)
	}
	for _, want := range []string{
		"<title>HOMIN-DEV — Social</title>",
		`<link rel="canonical" href="https://homin.dev/social"/>`,
		`content="https://homin.dev/social"`,
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
	t.Cleanup(func() { pages, prerenderShell, redirectByPath = origPages, origShell, origRedirects })

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

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
