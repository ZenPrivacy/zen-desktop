package rule

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/irbis-sh/zen-desktop/internal/networkrules/rulemodifiers"
	"github.com/irbis-sh/zen-desktop/internal/networkrules/rulemodifiers/removejsconstant"
)

// Rule represents modifiers of a rule.
type Rule struct {
	// string representation
	RawRule string
	// FilterName is the name of the filter that the rule belongs to.
	FilterName *string

	// mods is nil for rules without modifiers, which are most of them.
	// Keeping the modifiers behind a pointer shrinks Rule from 128 to 32 bytes.
	mods *ruleModifiers
}

// TODO: The split between And and Or modifiers is somewhat convoluted and exists only to support ContentType.
// Remove it by grouping multiple ContentTypes into a single modifier and evaluating all modifiers with AND logic.

// ruleModifiers is a rule's parsed modifiers.
//
// ModifierInterner shares one block between all rules with the same modifier text,
// so a block must not be modified after parsing. A feature that edits a rule's
// modifiers must copy the block first.
type ruleModifiers struct {
	// and are modifiers that must all match for the rule to apply.
	and []rulemodifiers.ConditionModifier
	// or are modifiers where at least one must match for the rule to apply.
	or      []rulemodifiers.ConditionModifier
	actions []rulemodifiers.ActionModifier
	queries []rulemodifiers.QueryModifier

	document  bool
	important bool
	all       bool
}

// Document reports whether the rule has the document modifier.
func (rm *Rule) Document() bool {
	return rm.mods != nil && rm.mods.document
}

// Important reports whether the rule has the important modifier.
func (rm *Rule) Important() bool {
	return rm.mods != nil && rm.mods.important
}

// All reports whether the rule has the all modifier.
// On primary rules, it implies Document. It is kept separate so that "@@...$all" exceptions,
// which embed Rule, still cancel non-document rules.
func (rm *Rule) All() bool {
	return rm.mods != nil && rm.mods.all
}

// andModifiers returns the condition modifiers that must all match for the rule to apply.
func (rm *Rule) andModifiers() []rulemodifiers.ConditionModifier {
	if rm.mods == nil {
		return nil
	}
	return rm.mods.and
}

// orModifiers returns the condition modifiers of which at least one must match for the rule to apply.
func (rm *Rule) orModifiers() []rulemodifiers.ConditionModifier {
	if rm.mods == nil {
		return nil
	}
	return rm.mods.or
}

// ActionModifiers returns the rule's action modifiers.
func (rm *Rule) ActionModifiers() []rulemodifiers.ActionModifier {
	if rm.mods == nil {
		return nil
	}
	return rm.mods.actions
}

// QueryModifiers returns the rule's query modifiers.
func (rm *Rule) QueryModifiers() []rulemodifiers.QueryModifier {
	if rm.mods == nil {
		return nil
	}
	return rm.mods.queries
}

// ParseModifiers parses modifiers into the rule. Call it once, on a freshly constructed Rule.
func (rm *Rule) ParseModifiers(modifiers []string) error {
	// Parse into a local block and attach it only if some modifier is not a noop,
	// so rules like "||example.com^$_" keep mods nil.
	var mods ruleModifiers
	empty := true

	for _, m := range modifiers {
		if len(m) == 0 {
			return errors.New("empty modifier")
		}

		// Noop modifier is ignored
		if isNoopModifier(m) {
			continue
		}
		empty = false

		name, hasValue := cutModifierName(m)

		var modifier rulemodifiers.Modifier
		var isOr bool // true if the modifier belongs to mods.or.

		if !hasValue {
			// Flag modifiers.
			switch name {
			case "document", "doc":
				mods.document = true
				continue
			case "important":
				mods.important = true
				continue
			case "xmlhttprequest",
				"xhr",
				"font",
				"subdocument",
				"image",
				"object",
				"script",
				"stylesheet",
				"media",
				"websocket",
				"ping",
				"other":
				modifier = &rulemodifiers.ContentTypeModifier{}
				isOr = true
			case "third-party":
				modifier = &rulemodifiers.ThirdPartyModifier{}
			case "removeparam":
				modifier = &rulemodifiers.RemoveParamModifier{}
			case "all":
				// TODO: should also act as "popup" modifier once it gets implemented
				mods.all = true
				continue
			default:
				return fmt.Errorf("unknown modifier %q", m)
			}
		} else {
			// Parametrised modifiers.
			switch name {
			case "domain":
				modifier = &rulemodifiers.DomainModifier{}
			case "method":
				modifier = &rulemodifiers.MethodModifier{}
			case "removeparam":
				modifier = &rulemodifiers.RemoveParamModifier{}
			case "header":
				modifier = &rulemodifiers.HeaderModifier{}
			case "removeheader":
				modifier = &rulemodifiers.RemoveHeaderModifier{}
			case "remove-js-constant":
				modifier = &removejsconstant.Modifier{}
			case "scramblejs":
				modifier = &rulemodifiers.ScrambleJSModifier{}
			case "jsonprune":
				modifier = &rulemodifiers.JSONPruneModifier{}
			default:
				return fmt.Errorf("unknown modifier %q", m)
			}
		}

		if err := modifier.Parse(m); err != nil {
			return err
		}

		switch typed := modifier.(type) {
		case rulemodifiers.ConditionModifier:
			if isOr {
				mods.or = append(mods.or, typed)
			} else {
				mods.and = append(mods.and, typed)
			}
		case rulemodifiers.ActionModifier:
			mods.actions = append(mods.actions, typed)
		case rulemodifiers.QueryModifier:
			mods.queries = append(mods.queries, typed)
		default:
			log.Fatalf("got unknown modifier type %T for modifier %s", modifier, m)
		}
	}

	if !empty {
		rm.mods = &mods
	}

	return nil
}

