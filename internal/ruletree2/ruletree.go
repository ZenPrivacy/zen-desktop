package ruletree2

type Data any

type Tree[T Data] struct {
	root *node[T]
}

func New[T Data]() *Tree[T] {
	return &Tree[T]{}
}

func (t *Tree[T]) Insert(s []token, v T) {
	var parent *node[T]
	n := t.root
	search := s
	for {
		if len(search) == 0 {
			if n.isLeaf() {
				n.leaf.val = append(n.leaf.val, v)
			} else {
				n.leaf = &leaf[T]{
					val: []T{v},
				}
			}
			return
		}

		parent = n
		n = n.getEdge(search[0])

		if n == nil {
			n := &node[T]{
				prefix: search,
				leaf: &leaf[T]{
					val: []T{v},
				},
			}
			parent.addEdge(edge[T]{
				label: search[0],
				node:  n,
			})
			return
		}

		commonPrefix := longestPrefix(search, n.prefix)
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		child := &node[T]{
			prefix: search[:commonPrefix],
		}
		parent.updateEdge(search[0], child)

		child.addEdge(edge[T]{
			label: n.prefix[commonPrefix],
			node:  n,
		})
		n.prefix = n.prefix[commonPrefix:]

		l := &leaf[T]{
			val: []T{v},
		}
		if commonPrefix == len(search) {
			child.leaf = l
		} else {
			n := &node[T]{
				leaf:   l,
				prefix: search[commonPrefix:],
			}
			child.addEdge(edge[T]{
				label: search[0],
				node:  n,
			})
		}
		return
	}
}

func (t *Tree[T]) Get(s string) []T {
	var data []T
	data = append(data, t.root.traverse(s, true)...)
	if dr := t.root.getEdge(tokenDomainRoot); dr != nil {
		data = append(data, dr.traverse(s, true)...)
	}
	if re := t.root.getEdge(tokenRootEnd); re != nil {
		data = append(data, re.traverse(s, true)...)
	}

	for 
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
