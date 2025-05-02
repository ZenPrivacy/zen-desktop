package rulemodifiers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ZenPrivacy/zen-desktop/internal/httprewrite"
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

func (m *JsonPruneModifier) ModifyRes(res *http.Response) (modified bool, err error) {
	if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		return false, nil
	}

	var touched bool

	err = httprewrite.BufferRewrite(res, func(src []byte) []byte {
		root, err := ajson.Unmarshal(src)
		if err != nil {
			return src
		}

		nodes, err := root.JSONPath(m.Path)
		if err != nil || len(nodes) == 0 {
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
		return false, err
	}

	return touched, nil
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
