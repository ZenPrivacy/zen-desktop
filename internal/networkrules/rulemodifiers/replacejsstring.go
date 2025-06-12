package rulemodifiers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"golang.org/x/net/html"
)

type ReplaceJSString struct {
	keys []string
}

// var _ rulemodifiers.ModifyingModifier = (*Modifier)(nil)

var replaceJSConstantRegex = regexp.MustCompile(`^replace-js-string=(.*)$`)

func (rc *ReplaceJSString) Parse(modifier string) error {
	match := replaceJSConstantRegex.FindStringSubmatch(modifier)
	if match == nil {
		return errors.New("invalid syntax")
	}

	rc.keys = strings.Split(match[1], "|")
	return nil
}

func (*ReplaceJSString) ModifyReq(*http.Request) bool {
	return false
}

func (rc *ReplaceJSString) ModifyRes(res *http.Response) (bool, error) {
	contentType := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false, nil
	}
	switch mediaType {
	case "text/html":
		if err := replaceInInlineHTML(res, rc.keys); err != nil {
			return false, fmt.Errorf("replace in inline HTML: %v", err)
		}
		return true, nil
	case "text/javascript":
		if err := replaceInJS(res, rc.keys); err != nil {
			return false, fmt.Errorf("replace in JS: %v", err)
		}
		return true, nil
	}
	return false, nil
}

func (rc *ReplaceJSString) Cancels(modifier Modifier) bool {
	return false
}

// replaceInInlineHTML replaces matched keys with unique random values in HTML responses.
func replaceInInlineHTML(res *http.Response, keys []string) error {
	return httprewrite.StreamRewrite(res, func(original io.ReadCloser, modified *io.PipeWriter) {
		defer original.Close()
		z := html.NewTokenizer(original)

	parseLoop:
		for {
			switch token := z.Next(); token {
			case html.ErrorToken:
				modified.CloseWithError(z.Err())
				break parseLoop
			case html.StartTagToken:
				modified.Write(z.Raw())
				if name, _ := z.TagName(); !bytes.Equal(name, []byte("script")) {
					continue parseLoop
				}
				next := z.Next()
				if next != html.TextToken {
					modified.Write(z.Raw())
					continue parseLoop
				}
				script := z.Raw()
				newScript, err := replaceKeysUniquely(script, keys)
				if err != nil {
					log.Printf("error randomizing JS constant for %q: %v", res.Request.URL, err)
					modified.Write(script)
					continue parseLoop
				}
				modified.Write(newScript)
			default:
				modified.Write(z.Raw())
			}
		}
	})
}

// replaceInJS replaces matched keys with unique random values in JS responses.
func replaceInJS(res *http.Response, keys []string) error {
	return httprewrite.BufferRewrite(res, func(src []byte) []byte {
		newScript, err := replaceKeysUniquely(src, keys)
		if err != nil {
			log.Printf("error randomizing JS constant for %q: %v", res.Request.URL, err)
			return src
		}
		return newScript
	})
}

// replaceKeysUniquely replaces each occurrence of keys with unique random strings.
func replaceKeysUniquely(script []byte, keys []string) ([]byte, error) {
	modifiedScript := string(script)
	for _, key := range keys {
		re := regexp.MustCompile(regexp.QuoteMeta(key))
		modifiedScript = re.ReplaceAllStringFunc(modifiedScript, func(_ string) string {
			return generateRandomString(10)
		})
	}
	return []byte(modifiedScript), nil
}

// generateRandomString returns a random alphanumeric string of given length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
