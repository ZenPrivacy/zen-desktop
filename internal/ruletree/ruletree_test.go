package ruletree_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/ZenPrivacy/zen-desktop/internal/ruletree"
)

func TestProfileHeap(t *testing.T) {
	if testing.Short() || os.Getenv("RULETREE_PROFILE") != "1" {
		t.Skip("skipping heap profile (set RULETREE_PROFILE=1 to run)")
	}

	var m runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m)
	t.Logf("Before: HeapAlloc=%s HeapSys=%s NumGC=%d\n",
		humanizeBytes(m.HeapAlloc),
		humanizeBytes(m.HeapSys),
		m.NumGC)

	tree := ruletree.NewRuleTree[*spyData]()

	for _, filename := range []string{"testdata/easylist.txt", "testdata/easyprivacy.txt"} {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := scanner.Text()

			if len(line) == 0 {
				continue
			}

			err := tree.Add(line, &spyData{})
			if err != nil {
				t.Errorf("add rule %q: %v", line, err)
			}
		}

		if err := scanner.Err(); err != nil {
			t.Fatalf("scan %s: %v", filename, err)
		}
	}

	defer runtime.KeepAlive(tree)

	runtime.GC()
	runtime.ReadMemStats(&m)
	t.Logf("After: HeapAlloc=%s HeapSys=%s NumGC=%d\n",
		humanizeBytes(m.HeapAlloc),
		humanizeBytes(m.HeapSys),
		m.NumGC)
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

func humanizeBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
