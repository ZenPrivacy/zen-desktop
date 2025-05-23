package scriptlet

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/hostmatch"
	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"github.com/ZenPrivacy/zen-desktop/internal/logger"
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

	var payload bytes.Buffer
	payload.Write(inj.bundle)
	payload.WriteString("(()=>{")
	for _, lst := range argLists {
		if err := lst.GenerateInjection(&payload); err != nil {
			return err
		}
	}
	payload.WriteString("})();")

	payloadBytes := payload.Bytes()
	if hasScriptControls(res.Header) {
		addHashToCSP(res.Header, sha256Base64(payloadBytes))
	}

	var injection bytes.Buffer
	injection.Write(scriptOpeningTag)
	injection.Write(payloadBytes)
	injection.Write(scriptClosingTag)

	// Appending the scriptlets bundle to the head of the document aligns with the behavior of uBlock Origin:
	// - https://github.com/gorhill/uBlock/blob/d7ae3a185eddeae0f12d07149c1f0ddd11fd0c47/platform/firefox/vapi-background-ext.js#L373-L375
	// - https://github.com/gorhill/uBlock/blob/d7ae3a185eddeae0f12d07149c1f0ddd11fd0c47/platform/chromium/vapi-background-ext.js#L223-L226
	if err := httprewrite.AppendHTMLHeadContents(res, injection.Bytes()); err != nil {
		return fmt.Errorf("append head contents: %w", err)
	}

	return nil
}

func addHashToCSP(h http.Header, hash string) {
	token := "'sha256-" + strings.TrimSpace(hash) + "'"

	values := h.Values("Content-Security-Policy")
	if len(values) == 0 {
		return
	}

	h.Del("Content-Security-Policy")
	for _, csp := range values {
		parts := strings.Split(csp, ";")
		patched := false

		for i, p := range parts {
			dir := strings.ToLower(strings.TrimSpace(p))
			switch {
			case strings.HasPrefix(dir, "script-src"):
				if !strings.Contains(p, token) {
					parts[i] = strings.TrimSpace(p) + " " + token
				}
				patched = true

			case strings.HasPrefix(dir, "default-src"):
				// if there is no script-src at all, we'll add the hash here
				if !patched && !strings.Contains(p, token) {
					parts[i] = strings.TrimSpace(p) + " " + token
				}
			}
		}

		h.Add("Content-Security-Policy", strings.Join(parts, ";"))
	}
}

func sha256Base64(b []byte) string {
	sum := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(sum[:])
}

func hasScriptControls(h http.Header) bool {
	for _, csp := range h.Values("Content-Security-Policy") {
		lc := strings.ToLower(csp)
		if strings.Contains(lc, "script-src") || strings.Contains(lc, "default-src") {
			return true
		}
	}
	return false
}
