package menu

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// MenuVisualCacheStats is a snapshot of one Menu instance's provider cache.
// It is intentionally limited to evidence needed by tests and benchmarks.
type MenuVisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

type menuVisualKey struct {
	overrideHash uint64
	overrideSet  bool
	compact      bool
}

type menuVisualCache struct {
	cache computed.ProviderCache[menuVisualKey, theme.MenuVisual]
}

func (c *menuVisualCache) stats() MenuVisualCacheStats {
	if c == nil {
		return MenuVisualCacheStats{}
	}
	cacheStats := c.cache.Stats()
	return MenuVisualCacheStats{
		Hits:               cacheStats.Hits,
		Misses:             cacheStats.Misses,
		ThemeInvalidations: cacheStats.ThemeInvalidations,
		KeyInvalidations:   cacheStats.KeyInvalidations,
	}
}

func menuOverrideKey(override *theme.MenuOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

// visual is the only production bridge from menu behavior to theme visuals.
// The widget supplies interaction-owned density; the provider owns
// declarations, metrics, colors, motion, and drawing helpers.
func (m Widget) visual(ctx *core.Context) theme.MenuVisual {
	return m.visualWithOverride(ctx, nil)
}

func (m Widget) visualWithOverride(ctx *core.Context, override *theme.MenuOverride) theme.MenuVisual {
	return m.visualWithCache(ctx, override, nil)
}

func (m Widget) visualWithCache(ctx *core.Context, override *theme.MenuOverride, cache *menuVisualCache) theme.MenuVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := menuOverrideKey(override)
	key := menuVisualKey{overrideHash: hash, overrideSet: set, compact: m.compact}
	return cache.resolve(active, key, func() theme.MenuVisual {
		return theme.ResolveVisual(active, theme.MenuRequest{
			Compact:  m.compact,
			Override: override,
		}, theme.ComponentVisualProvider.Menu, theme.ValidateMenuVisual)
	})
}

func (c *menuVisualCache) resolve(active *theme.Theme, key menuVisualKey, load func() theme.MenuVisual) theme.MenuVisual {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, key, load)
}
