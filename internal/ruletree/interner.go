package ruletree

import (
	"log"
	"sync"
)

// TokenInterner hands out a small integer for each unique string.
type TokenInterner struct {
	mu    sync.Mutex
	next  uint32
	ids   map[string]uint32
	names []string // optional reverse lookup if you ever need the string
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
	log.Printf("next token ID = %d\n", in.next)
	in.ids[s] = id
	in.names = append(in.names, s)
	return id
}

// Name returns the original string for an ID (if you ever need it).
func (in *TokenInterner) Name(id uint32) string {
	in.mu.Lock()
	defer in.mu.Unlock()
	if int(id) < len(in.names) {
		return in.names[id]
	}
	return ""
}
