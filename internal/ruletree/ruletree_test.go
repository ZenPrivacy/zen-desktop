/*
Benchmarks for building and querying the rule tree.

Run benchmarks with:

	// Run all benchmarks
	go test -bench=. ./internal/ruletree

	// Run BenchmarkMatch and export a memory profile
	go test -bench=Match$ -memprofile=mem.out ./internal/ruletree

	// Run BenchmarkLoadTree and export a CPU profile
	go test -bench=LoadTree -cpuprofile=cpu.out ./internal/ruletree

Inspect profile:

	go tool pprof -lines -focus=FindMatchingRulesReq mem.out

pprof tips:
  - -lines: show line-level metric attribution
  - -focus=FindMatchingRulesReq: restrict output to FindMatchingRulesReq; filters out setup/teardown noise
  - -ignore=runtime: hide nodes matching "runtime" (includes GC)
  - top: show top entries (usually somewhat hard to make sense of)
  - list <func>: show annotated source for the given function
  - web: generate an SVG call graph and open in browser
*/
package ruletree_test

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/ZenPrivacy/zen-desktop/internal/ruletree"
)

var (
	rnd         = rand.New(rand.NewSource(42))
	filterLists = []string{"testdata/easylist.txt", "testdata/easyprivacy.txt"}
)

func BenchmarkLoadTree(b *testing.B) {
	var rawLists [][]byte
	var totalBytes int64
	for _, filename := range filterLists {
		data, err := os.ReadFile(filename)
		if err != nil {
			b.Fatalf("read %s: %v", filename, err)
		}
		totalBytes += int64(len(data))
		rawLists = append(rawLists, data)
	}
	b.SetBytes(totalBytes)

	for b.Loop() {
		tree := ruletree.NewRuleTree[*spyData]()
		for _, data := range rawLists {
			scanner := bufio.NewScanner(bytes.NewReader(data))

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				if err := tree.Add(line, &spyData{}); err != nil {
					b.Fatalf("add rule %q: %v", line, err)
				}
			}

			if err := scanner.Err(); err != nil {
				b.Fatalf("scan: %v", err)
			}
		}
	}

	b.ReportAllocs()
}

func BenchmarkMatch(b *testing.B) {
	tree, err := loadTree()
	if err != nil {
		b.Fatalf("load tree: %v", err)
	}

	reqs, avgBytes, err := loadReqs()
	if err != nil {
		b.Fatalf("load reqs: %v", err)
	}
	b.SetBytes(avgBytes)

	var i int
	for b.Loop() {
		r := reqs[i]
		i++
		if i == len(reqs) {
			i = 0
		}
		tree.FindMatchingRulesReq(r)
	}

	b.ReportAllocs()
}

func BenchmarkMatchParallel(b *testing.B) {
	tree, err := loadTree()
	if err != nil {
		b.Fatalf("load tree: %v", err)
	}

	reqs, avgBytes, err := loadReqs()
	if err != nil {
		b.Fatalf("load reqs: %v", err)
	}
	b.SetBytes(avgBytes)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Custom rand Sources aren't thread-safe, and making parallel code
		// deterministic is hard anyway, so we just use the global rand.
		i := rand.Intn(len(reqs))
		for pb.Next() {
			r := reqs[i]
			i++
			if i == len(reqs) {
				i = 0
			}
			tree.FindMatchingRulesReq(r)
		}
	})

	b.ReportAllocs()
}

func loadTree() (*ruletree.RuleTree[*spyData], error) {
	tree := ruletree.NewRuleTree[*spyData]()

	for _, filename := range filterLists {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read %s: %v", filename, err)
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			err := tree.Add(line, &spyData{})
			if err != nil {
				return nil, fmt.Errorf("add rule %q: %v", line, err)
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan %s: %v", filename, err)
		}
	}
	return tree, nil
}

// loadReqs generates a list of HTTP requests from URLs in testdata/urls.txt
// and synthetic URLs. It returns the requests, average URL length in bytes,
// and any error encountered.
func loadReqs() ([]*http.Request, int64, error) {
	urls, err := loadURLs()
	if err != nil {
		return nil, 0, err
	}

	reqs := make([]*http.Request, len(urls))
	var totalURLBytes int64
	for i, u := range urls {
		reqs[i] = &http.Request{
			Method: http.MethodGet,
			URL:    u,
			Host:   u.Hostname(),
		}
		totalURLBytes += int64(len(u.String()))
	}
	avg := totalURLBytes / int64(len(urls))

	// Shuffle the elements to avoid ordering bias.
	rnd.Shuffle(len(reqs), func(i, j int) {
		reqs[i], reqs[j] = reqs[j], reqs[i]
	})

	return reqs, avg, nil
}

func loadURLs() ([]*url.URL, error) {
	const filename = "testdata/urls.txt"

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", filename, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var urls []*url.URL
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		u, err := url.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("invalid url: %s", line)
		}
		urls = append(urls, u)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %v", filename, err)
	}

	// Enrich with synthetic URLs
	urls = append(urls, genSyntheticURLs(len(urls))...)
	return urls, nil
}

func genSyntheticURLs(n int) []*url.URL {
	paths := []string{"ads", "static", "assets", "banner", "pixel", "collect", "metrics", "tag", "video", "js", "css", "img", "cdn", "widgets", "script", "log", "measure"}
	tlds := []string{"com", "net", "org", "io", "co"}
	params := []string{"utm_source", "utm_medium", "utm_campaign", "gclid", "fbclid", "adid", "track", "ref", "src"}
	out := make([]*url.URL, 0, n)

	for range n {
		var b strings.Builder
		if rnd.Float64() < 0.99 {
			b.WriteString("https://")
		} else {
			b.WriteString("http://")
		}

		labels := rnd.Intn(3) + 1
		for range labels {
			if rnd.Float64() < 0.5 {
				b.WriteString(paths[rnd.Intn(len(paths))])
			} else {
				b.WriteString(rndStr(rnd.Intn(7) + 3))
			}
			b.WriteByte('.')
		}
		b.WriteString(tlds[rnd.Intn(len(tlds))])
		b.WriteByte('/')

		segs := rnd.Intn(7)
		for range segs {
			if rnd.Float64() < 0.3 {
				b.WriteString(paths[rnd.Intn(len(paths))])
			} else {
				b.WriteString(rndStr(rnd.Intn(10) + 2))
			}
			b.WriteByte('/')
		}

		if rnd.Float64() < 0.7 {
			b.WriteByte('?')

			qn := rnd.Intn(4) + 1
			for j := 0; j < qn; j++ {
				if j > 0 {
					b.WriteByte('&')
				}
				if rnd.Float64() < 0.2 {
					b.WriteString(params[rnd.Intn(len(params))])
				} else {
					b.WriteString(rndStr(rnd.Intn(5) + 2))
				}
				b.WriteByte('=')
				b.WriteString(rndStr(rnd.Intn(8) + 3))
			}
		}

		if u, err := url.Parse(b.String()); err == nil {
			out = append(out, u)
		}
	}

	return out
}

func rndStr(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rnd.Intn(len(alphabet))]
	}
	return string(b)
}

type spyData struct {
	modifiers string
}

func (s *spyData) ShouldMatchRes(*http.Response) bool { return true }
func (s *spyData) ShouldMatchReq(*http.Request) bool  { return true }

func (s *spyData) ParseModifiers(modifiers string) error {
	s.modifiers = modifiers
	return nil
}
