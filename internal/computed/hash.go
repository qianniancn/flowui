package computed

import (
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func collectStyleLayers(
	ctx *core.Context,
	defaults, variant, size, custom style.Style,
) []style.Style {
	inherited := core.ActiveInheritedStylesReadOnly(ctx)
	active := core.ActiveStylesReadOnly(ctx)
	layers := make([]style.Style, 0, 4+len(inherited)+len(active))
	layers = append(layers, defaults)
	layers = append(layers, inherited...)
	layers = append(layers, variant, size)
	layers = append(layers, active...)
	layers = append(layers, custom)
	return layers
}

func hashStyleLayers(
	ctx *core.Context,
	state style.State,
	defaults, variant, size, custom style.Style,
) uint64 {
	// Keep this hash on the stack. This function runs once for every cached
	// style lookup in a real list, so constructing a temporary layer slice here
	// would show up directly in component benchmarks.
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)
	h := offset64
	mix := func(value uint64) {
		for range 8 {
			h ^= value & 0xff
			h *= prime64
			value >>= 8
		}
	}
	writeStyle := func(value style.Style) {
		mix(value.Hash64ForState(state))
	}
	writeStyle(defaults)
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		writeStyle(value)
	}
	writeStyle(variant)
	writeStyle(size)
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		writeStyle(value)
	}
	writeStyle(custom)
	return h
}

func styleLayersCacheSafe(ctx *core.Context, values ...style.Style) bool {
	for _, value := range values {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	return true
}

// StyleLayersCacheSafe reports whether a declaration can participate in a
// hash-keyed static resolve cache under the current style scope.
func StyleLayersCacheSafe(ctx *core.Context, values ...style.Style) bool {
	return styleLayersCacheSafe(ctx, values...)
}

// StyleScopeHash returns a stable hash for the current inherited and active
// style scope. Components can compute it once per layout and use it to
// invalidate a local resolved-style cache when the scope changes.
func StyleScopeHash(ctx *core.Context) uint64 {
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)
	hash := offset64
	mix := func(value uint64) {
		for range 8 {
			hash ^= value & 0xff
			hash *= prime64
			value >>= 8
		}
	}
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	return hash
}
