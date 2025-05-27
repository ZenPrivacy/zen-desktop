package ruletree

import (
	"log"
	"sync"
)

// TokenInterner hands out a small integer for each unique string.
type TokenInterner struct {
	mu   sync.Mutex
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
	in.mu.Lock()
	defer in.mu.Unlock()
	if id, ok := in.ids[s]; ok {
		return id
	}
	id := in.next
	in.next++
	in.ids[s] = id
	return id
}
