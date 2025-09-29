package extendedcss

import "regexp"

var (
	ruleRegex = regexp.MustCompile(`^.+#@?\??#.+$`)
)

// IsRule returns true if the given string is an extended CSS rule.
func IsRule(s string) bool {
	return ruleRegex.MatchString(s)
}
