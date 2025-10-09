package ruletree

// TrimSlice shrinks the capacity of a slice to its length and
// returns the new slice and number of capacity reductions made.
func TrimSlice[T any](s []T) ([]T, int) {
	if s == nil {
		return nil, 0
	}

	if len(s) == cap(s) {
		return s, 0
	}

	newSlice := make([]T, len(s))
	copy(newSlice, s)
	return newSlice, cap(s) - len(s)
}
