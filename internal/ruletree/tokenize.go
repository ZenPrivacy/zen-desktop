package ruletree

import "strings"

// token is a single byte of a tokenized pattern: either a literal character
// or one of the special tokens.
type token = byte

// Special tokens are encoded as ASCII control bytes, which URLs cannot contain
// and insertTree rejects in patterns, so they share a node's edge list with
// literal characters.
const (
	// tokenWildcard represents "*" and matches any set of characters.
	tokenWildcard token = 0x01
	// tokenDomainBoundary represents "||" and matches domain and subdomain roots.
	tokenDomainBoundary token = 0x02
	// tokenSeparator represents "^" and matches any character except a letter, digit, or _-.%.
	// It also matches the end of an address.
	tokenSeparator token = 0x03
	// tokenAnchor represents "|" and matches the beginning or the end of an address.
	tokenAnchor token = 0x04

	tokenMaxSpecial = tokenAnchor
)

func tokenize(s string) string {
	if !strings.ContainsAny(s, "*|^") {
		// Node prefixes are substrings of the returned string, so returning s
		// lets them share the caller's memory. The pattern is part of the
		// rule's raw text, which outlives the tree anyway.
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '*':
			if i == 0 || s[i-1] != '*' {
				b.WriteByte(tokenWildcard)
			}
		case '|':
			if i+1 < len(s) && s[i+1] == '|' {
				b.WriteByte(tokenDomainBoundary)
				i++
			} else {
				b.WriteByte(tokenAnchor)
			}
		case '^':
			b.WriteByte(tokenSeparator)
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
