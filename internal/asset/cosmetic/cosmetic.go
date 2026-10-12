package cosmetic

import (
	"regexp"

	"github.com/irbis-sh/zen-desktop/internal/asset/extendedcss"
)

var (
	// Domain lists never contain '#', so the first '#' always starts the
	// marker. If the domain part could match '#', example.com#@##ad would split
	// at the second ## into the domain "example.com#@" and a hiding rule for
	// "ad", and the exception would never apply.
	//
	// The ^ is required: FindStringSubmatch searches the whole line, so without
	// it ([^#]*)## would match from the '@' and read the domain as "@".
	ruleRegex          = regexp.MustCompile(`^[^#]*#@?#(.+)$`)
	primaryRuleRegex   = regexp.MustCompile(`^([^#]*)##(.*)`)
	exceptionRuleRegex = regexp.MustCompile(`^([^#]*)#@#(.+)`)
)

// IsRule returns true if the given string is a cosmetic rule.
func IsRule(s string) bool {
	match := ruleRegex.FindStringSubmatch(s)
	if match == nil {
		return false
	}

	body := match[1]
	// uBlock Origin uses the same syntax for cosmetic and extended CSS rules,
	// so we return false if the rule body contains any extended pseudo-classes.
	return !extendedcss.ExtPseudoClassRegex.MatchString(body)
}
