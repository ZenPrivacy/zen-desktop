package ruletree

import (
	"slices"
	"strings"

	"github.com/irbis-sh/zen-desktop/internal/trimslice"
)

// Compact moves bucketed domains into their compact sorted form and shrinks
// internal slice capacities to reduce memory usage.
func (t *Tree[T]) Compact() {
	t.freezeDomains()

	stack := []*node[T]{
		t.anchorRoot,
		t.domainBoundaryRoot,
		t.root,
	}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if n.isLeaf() {
			n.leaf = trimslice.TrimSlice(n.leaf)
		}

		n.edges = trimslice.TrimSlice(n.edges)
		n.prefix = trimslice.TrimSlice(n.prefix)

		if n.wildcard != nil {
			stack = append(stack, n.wildcard)
		}
		if n.separator != nil {
			stack = append(stack, n.separator)
		}
		if n.anchor != nil {
			stack = append(stack, n.anchor)
		}
		for _, e := range n.edges {
			stack = append(stack, e.node)
		}
	}
}

// freezeDomains moves the domains map into frozen.
func (t *Tree[T]) freezeDomains() {
	if len(t.domains) == 0 {
		return
	}

	n := len(t.frozen)
	for _, vals := range t.domains {
		n += len(vals)
	}
	frozen := append(make([]domainEntry[T], 0, n), t.frozen...)
	for key, vals := range t.domains {
		for _, v := range vals {
			frozen = append(frozen, domainEntry[T]{key, v})
		}
	}
	slices.SortFunc(frozen, func(a, b domainEntry[T]) int {
		return strings.Compare(a.key, b.key)
	})

	t.frozen = frozen
	// A fresh map, because clear() would keep the old one's slots allocated.
	t.domains = make(map[string][]T)
}
