package networkrules

import (
	"fmt"
	"testing"
)

func TestParseRuleParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		rawRule          string
		wantPattern      string
		wantModifiers    string
		wantHasModifiers bool
	}{
		{
			name:             "normal rule with modifiers",
			rawRule:          "||example.com^$script",
			wantPattern:      "||example.com^",
			wantModifiers:    "script",
			wantHasModifiers: true,
		},
		{
			name:             "regexp rule with dollar anchor",
			rawRule:          `/foo$/$script`,
			wantPattern:      `/foo$/`,
			wantModifiers:    "script",
			wantHasModifiers: true,
		},
		{
			name:             "empty pattern with modifiers",
			rawRule:          "$script",
			wantPattern:      "",
			wantModifiers:    "script",
			wantHasModifiers: true,
		},
		{
			name:             "regexp rule with escaped slash",
			rawRule:          `/foo\/bar/$script`,
			wantPattern:      `/foo\/bar/`,
			wantModifiers:    "script",
			wantHasModifiers: true,
		},
		{
			name:             "regexp rule with unescaped slash",
			rawRule:          `/example.com/[0-9a-z]*.php/$domain=example.com`,
			wantPattern:      `/example.com/[0-9a-z]*.php/`,
			wantModifiers:    "domain=example.com",
			wantHasModifiers: true,
		},
		{
			name:             "regexp rule with regexp modifier",
			rawRule:          `/^https:\/\/d[0-9a-z]{12,13}\.cloudfront\.net\/loader\.min\.js$/$script,third-party,header=etag:/^W\/"[0-9a-f]{32}"$/`,
			wantPattern:      `/^https:\/\/d[0-9a-z]{12,13}\.cloudfront\.net\/loader\.min\.js$/`,
			wantModifiers:    `script,third-party,header=etag:/^W\/"[0-9a-f]{32}"$/`,
			wantHasModifiers: true,
		},
		{
			name:             "slash-prefixed normal rule",
			rawRule:          `/path/like/filter$script`,
			wantPattern:      `/path/like/filter`,
			wantModifiers:    "script",
			wantHasModifiers: true,
		},
		{
			name:             "rule with trailing dollar",
			rawRule:          "||example.com^$",
			wantPattern:      "||example.com^",
			wantHasModifiers: true,
		},
		{
			name:        "regexp rule without modifiers",
			rawRule:     `/ads\d+/`,
			wantPattern: `/ads\d+/`,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("parses %s", tt.name), func(t *testing.T) {
			t.Parallel()

			gotPattern, gotModifiers, gotHasModifiers := parseRuleParts(tt.rawRule)
			if gotPattern != tt.wantPattern {
				t.Errorf("pattern = %q, want %q", gotPattern, tt.wantPattern)
			}
			if gotModifiers != tt.wantModifiers || gotHasModifiers != tt.wantHasModifiers {
				t.Errorf("modifiers = %q, %t, want %q, %t", gotModifiers, gotHasModifiers, tt.wantModifiers, tt.wantHasModifiers)
			}
		})
	}
}
