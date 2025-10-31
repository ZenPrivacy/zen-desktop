package csp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"golang.org/x/net/html"
)

const (
	cspHeader     = "Content-Security-Policy"
	cspReportOnly = "Content-Security-Policy-Report-Only"
)

type inlineKind int

const (
	InlineScript inlineKind = iota
	InlineStyle
)

// PatchHeaders mutates CSP headers and meta tags so an inline <script> or <style>
// element can run. Returns the nonce to place on the inline tag. Returns "" if
// neither headers required patching nor a CSP meta tag rewrite was scheduled.
func PatchHeaders(res *http.Response, kind inlineKind) (string, error) {
	if res == nil {
		return "", nil
	}

	nonce := newCSPNonce()

	metaPatched, err := patchMetaCSPs(res, nonce, kind)
	if err != nil {
		return "", fmt.Errorf("patch meta CSP: %w", err)
	}

	enforcedPatched := patchOneHeader(res.Header, cspHeader, nonce, kind)
	reportOnlyPatched := patchOneHeader(res.Header, cspReportOnly, nonce, kind)

	if !metaPatched && !enforcedPatched && !reportOnlyPatched {
		return "", nil
	}

	return nonce, nil
}

func patchOneHeader(h http.Header, key, nonce string, kind inlineKind) (patched bool) {
	lines := h.Values(key)
	if len(lines) == 0 {
		return
	}

	patchedLines, changed := patchPolicies(lines, nonce, kind)
	if changed {
		h.Del(key)
		for _, v := range patchedLines {
			h.Add(key, strings.TrimSpace(strings.Trim(v, " ;")))
		}
	}

	return changed
}

func patchPolicies(policies []string, nonce string, kind inlineKind) ([]string, bool) {
	if len(policies) == 0 {
		return policies, false
	}

	nonceToken := "'nonce-" + nonce + "'"
	changed := false
	// In case of multiple lines/policies, the browsers will select the most restrictive one.
	// For this reason, we modify each independently so they all allow the inline tag.
	// See more: https://content-security-policy.com/examples/multiple-csp-headers/.
	for i, line := range policies {
		rawDirs := strings.Split(line, ";")
		// Find the most specific directive governing this kind on this line/policy.
		bestIdx, bestName, bestPrio, bestValue := -1, "", 0, ""
		for j, raw := range rawDirs {
			d := strings.TrimSpace(raw)
			if d == "" {
				continue
			}
			name, value := cutDirective(d)
			if prio := directivePriority(kind, name); prio > bestPrio {
				bestIdx, bestName, bestPrio, bestValue = j, name, prio, value
			}
		}
		// No relevant directive on this line; leave it as-is.
		if bestIdx == -1 {
			continue
		}

		if allowsInline(kind, bestValue) {
			continue
		}

		var newValue string
		switch bestValue {
		case "'none'":
			newValue = nonceToken
		default:
			newValue = bestValue + " " + nonceToken
		}

		rawDirs[bestIdx] = bestName + " " + newValue
		policies[i] = strings.Join(rawDirs, ";")
		changed = true
	}

	if changed {
		for i, policy := range policies {
			policies[i] = strings.TrimSpace(strings.Trim(policy, " ;"))
		}
	}

	return policies, changed
}

