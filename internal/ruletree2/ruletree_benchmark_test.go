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
package ruletree2_test

import (
	"bufio"
	"bytes"
	"math/rand"
	"net/http"
	"os"
	"testing"

	"github.com/ZenPrivacy/zen-desktop/internal/ruletree2"
)

const baseSeed = 42

var (
	rnd         = rand.New(rand.NewSource(baseSeed)) // #nosec G404 -- Not used for cryptographic purposes.
	filterLists = []string{"testdata/easylist.txt", "testdata/easyprivacy.txt"}
)

func BenchmarkLoadTree(b *testing.B) {
	rawLists := make([][]byte, 0, len(filterLists))
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
		tree := ruletree2.New[*spyData]()
		for _, data := range rawLists {
			scanner := bufio.NewScanner(bytes.NewReader(data))

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				if err := tree.Insert(line, &spyData{}); err != nil {
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

type spyData struct {
	modifiers string
}

func (s *spyData) ShouldMatchRes(*http.Response) bool { return true }
func (s *spyData) ShouldMatchReq(*http.Request) bool  { return true }

func (s *spyData) ParseModifiers(modifiers string) error {
	s.modifiers = modifiers
	return nil
}
