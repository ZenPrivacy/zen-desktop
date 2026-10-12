package cosmetic

import "testing"

func TestAddRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rule string
		want spyEntry
	}{
		{"example.com###ad", spyEntry{exception: false, hostnamePatterns: "example.com", body: "#ad"}},
		{"so-net.ne.jp#@##ad-p3", spyEntry{exception: true, hostnamePatterns: "so-net.ne.jp", body: "#ad-p3"}},
		{"###ad", spyEntry{exception: false, hostnamePatterns: "", body: "#ad"}},
		{"#@##ad", spyEntry{exception: true, hostnamePatterns: "", body: "#ad"}},
		{"example.com##div#ad", spyEntry{exception: false, hostnamePatterns: "example.com", body: "div#ad"}},
		{`example.com##a[href="#top"]`, spyEntry{exception: false, hostnamePatterns: "example.com", body: `a[href="#top"]`}},
		{`example.com#@#a[href="#top"]`, spyEntry{exception: true, hostnamePatterns: "example.com", body: `a[href="#top"]`}},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			t.Parallel()

			store := &spyStore{}
			inj := &Injector{store: store}
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

func (s *spyStore) AddPrimaryRule(hostnamePatterns string, selector string) error {
	s.entries = append(s.entries, spyEntry{exception: false, hostnamePatterns: hostnamePatterns, body: selector})
	return nil
}

func (s *spyStore) AddExceptionRule(hostnamePatterns string, selector string) error {
	s.entries = append(s.entries, spyEntry{exception: true, hostnamePatterns: hostnamePatterns, body: selector})
	return nil
}

func (s *spyStore) Get(string) []string { return nil }
