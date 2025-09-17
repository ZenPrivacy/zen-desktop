package cssrule

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"golang.org/x/net/html"
)

func TestInjector(t *testing.T) {
	t.Parallel()

	t.Run("nonce in the <style> attribute matches the Content-Security-Policy header", func(t *testing.T) {
		t.Parallel()

		const initialHTML = "<!doctype html><html><head><meta charset='utf-8'></head><body><h1>hi</h1></body></html>"
		inj := NewInjector()
		if err := inj.AddRule("example.com#$#.ads{visibility:none!important;}"); err != nil {
			t.Fatalf("AddRule: %v", err)
		}

		req := &http.Request{
			URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/"},
		}

		res := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader([]byte(initialHTML))),
		}
		res.Header.Set("Content-Security-Policy", "default-src 'none'; style-src 'none'")
		res.Header.Set("Content-Type", "text/html; charset=utf-8")

		if err := inj.Inject(req, res); err != nil {
			t.Fatalf("Inject: %v", err)
		}

		finalBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("read injected body: %v", err)
		}

		doc, err := html.Parse(bytes.NewReader(finalBody))
		if err != nil {
			t.Fatalf("parse injected HTML: %v", err)
		}

		head := findNode(doc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "head" })
		if head == nil {
			t.Fatal("<head> not found in injected HTML")
		}

		style := findNode(head, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "style" })
		if style == nil {
			t.Fatal("injected <style> not found in <head>")
		}

		styleNonce := getAttr(style, "nonce")
		if styleNonce == "" {
			t.Fatal("injected <style> missing nonce attribute")
		}

		cspHeader := res.Header.Get("Content-Security-Policy")
		if cspHeader == "" {
			t.Fatal("CSP header not set")
		}

		nonce := extractNonceFromCSP(cspHeader)
		if nonce == "" {
			t.Fatalf("nonce not found in CSP header: %q", cspHeader)
		}

		if nonce != styleNonce {
			t.Fatalf("nonce mismatch: CSP=%q style=%q", nonce, styleNonce)
		}
	})
}

func findNode(n *html.Node, pred func(*html.Node) bool) *html.Node {
	if n == nil {
		return nil
	}
	if pred(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, pred); found != nil {
			return found
		}
	}
	return nil
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func extractNonceFromCSP(csp string) string {
	re := regexp.MustCompile(`'nonce-([A-Za-z0-9+/_=-]+)'`)
	m := re.FindStringSubmatch(csp)
	if m == nil {
		return ""
	}
	return m[1]
}
