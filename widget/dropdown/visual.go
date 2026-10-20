package dropdown

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// DropdownVisualCacheStats is a snapshot of the provider visual cache for one
// Dropdown instance. It is intentionally small so benchmarks can observe
// invalidation without exposing cache internals to applications.
type DropdownVisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

type dropdownVisualKey struct {
	overrideHash uint64
	overrideSet  bool
}

type dropdownVisualCache struct {
	cache computed.ProviderCache[dropdownVisualKey, theme.DropdownVisual]
}

func (c *dropdownVisualCache) stats() DropdownVisualCacheStats {
	if c == nil {
		return DropdownVisualCacheStats{}
	}
	cacheStats := c.cache.Stats()
	return DropdownVisualCacheStats{
		Hits:               cacheStats.Hits,
		Misses:             cacheStats.Misses,
		ThemeInvalidations: cacheStats.ThemeInvalidations,
		KeyInvalidations:   cacheStats.KeyInvalidations,
	}
}

func dropdownOverrideKey(override *theme.DropdownOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

// visual is the only production bridge from dropdown behavior to theme
// visuals. A layout pass resolves it once and passes the immutable plan to all
// trigger and overlay paths.
func (d Widget) visual(ctx *core.Context, override *theme.DropdownOverride, cache *dropdownVisualCache) theme.DropdownVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := dropdownOverrideKey(override)
	key := dropdownVisualKey{overrideHash: hash, overrideSet: set}
	return cache.resolve(active, key, func() theme.DropdownVisual {
		return theme.ResolveVisual(active, theme.DropdownRequest{Override: override},
			theme.ComponentVisualProvider.Dropdown, theme.ValidateDropdownVisual)
	})
}

func (c *dropdownVisualCache) resolve(active *theme.Theme, key dropdownVisualKey, load func() theme.DropdownVisual) theme.DropdownVisual {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, key, load)
}
