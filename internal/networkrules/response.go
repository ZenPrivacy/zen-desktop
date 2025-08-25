package networkrules

import (
	"bytes"
	_ "embed"
	"io"
	"net/http"
	"strconv"
	"text/template"
)

// CreateBlockResponse creates a response for a blocked request.
func (nr *NetworkRules) CreateBlockResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: http.StatusForbidden,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Proto:      req.Proto,
	}
}

type BlockInfo struct {
	URL        string
	Rule       string
	FilterList string
}

//go:embed blockpage.html
var blockPageTpl string

var blockTmpl = template.Must(template.New("block").Parse(blockPageTpl))

func (nr *NetworkRules) CreateBlockResponseNew(req *http.Request, data BlockInfo) *http.Response {
	var buf bytes.Buffer
	_ = blockTmpl.Execute(&buf, data)

	h := make(http.Header)
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Length", strconv.Itoa(buf.Len()))
	h.Set("X-Blocked-By", "Zen")

	if data.Rule != "" {
		h.Set("X-Block-Rule", data.Rule)
	}

	if data.FilterList != "" {
		h.Set("X-Block-List", data.FilterList)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
		Request:    req,
	}
}

// CreateRedirectResponse creates a response for a redirected request.
func (nr *NetworkRules) CreateRedirectResponse(req *http.Request, to string) *http.Response {
	return &http.Response{
		// The use of 307 Temporary Redirect instead of 308 Permanent Redirect is intentional.
		// 308's can be cached by clients, which might cause issues in cases of erroneous redirects, changing filter rules, etc.
		StatusCode: http.StatusTemporaryRedirect,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Proto:      req.Proto,
		Header: http.Header{
			"Location": []string{to},
		},
	}
}
