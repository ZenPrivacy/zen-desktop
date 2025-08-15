package scriptlet

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	cspHeader     = "Content-Security-Policy"
	cspReportOnly = "Content-Security-Policy-Report-Only"
)

var directivePriority = map[string]int{
	"default-src":     1, // fallback
	"script-src":      2,
	"script-src-elem": 3, // most specific for <script> elements
}

// patchCSPHeaders mutates headers so an inline <script nonce=...> can run.
// Returns the nonce to place on the <script> tag.
func patchCSPHeaders(h http.Header) (nonce string) {
	// If there is no CSP at all, nothing to patch; return empty nonce.
	if len(h.Values(cspHeader)) == 0 && len(h.Values(cspReportOnly)) == 0 {
		return ""
	}
	n := newCSPNonce()

	enforcedPatched := patchOneHeader(h, cspHeader, n)
	reportOnlyPatched := patchOneHeader(h, cspReportOnly, n)

	if !enforcedPatched && !reportOnlyPatched {
		return ""
	}

	return n
}

func patchOneHeader(h http.Header, key, nonce string) (patched bool) {
	lines := h.Values(key)
	if len(lines) == 0 {
		return
	}

	nonceToken := "'nonce-" + nonce + "'"
	var changed bool

	// In case of multiple lines/policies, the browsers will select the most restrictive one.
	// For this reason, we modify each independently so they all allow the <script>.
	// See more: https://content-security-policy.com/examples/multiple-csp-headers/.
	for i, line := range lines {
		rawDirs := strings.Split(line, ";")

		// Find most specific directive controlling <script> elements on this line/policy.
		bestIdx := -1
		bestName := ""
		bestPrio := 0
		bestValue := ""

		for j, raw := range rawDirs {
			d := strings.TrimSpace(raw)
			if d == "" {
				continue
			}
			name, value := cutDirective(d)
			prio, ok := directivePriority[name]
			if !ok {
				continue
			}
			if prio > bestPrio {
				bestIdx, bestName, bestPrio, bestValue = j, name, prio, value
			}
		}

		// No relevant directive on this line; leave it as-is.
		if bestIdx == -1 {
			continue
		}

		// If policy already allows inline <script> elements, do nothing.
		if allowsInline(bestValue) {
			continue
		}

		var newValue string
		if bestValue == "'none'" {
			newValue = nonceToken
		} else {
			newValue = bestValue + " " + nonceToken
		}

		rawDirs[bestIdx] = bestName + " " + newValue
		lines[i] = strings.Join(rawDirs, ";")
		changed = true
	}

	if changed {
		h.Del(key)
		for _, v := range lines {
			h.Add(key, strings.TrimSpace(strings.Trim(v, " ;")))
		}
	}

	return changed
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
func allowsInline(sourceList string) bool {
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
			return false
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
