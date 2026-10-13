package proxy

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/irbis-sh/zen-desktop/internal/fetchmeta"
	"github.com/irbis-sh/zen-desktop/internal/i18n"
	"github.com/irbis-sh/zen-desktop/internal/pagestyle"
)

//go:embed errorpage.html
var errorPageHTML string

var errorPageTmpl = template.Must(template.New("errorpage").Parse(errorPageHTML))

type errorPageData struct {
	SharedStyle template.CSS
	L           *i18n.Catalogue
	Title       string
	Message     string
	// Detail is the raw error string, shown in a collapsed section for debugging.
	// It may contain the request URL, which is fine: the page is only served to
	// the client that requested that URL. It isn't translated: it's for bug
	// reports, and OS and network error text can't be translated usefully.
	Detail string
}

// errorKind selects an error page's title and message, from the
// errorPage.<kind>.title and errorPage.<kind>.message catalogue keys.
type errorKind string

const (
	errorKindTimeout     errorKind = "timeout"
	errorKindNotFound    errorKind = "notFound"
	errorKindRefused     errorKind = "refused"
	errorKindReset       errorKind = "reset"
	errorKindTLS         errorKind = "tls"
	errorKindUnreachable errorKind = "unreachable"
	errorKindInternal    errorKind = "internal"
)

// writeUpstreamError responds to a failed upstream request: a friendly HTML error
// page for browser navigations, a plain-text 502 otherwise.
func writeUpstreamError(w http.ResponseWriter, req *http.Request, err error) {
	if isTLSError(err) {
		// connectHandler adds hosts that fail TLS to the transparent list, which
		// only takes effect on a fresh CONNECT. Close the client connection so a
		// retry cannot reuse the intercepted session and reconnects unfiltered.
		// Go's HTTP/2 server translates this header into a GOAWAY frame.
		w.Header().Set("Connection", "close")
	}

	if !fetchmeta.IsLikelyUserNavigation(req) {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeErrorPage(w, classifyUpstreamError(err), err.Error())
}

// writeFilterError responds to an internal filtering failure: a friendly HTML error
// page for browser navigations, a plain-text 502 otherwise.
func writeFilterError(w http.ResponseWriter, req *http.Request, err error) {
	if !fetchmeta.IsLikelyUserNavigation(req) {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeErrorPage(w, errorKindInternal, err.Error())
}

// classifyUpstreamError maps an upstream request error to the kind of error
// page to show.
func classifyUpstreamError(err error) errorKind {
	var dnsErr *net.DNSError
	var netErr net.Error
	switch {
	// Timeouts are checked before DNS errors so that a slow resolver reports
	// "took too long" rather than "server not found": url.Error and net.DNSError
	// both surface timeouts through the net.Error interface.
	case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()):
		return errorKindTimeout
	case errors.As(err, &dnsErr):
		return errorKindNotFound
	case errorsIsAny(err, connRefusedErrs):
		return errorKindRefused
	case errorsIsAny(err, connResetErrs):
		return errorKindReset
	case isTLSError(err):
		// writeUpstreamError closes the client connection on TLS errors, and
		// connectHandler puts the host on the transparent list, so a retry over a
		// fresh CONNECT reconnects without interception. errorPage.tls.message
		// says "may" rather than "will": the plain-HTTP path serves it without a
		// transparent-list entry.
		return errorKindTLS
	default:
		return errorKindUnreachable
	}
}

// errorsIsAny reports whether errors.Is matches err against any of targets.
func errorsIsAny(err error, targets []error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func writeErrorPage(w http.ResponseWriter, kind errorKind, detail string) {
	l := i18n.Current()
	var buf bytes.Buffer
	err := errorPageTmpl.Execute(&buf, errorPageData{
		SharedStyle: pagestyle.Shared,
		L:           l,
		Title:       l.T("errorPage." + string(kind) + ".title"),
		Message:     l.T("errorPage." + string(kind) + ".message"),
		Detail:      detail,
	})
	if err != nil {
		log.Printf("error executing error page template: %v", err)
		http.Error(w, detail, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusBadGateway)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("error writing error page: %v", err)
	}
}
