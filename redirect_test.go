package main

import "testing"

var (
	testCase = []string{
		"",
		"/",
		"/wiki/doku.php",
		"/blog/tags/하프라이프/index.xml",
		"/blog/ananke/css/main.min.css",
	}
)

func BenchmarkGenFakeSubdomain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, t := range testCase {
			getSubdomain(t)
		}
	}
}

func BenchmarkGenFakeSubdomain2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, t := range testCase {
			getSubdomain2(t)
		}
	}
}

func TestGetSudomain2Good(t *testing.T) {
	for _, v := range testCase {
		a1, b1 := getSubdomain(v)
		a2, b2 := getSubdomain2(v)
		if !(a1 == a2 && b1 == b2) {
			t.Errorf("v=%v: a1=%v, a2=%v, b1=%v, b2=%v",
				v, a1, a2, b1, b2,
			)
		}
	}
}
