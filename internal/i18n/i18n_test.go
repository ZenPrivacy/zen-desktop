package i18n

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		want   string
	}{
		{"de-DE", "de-DE"},
		{"de", "de-DE"},
		{"en", "en-US"},
		{"zh", "zh-CN"},
		{"zh-CN", "zh-CN"},
		{"zh-TW", "zh-TW"},
		{"xx", "en-US"},
		{"", "en-US"},
	}

	for _, tt := range tests {
		if got := resolve(tt.locale).Lang(); got != tt.want {
			t.Errorf("resolve(%q).Lang() = %q, want %q", tt.locale, got, tt.want)
		}
	}
}

func TestBareLocalesPointToCatalogues(t *testing.T) {
	t.Parallel()

	for bare, locale := range bareLocales {
		if _, ok := catalogues[locale]; !ok {
			t.Errorf("bareLocales[%q] = %q, which has no catalogue", bare, locale)
		}
	}
}

// TestCataloguesMatchFrontend fails when a locale is added to the frontend but
// not here, which would otherwise leave the browser pages in English for it.
func TestCataloguesMatchFrontend(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("../../frontend/src/i18n/locales/*.json")
	if err != nil {
		t.Fatalf("list frontend catalogues: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("found no frontend catalogues")
	}

	for _, file := range files {
		locale := strings.TrimSuffix(filepath.Base(file), ".json")
		if _, ok := catalogues[locale]; !ok {
			t.Errorf("%s: no catalogue in internal/i18n/locales", locale)
		}
		lang, _, _ := strings.Cut(locale, "-")
		if _, ok := bareLocales[lang]; !ok {
			t.Errorf("%s: bareLocales has no entry for %q", locale, lang)
		}
	}
}

func TestCataloguesMatchEnglish(t *testing.T) {
	t.Parallel()

	en := catalogues[fallbackLocale].messages
	for locale, c := range catalogues {
		for key := range en {
			if _, ok := c.messages[key]; !ok {
				t.Errorf("%s: %s is missing", locale, key)
			}
		}
		for key, s := range c.messages {
			if _, ok := en[key]; !ok {
				t.Errorf("%s: %s isn't in en-US", locale, key)
			}
			if s == "" {
				t.Errorf("%s: %s is empty", locale, key)
			}
		}
	}
}

func TestT(t *testing.T) {
	t.Parallel()

	c := &Catalogue{locale: "xx", messages: map[string]string{"blockPage.allow": "Translated"}}

	if got := c.T("blockPage.allow"); got != "Translated" {
		t.Errorf("T(present key) = %q, want %q", got, "Translated")
	}
	if got := c.T("blockPage.title"); got != "Request blocked" {
		t.Errorf("T(key missing from catalogue) = %q, want the en-US text", got)
	}
	if got := c.T("no.such.key"); got != "no.such.key" {
		t.Errorf("T(key missing everywhere) = %q, want the key", got)
	}
}
