package rule

import "sync"

// ModifierInterner shares parsed modifiers between rules with the same modifier text.
// Filter lists repeat modifier text a lot: AdGuard Spyware alone has over 200k
// email tracking-pixel rules of the form "email.example.com$image", which then
// use one block instead of 200k. It is safe for concurrent use.
//
// A nil *ModifierInterner is valid and shares nothing.
type ModifierInterner struct {
	mu sync.Mutex
	m  map[string]*ruleModifiers
}

func NewModifierInterner() *ModifierInterner {
	return &ModifierInterner{m: make(map[string]*ruleModifiers)}
}

// Parse parses rawModifiers, the text after '$', into rm, a freshly constructed Rule.
func (in *ModifierInterner) Parse(rm *Rule, rawModifiers string) error {
	if in == nil {
		return rm.ParseModifiers(splitModifiers(rawModifiers))
	}

	// Keying on the raw text skips splitting on a hit. Escaping variants of
	// the same modifiers get separate blocks, which only costs a missed share.
	in.mu.Lock()
	mods, ok := in.m[rawModifiers]
	in.mu.Unlock()
	if ok {
		rm.mods = mods
		return nil
	}

	if err := rm.ParseModifiers(splitModifiers(rawModifiers)); err != nil {
		return err
	}

	// Parsing outside the lock keeps lists parsing in parallel. Two goroutines
	// may parse the same text at once: the later store wins and the earlier
	// rules keep an identical block of their own, which only costs a missed share.
	in.mu.Lock()
	in.m[rawModifiers] = rm.mods
	in.mu.Unlock()
	return nil
}
