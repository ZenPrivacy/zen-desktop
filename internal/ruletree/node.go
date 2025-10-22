package ruletree

import (
	"sort"
)

type leaf[T Data] struct {
	val []T
}

type edge[T Data] struct {
	label token
	node  *node[T]
}

type node[T Data] struct {
	// leaf stores a possible leaf.
	leaf *leaf[T]

	// prefix is the common prefix.
	prefix []token

	edges []edge[T]
}

func (n *node[T]) isLeaf() bool {
	return n.leaf != nil
}

func (n *node[T]) addEdge(e edge[T]) {
	idx := sort.Search(len(n.edges), func(i int) bool {
		return n.edges[i].label >= e.label
	})

	n.edges = append(n.edges, edge[T]{})
	copy(n.edges[idx+1:], n.edges[idx:])
	n.edges[idx] = e
}

func (n *node[T]) updateEdge(label token, node *node[T]) {
	idx := sort.Search(len(n.edges), func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < len(n.edges) && n.edges[idx].label == label {
		n.edges[idx].node = node
	}
}

func (n *node[T]) getEdge(label token) *node[T] {
	idx := sort.Search(len(n.edges), func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < len(n.edges) && n.edges[idx].label == label {
		return n.edges[idx].node
	}
	return nil
}

func (n *node[T]) traverse(url string) []T {
	var data []T

	sep := n.getEdge(tokenSeparator)

	if len(url) == 0 {
		if re := n.getEdge(tokenAnchor); re != nil && re.isLeaf() {
			data = append(data, re.leaf.val...)
		}
		if sep != nil && sep.isLeaf() {
			data = append(data, sep.leaf.val...)
		}
		return data
	}

	wild := n.getEdge(tokenWildcard)

	var traversePrefix func(prefix []token, url string)
	traversePrefix = func(prefix []token, url string) {
		if len(prefix) == 0 {
			if n.isLeaf() {
				data = append(data, n.leaf.val...)
			}
			if url != "" {
				firstCh := url[0]
				if isSeparator(firstCh) && sep != nil {
					data = append(data, sep.traverse(url)...)
				}
				if wild != nil {
					data = append(data, wild.traverse(url)...)
				}
				if ch := n.getEdge(token(firstCh)); ch != nil {
					data = append(data, ch.traverse(url)...)
				}
			}
			return
		}
		if len(url) == 0 {
			if n.isLeaf() && len(prefix) == 1 && (prefix[0] == tokenAnchor || prefix[0] == tokenSeparator || prefix[0] == tokenWildcard) {
				data = append(data, n.leaf.val...)
			}
			return
		}

		switch prefix[0] {
		case tokenWildcard:
			if len(prefix) == 1 {
				for i := range len(url) {
					traversePrefix(nil, url[i:])
				}
			} else {
				nextTok := prefix[1]
				if nextTok == tokenAnchor {
					traversePrefix(prefix[1:], "")
					return
				}
				for i := range len(url) {
					switch nextTok {
					case tokenSeparator:
						if isSeparator(url[i]) {
							traversePrefix(prefix[1:], url[i:])
						}
					default:
						if url[i] == byte(nextTok) {
							traversePrefix(prefix[1:], url[i:])
						}
					}
				}
			}
		case tokenSeparator:
			switch isSeparator(url[0]) {
			case true:
				traversePrefix(prefix[1:], url[1:])
				traversePrefix(prefix, url[1:]) // Separator may consume multiple subsequent "separator" characters
			case false:
				return
			}
		default:
			if prefix[0] == token(url[0]) {
				traversePrefix(prefix[1:], url[1:])
			}
		}
	}

	traversePrefix(n.prefix, url)

	return data
}

var separators [256]bool

func init() {
	for _, ch := range "~:/?#[]@!$&'()*+,;=" {
		separators[ch] = true
	}
}

func isSeparator(char byte) bool {
	return separators[char]
}
