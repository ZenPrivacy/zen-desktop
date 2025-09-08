package extendedcss

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
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

	injectionTmp = template.Must(template.New("injection").Parse(`<script{{if .Nonce}} nonce="{{.Nonce}}"{{end}}>{{.Bundle}}(()=>{window.extendedCSS("{{.Rules}}")})();</script>`))
)

type store interface {
	AddPrimaryRule(hostnamePatterns string, body string) error
	AddExceptionRule(hostnamePatterns string, body string) error
	Get(hostname string) []string
}

// Injector injects extended CSS rules into HTML HTTP responses.
type Injector struct {
	// bundle contains the extended CSS JS bundle.
	bundle template.JS
	// store stores and retrieves extended CSS rules by hostname.
	store store
}

func NewInjectorWithDefaults() (*Injector, error) {
	store := hostmatch.NewHostMatcher[string]()
	return newInjector(defaultExtendedCSSBundle, store)
}

func newInjector(bundleData []byte, store store) (*Injector, error) {
	if bundleData == nil {
		return nil, errors.New("bundleData is nil")
	}
	if store == nil {
		return nil, errors.New("store is nil")
	}

	return &Injector{
		bundle: template.JS(bundleData),
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
	err := injectionTmp.Execute(&injection, struct {
		Nonce  string
		Bundle template.JS
		Rules  string
	}{
		Nonce:  nonce,
		Bundle: inj.bundle,
		Rules:  strings.Join(rules, "\n"),
	})
	if err != nil {
		return fmt.Errorf("execute template: %v", err)
	}

	if err := httprewrite.AppendHTMLHeadContents(res, injection.Bytes()); err != nil {
		return fmt.Errorf("append head contents: %v", err)
	}

	return nil
}
