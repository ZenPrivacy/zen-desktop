package rulemodifiers

import (
	"net/http"
	"testing"
)

func TestHeaderModifier(t *testing.T) {
	t.Parallel()

	t.Run("fails on empty modifier", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse(""); err == nil {
			t.Error("headerModifier.Parse(\"\") = nil, want error")
		}
	})

	t.Run("fails on missing specifier", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header="); err == nil {
			t.Error("headerModifier.Parse(\"header=\") = nil, want error")
		}
	})

	t.Run("fails on invalid specifier", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=one:two:three"); err == nil {
			t.Error("headerModifier.Parse(\"header=one:two:three\") = nil, want error")
		}
	})

	t.Run("matches response by header name", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=content-type"); err != nil {
			t.Fatalf("headerModifier.Parse(\"header=content-type\") = %v, want nil", err)
		}

		res := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}

		res.Header.Del("Content-Type")
		if m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = true, want false")
		}
	})

	t.Run("matches response by canonicalized header name", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=etag"); err != nil {
			t.Fatalf("headerModifier.Parse(\"header=etag\") = %v, want nil", err)
		}

		res := &http.Response{
			Header: http.Header{
				"Etag": []string{`"abc"`},
			},
		}
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}
	})

	t.Run("matches response by weirdly capitalized header name", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=sET-CoOKIe"); err != nil {
			t.Fatalf("headerModifier.Parse(\"header=sET-CoOKIe\") = %v, want nil", err)
		}

		res := &http.Response{
			Header: http.Header{
				"Set-Cookie": []string{"session=abc"},
			},
		}
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}
	})

	t.Run("matches any response header value", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=sET-CoOKIe:/session=/"); err != nil {
			t.Fatalf("headerModifier.Parse(\"header=sET-CoOKIe:/session=/\") = %v, want nil", err)
		}

		res := &http.Response{
			Header: http.Header{
				"Set-Cookie": []string{"theme=dark", "session=abc"},
			},
		}
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}
	})

	t.Run("matches response by header name and exact value", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=Content-Type:application/json"); err != nil {
			t.Fatalf("headerModifier.Parse(\"header=Content-Type:application/json\") = %v, want nil", err)
		}

		res := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}

		res.Header.Set("Content-Type", "application/xml")
		if m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = true, want false")
		}

		res.Header.Del("Content-Type")
		if m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = true, want false")
		}
	})

	t.Run("matches response by header name and regexp value", func(t *testing.T) {
		t.Parallel()
		m := HeaderModifier{}
		if err := m.Parse("header=Content-Type:/application/i"); err != nil {
			t.Fatalf("headerModifier.Parse(\"header=Content-Type:/application/i\") = %v, want nil", err)
		}

		res := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}

		res.Header.Set("Content-Type", "application/xml")
		if !m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = false, want true")
		}

		res.Header.Set("Content-Type", "text/plain")
		if m.ShouldMatchRes(res) {
			t.Error("headerModifier.ShouldMatchRes(res) = true, want false")
		}
	})
}
