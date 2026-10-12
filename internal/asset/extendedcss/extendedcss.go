package extendedcss

import "regexp"

var (
	// ExtPseudoClassRegex matches extended pseudo-classes.
	ExtPseudoClassRegex = regexp.MustCompile(`:(?:has-text|contains|matches-attr|matches-css(?:-before|-after)?|matches-media|matches-path|matches-prop(?:erty)?|min-text-length|others|upward|xpath|nth-ancestor|watch-attr|remove|style|-abp-(?:contains|has))`)

	// Domain lists never contain '#', so the first '#' always starts the
	// marker. If the domain part could match '#', example.com#@##ad:has-text(x)
	// would split at the second ## into the domain "example.com#@" and a rule
	// hiding "ad:has-text(x)", and the exception would never apply.
	//
	// The ^ is required: FindStringSubmatch searches the whole line, so without
	// it ([^#]+)#\??# would match from the '@' and read the domain as "@".
	//
	// Unlike the cosmetic regexes, the domain part is still required, so
	// generic rules such as ##div:has-text(x) are not handled here yet.
	ruleRegex          = regexp.MustCompile(`^[^#]+#@?\??#.+$`)
	primaryRuleRegex   = regexp.MustCompile(`^([^#]+)#\??#(.+)`)
	exceptionRuleRegex = regexp.MustCompile(`^([^#]+)#@\??#(.+)`)
)

// IsRule returns true if the given string is an extended CSS rule.
func IsRule(s string) bool {
	return ruleRegex.MatchString(s)
}
