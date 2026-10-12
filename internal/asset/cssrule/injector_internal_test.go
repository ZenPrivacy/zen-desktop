package cssrule

import "testing"

func TestAddRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rule string
		want spyEntry
	}{
		{"safelink.asia#$##seo { display: none !important; }", spyEntry{exception: false, hostnamePatterns: "safelink.asia", body: "#seo { display: none !important; }"}},
		{"example.com#@$##ad { color: red; }", spyEntry{exception: true, hostnamePatterns: "example.com", body: "#ad { color: red; }"}},
		{"#$##ad { color: red; }", spyEntry{exception: false, hostnamePatterns: "", body: "#ad { color: red; }"}},
		{`example.com#$#a[href="#top"] { color: red; }`, spyEntry{exception: false, hostnamePatterns: "example.com", body: `a[href="#top"] { color: red; }`}},
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

func (s *spyStore) AddPrimaryRule(hostnamePatterns string, css string) error {
	s.entries = append(s.entries, spyEntry{exception: false, hostnamePatterns: hostnamePatterns, body: css})
	return nil
}

func (s *spyStore) AddExceptionRule(hostnamePatterns string, css string) error {
	s.entries = append(s.entries, spyEntry{exception: true, hostnamePatterns: hostnamePatterns, body: css})
	return nil
}

func (s *spyStore) Get(string) []string { return nil }
