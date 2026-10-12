package jsrule

import "testing"

func TestAddRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rule string
		want spyEntry
	}{
		{"example.com#%#window.x = 1;", spyEntry{exception: false, hostnamePatterns: "example.com", body: "window.x = 1;"}},
		{`example.com#%#var s = "#%#";`, spyEntry{exception: false, hostnamePatterns: "example.com", body: `var s = "#%#";`}},
		{`example.com#@%#var s = "#@%#";`, spyEntry{exception: true, hostnamePatterns: "example.com", body: `var s = "#@%#";`}},
		{"#%#window.x = 1;", spyEntry{exception: false, hostnamePatterns: "", body: "window.x = 1;"}},
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

func (s *spyStore) AddPrimaryRule(hostnamePatterns string, script string) error {
	s.entries = append(s.entries, spyEntry{exception: false, hostnamePatterns: hostnamePatterns, body: script})
	return nil
}

func (s *spyStore) AddExceptionRule(hostnamePatterns string, script string) error {
	s.entries = append(s.entries, spyEntry{exception: true, hostnamePatterns: hostnamePatterns, body: script})
	return nil
}

func (s *spyStore) Get(string) []string { return nil }
