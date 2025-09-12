package csp

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPatchHeaders(t *testing.T) {
	t.Parallel()

	t.Run("don't modify header and return empty nonce if there is no CSP header", func(t *testing.T) {
		t.Parallel()

		h := http.Header{}
		nonce := PatchHeaders(h, InlineScript)

		if nonce != "" {
			t.Fatalf("expected empty nonce when no CSP present, got %q", nonce)
		}
		if got := h.Values("Content-Security-Policy"); len(got) != 0 {
			t.Fatalf("headers should be unchanged, got %v", got)
		}
	})

	t.Run("replace 'none' in most specific", func(t *testing.T) {
		t.Parallel()

		h := http.Header{}
		h.Add("Content-Security-Policy", "script-src-elem 'none'")

		nonce := PatchHeaders(h, InlineScript)
		if nonce == "" {
			t.Fatalf("expected nonce to be returned")
		}
		token := "'nonce-" + nonce + "'"

		got := strings.Join(h.Values("Content-Security-Policy"), ", ")
		expected := fmt.Sprintf("script-src-elem %s", token)
		if got != expected {
			t.Fatalf("expected header value %q, got %q", expected, got)
		}
	})
}

func TestPatchHeaders_NoncePriority_Script(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		cspLine       string
		wantNonce     bool
		wantDirective string
	}{
		{
			name:          "script-src-elem is most specific",
			cspLine:       "default-src 'self'; script-src 'self'; script-src-elem 'self'",
			wantNonce:     true,
			wantDirective: "script-src-elem",
		},
		{
			name:          "script-src fallback",
			cspLine:       "object-src 'none'; script-src 'self'",
			wantNonce:     true,
			wantDirective: "script-src",
		},
		{
			name:          "default-src fallback",
			cspLine:       "default-src 'self'",
			wantNonce:     true,
			wantDirective: "default-src",
		},
		{
			name:      "no blocking directives -> no nonce needed",
			cspLine:   "img-src *; object-src 'none'",
			wantNonce: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := http.Header{}
			h.Add("Content-Security-Policy", tc.cspLine)

			nonce := PatchHeaders(h, InlineScript)
			if tc.wantNonce && nonce == "" {
				t.Fatalf("expected nonce, got empty")
			}
			if !tc.wantNonce && nonce != "" {
				t.Fatalf("did not expect nonce, got %q", nonce)
			}
			if tc.wantNonce {
				if !dirHasNonce(h, tc.wantDirective, nonce) {
					t.Fatalf("nonce not placed in %s\nheader: %s",
						tc.wantDirective, h.Get("Content-Security-Policy"))
				}
			}
		})
	}
}

func TestPatchHeaders_NoncePriority_Style(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		cspLine       string
		wantNonce     bool
		wantDirective string
	}{
		{
			name:          "style-src-elem is most specific",
			cspLine:       "default-src 'self'; style-src 'self'; style-src-elem 'self'",
			wantNonce:     true,
			wantDirective: "style-src-elem",
		},
		{
			name:          "style-src fallback",
			cspLine:       "object-src 'none'; style-src 'self'",
			wantNonce:     true,
			wantDirective: "style-src",
		},
		{
			name:          "default-src fallback",
			cspLine:       "default-src 'self'",
			wantNonce:     true,
			wantDirective: "default-src",
		},
		{
			name:      "no blocking directives -> no nonce",
			cspLine:   "img-src *; object-src 'none'",
			wantNonce: false,
		},
	}

	for _, tc := range cases {
		h := http.Header{}
		h.Add("Content-Security-Policy", tc.cspLine)

		nonce := PatchHeaders(h, InlineStyle)

		if tc.wantNonce && nonce == "" {
			t.Errorf("%s: expected nonce, got empty", tc.name)
			continue
		}
		if !tc.wantNonce && nonce != "" {
			t.Errorf("%s: did not expect nonce, got %q", tc.name, nonce)
			continue
		}
		if !tc.wantNonce {
			continue
		}

		token := "'nonce-" + nonce + "'"
		found := false
		for _, line := range h.Values("Content-Security-Policy") {
			if strings.Contains(strings.ToLower(line), tc.wantDirective) && strings.Contains(line, token) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: nonce not placed in %s; header: %s", tc.name, tc.wantDirective, strings.Join(h.Values("Content-Security-Policy"), " | "))
		}
	}
}

func dirHasNonce(h http.Header, dir, nonce string) bool {
	token := "'nonce-" + nonce + "'"
	lines := h.Values("Content-Security-Policy")

	for _, line := range lines {
		rawDirs := strings.SplitSeq(line, ";")

		for raw := range rawDirs {
			d := strings.TrimSpace(raw)
			if d == "" {
				continue
			}
			name, value := cutDirective(d)
			if name == dir && strings.Contains(value, token) {
				return true
			}
		}
	}
	return false
}
