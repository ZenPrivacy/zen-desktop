package extendedcss

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/hostmatch"
	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"github.com/ZenPrivacy/zen-desktop/internal/logger"
	"github.com/google/uuid"
)

var (
	//go:embed bundle.js
	defaultExtendedCSSBundle []byte
	scriptOpeningTag         = []byte("<script>")
	scriptClosingTag         = []byte("</script>")
)

type store interface {
	AddPrimaryRule(hostnamePatterns string, body string) error
	Get(hostname string) []string
}

// Injector injects extended-css rules into HTML HTTP responses.
type Injector struct {
	// bundle contains the extended-css JS bundle.
	bundle []byte
	// store stores and retrieves extended-css rules by hostname.
	store store
}

func NewInjectorWithDefaults() (*Injector, error) {
	store := hostmatch.NewHostMatcher[string]()
	return newInjector(defaultExtendedCSSBundle, store)
}

// newInjector creates a new Injector with the embedded extended-css bundle.
func newInjector(bundleData []byte, store store) (*Injector, error) {
	if bundleData == nil {
		return nil, errors.New("bundleData is nil")
	}
	if store == nil {
		return nil, errors.New("store is nil")
	}

	return &Injector{
		bundle: bundleData,
		store:  store,
	}, nil
}

// AddRule adds an extended-css rule to the injector.
func (inj *Injector) AddRule(rule string) error {
	parsed, err := parseRule(rule)
	if err != nil {
		return fmt.Errorf("parse extended-css rule: %v", err)
	}

	if err := inj.store.AddPrimaryRule(parsed.hostnamePatterns, parsed.selector); err != nil {
		return fmt.Errorf("add extended-css rule to store: %v", err)
	}

	return nil
}

// Inject injects extended-css rules into a given HTTP HTML response.
//
// On error, the caller may proceed as if the function had not been called.
func (inj *Injector) Inject(req *http.Request, res *http.Response) error {
	hostname := req.URL.Hostname()
	rules := inj.store.Get(hostname)
	log.Printf("got %d extended-css rules for %q", len(rules), logger.Redacted(hostname))
	if len(rules) == 0 {
		return nil
	}

	nonce := ""
	if hasScriptControls(res.Header) {
		nonce = uuid.NewString()
		addNonceToCSP(res.Header, nonce)
	}

	var injection bytes.Buffer
	if nonce == "" {
		injection.Write(scriptOpeningTag)
	} else {
		fmt.Fprintf(&injection, `<script nonce="%s">`, nonce)
	}
	injection.Write(inj.bundle)
	injection.WriteString("(()=>{window.extendedCSS(")

	rulesString := strings.Join(rules, "\n")
	fmt.Fprintf(&injection, `%q`, rulesString)

	injection.WriteString(")})();")
	injection.Write(scriptClosingTag)

	if err := httprewrite.AppendHTMLHeadContents(res, injection.Bytes()); err != nil {
		return fmt.Errorf("append head contents: %w", err)
	}

	return nil
}

func addNonceToCSP(h http.Header, nonce string) {
	const key = "Content-Security-Policy"
	lines := h[key]
	if len(lines) == 0 {
		return
	}

	// https://w3c.github.io/webappsec-csp/#directive-fallback-list
	prio := []string{"script-src-elem", "script-src", "default-src"}

	lineIdx, dirMatch := -1, ""
outer:
	for _, dir := range prio {
		for i, l := range lines {
			if strings.Contains(strings.ToLower(l), dir) {
				lineIdx, dirMatch = i, dir
				break outer
			}
		}
	}
	if lineIdx == -1 {
		return
	}

	parts := strings.Split(lines[lineIdx], ";")
	for i, p := range parts {
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(p)), dirMatch) {
			continue
		}

		token := " 'nonce-" + nonce + "'"
		switch {
		case strings.Contains(strings.ToLower(p), "'unsafe-inline'"):
			// Intentionally empty. 'unsafe-inline' allows the execution of inline scripts, and is incompatible with 'nonce-' directives.
		case strings.Contains(strings.ToLower(p), "'none'"):
			parts[i] = strings.Replace(p, "'none'", token, 1)
		default:
			parts[i] = strings.TrimSpace(p) + token
		}
		break
	}
	h[key][lineIdx] = strings.Join(parts, "; ")
}

func hasScriptControls(h http.Header) bool {
	for _, csp := range h.Values("Content-Security-Policy") {
		lc := strings.ToLower(csp)
		if strings.Contains(lc, "script-src-elem") || strings.Contains(lc, "script-src") || strings.Contains(lc, "default-src") {
			return true
		}
	}
	return false
}
