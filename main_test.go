package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/suapapa/site-ingress/internal/ingress"
)

func TestTidyPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"foo", "/foo"},
		{"/foo", "/foo"},
		{"/foo/", "/foo"},
		{"foo/", "/foo"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := tidyPath(tt.in); got != tt.want {
				t.Errorf("tidyPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildRedirectIndex(t *testing.T) {
	links := ingress.Links{
		"": []*ingress.Link{{Name: "gh", Link: "https://github.com"}},
	}
	idx := buildRedirectIndex(links)
	if got := idx["/gh"]; got != "https://github.com" {
		t.Fatalf("idx[/gh] = %q", got)
	}
}

func TestBuildRedirectIndex_firstSortedPrefixWins(t *testing.T) {
	links := ingress.Links{
		"z": []*ingress.Link{{Name: "dup", Link: "second"}},
		"a": []*ingress.Link{{Name: "dup", Link: "first"}},
	}
	idx := buildRedirectIndex(links)
	if got := idx["/dup"]; got != "first" {
		t.Fatalf("idx[/dup] = %q, want first prefix in sort order to win", got)
	}
}

// TestSitemapLoc pins the fix for the broken <loc> values: an absolute link must survive
// as-is, instead of being mangled by path.Join into /https:/homin.dev/resume.
func TestSitemapLoc(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		link   string
		want   string
		ok     bool
	}{
		{"absolute keeps scheme and trailing slash", "/", "https://homin.dev/resume/", "https://homin.dev/resume/", true},
		{"absolute under non-root prefix is not prefixed", "/archive", "https://homin.dev/blog/", "https://homin.dev/blog/", true},
		{"absolute without trailing slash", "/", "https://homin.dev/social", "https://homin.dev/social", true},
		{"relative is resolved against siteURL and prefix", "/x", "foo/bar", "https://homin.dev/x/foo/bar", true},
		{"root-relative is resolved against siteURL", "", "/abs/path", "https://homin.dev/abs/path", true},
		{"mailto is rejected", "/", "mailto:i@homin.dev", "", false},
		{"javascript is rejected", "/", "javascript:alert(1)", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := sitemapLoc(tt.prefix, tt.link)
			if ok != tt.ok {
				t.Fatalf("sitemapLoc(%q, %q) ok = %v, want %v (loc = %q)", tt.prefix, tt.link, ok, tt.ok, got)
			}
			if got != tt.want {
				t.Errorf("sitemapLoc(%q, %q) = %q, want %q", tt.prefix, tt.link, got, tt.want)
			}
		})
	}
}

func TestSitemapHandler(t *testing.T) {
	orig := siteLinks
	t.Cleanup(func() { siteLinks = orig })

	siteLinks = ingress.Links{
		"/": {
			{Name: "/resume", Link: "https://homin.dev/resume/", SiteMap: true},
			{Name: "/gh", Link: "https://github.com/suapapa"},          // no site_map flag
			{Name: "/mail", Link: "mailto:i@homin.dev", SiteMap: true}, // not crawlable
		},
		"/archive": {
			{Name: "/blog", Link: "https://homin.dev/blog/", SiteMap: true},
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/sitemap.xml", sitemapHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/xml") {
		t.Errorf("Content-Type = %q", ct)
	}

	body := w.Body.String()
	for _, want := range []string{
		"<loc>https://homin.dev/resume/</loc>",
		"<loc>https://homin.dev/blog/</loc>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("sitemap missing %q\n%s", want, body)
		}
	}
	for _, bad := range []string{"/https:/", "mailto:", "github.com"} {
		if strings.Contains(body, bad) {
			t.Errorf("sitemap must not contain %q\n%s", bad, body)
		}
	}
}

// TestSitemapHandler_deterministic guards against Go's random map iteration order leaking
// into the generated document (which would make ETag/cache behaviour unstable).
func TestSitemapHandler_deterministic(t *testing.T) {
	orig := siteLinks
	t.Cleanup(func() { siteLinks = orig })

	siteLinks = ingress.Links{}
	for _, p := range []string{"/a", "/b", "/c", "/d", "/e", "/f"} {
		siteLinks[p] = []*ingress.Link{{Name: p + "/x", Link: "https://homin.dev" + p + "/", SiteMap: true}}
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/sitemap.xml", sitemapHandler)

	render := func() string {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
		return w.Body.String()
	}

	first := render()
	for i := 0; i < 20; i++ {
		if got := render(); got != first {
			t.Fatalf("sitemap output is not deterministic\n--- first ---\n%s\n--- got ---\n%s", first, got)
		}
	}
}

func BenchmarkRedirectLookup(b *testing.B) {
	idx := buildRedirectIndex(ingress.Links{
		"": {
			{Name: "a", Link: "1"},
			{Name: "b", Link: "2"},
			{Name: "target", Link: "https://example.com"},
		},
	})
	dest := tidyPath("target")
	if idx[dest] == "" {
		b.Fatal("expected indexed path")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = idx[dest]
	}
}
