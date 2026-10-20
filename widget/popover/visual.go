package popover

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// PopoverVisualCacheStats is a snapshot of one Popover instance's visual
// cache. It is intentionally limited to evidence needed by tests and
// benchmarks; the cache remains an implementation detail of behavior state.
type PopoverVisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

type popoverVisualKey struct {
	overrideHash uint64
	overrideSet  bool
}

type popoverVisualCache struct {
	cache computed.ProviderCache[popoverVisualKey, theme.PopoverVisual]
}

func (c *popoverVisualCache) stats() PopoverVisualCacheStats {
	if c == nil {
		return PopoverVisualCacheStats{}
	}
	cacheStats := c.cache.Stats()
	return PopoverVisualCacheStats{
		Hits:               cacheStats.Hits,
		Misses:             cacheStats.Misses,
		ThemeInvalidations: cacheStats.ThemeInvalidations,
		KeyInvalidations:   cacheStats.KeyInvalidations,
	}
}

func popoverOverrideKey(override *theme.PopoverOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (p Widget) visualWithCache(ctx *core.Context, override *theme.PopoverOverride, cache *popoverVisualCache) theme.PopoverVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := popoverOverrideKey(override)
	key := popoverVisualKey{overrideHash: hash, overrideSet: set}
	return cache.resolve(active, key, func() theme.PopoverVisual {
		return theme.ResolveVisual(active, theme.PopoverRequest{Override: override},
			theme.ComponentVisualProvider.Popover, theme.ValidatePopoverVisual)
	})
}

func (c *popoverVisualCache) resolve(active *theme.Theme, key popoverVisualKey, load func() theme.PopoverVisual) theme.PopoverVisual {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, key, load)
}
