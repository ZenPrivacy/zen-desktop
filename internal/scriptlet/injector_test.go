package scriptlet_test

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/ZenPrivacy/zen-desktop/internal/scriptlet"
	"golang.org/x/net/html"
)

func TestInjectorPublic(t *testing.T) {
	t.Parallel()

	t.Run("makes an HTML-standards compliant injection with a generic scriptlet", func(t *testing.T) {
		t.Parallel()

		i := newInjector(t)
		err := i.AddRule(`#%#//scriptlet('prevent-xhr', 'example.com')`, false)
		if err != nil {
			t.Fatalf("failed to add rule: %v", err)
		}

		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		res := newBlankHTTPResponse(t)

		if err := i.Inject(req, res); err != nil {
			t.Errorf("failed to inject: %v", err)
		}

		if !hasScriptTag(t, res.Body) {
			t.Error("expected response body to contain at least one <script> tag, got 0")
		}
	})

	t.Run("makes an HTML-standards compliant injection with a hostname-specific scriptlet", func(t *testing.T) {
		t.Parallel()

		i := newInjector(t)
		err := i.AddRule(`news.example.com#%#//scriptlet('prevent-xhr', 'example.com')`, false)
		if err != nil {
			t.Fatalf("failed to add rule: %v", err)
		}

		req, err := http.NewRequest("GET", "http://news.example.com/frontpage", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		res := newBlankHTTPResponse(t)

		if err := i.Inject(req, res); err != nil {
			t.Errorf("failed to inject: %v", err)
		}

		if !hasScriptTag(t, res.Body) {
			t.Error("expected response body to contain at least one <script> tag, got 0")
		}
	})

	t.Run("doesn't inject scriptlets into a response without a matching rule", func(t *testing.T) {
		t.Parallel()

		i := newInjector(t)
		err := i.AddRule(`example.com#%#//scriptlet('prevent-xhr', 'example.com')`, false)
		if err != nil {
			t.Fatalf("failed to add rule: %v", err)
		}

		req, err := http.NewRequest("GET", "http://notexamplecom.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		res := newBlankHTTPResponse(t)

		if err := i.Inject(req, res); err != nil {
			t.Errorf("failed to inject: %v", err)
		}

		if hasScriptTag(t, res.Body) {
			t.Error("expected response body to contain 0 <script> tags, got 1")
		}
	})

	t.Run("dont add nonce to response without CSP header", func(t *testing.T) {
		t.Parallel()

		i := newInjector(t)
		err := i.AddRule(`example.com#%#//scriptlet('prevent-xhr', 'example.com')`, false)
		if err != nil {
			t.Fatalf("failed to add rule: %v", err)
		}

		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		res := newBlankHTTPResponse(t)

		if err := i.Inject(req, res); err != nil {
			t.Errorf("failed to inject: %v", err)
		}

		if !hasScriptTag(t, res.Body) {
			t.Fatalf("expected response body to contain at least one <script> tag, got 0")
		}

		if nonce := nonceFromBody(t, res.Body); nonce != "" {
			t.Fatalf("unexpected nonce attribute %q in <script>", nonce)
		}
	})
}

var nonceRx = regexp.MustCompile(`<script[^>]*\snonce="([^"]+)"`)

func TestInject_NoncePriority(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		csp           string
		wantNonce     bool
		wantDirective string
	}{
		{
			name:          "script-src-elem wins",
			csp:           "default-src 'self'; script-src 'self'; script-src-elem 'self'",
			wantNonce:     true,
			wantDirective: "script-src-elem",
		},
		{
			name:          "script-src fallback",
			csp:           "object-src 'none'; script-src 'self'",
			wantNonce:     true,
			wantDirective: "script-src",
		},
		{
			name:          "default-src fallback",
			csp:           "default-src 'self'",
			wantNonce:     true,
			wantDirective: "default-src",
		},
		{
			name:      "no blocking directives → no nonce",
			csp:       "img-src *; object-src 'none'",
			wantNonce: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("GET", "https://example.com/", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			res := newBlankHTTPResponse(t)
			res.Header.Add("Content-Security-Policy", tc.csp)

			i := newInjector(t)
			err = i.AddRule(`#%#//scriptlet('prevent-xhr','example.com')`, false)
			if err != nil {
				t.Fatalf("failed to add rule: %v", err)
			}

			if err := i.Inject(req, res); err != nil {
				t.Fatalf("inject: %v", err)
			}

			nonce := nonceFromBody(t, res.Body)
			if tc.wantNonce && nonce == "" {
				t.Errorf("expected nonce attribute in <script>, got none")
			}
			if !tc.wantNonce && nonce != "" {
				t.Errorf("did not expect nonce attribute, got %q", nonce)
			}
			if tc.wantNonce && !dirHasNonce(res.Header, tc.wantDirective, nonce) {
				t.Errorf("nonce not placed in %s directive\nheader: %s", tc.wantDirective, res.Header.Get("Content-Security-Policy"))
			}
		})
	}
}

func hasScriptTag(t *testing.T, body io.ReadCloser) bool {
	t.Helper()
	doc, err := html.Parse(body)
	if err != nil {
		t.Errorf("failed to parse response body after injection: %v", err)
	}

	var metScriptTag bool
	nodeStack := []*html.Node{doc}
	var currNode *html.Node
	for len(nodeStack) > 0 {
		currNode = nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]
		if currNode.Type == html.ElementNode && currNode.Data == "script" {
			metScriptTag = true
			break
		}

		for c := currNode.FirstChild; c != nil; c = c.NextSibling {
			nodeStack = append(nodeStack, c)
		}
	}
	return metScriptTag
}

func newBlankHTTPResponse(t *testing.T) *http.Response {
	t.Helper()
	body := io.NopCloser(strings.NewReader(`<html><head></head></html>`))
	header := http.Header{
		"Content-Type": []string{"text/html; charset=UTF-8"},
	}
	return &http.Response{
		Body:   body,
		Header: header,
	}
}

func newInjector(t *testing.T) *scriptlet.Injector {
	t.Helper()
	injector, err := scriptlet.NewInjectorWithDefaults()
	if err != nil {
		t.Fatalf("failed to create injector: %v", err)
	}
	return injector
}

func nonceFromBody(t *testing.T, r io.Reader) string {
	t.Helper()
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	m := nonceRx.FindSubmatch(raw)
	if len(m) < 2 {
		return ""
	}
	return string(m[1])
}

func dirHasNonce(h http.Header, dir, nonce string) bool {
	token := "'nonce-" + nonce + "'"
	for _, l := range h.Values("Content-Security-Policy") {
		if strings.Contains(strings.ToLower(l), dir) &&
			strings.Contains(l, token) {
			return true
		}
	}
	return false
}
