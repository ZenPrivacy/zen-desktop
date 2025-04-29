package removejsconstant

import (
	"errors"
	"log"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/networkrules/rulemodifiers"
)

type Modifier struct {
	keys [][]string
}

var _ rulemodifiers.ModifyingModifier = (*Modifier)(nil)

var removeJSConstantRegex = regexp.MustCompile(`^remove-js-constant=(.*)$`)

// $removeconstant=key1,key2,key3
func (rc *Modifier) Parse(modifier string) error {
	match := removeJSConstantRegex.FindStringSubmatch(modifier)
	if match == nil {
		return errors.New("invalid syntax")
	}

	keys := strings.Split(match[1], "|")
	rc.keys = make([][]string, len(keys))
	for i := range keys {
		rc.keys[i] = strings.Split(keys[i], ".")
	}
	return nil
}

// ModifyReq implements ModifyingModifier.
func (*Modifier) ModifyReq(*http.Request) bool {
	return false
}

func (rc *Modifier) ModifyRes(res *http.Response) bool {
	contentType := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	if mediaType == "text/html" {
		if err := injectConstantRemover(res, rc.keys); err != nil {
			log.Printf("remove-js-constant error: %v", err)
			return false
		}
		return true
	}
	return false
}

func (rc *Modifier) Cancels(modifier rulemodifiers.Modifier) bool {
	rc2, ok := modifier.(*Modifier)
	if !ok {
		return false
	}

	if len(rc.keys) != len(rc2.keys) {
		return false
	}
	for i := range rc.keys {
		if len(rc.keys[i]) != len(rc2.keys[i]) {
			return false
		}
		for j := range rc.keys[i] {
			if rc.keys[i][j] != rc2.keys[i][j] {
				return false
			}
		}
	}
	return true
}
