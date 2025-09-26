package ruletree_test

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	"github.com/ZenPrivacy/zen-desktop/internal/ruletree"
)

var rnd = rand.New(rand.NewSource(1))

func BenchmarkLoadTree(b *testing.B) {
	for b.Loop() {
		_, err := loadTree()
		if err != nil {
			b.Fatalf("load tree: %v", err)
		}
	}

	b.ReportAllocs()
}

func BenchmarkMatch(b *testing.B) {
	checkStr := rndStr(5)
	if checkStr != "XVlBz" {
		b.Fatalf("random generator seed mismatch: %s", checkStr)
	}

	tree, err := loadTree()
	if err != nil {
		b.Fatalf("load tree: %v", err)
	}

	writeHeapProfile(b, "test_heap_before.pprof")
	for b.Loop() {
		tree.FindMatchingRulesReq(genRequest())
	}
	writeHeapProfile(b, "test_heap_after.pprof")
	runtime.KeepAlive(tree)

	b.ReportAllocs()
}

func loadTree() (*ruletree.RuleTree[*spyData], error) {
	tree := ruletree.NewRuleTree[*spyData]()

	for _, filename := range []string{"testdata/easylist.txt", "testdata/easyprivacy.txt"} {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read %s: %v", filename, err)
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := scanner.Text()

			if len(line) == 0 {
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

func rndStr(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}

func genRequest() *http.Request {
	url, _ := url.Parse(fmt.Sprintf("https://%s.example.com/%s", rndStr(10), rndStr(20)))
	return &http.Request{
		Method: "GET",
		URL:    url,
		Host:   "example.com",
	}
}

func writeHeapProfile(b *testing.B, name string) {
	f, err := os.Create(name)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		b.Fatal(err)
	}
}

type spyData struct {
	modifiers string
}

func (s *spyData) ShouldMatchRes(*http.Response) bool {
	return true
}

func (s *spyData) ShouldMatchReq(*http.Request) bool {
	return true
}

func (s *spyData) ParseModifiers(modifiers string) error {
	s.modifiers = modifiers
	return nil
}
