package runtime

import (
	"hash/fnv"

	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/style"
)

func hashStyleLayers(
	ctx *frame.Context,
	state style.StyleState,
	defaults, variant, size, custom style.Style,
) uint64 {
	h := fnv.New64a()
	writeStyle := func(value style.Style) {
		h.Write(uint64ToBytes(value.Hash64ForState(state)))
	}
	writeStyle(defaults)
	for _, value := range frame.ActiveInheritedStylesReadOnly(ctx) {
		writeStyle(value)
	}
	writeStyle(variant)
	writeStyle(size)
	for _, value := range frame.ActiveStylesReadOnly(ctx) {
		writeStyle(value)
	}
	writeStyle(custom)

	return h.Sum64()
}

func styleLayersCacheSafe(ctx *frame.Context, values ...style.Style) bool {
	for _, value := range values {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range frame.ActiveInheritedStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range frame.ActiveStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	return true
}

func uint64ToBytes(v uint64) []byte {
	return []byte{
		byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
		byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56),
	}
}
