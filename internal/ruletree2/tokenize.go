package ruletree2

type token uint16

const (
	// tokenWildcard represents "*" and matches any set of characters.
	tokenWildcard token = (2 << 7)
	// tokenDomainRoot represents "||" and matches domain and subdomain roots.
	tokenDomainRoot token = (2 << 7) + iota
	// tokenSeparator represents "^" and matches any character except a letter, digit, or _-.%.
	tokenSeparator token = (2 << 7) + iota
	// tokenStartEnd represents "|" and matches the beginning or the end of an address.
	tokenStartEnd token = (2 << 7) + iota
)

func tokenize(s string) []token {
	var tokens []token
	for i := 0; i < len(s); {
		switch s[i] {
		case '*':
			tokens = append(tokens, tokenWildcard)
		case '|':
			switch {
			case i+1 < len(s) && s[i+1] == '|':
				tokens = append(tokens, tokenDomainRoot)
				i++
			default:
				tokens = append(tokens, tokenStartEnd)
			}
		case '^':
			tokens = append(tokens, tokenSeparator)
		default:
			tokens = append(tokens, token(s[i]))
		}
		i++
	}
	return tokens
}
