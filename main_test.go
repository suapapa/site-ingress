package main

import (
	"testing"

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
