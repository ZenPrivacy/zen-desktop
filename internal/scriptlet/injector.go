package scriptlet

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
	defaultScriptletsBundle []byte
	scriptOpeningTag        = []byte("<script>")
	scriptClosingTag        = []byte("</script>")
)

type store interface {
	AddPrimaryRule(hostnamePatterns string, body argList) error
	AddExceptionRule(hostnamePatterns string, body argList) error
	Get(hostname string) []argList
}

// Injector injects scriptlets into HTML HTTP responses.
type Injector struct {
	// bundle contains the scriptlets JS bundle.
	bundle []byte
	// store stores and retrieves scriptlets by hostname.
	store store
}

func NewInjectorWithDefaults() (*Injector, error) {
	store := hostmatch.NewHostMatcher[argList]()
	return newInjector(defaultScriptletsBundle, store)
}

// newInjector creates a new Injector with the embedded scriptlets.
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

// Inject injects scriptlets into a given HTTP HTML response.
//
// On error, the caller may proceed as if the function had not been called.
func (inj *Injector) Inject(req *http.Request, res *http.Response) error {
	hostname := req.URL.Hostname()
	argLists := inj.store.Get(hostname)
	log.Printf("got %d scriptlets for %q", len(argLists), logger.Redacted(hostname))
	if len(argLists) == 0 {
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
	injection.WriteString("(()=>{")
	for _, argLst := range argLists {
		if err := argLst.GenerateInjection(&injection); err != nil {
			return fmt.Errorf("generate injection for scriptlet %q: %v", argLst, err)
		}
	}
	injection.WriteString("})();")
	injection.Write(scriptClosingTag)

	// Appending the scriptlets bundle to the head of the document aligns with the behavior of uBlock Origin:
	// - https://github.com/gorhill/uBlock/blob/d7ae3a185eddeae0f12d07149c1f0ddd11fd0c47/platform/firefox/vapi-background-ext.js#L373-L375
	// - https://github.com/gorhill/uBlock/blob/d7ae3a185eddeae0f12d07149c1f0ddd11fd0c47/platform/chromium/vapi-background-ext.js#L223-L226
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
		p = strings.ToLower(p)
		switch {
		case strings.Contains(p, "'unsafe-inline'") && !strings.Contains(p, "'strict-dynamic'"):
			// Intentionally empty. 'unsafe-inline' allows the execution of inline scripts, and is incompatible with 'nonce-' directives.
		case strings.Contains(p, "'none'"):
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
