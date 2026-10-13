package networkrules

import (
	"sync"
	"testing"
)

// TestRuleStoreConcurrentInsertAndGet is meant for -race: the block page's
// allow button inserts rules of every pattern kind while requests are matched.
func TestRuleStoreConcurrentInsertAndGet(t *testing.T) {
	t.Parallel()

	s := newRuleStore[string]()
	s.Compact()

	var wg sync.WaitGroup
	wg.Go(func() {
		for range 200 {
			for _, pattern := range []string{"", `/banner\d+/`, "||ads.example.com^", "/ads/*"} {
				if err := s.Insert(pattern, pattern); err != nil {
					t.Errorf("insert %q: %v", pattern, err)
					return
				}
			}
		}
	})
	for range 4 {
		wg.Go(func() {
			for range 200 {
				s.Get("https://ads.example.com/ads/banner1")
			}
		})
	}
	wg.Wait()
}
