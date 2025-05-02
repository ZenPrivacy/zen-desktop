package rulemodifiers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/spyzhov/ajson"
)

type JsonPruneModifier struct {
	Path string
}

var _ ModifyingModifier = (*JsonPruneModifier)(nil)

var ErrInvalidJsonPruneModifier = errors.New("invalid jsonprune modifier")

func (m *JsonPruneModifier) Parse(modifier string) error {
	if !strings.HasPrefix(modifier, "jsonprune=") {
		return ErrInvalidJsonPruneModifier
	}
	raw := strings.TrimPrefix(modifier, "jsonprune=")

	// Unescape \$ and \,
	raw = strings.ReplaceAll(raw, `\,`, `,`)
	raw = strings.ReplaceAll(raw, `\.`, `.`)
	raw = strings.ReplaceAll(raw, `\$`, `$`)
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return ErrInvalidJsonPruneModifier
	}

	m.Path = raw
	return nil
}

func (m *JsonPruneModifier) ModifyRes(res *http.Response) (modified bool) {
	if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		return false
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	root, err := ajson.Unmarshal(bodyBytes)
	if err != nil {
		res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		return false
	}

	nodes, err := root.JSONPath(m.Path)
	if err != nil {
		res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		return false
	}

	touched := false
	for _, node := range nodes {
		if err := node.Delete(); err == nil {
			touched = true
		}
	}

	if !touched {
		res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		return false
	}

	newBody, err := ajson.Marshal(root)
	if err != nil {
		res.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // fail-safe fallback
		return false
	}

	res.Body = io.NopCloser(bytes.NewReader(newBody))
	res.ContentLength = int64(len(newBody))
	res.Header.Set("Content-Length", strconv.Itoa(len(newBody)))

	return true
}

func (m *JsonPruneModifier) ModifyReq(req *http.Request) bool {
	return false
}

func (m *JsonPruneModifier) Cancels(other Modifier) bool {
	if o, ok := other.(*JsonPruneModifier); ok {
		return m.Path == o.Path
	}
	return false
}
