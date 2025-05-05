package scriptlet

import (
	"errors"
	"fmt"
	"regexp"
)

// TODO: rethink and reimplement trusted rule handling.

var (
	// RuleRegex matches patterns for scriptlet rules in two formats:
	// 1. `#%#//scriptlet` or `#@%#//scriptlet` for canonical rules.
	// 2. `##+js` or `#@#+js` for uBlock-style rules.
	RuleRegex = regexp.MustCompile(`(?:#@?%#\/\/scriptlet)|(?:#@?#\+js)`)

	canonicalPrimary        = regexp.MustCompile(`(.*)#%#\/\/scriptlet\((.+)\)`)
	canonicalExceptionRegex = regexp.MustCompile(`(.*)#@%#\/\/scriptlet\((.+)\)`)
	ublockPrimaryRegex      = regexp.MustCompile(`(.*)##\+js\((.+)\)`)
	ublockExceptionRegex    = regexp.MustCompile(`(.*)#@#\+js\((.+)\)`)
	errUnsupportedSyntax    = errors.New("unsupported syntax")
)

func (inj *Injector) AddRule(rule string, _ bool) error {
	if match := canonicalPrimary.FindStringSubmatch(rule); match != nil {
		normalized, err := argList(match[2]).Normalize()
		if err != nil {
			return fmt.Errorf("normalize scriptlet body: %w", err)
		}
		inj.store.AddPrimaryRule(match[1], normalized)
	} else if match := canonicalExceptionRegex.FindStringSubmatch(rule); match != nil {
		normalized, err := argList(match[2]).Normalize()
		if err != nil {
			return fmt.Errorf("normalize scriptlet body: %w", err)
		}
		inj.store.AddExceptionRule(match[1], normalized)
	} else if match := ublockPrimaryRegex.FindStringSubmatch(rule); match != nil {
		normalized, err := argList(match[2]).ConvertUboToCanonical().Normalize()
		if err != nil {
			return fmt.Errorf("normalize scriptlet body: %w", err)
		}
		inj.store.AddPrimaryRule(match[1], normalized)
	} else if match := ublockExceptionRegex.FindStringSubmatch(rule); match != nil {
		normalized, err := argList(match[2]).ConvertUboToCanonical().Normalize()
		if err != nil {
			return fmt.Errorf("normalize scriptlet body: %w", err)
		}
		inj.store.AddExceptionRule(match[1], normalized)
	} else {
		return errUnsupportedSyntax
	}

	return nil
}
