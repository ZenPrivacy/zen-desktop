package removejsconstant

import (
	"bytes"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"github.com/ZenPrivacy/zen-desktop/internal/networkrules/rulemodifiers"
	"golang.org/x/net/html"
)

type Modifier struct {
	keys [][]string
}

var _ rulemodifiers.ModifyingModifier = (*Modifier)(nil)

var removeJSConstantRegex = regexp.MustCompile(`^remove-js-constant=(.*)$`)

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

// ModifyReq implements [rulemodifiers.ModifyingModifier].
func (*Modifier) ModifyReq(*http.Request) bool {
	return false
}

func (rc *Modifier) ModifyRes(res *http.Response) bool {
	contentType := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	switch mediaType {
	case "text/html":
		if err := removeFromInlineHTML(res, rc.keys); err != nil {
			log.Printf("remove-js-constant error: %v", err)
			return false
		}
		return true
	case "text/javascript":
		if err := removeFromJS(res, rc.keys); err != nil {
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

// removeFromInlineHTML removes the specified JS constants from inline scripts in a HTML response.
func removeFromInlineHTML(res *http.Response, keys [][]string) error {
	return httprewrite.RewriteBody(res, func(original io.ReadCloser, modified *io.PipeWriter) {
		defer original.Close()

		z := html.NewTokenizer(original)

	parse:
		for {
			switch token := z.Next(); token {
			case html.ErrorToken:
				modified.CloseWithError(z.Err())
				break parse
			case html.StartTagToken:
				modified.Write(z.Raw())
				if name, _ := z.TagName(); !bytes.Equal(name, []byte("script")) {
					continue parse
				}
				next := z.Next()
				if next != html.TextToken {
					modified.Write(z.Raw())
					continue parse
				}
				script := z.Raw()

				newScript, err := stripKeys(script, keys)

				if err != nil {
					log.Printf("error removing JS constant for %q: %v", res.Request.URL, err)
					modified.Write(script)
					continue parse
				}
				modified.Write(newScript)
			default:
				modified.Write(z.Raw())
			}
		}
	})
}

// removeFromJS removes the specified JS constant from a JS response.
func removeFromJS(res *http.Response, keys [][]string) error {
	return httprewrite.RewriteBody(res, func(original io.ReadCloser, modified *io.PipeWriter) {
		defer original.Close()

		script, err := io.ReadAll(original)
		if err != nil {
			modified.CloseWithError(err)
			return
		}
		newScript, err := stripKeys(script, keys)
		if err != nil {
			log.Printf("error removing JS constant for %q: %v", res.Request.URL, err)
			modified.Write(script)
			modified.Close()
			return
		}

		modified.Write(newScript)
		modified.Close()
	})
}
