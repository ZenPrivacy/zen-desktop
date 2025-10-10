package csp

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPatchHeaders(t *testing.T) {
	t.Parallel()

	t.Run("don't modify header and return empty nonce if there is no CSP header", func(t *testing.T) {
		t.Parallel()

		res := &http.Response{Header: http.Header{}, Body: http.NoBody}
		nonce, err := PatchHeaders(res, InlineScript)
		if err != nil {
			t.Fatalf("patch headers: %v", err)
		}

		if nonce != "" {
			t.Fatalf("expected empty nonce when no CSP present, got %q", nonce)
		}
		if got := res.Header.Values("Content-Security-Policy"); len(got) != 0 {
			t.Fatalf("headers should be unchanged, got %v", got)
		}
	})

	t.Run("replace 'none' in most specific", func(t *testing.T) {
		t.Parallel()

		res := &http.Response{Header: http.Header{}, Body: http.NoBody}
		res.Header.Add("Content-Security-Policy", "script-src-elem 'none'")

		nonce, err := PatchHeaders(res, InlineScript)
		if err != nil {
			t.Fatalf("patch headers: %v", err)
		}
		if nonce == "" {
			t.Fatalf("expected nonce to be returned")
		}
		token := "'nonce-" + nonce + "'"

		got := strings.Join(res.Header.Values("Content-Security-Policy"), ", ")
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

			res := &http.Response{Header: http.Header{}, Body: http.NoBody}
			res.Header.Add("Content-Security-Policy", tc.cspLine)

			nonce, err := PatchHeaders(res, InlineScript)
			if err != nil {
				t.Fatalf("patch headers: %v", err)
			}
			if tc.wantNonce && nonce == "" {
				t.Fatalf("expected nonce, got empty")
			}
			if !tc.wantNonce && nonce != "" {
				t.Fatalf("did not expect nonce, got %q", nonce)
			}
			if tc.wantNonce {
				if !dirHasNonce(res.Header, tc.wantDirective, nonce) {
					t.Fatalf("nonce not placed in %s\nheader: %s",
						tc.wantDirective, res.Header.Get("Content-Security-Policy"))
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
		res := &http.Response{Header: http.Header{}, Body: http.NoBody}
		res.Header.Add("Content-Security-Policy", tc.cspLine)

		nonce, err := PatchHeaders(res, InlineStyle)
		if err != nil {
			t.Fatalf("patch headers: %v", err)
		}

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
		for _, line := range res.Header.Values("Content-Security-Policy") {
			if strings.Contains(strings.ToLower(line), tc.wantDirective) && strings.Contains(line, token) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: nonce not placed in %s; header: %s", tc.name, tc.wantDirective, strings.Join(res.Header.Values("Content-Security-Policy"), " | "))
		}
	}
}

func TestPatchHeaders_Meta(t *testing.T) {
	t.Parallel()

	t.Run("meta only", func(t *testing.T) {
		t.Parallel()

		htmlBody := `<html><head><meta http-equiv="Content-Security-Policy" content="script-src 'none'"></head><body></body></html>`
		res := &http.Response{
			Header: http.Header{},
			Body:   io.NopCloser(strings.NewReader(htmlBody)),
		}
		res.Header.Set("Content-Type", "text/html; charset=utf-8")

		nonce, err := PatchHeaders(res, InlineScript)
		if err != nil {
			t.Fatalf("patch headers: %v", err)
		}
		if nonce == "" {
			t.Fatalf("expected nonce to be returned")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		res.Body.Close()

		token := "'nonce-" + nonce + "'"
		escapedToken := strings.ReplaceAll(token, "'", "&#39;")
		bodyStr := string(body)
		if !strings.Contains(bodyStr, token) && !strings.Contains(bodyStr, escapedToken) {
			t.Fatalf("expected meta CSP to contain %q (or %q), body: %s", token, escapedToken, bodyStr)
		}
	})

	t.Run("header and meta", func(t *testing.T) {
		t.Parallel()

		htmlBody := `<html><head><meta http-equiv="Content-Security-Policy" content="script-src 'none'"></head><body></body></html>`
		res := &http.Response{
			Header: http.Header{},
			Body:   io.NopCloser(strings.NewReader(htmlBody)),
		}
		res.Header.Set("Content-Type", "text/html; charset=utf-8")
		res.Header.Add("Content-Security-Policy", "script-src 'none'")

		nonce, err := PatchHeaders(res, InlineScript)
		if err != nil {
			t.Fatalf("patch headers: %v", err)
		}
		if nonce == "" {
			t.Fatalf("expected nonce to be returned")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		res.Body.Close()

		token := "'nonce-" + nonce + "'"
		escapedToken := strings.ReplaceAll(token, "'", "&#39;")
		bodyStr := string(body)
		if !strings.Contains(bodyStr, token) && !strings.Contains(bodyStr, escapedToken) {
			t.Fatalf("expected meta CSP to contain %q (or %q), body: %s", token, escapedToken, bodyStr)
		}

		if !dirHasNonce(res.Header, "script-src", nonce) {
			t.Fatalf("expected header to contain nonce; header: %s", res.Header.Get("Content-Security-Policy"))
		}
	})
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
