package ruletree2

type token uint16

func (t token) matches(ch byte) bool {
	switch t {
	case tokenWildcard:
		return true
	case tokenSeparator:
		return isSeparator(ch)
	default:
		return t == token(ch)
	}
}

const (
	tokenWildcard   token = (2 << 7)
	tokenDomainRoot token = (2 << 7) + iota
	tokenSeparator  token = (2 << 7) + iota
	tokenRootEnd    token = (2 << 7) + iota
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
				tokens = append(tokens, tokenRootEnd)
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
