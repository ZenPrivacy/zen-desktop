package rulemodifiers

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type removeparamKind int8

const (
	removeparamKindGeneric removeparamKind = iota
	removeparamKindRegexp
	removeparamKindRegexpInverse
	removeparamKindExact
	removeparamKindExactInverse
)

type RemoveParamModifier struct {
	kind   removeparamKind
	param  string
	regexp *regexp.Regexp
}

var _ ModifyingModifier = (*RemoveParamModifier)(nil)

func (rm *RemoveParamModifier) Parse(modifier string) error {
	if modifier == "removeparam" {
		rm.kind = removeparamKindGeneric
		return nil
	}

	eqIndex := strings.IndexByte(modifier, '=')
	if eqIndex == -1 || eqIndex == len(modifier)-1 {
		return errors.New("invalid syntax")
	}
	value := modifier[eqIndex+1:]

	var inverse bool
	if value[0] == '~' {
		inverse = true
		value = value[1:]
	}

	regexp, err := parseRegexp(value)
	if err != nil {
		return fmt.Errorf("parse regexp: %w", err)
	}
	if regexp != nil {
		if inverse {
			rm.kind = removeparamKindRegexpInverse
		} else {
			rm.kind = removeparamKindRegexp
		}
		rm.regexp = regexp
		return nil
	}

	if inverse {
		rm.kind = removeparamKindExactInverse
		rm.param = value
	} else {
		rm.kind = removeparamKindExact
		rm.param = value
	}
	return nil
}

func (rm *RemoveParamModifier) ModifyReq(req *http.Request) (modified bool) {
	query := req.URL.Query()

	switch rm.kind {
	case removeparamKindGeneric:
		for param := range query {
			query.Del(param)
			modified = true
		}
	case removeparamKindRegexp:
		for param, values := range query {
			filtered := values[:0]
			for _, v := range values {
				// Regexp rules match the entire query parameter, not just the name.
				if rm.regexp.MatchString(param + "=" + v) {
					modified = true
				} else {
					filtered = append(filtered, v)
				}
			}
			query[param] = filtered
		}
	case removeparamKindRegexpInverse:
		for param, values := range query {
			filtered := values[:0]
			for _, v := range values {
				// Regexp rules match the entire query parameter, not just the name.
				if !rm.regexp.MatchString(param + "=" + v) {
					modified = true
				} else {
					filtered = append(filtered, v)
				}
			}
			query[param] = filtered
		}
	case removeparamKindExact:
		for param := range query {
			if param == rm.param {
				query.Del(param)
				modified = true
			}
		}
	case removeparamKindExactInverse:
		for param := range query {
			if param != rm.param {
				query.Del(param)
				modified = true
			}
		}
	}

	if modified {
		req.URL.RawQuery = query.Encode()
	}
	return modified
}

// ModifyRes implements [ModifyingModifier].
func (rm *RemoveParamModifier) ModifyRes(*http.Response) (modified bool, err error) {
	return false, nil
}

func (rm *RemoveParamModifier) Cancels(modifier Modifier) bool {
	other, ok := modifier.(*RemoveParamModifier)
	if !ok {
		return false
	}

	if other.kind != rm.kind || other.param != rm.param {
		return false
	}

	if rm.regexp == nil && other.regexp == nil {
		return true
	}
	if rm.regexp == nil || other.regexp == nil {
		return false
	}
	return rm.regexp.String() == other.regexp.String()
}
