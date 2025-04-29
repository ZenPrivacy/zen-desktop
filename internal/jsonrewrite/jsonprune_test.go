package jsonrewrite_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ZenPrivacy/zen-desktop/internal/jsonrewrite"
)

func makeResponse(body string) *http.Response {
	return &http.Response{
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(data)
}

func TestModifyJSONResponse(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		remove         string
		required       string
		expectedOutput string
	}{
		{
			name:           "removes root object prop",
			input:          `{"a": 123, "b": 321}`,
			remove:         "a",
			expectedOutput: `{"b":321}`,
		},
		{
			name:           "removes root object prop only if required prop is matched",
			input:          `{"a": 123, "b": 321}`,
			remove:         "a",
			required:       "b",
			expectedOutput: `{"b":321}`,
		},
		{
			name:           "does not remove if required prop not matched",
			input:          `{"a": 123, "c": 321}`,
			remove:         "a",
			required:       "b",
			expectedOutput: `{"a":123,"c":321}`,
		},
		{
			name:           "removes deeply nested properties",
			input:          `{"a":{"b":{"c":{"d":{"e":{"f":{"g":123,"h":321}}}}}}}`,
			remove:         "a.b.c.d.e.f.g a.b.c.d.e.f.h",
			expectedOutput: `{"a":{"b":{"c":{"d":{"e":{"f":{}}}}}}}`,
		},
		{
			name:           "removes array properties with []",
			input:          `{"a":[{"b":"keep","z":"remove"},{"c":"keep","z":"remove"}]}`,
			remove:         "a.[].z",
			expectedOutput: `{"a":[{"b":"keep"},{"c":"keep"}]}`,
		},
		{
			name:           "null input remains null",
			input:          `null`,
			remove:         "a",
			expectedOutput: `null`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := makeResponse(tc.input)

			err := jsonrewrite.ModifyJSONResponse(resp, tc.remove, tc.required)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			body := readResponseBody(t, resp)
			// normalize JSON strings
			var normalizedExpected, normalizedActual any
			if err := json.Unmarshal([]byte(tc.expectedOutput), &normalizedExpected); err != nil {
				t.Fatal("invalid expected JSON")
			}
			if err := json.Unmarshal([]byte(body), &normalizedActual); err != nil {
				t.Fatal("invalid actual JSON")
			}
			if !equals(normalizedExpected, normalizedActual) {
				t.Errorf("got %s, want %s", body, tc.expectedOutput)
			}
		})
	}
}

func equals(a, b any) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return bytes.Equal(aJSON, bJSON)
}
