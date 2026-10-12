package asset

import (
	"bytes"
	"testing"
)

func TestAddRuleRoutesByFirstMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rule     string
		wantPath string
		hostname string
		wantBody string
	}{
		{"so-net.ne.jp###ad-p3", cosmeticCSSPath, "so-net.ne.jp", "#ad-p3"},
		{"lunar.az#?##dle-content:has(.x)", extendedCSSPath, "lunar.az", "#dle-content:has(.x)"},
		{"safelink.asia#$##seo { color: red; }", cssRulePath, "safelink.asia", "#seo { color: red; }"},
		{`example.com##a[title="##+js(x)"]`, cosmeticCSSPath, "example.com", `a[title="##+js(x)"]`},
		{`example.com#%#var s = "##";`, jsRulePath, "example.com", `var s = "##";`},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			t.Parallel()

			e := newTestEngine(t)
			handled, err := e.AddRule(tt.rule, true)
			if err != nil {
				t.Fatalf("AddRule(%q): %v", tt.rule, err)
			}
			if !handled {
				t.Fatalf("AddRule(%q) not handled", tt.rule)
			}

			for _, path := range []string{cosmeticCSSPath, cssRulePath, extendedCSSPath, jsRulePath} {
				body, err := e.assetBytes(tt.hostname, path)
				if err != nil {
					t.Fatalf("assetBytes(%q, %q): %v", tt.hostname, path, err)
				}
				got := bytes.Contains(body, []byte(tt.wantBody))
				if want := path == tt.wantPath; got != want {
					t.Errorf("%s for %q contains %q = %v, want %v", path, tt.hostname, tt.wantBody, got, want)
				}
			}
		})
	}
}

func TestAddRuleLeavesCSSRuleFromUntrustedList(t *testing.T) {
	t.Parallel()

	e := newTestEngine(t)
	rule := "safelink.asia#$##seo { color: red; }"
	handled, err := e.AddRule(rule, false)
	if err != nil {
		t.Fatalf("AddRule(%q): %v", rule, err)
	}
	if handled {
		t.Errorf("AddRule(%q) handled a CSS rule from an untrusted list", rule)
	}
}

func TestExceptionWithIDSelectorCancelsGenericRule(t *testing.T) {
	t.Parallel()

	e := newTestEngine(t)
	for _, rule := range []string{"###ad-p3", "so-net.ne.jp#@##ad-p3"} {
		if _, err := e.AddRule(rule, false); err != nil {
			t.Fatalf("AddRule(%q): %v", rule, err)
		}
	}

	if body, _ := e.assetBytes("so-net.ne.jp", cosmeticCSSPath); body != nil {
		t.Errorf("cosmetic CSS for so-net.ne.jp = %q, want none", body)
	}
	if body, _ := e.assetBytes("example.org", cosmeticCSSPath); !bytes.Contains(body, []byte("#ad-p3")) {
		t.Errorf("cosmetic CSS for example.org = %q, want it to hide #ad-p3", body)
	}
}

func newTestEngine(t *testing.T) *Engine {
	t.Helper()

	e, err := NewEngine("zen.test")
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return e
}
