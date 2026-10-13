package ruletree

import (
	"bufio"
	"os"
	"testing"
)

func TestDomainKey(t *testing.T) {
	t.Parallel()

	accepted := map[string]string{
		"||example.com^":      "example.com",
		"||a^":                "a",
		"||ads-1.my_cdn.net^": "ads-1.my_cdn.net",
	}
	for pattern, want := range accepted {
		if got, ok := domainKey(pattern); !ok || got != want {
			t.Errorf("domainKey(%q) = %q, %v; want %q, true", pattern, got, ok, want)
		}
	}

	rejected := []string{
		"||^",
		"|example.com^",
		"||example.com",
		"||example.com^|",
		"||Example.com^",
		"||.example.com^",
		"||example.com.^",
		"||*.example.com^",
		"||example.com/ads^",
		"||example.com:8080^",
		"||exa%6dple.com^",
	}
	for _, pattern := range rejected {
		if got, ok := domainKey(pattern); ok {
			t.Errorf("domainKey(%q) = %q, true; want false", pattern, got)
		}
	}
}

// TestDomainBucketMatchesTree checks that patterns stored in the domain bucket
// match exactly the URLs they would match if stored in the prefix tree.
func TestDomainBucketMatchesTree(t *testing.T) {
	t.Parallel()

	lines := append(readLines(t, "testdata/easylist.txt"), readLines(t, "testdata/easyprivacy.txt")...)
	urls := readLines(t, "testdata/urls.txt")
	// ||ad.doubleclick.net^ is in easylist.txt. These cover how the host can
	// end, and dots outside the host.
	urls = append(urls,
		"https://ad.doubleclick.net",
		"https://ad.doubleclick.net/x",
		"https://ad.doubleclick.net?x=1",
		"https://ad.doubleclick.net#x",
		"https://ad.doubleclick.net:8080/",
		"https://ad.doubleclick.net;x",
		"https://ad.doubleclick.net.",
		"https://ad.doubleclick.net.evil.com/",
		"https://ad.doubleclick.net%2f/",
		"https://ad.doubleclick.net-x/",
		"https://xad.doubleclick.net/",
		"https://a.b.ad.doubleclick.net/x",
		"https://example.com#a.ad.doubleclick.net",
		"https://example.com?a.ad.doubleclick.net",
		"https://example.com/a.ad.doubleclick.net",
		"https://user:pass@ad.doubleclick.net/",
		"https://[::1]/",
		"ad.doubleclick.net",
	)

	reference := New[string]()
	bucketed := New[string]()
	for _, line := range lines {
		reference.insertTree(line, line)
		bucketed.Insert(line, line)
	}
	if n := len(bucketed.domains); n < 1000 {
		t.Fatalf("only %d test domains went to the domain bucket", n)
	}

	compare := func(stage string) {
		for _, u := range urls {
			want := reference.Get(u)
			got := bucketed.Get(u)
			if !equalSets(got, want) {
				t.Errorf("%s: Get(%q): got=%v, want=%v", stage, u, got, want)
			}
		}
	}

	compare("unfrozen")

	bucketed.Compact()
	if len(bucketed.domains) != 0 || len(bucketed.frozen) == 0 {
		t.Fatalf("Compact left %d map entries and %d frozen entries", len(bucketed.domains), len(bucketed.frozen))
	}
	compare("frozen")
}

func TestDomainBucketInsertAfterCompact(t *testing.T) {
	t.Parallel()

	tr := New[string]()
	tr.Insert("||example.com^", "A")
	tr.Insert("||ads.example.com^", "B")
	tr.Insert("||zzz.net^", "C")
	tr.Compact()

	tr.Insert("||example.com^", "D")
	tr.Insert("||tracker.net^", "E")

	check := func(stage string) {
		cases := map[string][]string{
			"https://example.com/":      {"A", "D"},
			"https://ads.example.com/x": {"A", "B", "D"},
			"https://tracker.net":       {"E"},
			"https://zzz.net/":          {"C"},
		}
		for u, want := range cases {
			if got := tr.Get(u); !equalSets(got, want) {
				t.Errorf("%s: Get(%q): got=%v, want=%v", stage, u, got, want)
			}
		}
	}

	check("before second compact")
	tr.Compact()
	check("after second compact")
}

func readLines(t *testing.T, name string) []string {
	t.Helper()

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("open %q: %v", name, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %q: %v", name, err)
	}
	return lines
}
