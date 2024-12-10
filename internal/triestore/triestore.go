package triestore

import (
	"strings"
	"sync"
)

type node[T comparable] struct {
	children map[string]*node[T]
	data     []T
}

func (n *node[T]) findOrAddChild(segment string) *node[T] {
	if n.children == nil {
		newChild := &node[T]{}
		n.children = map[string]*node[T]{
			segment: newChild,
		}
		return newChild
	}

	existingChild, ok := n.children[segment]
	if ok {
		return existingChild
	}

	newChild := &node[T]{}
	n.children[segment] = newChild
	return newChild
}

func (n *node[T]) getMatchingData(segments []string, isWildcard bool) []T {
	if len(segments) == 0 {
		return n.data
	}

	var data []T
	if isWildcard {
		// Wildcards can consume as many segments as possible.
		data = append(data, n.getMatchingData(segments[1:], true)...)
	}

	if wildcardChild, ok := n.children["*"]; ok {
		data = append(data, wildcardChild.getMatchingData(segments[1:], true)...)
	}
	if exactChild, ok := n.children[segments[0]]; ok {
		data = append(data, exactChild.getMatchingData(segments[1:], false)...)
	}

	return data
}

type TrieStore[T comparable] struct {
	mu sync.RWMutex
	// universal data is data not tied to a specific hostname.
	universalData []T
	root          *node[T]
}

func NewTrieStore[T comparable]() *TrieStore[T] {
	return &TrieStore[T]{
		root: &node[T]{},
	}
}

func (ts *TrieStore[T]) Add(hostnames []string, data T) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(hostnames) == 0 {
		ts.universalData = append(ts.universalData, data)
		return
	}

	for _, hostname := range hostnames {
		segments := strings.Split(hostname, ".")

		node := ts.root
		for _, segment := range segments {
			node = node.findOrAddChild(segment)
		}
		node.data = append(node.data, data)
	}
}

func (ts *TrieStore[T]) Get(hostname string) []T {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	segments := strings.Split(hostname, ".")
	return append(ts.root.getMatchingData(segments, false), ts.universalData...)
}
