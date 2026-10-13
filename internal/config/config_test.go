package config

import (
	"testing"

	"github.com/irbis-sh/zen-desktop/internal/i18n"
)

// TestLocaleReachesI18n checks that the browser pages follow the locale the user
// picks, both when it changes and after a restart.
func TestLocaleReachesI18n(t *testing.T) {
	prevConfigDir := ConfigDir
	ConfigDir = t.TempDir()
	t.Cleanup(func() {
		ConfigDir = prevConfigDir
		i18n.SetLocale("")
	})

	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.SetLocale("de"); err != nil {
		t.Fatalf("SetLocale: %v", err)
	}
	if got := i18n.Current().Lang(); got != "de-DE" {
		t.Errorf("after SetLocale, i18n locale = %q, want de-DE", got)
	}

	i18n.SetLocale("")
	if _, err := New(); err != nil {
		t.Fatalf("New after restart: %v", err)
	}
	if got := i18n.Current().Lang(); got != "de-DE" {
		t.Errorf("after restart, i18n locale = %q, want de-DE", got)
	}
}