func patchMetaCSPs(res *http.Response, nonce string, kind inlineKind) (bool, error) {
	if res.Body == nil || res.Body == http.NoBody {
		return false, nil
	}

	err := httprewrite.StreamRewrite(res, func(src io.ReadCloser, dst *io.PipeWriter) {
		defer src.Close()

		z := html.NewTokenizer(src)
		var foundBody bool
	loop:
		for {
			tt := z.Next()
			switch tt {
			case html.ErrorToken:
				dst.CloseWithError(z.Err())
				return
			case html.StartTagToken, html.SelfClosingTagToken:
				raw := append([]byte(nil), z.Raw()...)
				tok := z.Token()

				if strings.EqualFold(tok.Data, "meta") {
					httpEquiv := ""
					contentIdx := -1
					for i, attr := range tok.Attr {
						switch strings.ToLower(attr.Key) {
						case "http-equiv":
							httpEquiv = attr.Val
						case "content":
							contentIdx = i
						}
					}

					if strings.EqualFold(httpEquiv, "content-security-policy") && contentIdx >= 0 {
						patchedPolicies, changed := patchPolicies([]string{tok.Attr[contentIdx].Val}, nonce, kind)
						if changed {
							tok.Attr[contentIdx].Val = patchedPolicies[0]
							if _, err := dst.Write([]byte(tok.String())); err != nil {
								dst.CloseWithError(err)
								return
							}
							continue
						}
					}
				}

				if _, err := dst.Write(raw); err != nil {
					dst.CloseWithError(err)
					return
				}
				if tt == html.StartTagToken && strings.EqualFold(tok.Data, "body") {
					foundBody = true
					break loop
				}
			case html.EndTagToken:
				raw := z.Raw()
				if _, err := dst.Write(raw); err != nil {
					dst.CloseWithError(err)
					return
				}
				tok := z.Token()
				// The head is over, no more meta tags.
				if strings.EqualFold(tok.Data, "head") {
					foundBody = true
					break loop
				}
			default:
				if _, err := dst.Write(z.Raw()); err != nil {
					dst.CloseWithError(err)
					return
				}
			}
		}

		if foundBody {
			if _, err := io.Copy(dst, src); err != nil {
				dst.CloseWithError(err)
				return
			}
		}

		dst.Close()
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

// cutDirective splits "name [value...]" -> (lowercased name, value without leading and trailing whitespace).
func cutDirective(s string) (string, string) {
	name, rest, ok := strings.Cut(s, " ")
	if !ok {
		return strings.ToLower(name), ""
	}
	return strings.ToLower(name), strings.TrimSpace(rest)
}

// newCSPNonce returns a cryptographically random base64 string.
func newCSPNonce() string {
	// From https://www.w3.org/TR/CSP3/#security-nonces:
	// The generated value SHOULD be at least 128 bits long (before encoding), and
	// SHOULD be generated via a cryptographically secure random number generator in order to ensure that the value is difficult for an attacker to predict.
	// The code below satisfies both of these requirements.
	var b [18]byte // 144 bits
	rand.Read(b[:])
	return base64.StdEncoding.EncodeToString(b[:])
}

// allowsInline implements CSP3 "Does a source list allow all inline behavior for type?" algorithm.
// True iff 'unsafe-inline' is present AND there is NO nonce/hash AND NO 'strict-dynamic'.
//
// Reference: https://www.w3.org/TR/CSP3/#allow-all-inline
func allowsInline(kind inlineKind, sourceList string) bool {
	sourceList = strings.TrimSpace(sourceList)
	if sourceList == "" {
		return false
	}
	tokens := strings.Fields(sourceList)

	var unsafeInline bool
	for _, t := range tokens {
		switch t {
		case "'unsafe-inline'":
			unsafeInline = true
		case "'strict-dynamic'":
			if kind == InlineScript {
				return false
			}
		default:
			if isNonceOrHashSource(t) {
				return false
			}
		}
	}
	return unsafeInline
}

func isNonceOrHashSource(t string) bool {
	if len(t) < 3 || t[0] != '\'' || t[len(t)-1] != '\'' {
		return false
	}
	inner := t[1 : len(t)-1]
	return strings.HasPrefix(inner, "nonce-") ||
		strings.HasPrefix(inner, "sha256-") ||
		strings.HasPrefix(inner, "sha384-") ||
		strings.HasPrefix(inner, "sha512-")
}

func directivePriority(kind inlineKind, name string) int {
	switch kind {
	case InlineScript:
		switch name {
		case "script-src-elem":
			return 3
		case "script-src":
			return 2
		case "default-src":
			return 1
		}
	case InlineStyle:
		switch name {
		case "style-src-elem":
			return 3
		case "style-src":
			return 2
		case "default-src":
			return 1
		}
	}
	return 0
}