// splitModifiers splits by unescaped commas.
// Empty fields are preserved (like strings.Split).
func splitModifiers(s string) []string {
	var res []string
	var b strings.Builder
	escaped := false

	for _, r := range s {
		switch r {
		case '\\':
			if escaped {
				b.WriteRune('\\')
			}
			escaped = !escaped
		case ',':
			if escaped {
				b.WriteRune(',')
				escaped = false
			} else {
				res = append(res, b.String())
				b.Reset()
			}
		default:
			if escaped {
				b.WriteRune('\\')
				escaped = false
			}
			b.WriteRune(r)
		}
	}

	if escaped {
		b.WriteRune('\\')
	}
	res = append(res, b.String())
	return res
}

// isNoopModifier returns true if modifier is one or more underscores.
func isNoopModifier(modifier string) bool {
	for i := 0; i < len(modifier); i++ {
		if modifier[i] != '_' {
			return false
		}
	}
	return true
}

func cutModifierName(modifier string) (name string, hasValue bool) {
	if len(modifier) > 0 && modifier[0] == '~' {
		modifier = modifier[1:]
	}
	name, _, hasValue = strings.Cut(modifier, "=")
	return name, hasValue
}

// ShouldMatchReq returns true if the rule should match the request.
func (rm *Rule) ShouldMatchReq(req *http.Request) bool {
	if req.Header.Get("Sec-Fetch-User") == "?1" && req.Header.Get("Sec-Fetch-Dest") == "document" && !rm.Document() && !rm.All() {
		return false
	}

	return rm.ModifiersMatchReq(req)
}

// ModifiersMatchReq returns true if the rule's matching modifiers match the request.
func (rm *Rule) ModifiersMatchReq(req *http.Request) bool {
	// andModifiers: All must match.
	for _, m := range rm.andModifiers() {
		if !m.ShouldMatchReq(req) {
			return false
		}
	}

	// orModifiers: At least one must match.
	if len(rm.orModifiers()) > 0 {
		for _, m := range rm.orModifiers() {
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
	return rm.ModifiersMatchRes(res)
}

// ModifiersMatchRes returns true if the rule's matching modifiers match the response.
func (rm *Rule) ModifiersMatchRes(res *http.Response) bool {
	for _, m := range rm.andModifiers() {
		if !m.ShouldMatchRes(res) {
			return false
		}
	}

	if len(rm.orModifiers()) > 0 {
		for _, m := range rm.orModifiers() {
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
	return len(rm.ActionModifiers()) == 0 && len(rm.QueryModifiers()) == 0
}

// ModifyReq modifies a request. Returns true if the request was modified.
func (rm *Rule) ModifyReq(req *http.Request) (modified bool) {
	for _, modifier := range rm.ActionModifiers() {
		if modifier.ModifyReq(req) {
			modified = true
		}
	}

	return modified
}

// ModifyReqQuery modifies a request query. Returns true if the query was modified.
func (rm *Rule) ModifyReqQuery(query url.Values) (modified bool) {
	for _, qm := range rm.QueryModifiers() {
		if qm.ModifyQuery(query) {
			modified = true
		}
	}

	return modified
}

// ModifyRes modifies a response. Returns true if the response was modified.
func (rm *Rule) ModifyRes(res *http.Response) (modified bool, err error) {
	for _, modifier := range rm.ActionModifiers() {
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
