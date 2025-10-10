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
			"abc123",
			[]token{'a', 'b', 'c', '1', '2', '3'},
		},
		{
			"*",
			[]token{tokenWildcard},
		},
		{
			"a*b",
			[]token{'a', tokenWildcard, 'b'},
		},
		{
			"*a*",
			[]token{tokenWildcard, 'a', tokenWildcard},
		},
		{
			"||",
			[]token{tokenDomainRoot},
		},
		{
			"|||||",
			[]token{tokenDomainRoot, tokenDomainRoot, tokenStartEnd},
		},
		{
			"||example.com",
			[]token{tokenDomainRoot, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
		},
		{
			"|",
			[]token{tokenStartEnd},
		},
		{
			"example|",
			[]token{'e', 'x', 'a', 'm', 'p', 'l', 'e', tokenStartEnd},
		},
		{
			"^",
			[]token{tokenSeparator},
		},
		{
			"a^b",
			[]token{'a', tokenSeparator, 'b'},
		},
		{
			"*||^|",
			[]token{tokenWildcard, tokenDomainRoot, tokenSeparator, tokenStartEnd},
		},
		{
			"a*b||c^d|e",
			[]token{'a', tokenWildcard, 'b', tokenDomainRoot, 'c', tokenSeparator, 'd', tokenStartEnd, 'e'},
		},
		{
			"||example.com/ads/*",
			[]token{tokenDomainRoot, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm', '/', 'a', 'd', 's', '/', tokenWildcard},
		},
	}

	for _, test := range tests {
		if got := tokenize(test.s); !reflect.DeepEqual(got, test.tokens) {
			t.Errorf("Tokenize(%q) = %#v, want %#v", test.s, got, test.tokens)
		}
	}
}
