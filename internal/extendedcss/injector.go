package extendedcss

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/csp"
	"github.com/ZenPrivacy/zen-desktop/internal/hostmatch"
	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"github.com/ZenPrivacy/zen-desktop/internal/logger"
)

var (
	// RuleRegex matches extended CSS rules.
	RuleRegex = regexp.MustCompile(`.+?#@?\?#.+$`)

	primaryRuleRegex   = regexp.MustCompile(`(.+?)#\?#(.+)`)
	exceptionRuleRegex = regexp.MustCompile(`(.+?)#@\?#(.+)`)

	//go:embed bundle.js
	defaultExtendedCSSBundle []byte
	scriptOpeningTag         = []byte("<script>")
	scriptClosingTag         = []byte("</script>")
)

type store interface {
	AddPrimaryRule(hostnamePatterns string, body string) error
	AddExceptionRule(hostnamePatterns string, body string) error
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

// AddRule adds an extended CSS rule to the injector.
func (inj *Injector) AddRule(rule string) error {
	if match := primaryRuleRegex.FindStringSubmatch(rule); match != nil {
		hostnamePatters := match[1]
		selector := match[2]
		if err := inj.store.AddPrimaryRule(hostnamePatters, selector); err != nil {
			return fmt.Errorf("add primary rule: %v", err)
		}
		return nil
	} else if match := exceptionRuleRegex.FindStringSubmatch(rule); match != nil {
		hostnamePatterns := match[1]
		selector := match[2]
		if err := inj.store.AddExceptionRule(hostnamePatterns, selector); err != nil {
			return fmt.Errorf("add exception rule: %v", err)
		}
		return nil
	} else {
		return errors.New("unknown rule format")
	}
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

	nonce := csp.PatchHeaders(res.Header, csp.InlineScript)

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
