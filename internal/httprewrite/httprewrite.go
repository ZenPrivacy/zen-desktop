// Package httprewrite provides utilities for streaming rewrites of HTTP responses.
package httprewrite

import (
	"bytes"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

// PrependBodyContents allows to prepend the contents of the <body> tag in an HTTP text/html response.
//
// On error, the response body is unchanged and the caller may proceed as if the function had not been called.
func PrependBodyContents(res *http.Response, prependWith []byte) error {
	return RewriteBody(res, func(original io.ReadCloser, modified *io.PipeWriter) {
		defer original.Close()

		z := html.NewTokenizer(original)
	outer:
		for {
			switch token := z.Next(); token {
			case html.ErrorToken:
				modified.CloseWithError(z.Err())
				break outer
			case html.StartTagToken:
				modified.Write(z.Raw())
				if name, _ := z.TagName(); bytes.Equal(name, []byte("body")) {
					modified.Write(prependWith)
					modified.Write(z.Buffered())
					// Directly copy the remaining content, without the overhead of tokenization.
					_, err := io.Copy(modified, original)
					modified.CloseWithError(err)
					break outer
				}
			default:
				modified.Write(z.Raw())
			}
		}
	})
}

// AppendHeadContents allows to append the contents of the <head> tag in an HTTP text/html response.
//
// On error, the response body is unchanged and the caller may proceed as if the function had not been called.
func AppendHeadContents(res *http.Response, appendWith []byte) error {
	return RewriteBody(res, func(original io.ReadCloser, modified *io.PipeWriter) {
		defer original.Close()

		z := html.NewTokenizer(original)
	outer:
		for {
			switch token := z.Next(); token {
			case html.ErrorToken:
				modified.CloseWithError(z.Err())
				break outer
			case html.EndTagToken:
				if name, _ := z.TagName(); bytes.Equal(name, []byte("head")) {
					modified.Write(appendWith)
					modified.Write(z.Raw())
					modified.Write(z.Buffered())
					// Directly copy the remaining content, without the overhead of tokenization.
					_, err := io.Copy(modified, original)
					modified.CloseWithError(err)
					break outer
				}
				modified.Write(z.Raw())
			default:
				modified.Write(z.Raw())
			}
		}
	})
}
