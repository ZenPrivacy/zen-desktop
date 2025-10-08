package ruletree2

import (
	"sort"
)

type node[T Data] struct {
	// leaf stores a possible leaf.
	leaf *leaf[T]

	// prefix is the common prefix.
	prefix []token

	edges edges[T]
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
		return
	}
	panic("updating missing edge")
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
		if re := n.getEdge(tokenStartEnd); re != nil && re.isLeaf() {
			data = append(data, re.leaf.val...)
		}
		if sep != nil && sep.isLeaf() {
			data = append(data, sep.leaf.val...)
		}
		return data
	}

	wild := n.getEdge(tokenWildcard)

	proceed := func(url string) {
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

	var traversePrefix func(prefix []token, url string)
	traversePrefix = func(prefix []token, url string) {
		if len(prefix) == 0 {
			if n.isLeaf() {
				data = append(data, n.leaf.val...)
			}
			proceed(url)
			return
		}
		if len(url) == 0 {
			if n.isLeaf() && len(prefix) == 1 && (prefix[0] == tokenStartEnd || prefix[0] == tokenSeparator) {
				data = append(data, n.leaf.val...)
			}
			return
		}

		switch prefix[0] {
		case tokenWildcard:
			traversePrefix(prefix[1:], url)     // Wildcard can match zero chars,
			traversePrefix(prefix[1:], url[1:]) // one,
			traversePrefix(prefix, url[1:])     // or many.
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

type leaf[T Data] struct {
	val []T
}

type edge[T Data] struct {
	label token
	node  *node[T]
}

type edges[T Data] []edge[T]

func (e edges[T]) Len() int {
	return len(e)
}

func (e edges[T]) Less(i, j int) bool {
	return e[i].label < e[j].label
}

func (e edges[T]) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges[T]) Sort() {
	sort.Sort(e)
}

func isSeparator(char byte) bool {
	return !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-' || char == '.' || char == '%')
}
