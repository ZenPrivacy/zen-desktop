package ruletree2

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s      string
		tokens []token
	}{
		{
			s:      "abc123",
			tokens: []token{'a', 'b', 'c', '1', '2', '3'},
		},
		{
			s:      "*",
			tokens: []token{tokenWildcard},
		},
		{
			s:      "a*b",
			tokens: []token{'a', tokenWildcard, 'b'},
		},
		{
			s:      "*a*",
			tokens: []token{tokenWildcard, 'a', tokenWildcard},
		},
		{
			s:      "||",
			tokens: []token{tokenDomainRoot},
		},
		{
			s:      "|||||",
			tokens: []token{tokenDomainRoot, tokenDomainRoot, tokenRootEnd},
		},
		{
			s:      "||example.com",
			tokens: []token{tokenDomainRoot, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
		},
		{
			s:      "|",
			tokens: []token{tokenRootEnd},
		},
		{
			s:      "example|",
			tokens: []token{'e', 'x', 'a', 'm', 'p', 'l', 'e', tokenRootEnd},
		},
		{
			s:      "^",
			tokens: []token{tokenSeparator},
		},
		{
			s:      "a^b",
			tokens: []token{'a', tokenSeparator, 'b'},
		},
		{
			s:      "*||^|",
			tokens: []token{tokenWildcard, tokenDomainRoot, tokenSeparator, tokenRootEnd},
		},
		{
			s:      "a*b||c^d|e",
			tokens: []token{'a', tokenWildcard, 'b', tokenDomainRoot, 'c', tokenSeparator, 'd', tokenRootEnd, 'e'},
		},
	}

	for _, test := range tests {
		if got := tokenize(test.s); !reflect.DeepEqual(got, test.tokens) {
			t.Errorf("Tokenize(%q) = %#v, want %#v", test.s, got, test.tokens)
		}
	}
}
