package extendedcss

import "testing"

func TestAddRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rule string
		want spyEntry
	}{
		{"lunar.az#?##dle-content > .fulltitle:has(.x)", spyEntry{exception: false, hostnamePatterns: "lunar.az", body: "#dle-content > .fulltitle:has(.x)"}},
		{"example.com#@?##ad:has(.x)", spyEntry{exception: true, hostnamePatterns: "example.com", body: "#ad:has(.x)"}},
		{"example.com###ad:has-text(x)", spyEntry{exception: false, hostnamePatterns: "example.com", body: "#ad:has-text(x)"}},
		{"example.com#@##ad:has-text(x)", spyEntry{exception: true, hostnamePatterns: "example.com", body: "#ad:has-text(x)"}},
		{"example.com##div:has-text(#top)", spyEntry{exception: false, hostnamePatterns: "example.com", body: "div:has-text(#top)"}},
		{`example.com#?#a[href="#top"]:has(.x)`, spyEntry{exception: false, hostnamePatterns: "example.com", body: `a[href="#top"]:has(.x)`}},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			t.Parallel()

			store := &spyStore{}
			inj, err := newInjector([]byte{}, store)
			if err != nil {
				t.Fatalf("newInjector: %v", err)
			}
			if err := inj.AddRule(tt.rule); err != nil {
				t.Fatalf("AddRule(%q): %v", tt.rule, err)
			}

			if len(store.entries) != 1 || store.entries[0] != tt.want {
				t.Errorf("AddRule(%q) stored %+v, want [%+v]", tt.rule, store.entries, tt.want)
			}
		})
	}
}

type spyEntry struct {
	exception        bool
	hostnamePatterns string
	body             string
}

type spyStore struct {
	entries []spyEntry
}

func (s *spyStore) AddPrimaryRule(hostnamePatterns string, body string) error {
	s.entries = append(s.entries, spyEntry{exception: false, hostnamePatterns: hostnamePatterns, body: body})
	return nil
}

func (s *spyStore) AddExceptionRule(hostnamePatterns string, body string) error {
	s.entries = append(s.entries, spyEntry{exception: true, hostnamePatterns: hostnamePatterns, body: body})
	return nil
}

func (s *spyStore) Get(string) []string { return nil }
