package rule

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/networkrules/rulemodifiers"
	"github.com/ZenPrivacy/zen-desktop/internal/networkrules/rulemodifiers/removejsconstant"
)

// Rule represents modifiers of a rule.
type Rule struct {
	// string representation
	RawRule string
	// FilterName is the name of the filter that the rule belongs to.
	FilterName *string

	MatchingModifiers  matchingModifiers
	ModifyingModifiers []rulemodifiers.ModifyingModifier

	// Document shows if rule has Document modifier.
	Document bool
}

type matchingModifiers struct {
	// AndModifiers should be matched together.
	AndModifiers []rulemodifiers.MatchingModifier
	// OrModifiers should be matched if one of them is matched.
	OrModifiers []rulemodifiers.MatchingModifier
}

func (rm *Rule) ParseModifiers(modifiers string) error {
	if len(modifiers) == 0 {
		return nil
	}

	for _, m := range splitModifiersEscaped(modifiers) {
		if len(m) == 0 {
			return fmt.Errorf("empty modifier")
		}

		isKind := func(kind string) bool {
			if len(m) > 0 && m[0] == '~' {
				return strings.HasPrefix(m[1:], kind)
			}
			return strings.HasPrefix(m, kind)
		}

		if isKind("document") || isKind("doc") {
			rm.Document = true
			continue
		}

		var modifier rulemodifiers.Modifier
		var isOr bool // true if modifier belongs to OrModifiers; false if it belongs to AndModifiers
		switch {
		case isKind("domain"):
			modifier = &rulemodifiers.DomainModifier{}
		case isKind("method"):
			modifier = &rulemodifiers.MethodModifier{}
		case isKind("xmlhttprequest"),
			isKind("xhr"),
			isKind("font"),
			isKind("subdocument"),
			isKind("image"),
			isKind("object"),
			isKind("script"),
			isKind("stylesheet"),
			isKind("media"),
			isKind("other"):
			modifier = &rulemodifiers.ContentTypeModifier{}
			isOr = true
		case isKind("third-party"):
			modifier = &rulemodifiers.ThirdPartyModifier{}
		case isKind("removeparam"):
			modifier = &rulemodifiers.RemoveParamModifier{}
		case isKind("header"):
			modifier = &rulemodifiers.HeaderModifier{}
		case isKind("removeheader"):
			modifier = &rulemodifiers.RemoveHeaderModifier{}
		case isKind("remove-js-constant"):
			modifier = &removejsconstant.Modifier{}
		case isKind("jsonprune"):
			modifier = &rulemodifiers.JSONPruneModifier{}
		case isKind("all"):
			// TODO: should act as "popup" modifier once it gets implemented
			continue
		default:
			return fmt.Errorf("unknown modifier %q", m)
		}

		if err := modifier.Parse(m); err != nil {
			return err
		}

		if matchingModifier, ok := modifier.(rulemodifiers.MatchingModifier); ok {
			if isOr {
				rm.MatchingModifiers.OrModifiers = append(rm.MatchingModifiers.OrModifiers, matchingModifier)
			} else {
				rm.MatchingModifiers.AndModifiers = append(rm.MatchingModifiers.AndModifiers, matchingModifier)
			}
		} else if modifyingModifier, ok := modifier.(rulemodifiers.ModifyingModifier); ok {
			rm.ModifyingModifiers = append(rm.ModifyingModifiers, modifyingModifier)
		} else {
			log.Fatalf("got unknown modifier type %T for modifier %s", modifier, m)
		}
	}

	return nil
}

// ShouldMatchReq returns true if the rule should match the request.
func (rm *Rule) ShouldMatchReq(req *http.Request) bool {
	if req.Header.Get("Sec-Fetch-User") == "?1" && req.Header.Get("Sec-Fetch-Dest") == "document" && !rm.Document {
		return false
	}

	// AndModifiers: All must match.
	for _, m := range rm.MatchingModifiers.AndModifiers {
		if !m.ShouldMatchReq(req) {
			return false
		}
	}

	// OrModifiers: At least one must match.
	if len(rm.MatchingModifiers.OrModifiers) > 0 {
		for _, m := range rm.MatchingModifiers.OrModifiers {
			if m.ShouldMatchReq(req) {
				return true
			}
		}
		return false
	}

	return true
}

// ShouldMatchRes returns true if the rule should match the response.
func (rm *Rule) ShouldMatchRes(res *http.Response) bool {
	// maybe add sec-fetch logic too
	for _, m := range rm.MatchingModifiers.AndModifiers {
		if !m.ShouldMatchRes(res) {
			return false
		}
	}

	if len(rm.MatchingModifiers.OrModifiers) > 0 {
		for _, m := range rm.MatchingModifiers.OrModifiers {
			if m.ShouldMatchRes(res) {
				return true
			}
		}
		return false
	}

	return true
}

// ShouldBlockReq returns true if the request should be blocked.
func (rm *Rule) ShouldBlockReq(*http.Request) bool {
	return len(rm.ModifyingModifiers) == 0
}

// ModifyReq modifies a request. Returns true if the request was modified.
func (rm *Rule) ModifyReq(req *http.Request) (modified bool) {
	for _, modifier := range rm.ModifyingModifiers {
		if modifier.ModifyReq(req) {
			modified = true
		}
	}

	return modified
}

// ModifyRes modifies a response. Returns true if the response was modified.
func (rm *Rule) ModifyRes(res *http.Response) (modified bool, err error) {
	for _, modifier := range rm.ModifyingModifiers {
		m, err := modifier.ModifyRes(res)
		if err != nil {
			return false, fmt.Errorf("modify response: %w", err)
		}
		if m {
			modified = true
		}
	}

	return modified, nil
}

func splitModifiersEscaped(modifiers string) []string {
	var res []string
	var current string
	var escaped bool
	for _, c := range modifiers {
		switch c {
		case '\\':
			escaped = !escaped
			continue
		case ',':
			if escaped {
				current += string(c)
				escaped = false
				continue
			}
			res = append(res, current)
			current = ""
		default:
			current += string(c)
			escaped = false
		}
	}
	if len(current) > 0 {
		res = append(res, current)
	}
	return res
}
