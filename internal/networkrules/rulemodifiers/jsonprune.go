package rulemodifiers

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
	"github.com/spyzhov/ajson"
)

type JsonPruneModifier struct {
	// commands is a parsed sequence representing the JSONPath expression.
	commands []string
}

var _ ModifyingModifier = (*JsonPruneModifier)(nil)

var ErrInvalidJsonPruneModifier = errors.New("invalid jsonprune modifier")

func (m *JsonPruneModifier) Parse(modifier string) error {
	if !strings.HasPrefix(modifier, "jsonprune=") {
		return ErrInvalidJsonPruneModifier
	}
	raw := strings.TrimPrefix(modifier, "jsonprune=")
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return ErrInvalidJsonPruneModifier
	}

	commands, err := ajson.ParseJSONPath(raw)
	if err != nil {
		return fmt.Errorf("parse JSONPath: %w", err)
	}

	m.commands = commands
	return nil
}

func (m *JsonPruneModifier) ModifyRes(res *http.Response) (modified bool, err error) {
	if !isJSONResponse(res) {
		return false, nil
	}

	var touched bool
	err = httprewrite.BufferRewrite(res, func(src []byte) []byte {
		root, err := ajson.Unmarshal(src)
		if err != nil {
			return src
		}

		nodes, err := ajson.ApplyJSONPath(root, m.commands)
		if err != nil {
			return src
		}

		for _, node := range nodes {
			if err := node.Delete(); err == nil {
				touched = true
			}
		}

		if !touched {
			return src
		}

		newBody, err := ajson.Marshal(root)
		if err != nil {
			return src
		}
		return newBody
	})

	if err != nil {
		return false, fmt.Errorf("buffer rewrite: %w", err)
	}

	return touched, nil
}

func (m *JsonPruneModifier) ModifyReq(req *http.Request) bool {
	return false
}

func (m *JsonPruneModifier) Cancels(other Modifier) bool {
	o, ok := other.(*JsonPruneModifier)
	if !ok {
		return false
	}

	if len(m.commands) != len(o.commands) {
		return false
	}

	for i := range m.commands {
		if m.commands[i] != o.commands[i] {
			return false
		}
	}

	return true
}

func isJSONResponse(res *http.Response) bool {
	contentType := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	if mediaType != "application/json" {
		return false
	}

	return true
}
