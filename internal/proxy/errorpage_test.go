package proxy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"syscall"
	"testing"
)

func TestClassifyUpstreamError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want errorKind
	}{
		{
			name: "DNS failure",
			err:  wrapped(&net.DNSError{Err: "no such host", Name: "example.com", IsNotFound: true}),
			want: errorKindNotFound,
		},
		{
			name: "timeout",
			err:  wrapped(context.DeadlineExceeded),
			want: errorKindTimeout,
		},
		{
			name: "DNS timeout",
			err:  wrapped(&net.DNSError{Err: "i/o timeout", Name: "example.com", IsTimeout: true}),
			want: errorKindTimeout,
		},
		{
			name: "connection refused",
			err:  wrapped(syscall.ECONNREFUSED),
			want: errorKindRefused,
		},
		{
			name: "connection reset",
			err:  wrapped(syscall.ECONNRESET),
			want: errorKindReset,
		},
		{
			name: "TLS failure",
			err:  wrapped(errors.New("tls: failed to verify certificate")),
			want: errorKindTLS,
		},
		{
			name: "unrecognized error",
			err:  errors.New("received body length does not match content length"),
			want: errorKindUnreachable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := classifyUpstreamError(tt.err)
			if got != tt.want {
				t.Errorf("classifyUpstreamError() = %q, want %q", got, tt.want)
			}

			rec := httptest.NewRecorder()
			writeErrorPage(rec, got, "detail")
			checkNoRawKeys(t, rec.Body.String())
		})
	}
}

func TestWriteUpstreamError(t *testing.T) {
	t.Parallel()

	t.Run("serves HTML error page to browser navigations", func(t *testing.T) {
		t.Parallel()

		req := newNavigationRequest(t)
		rec := httptest.NewRecorder()
		writeUpstreamError(rec, req, wrapped(&net.DNSError{Err: "no such host", Name: "example.com"}))

		if rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Errorf("Content-Type = %q, want text/html", ct)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("Cache-Control = %q, want no-store", cc)
		}

		if got := rec.Header().Get("Connection"); got != "" {
			t.Errorf("Connection = %q, want empty for non-TLS errors", got)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Retry") {
			t.Error("body does not contain a retry button")
		}
		if !strings.Contains(body, "no such host") {
			t.Error("body does not contain the error detail")
		}
		if !strings.Contains(body, `lang="en-US"`) {
			t.Error(`body does not contain lang="en-US"`)
		}
		// Browsers replace error page bodies shorter than 512 bytes with their own.
		if len(body) <= 512 {
			t.Errorf("body is %d bytes, must exceed 512 for browsers to render it", len(body))
		}
	})

	t.Run("serves plain text to non-navigation requests", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")

		rec := httptest.NewRecorder()
		upstreamErr := wrapped(syscall.ECONNREFUSED)
		writeUpstreamError(rec, req, upstreamErr)

		if rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
			t.Errorf("Content-Type = %q, want text/plain", ct)
		}
		if body := rec.Body.String(); body != upstreamErr.Error()+"\n" {
			t.Errorf("body = %q, want the raw error", body)
		}
	})

	t.Run("closes the connection on TLS errors", func(t *testing.T) {
		t.Parallel()

		req := newNavigationRequest(t)
		rec := httptest.NewRecorder()
		writeUpstreamError(rec, req, wrapped(errors.New("tls: failed to verify certificate")))

		if got := rec.Header().Get("Connection"); got != "close" {
			t.Errorf("Connection = %q, want %q", got, "close")
		}
	})

	t.Run("escapes HTML in error detail", func(t *testing.T) {
		t.Parallel()

		req := newNavigationRequest(t)
		rec := httptest.NewRecorder()
		writeUpstreamError(rec, req, errors.New(`<script>alert("xss")</script>`))

		body := rec.Body.String()
		if strings.Contains(body, `<script>alert`) {
			t.Error("body contains unescaped error detail")
		}
		if !strings.Contains(body, "&lt;script&gt;") {
			t.Error("body does not contain the escaped error detail")
		}
	})
}

func TestWriteFilterError(t *testing.T) {
	t.Parallel()

	req := newNavigationRequest(t)
	rec := httptest.NewRecorder()
	writeFilterError(rec, req, errors.New("rewrite html: unexpected EOF"))

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "internal error") {
		t.Error("body does not mention an internal error")
	}
	checkNoRawKeys(t, body)
}

// wrapped wraps err the way http.Client returns errors from a round trip.
func wrapped(err error) error {
	return &url.Error{Op: "Get", URL: "https://example.com/", Err: err}
}

func newNavigationRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	return req
}

// checkNoRawKeys fails if the page shows a catalogue key instead of its text.
// Checking en-US is enough: T falls back to en-US, and the i18n tests check that
// every catalogue has the same keys as en-US.
func checkNoRawKeys(t *testing.T, body string) {
	t.Helper()

	for _, prefix := range []string{"common.", "errorPage."} {
		if strings.Contains(body, prefix) {
			t.Errorf("body contains an untranslated %s key", prefix)
		}
	}
}
