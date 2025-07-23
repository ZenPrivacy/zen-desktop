package ruletree

import (
	"sync"
)

// TokenInterner hands out a small integer for each unique string.
type TokenInterner struct {
	mu   sync.RWMutex
	next uint32
	ids  map[string]uint32
}

func NewTokenInterner() *TokenInterner {
	return &TokenInterner{
		ids: make(map[string]uint32),
	}
}

// Intern returns the unique ID for s, assigning a new one if needed.
func (in *TokenInterner) Intern(s string) uint32 {
	in.mu.RLock()
	if id, ok := in.ids[s]; ok {
		in.mu.RUnlock()
		return id
	}
	in.mu.RUnlock()

	in.mu.Lock()
	defer in.mu.Unlock()

	// Another goroutine may have inserted it between RUnlock() and Lock(),
	// so check again:
	if id, ok := in.ids[s]; ok {
		return id
	}

	// Now assign a new ID
	id := in.next
	in.next++
	in.ids[s] = id
	return id
}
