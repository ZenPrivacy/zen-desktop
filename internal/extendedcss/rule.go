package extendedcss

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// RuleRegex matches extended-css rules.
	RuleRegex = regexp.MustCompile(`^([^#]*?)#\?#(.+)$`)
)

type parsedRule struct {
	hostnamePatterns string
	selector         string
}

// parseRule parses an extended-css rule and returns the hostname patterns and selector.
func parseRule(rule string) (*parsedRule, error) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return nil, errors.New("empty rule")
	}

	matches := RuleRegex.FindStringSubmatch(rule)
	if len(matches) != 3 {
		return nil, errors.New("invalid extended-css rule format")
	}

	hostnamePatterns := strings.TrimSpace(matches[1])
	selector := strings.TrimSpace(matches[2])

	if selector == "" {
		return nil, errors.New("empty selector")
	}

	// If no hostname patterns are provided, default to all hosts
	if hostnamePatterns == "" {
		hostnamePatterns = "*"
	}

	return &parsedRule{
		hostnamePatterns: hostnamePatterns,
		selector:         selector,
	}, nil
}
