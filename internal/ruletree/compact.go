package ruletree

// Compact shrinks internal slice capacities to reduce memory usage.
func (t *Tree[T]) Compact() {
	t.insertMu.Lock()
	defer t.insertMu.Unlock()

	t.generic = trimSlice(t.generic)

	stack := []*node[T]{
		t.anchorRoot,
		t.domainBoundaryRoot,
		t.root,
	}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if l := n.leaf; l != nil {
			l.val = trimSlice(l.val)
		}

		n.edges = trimSlice(n.edges)
		n.prefix = trimSlice(n.prefix)

		for _, e := range n.edges {
			stack = append(stack, e.node)
		}
	}
}

func trimSlice[T any](s []T) []T {
	if len(s) == cap(s) {
		return s
	}

	newSlice := make([]T, len(s))
	copy(newSlice, s)
	return newSlice
}
