package state

// memoEntry holds a cached computation result with its invalidation tokens.
type memoEntry struct {
	themeGen  uint64
	inputHash uint64
	value     any
}

// Memo provides memoization for expensive computations. Results are cached
// by component identity and invalidated when theme or inputs change.
type Memo[T any] struct {
	store     *Store
	identity  Identity
	themeGen  uint64
	inputHash uint64
}

// UseMemo returns a memoization handle for the given identity and inputs.
//
// Parameters:
//   - store: the state store (usually from frame.Context)
//   - key: component key (from frame.ClaimKey)
//   - slot: memo slot name (e.g. "style-resolve", "chart-path")
//   - themeGen: theme generation (bumped when theme changes)
//   - inputHash: hash of all inputs affecting the result
//
// The memo is valid only for the current frame. Call UseMemo each frame,
// then use Get/Set or Compute.
func UseMemo[T any](store *Store, key, slot string, themeGen, inputHash uint64) *Memo[T] {
	return &Memo[T]{
		store:     store,
		identity:  Identity{Key: key, Slot: slot},
		themeGen:  themeGen,
		inputHash: inputHash,
	}
}

// Get attempts to retrieve the cached result. Returns (result, true) on cache
// hit, or (zero, false) on cache miss.
//
// Cache hits require both themeGen and inputHash to match the cached entry.
func (m *Memo[T]) Get() (T, bool) {
	if m.store == nil {
		var zero T
		return zero, false
	}

	entry, ok := Peek[memoEntry](m.store, m.identity)
	if !ok {
		var zero T
		return zero, false
	}

	// Cache hit requires exact match on both tokens
	if entry.themeGen != m.themeGen || entry.inputHash != m.inputHash {
		var zero T
		return zero, false
	}

	result, ok := entry.value.(T)
	if !ok {
		// Type mismatch should not happen if slot names are unique
		var zero T
		return zero, false
	}

	return result, true
}

// Set stores a new result in the cache. The result will be returned by
// future Get calls until the theme or inputs change.
func (m *Memo[T]) Set(value T) {
	if m.store == nil {
		return
	}

	entry := Use(m.store, m.identity, func() *memoEntry {
		return &memoEntry{}
	})

	entry.themeGen = m.themeGen
	entry.inputHash = m.inputHash
	entry.value = value
}

// Compute is a convenience method that gets the cached result or computes it.
//
// Usage:
//
//	result := memo.Compute(func() ExpensiveResult {
//	    return computeExpensive(...)
//	})
//
// The function fn is called only on cache miss. Its result is cached for
// future calls with the same themeGen and inputHash.
func (m *Memo[T]) Compute(fn func() T) T {
	if result, ok := m.Get(); ok {
		return result
	}
	result := fn()
	m.Set(result)
	return result
}
