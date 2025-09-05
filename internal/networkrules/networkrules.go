package networkrules

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/networkrules/exceptionrule"
	"github.com/ZenPrivacy/zen-desktop/internal/networkrules/rule"
	"github.com/ZenPrivacy/zen-desktop/internal/ruletree"
)

var (
	// exceptionRegex matches exception rules.
	exceptionRegex = regexp.MustCompile(`^@@`)

	reHosts       = regexp.MustCompile(`^(?:0\.0\.0\.0|127\.0\.0\.1)\s(.+)`) // Replace ' ' with \s to support edge-cases mentioned below
	reHostsIgnore = regexp.MustCompile(`^(?:0\.0\.0\.0|broadcasthost|local|localhost(?:\.localdomain)?|ip6-\w+)$`)
)

type NetworkRules struct {
	regularRuleTree   *ruletree.RuleTree[*rule.Rule]
	exceptionRuleTree *ruletree.RuleTree[*exceptionrule.ExceptionRule]
}

func NewNetworkRules() *NetworkRules {
	regularTree := ruletree.NewRuleTree[*rule.Rule]()
	exceptionTree := ruletree.NewRuleTree[*exceptionrule.ExceptionRule]()

	return &NetworkRules{
		regularRuleTree:   regularTree,
		exceptionRuleTree: exceptionTree,
	}
}

func (nr *NetworkRules) ParseRule(rawRule string, filterName *string) (isException bool, err error) {
	if matches := reHosts.FindStringSubmatch(rawRule); matches != nil {
		hostsField := matches[1]
		if commentIndex := strings.IndexByte(hostsField, '#'); commentIndex != -1 {
			hostsField = hostsField[:commentIndex]
		}

		// An IP address may be followed by multiple hostnames.
		//
		// As stated in https://man.freebsd.org/cgi/man.cgi?hosts(5):
		// "Items are separated by any number of blanks and/or tab characters."
		hosts := strings.Fields(hostsField)

		for _, host := range hosts {
			if reHostsIgnore.MatchString(host) {
				continue
			}

			r := fmt.Sprintf("||%s^$document", host)
			if err := nr.regularRuleTree.Add(r, &rule.Rule{
				RawRule:    r,
				FilterName: filterName,
			}); err != nil {
				return false, fmt.Errorf("add host rule: %w", err)
			}
		}

		return false, nil
	}

	if exceptionRegex.MatchString(rawRule) {
		return true, nr.exceptionRuleTree.Add(rawRule[2:], &exceptionrule.ExceptionRule{
			RawRule:    rawRule,
			FilterName: filterName,
		})
	}

	return false, nr.regularRuleTree.Add(rawRule, &rule.Rule{
		RawRule:    rawRule,
		FilterName: filterName,
	})
}

func (nr *NetworkRules) ModifyRes(req *http.Request, res *http.Response) ([]rule.Rule, error) {
	regularRules := nr.regularRuleTree.FindMatchingRulesRes(req, res)
	if len(regularRules) == 0 {
		return nil, nil
	}

	exceptions := nr.exceptionRuleTree.FindMatchingRulesRes(req, res)

	var appliedRules []rule.Rule
outer:
	for _, r := range regularRules {
		for _, ex := range exceptions {
			if ex.Cancels(r) {
				continue outer
			}
		}

		m, err := r.ModifyRes(res)
		if err != nil {
			return nil, fmt.Errorf("apply %q: %v", r.RawRule, err)
		}
		if m {
			appliedRules = append(appliedRules, *r)
		}
	}

	return appliedRules, nil
}

func (nr *NetworkRules) ModifyReq(req *http.Request) (appliedRules []rule.Rule, shouldBlock bool, redirectURL string) {
	regularRules := nr.regularRuleTree.FindMatchingRulesReq(req)
	if len(regularRules) == 0 {
		return nil, false, ""
	}

	exceptions := nr.exceptionRuleTree.FindMatchingRulesReq(req)
	initialURL := req.URL.String()
outer:
	for _, r := range regularRules {
		for _, ex := range exceptions {
			if ex.Cancels(r) {
				continue outer
			}
		}
		if r.ShouldBlockReq(req) {
			return []rule.Rule{*r}, true, ""
		}
		if r.ModifyReq(req) {
			appliedRules = append(appliedRules, *r)
		}
	}

	finalURL := req.URL.String()

	if initialURL != finalURL {
		return appliedRules, false, finalURL
	}

	return appliedRules, false, ""
}
