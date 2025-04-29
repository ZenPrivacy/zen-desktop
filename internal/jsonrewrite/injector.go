package jsonrewrite

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/hostmatch"
	"github.com/ZenPrivacy/zen-desktop/internal/logger"
)

var ruleLineRegex = regexp.MustCompile(`^([^#]+)#%#//scriptlet\(\s*'([^']+)'\s*,\s*'([^']*)'(?:\s*,\s*'([^']*)')?(?:\s*,\s*'([^']*)')?\s*\)$`)

// Rule represents a parsed json-prune rule.
type Rule struct {
	PropsToRemove string
	RequiredProps string
	Stack         string
}

// ParsedRule is a container for hosts and a rule.
type ParsedRule struct {
	Hosts []string
	Rule  Rule
}

var (
	RuleRegex = regexp.MustCompile(`(?:#%#//scriptlet|##\+js)\((?:'|")json-prune`)

	primaryRuleRegex   = regexp.MustCompile(`(.*?)#%#(//scriptlet\(.*)`)
	exceptionRuleRegex = regexp.MustCompile(`(.*?)#@%#(//scriptlet\(.*)`)
)

type Injector struct {
	store hostmatch.HostMatcher[Rule]
}

func NewInjector() *Injector {
	return &Injector{
		store: *hostmatch.NewHostMatcher[Rule](),
	}
}

// ParseRule parses a raw json-prune rule line into a ParsedRule.
func ParseRule(raw string) (*ParsedRule, error) {
	match := ruleLineRegex.FindStringSubmatch(raw)
	if match == nil {
		return nil, errors.New("invalid json-prune rule format")
	}

	hostsPart := match[1]
	scriptletName := match[2]
	propsToRemove := match[3]
	requiredProps := match[4]
	stack := match[5]

	if !strings.HasPrefix(scriptletName, "json-prune") {
		return nil, errors.New("unsupported scriptlet: " + scriptletName)
	}

	hosts := strings.Split(hostsPart, ",")

	return &ParsedRule{
		Hosts: hosts,
		Rule: Rule{
			PropsToRemove: propsToRemove,
			RequiredProps: requiredProps,
			Stack:         stack,
		},
	}, nil
}

func (inj *Injector) AddRule(rule string) error {
	if match := primaryRuleRegex.FindStringSubmatch(rule); match != nil {
		parsed, err := ParseRule(rule)
		if err != nil {
			return fmt.Errorf("parse primary rule: %w", err)
		}
		for _, host := range parsed.Hosts {
			if err := inj.store.AddPrimaryRule(host, parsed.Rule); err != nil {
				return fmt.Errorf("add primary rule: %w", err)
			}
		}
		return nil
	}

	if match := exceptionRuleRegex.FindStringSubmatch(rule); match != nil {
		parsed, err := ParseRule(rule)
		if err != nil {
			return fmt.Errorf("parse exception rule: %w", err)
		}
		for _, host := range parsed.Hosts {
			if err := inj.store.AddExceptionRule(host, parsed.Rule); err != nil {
				return fmt.Errorf("add exception rule: %w", err)
			}
		}
		return nil
	}

	return errors.New("unsupported json-prune rule syntax")
}

func (inj *Injector) Inject(req *http.Request, res *http.Response) error {
	hostname := req.URL.Hostname()
	rules := inj.store.Get(hostname)
	log.Printf("got %d json-prune rules for %q", len(rules), logger.Redacted(hostname))
	if len(rules) == 0 {
		return nil
	}

	if len(rules) > 0 {
		for _, r := range rules {
			fmt.Printf("json-prune for %s: %s\n", hostname, r)
		}
	}

	for _, rule := range rules {
		if err := ModifyJSONResponse(res, rule.PropsToRemove, rule.RequiredProps); err != nil {
			return fmt.Errorf("inject json-prune: %w", err)
		}
	}

	return nil
}
