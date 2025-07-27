package scriptlet

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	nonceRe = regexp.MustCompile(`'nonce-([^']+)'`)
)

// patchCSPHeaders mutates header as needed and tells the caller whether injection is
// possible. It returns the nonce to place on the <script> tag (empty string if
// no nonce is required) and ok=false when inline injection would be blocked
// (e.g., multiple conflicting nonces).
func patchCSPHeaders(h http.Header) (nonce string, ok bool) {
	nonces := collectNonces(h)
	switch len(nonces) {
	case 0:
		if needsNonce(h) {
			n := uuid.NewString()
			addNonceToCSP(h, n)
			return n, true
		}
		// inline allowed, but no nonce needed
		return "", true
	case 1:
		// get the single element
		for n := range nonces {
			return n, true
		}
	}

	// multiple distinct nonces -> skip
	// Browsers enforce *all* CSP headers: an inline script can satisfy
	// only one nonce, so with two different nonces every inline script
	// is blocked (CSP 2 §3.1, CSP 3 §7.2.3) - https://www.w3.org/TR/CSP2
	return "", false
}

// addNonceToCSP appends nonce *only* to directives that don't already contain a
// nonce and aren't "'none'".
func addNonceToCSP(h http.Header, nonce string) {
	const key = "Content-Security-Policy"
	lines := h.Values(key)
	if len(lines) == 0 {
		return
	}

	for i, line := range lines {
		changed := false
		dirs := strings.Split(line, ";")
		for j, d := range dirs {
			ld := strings.ToLower(strings.TrimSpace(d))

			if !strings.HasPrefix(ld, "script-src-elem") &&
				!strings.HasPrefix(ld, "script-src") &&
				!strings.HasPrefix(ld, "default-src") {
				continue
			}
			if strings.Contains(ld, "'none'") ||
				strings.Contains(ld, "'nonce-") {
				// leave directives that already forbid all or already have a nonce
				continue
			}
			token := "'nonce-" + nonce + "'"
			if !strings.Contains(d, token) {
				dirs[j] = strings.TrimSpace(d) + " " + token
				changed = true
			}
		}
		if changed {
			lines[i] = strings.Join(dirs, ";")
		}
	}
	h[key] = lines
}

// collectNonces returns distinct nonce values from script-related directives only.
func collectNonces(h http.Header) map[string]struct{} {
	out := make(map[string]struct{})
	for _, line := range h.Values("Content-Security-Policy") {
		for raw := range strings.SplitSeq(line, ";") {
			dir := strings.TrimSpace(raw)
			if dir == "" {
				continue
			}
			name := strings.ToLower(strings.Fields(dir)[0])
			if name != "script-src" && name != "script-src-elem" && name != "default-src" {
				continue
			}
			for _, m := range nonceRe.FindAllStringSubmatch(dir, -1) {
				out[m[1]] = struct{}{}
			}
		}
	}
	return out
}

// inlineAllowed reports whether a directive explicitly allows inline scripts
// *without* a nonce and *without* 'strict‑dynamic'.
func inlineAllowed(dir string) bool {
	lc := strings.ToLower(dir)
	if !strings.Contains(lc, "'unsafe-inline'") {
		return false
	}
	blockers := []string{"'strict-dynamic'", "'nonce-", "'sha256-", "'sha384-", "'sha512-'"}
	for _, b := range blockers {
		if strings.Contains(lc, b) {
			return false
		}
	}
	return true
}

// needsNonce decides whether a fresh nonce is required for inline JS to run.
// It returns false when inline execution is already allowed (DuckDuckGo case).
func needsNonce(h http.Header) bool {
	for _, line := range h.Values("Content-Security-Policy") {
		for _, dirName := range []string{"script-src-elem", "script-src", "default-src"} {
			lc := strings.ToLower(line)
			pos := strings.Index(lc, dirName)
			if pos == -1 {
				continue
			}
			end := strings.Index(line[pos:], ";")
			if end == -1 {
				end = len(line)
			} else {
				end += pos
			}
			dir := line[pos:end]
			return !inlineAllowed(dir)
		}
	}

	// No script controls means inline is implicitly allowed.
	return false
}
