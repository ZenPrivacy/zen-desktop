package ruletree2

import (
	"testing"
)

// asSet turns a slice into a set (ignores duplicates).
func asSet(xs []string) map[string]struct{} {
	m := make(map[string]struct{}, len(xs))
	for _, x := range xs {
		m[x] = struct{}{}
	}
	return m
}

func equalSets(a, b []string) bool {
	am := asSet(a)
	bm := asSet(b)
	if len(am) != len(bm) {
		return false
	}
	for k := range am {
		if _, ok := bm[k]; !ok {
			return false
		}
	}
	return true
}

func TestInsert_EmptyPatternError(t *testing.T) {
	tr := New[string]()
	if err := tr.Insert("", "X"); err == nil {
		t.Fatalf("expected error on empty pattern, got nil")
	}
}

func TestGet_Table(t *testing.T) {
	type tc struct {
		name   string
		rules  []string
		url    string
		expect []string
	}
	tests := []tc{
		// --- Wildcards ---
		{
			name:   "Wildcard matches zero length",
			rules:  []string{"ab*cd"},
			url:    "abcd",
			expect: []string{"ab*cd"},
		},
		{
			name:   "Wildcard matches many chars",
			rules:  []string{"a*c"},
			url:    "abbbbbc",
			expect: []string{"a*c"},
		},

		// --- Separators (^ ) ---
		{
			name:   "Separator matches '?' after token",
			rules:  []string{"ads^"},
			url:    "http://example.com/ads?x=1",
			expect: []string{"ads^"},
		},
		{
			name:   "Separator matches multiple separator chars",
			rules:  []string{"ads^"},
			url:    "http://example.com/ads???x=1",
			expect: []string{"ads^"},
		},
		{
			name:   "Separator matches end of address",
			rules:  []string{"ads^"},
			url:    "http://example.com/ads",
			expect: []string{"ads^"},
		},
		{
			name:   "Separator must not match letters/digits/_-.%",
			rules:  []string{"ads^"},
			url:    "http://example.com/adsx",
			expect: []string{},
		},

		// --- Start/End anchors (|) ---
		{
			name:   "Start anchor matches only at beginning",
			rules:  []string{"|http://example.org"},
			url:    "http://example.org/page",
			expect: []string{"|http://example.org"},
		},
		{
			name:   "Start anchor does not match when URL contains the string later",
			rules:  []string{"|http://example.org"},
			url:    "http://domain.com/?url=http://example.org",
			expect: []string{},
		},
		{
			name:   "End anchor matches exact suffix",
			rules:  []string{"swf|"},
			url:    "http://example.com/annoyingflash.swf",
			expect: []string{"swf|"},
		},
		{
			name:   "End anchor does not match when there are trailing chars",
			rules:  []string{"swf|"},
			url:    "http://example.com/swf/index.html",
			expect: []string{},
		},

		// --- Domain-root (||) basics ---
		{
			name:   "Domain-root matches main domain path",
			rules:  []string{"||example.com/ads/*"},
			url:    "http://example.com/ads/banner.jpg",
			expect: []string{"||example.com/ads/*"},
		},
		{
			name:   "Domain-root matches subdomain path",
			rules:  []string{"||example.com/ads/*"},
			url:    "https://sub.example.com/ads/otherbanner.jpg",
			expect: []string{"||example.com/ads/*"},
		},
		{
			name:   "Domain-root matches across protocols",
			rules:  []string{"||example.com/ads/*"},
			url:    "wss://example.com/ads/x",
			expect: []string{"||example.com/ads/*"},
		},
		{
			name:   "Domain-root does not match lookalike domain",
			rules:  []string{"||example.com/ads/*"},
			url:    "http://notexample.com/ads/x",
			expect: []string{},
		},

		// --- Domain/root boundary with separator ---
		{
			name:   "Domain-root + separator matches end-of-host",
			rules:  []string{"||example.com^"},
			url:    "https://sub.example.com",
			expect: []string{"||example.com^"},
		},
		{
			name:   "Domain-root + separator matches slash after host",
			rules:  []string{"||example.com^"},
			url:    "https://sub.example.com/path",
			expect: []string{"||example.com^"},
		},
		{
			name:   "Domain-root + separator does not match badexample.com",
			rules:  []string{"||example.com^"},
			url:    "https://badexample.com/",
			expect: []string{},
		},

		// --- Intersecting (overlapping) rule sets (>5 rules) ---
		{
			name: "Intersecting rules - only some match",
			rules: []string{
				"||example.com/*",
				"||example.com/ads/*",
				"|http://example.com/ads/top",
				"*/top*",
				"ads^",
				"swf|",
				"|https://sub.example.com/strict|",
			},
			url: "http://sub.example.com/ads/top?x=1",
			expect: []string{
				"||example.com/*",
				"||example.com/ads/*",
				"*/top*",
				"ads^",
			},
		},
		{
			name: "Intersecting rules - anchored exact vs broad domain",
			rules: []string{
				"||example.com^",
				"||example.com/ads/*",
				"||example.net^",
				"|https://example.com/login",
				"|https://sub.example.com/strict|",
				"*/strict*",
			},
			url: "https://sub.example.com/strict",
			expect: []string{
				"||example.com^",
				"|https://sub.example.com/strict|",
				"*/strict*",
			},
		},

		// --- Another intersecting set (>5 rules), mixed schemes ---
		{
			name: "Intersecting rules - login page on https",
			rules: []string{
				"||example.com^",
				"||example.com/ads/*",
				"||tracking.example.com/*",
				"|https://example.com/login|",
				"|http://example.com/",
				"*/login*",
			},
			url: "https://example.com/login",
			expect: []string{
				"||example.com^",
				"|https://example.com/login|",
				"*/login*",
			},
		},

		// --- Plain substring patterns without anchors ---
		{
			name:   "Plain substring matches anywhere (no anchors)",
			rules:  []string{"http://example.org"},
			url:    "http://domain.com/?u=http://example.org",
			expect: []string{"http://example.org"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New[string]()
			for _, r := range tt.rules {
				if err := tr.Insert(r, r); err != nil {
					t.Fatalf("Insert(%q) error: %v", r, err)
				}
			}
			if tt.name == "Intersecting rules - only some match" {
				println("here")
			}
			got := tr.Get(tt.url)
			if !equalSets(got, tt.expect) {
				t.Fatalf("url=%q\nrules=%v\ngot=%v\nwant=%v", tt.url, tt.rules, got, tt.expect)
			}
		})
	}
}

func TestInsert_DuplicateValuesAppend(t *testing.T) {
	tr := New[string]()
	// same pattern, two different IDs
	if err := tr.Insert("||example.com/ads/*", "R1a"); err != nil {
		t.Fatal(err)
	}
	if err := tr.Insert("||example.com/ads/*", "R1b"); err != nil {
		t.Fatal(err)
	}
	got := tr.Get("http://example.com/ads/x")
	set := asSet(got)
	if _, ok := set["R1a"]; !ok {
		t.Fatalf("missing R1a in %v", got)
	}
	if _, ok := set["R1b"]; !ok {
		t.Fatalf("missing R1b in %v", got)
	}
}
