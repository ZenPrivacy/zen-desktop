package ruletree2

import (
	"errors"
	"strings"
	"sync"
)

type Data any

type Tree[T Data] struct {
	insertMu sync.Mutex

	root       *node[T]
	domainRoot *node[T]
	startRoot  *node[T]
}

func New[T Data]() *Tree[T] {
	return &Tree[T]{
		root:       &node[T]{},
		domainRoot: &node[T]{},
		startRoot:  &node[T]{},
	}
}

func (t *Tree[T]) Insert(pattern string, v T) error {
	if pattern == "" {
		return errors.New("empty pattern")
	}

	var parent *node[T]
	var n *node[T]

	tokens := tokenize(pattern)

	t.insertMu.Lock()
	defer t.insertMu.Unlock()

	switch tokens[0] {
	case tokenDomainBoundary:
		n, tokens = t.domainRoot, tokens[1:]
	case tokenAnchor:
		n, tokens = t.startRoot, tokens[1:]
	default:
		n = t.root
	}

	for {
		if len(tokens) == 0 {
			if n.isLeaf() {
				n.leaf.val = append(n.leaf.val, v)
			} else {
				n.leaf = &leaf[T]{
					val: []T{v},
				}
			}
			return nil
		}

		parent = n
		n = n.getEdge(tokens[0])

		if n == nil {
			n := &node[T]{
				prefix: tokens,
				leaf: &leaf[T]{
					val: []T{v},
				},
			}
			parent.addEdge(edge[T]{
				label: tokens[0],
				node:  n,
			})
			return nil
		}

		commonPrefix := longestPrefix(tokens, n.prefix)
		if commonPrefix == len(n.prefix) {
			tokens = tokens[commonPrefix:]
			continue
		}

		child := &node[T]{
			prefix: tokens[:commonPrefix],
		}
		parent.updateEdge(tokens[0], child)

		child.addEdge(edge[T]{
			label: n.prefix[commonPrefix],
			node:  n,
		})
		n.prefix = n.prefix[commonPrefix:]

		l := &leaf[T]{
			val: []T{v},
		}
		if commonPrefix == len(tokens) {
			child.leaf = l
		} else {
			n := &node[T]{
				leaf:   l,
				prefix: tokens[commonPrefix:],
			}
			child.addEdge(edge[T]{
				label: tokens[commonPrefix],
				node:  n,
			})
		}
		return nil
	}
}

func (t *Tree[T]) Get(url string) []T {
	var data []T
	data = append(data, t.root.traverse(url)...)
	data = append(data, t.startRoot.traverse(url)...)

	inScheme := true
	inHost := false
	for i, c := range url {
		data = append(data, t.root.traverse(url[i:])...)

		if inHost {
			if c == '.' {
				data = append(data, t.domainRoot.traverse(url[i+1:])...)
			}

			if c == '/' {
				inHost = false
			}
		}

		if inScheme {
			if strings.HasSuffix(url[:i+1], "://") {
				data = append(data, t.domainRoot.traverse(url[i+1:])...)
				inScheme = false
				inHost = true
			}
		}
	}

	return data
}

func longestPrefix(a, b []token) int {
	max := len(a)
	if l := len(b); l < max {
		max = l
	}
	for i := range max {
		if a[i] != b[i] {
			return i
		}
	}
	return max
}
