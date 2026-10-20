package tooltip

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// TooltipVisualCacheStats is a snapshot of one tooltip instance's provider
// cache. It exposes only evidence needed by tests and benchmarks.
type TooltipVisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

type tooltipVisualKey struct {
	overrideHash uint64
	overrideSet  bool
}

type tooltipVisualCache struct {
	cache computed.ProviderCache[tooltipVisualKey, theme.TooltipVisual]
}

func (c *tooltipVisualCache) stats() TooltipVisualCacheStats {
	if c == nil {
		return TooltipVisualCacheStats{}
	}
	stats := c.cache.Stats()
	return TooltipVisualCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func tooltipOverrideKey(override *theme.TooltipOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func resolveTooltipVisual(active *theme.Theme, override *theme.TooltipOverride) theme.TooltipVisual {
	return theme.ResolveVisual(active, theme.TooltipRequest{Override: override},
		theme.ComponentVisualProvider.Tooltip, theme.ValidateTooltipVisual)
}

func (c *tooltipVisualCache) resolve(active *theme.Theme, key tooltipVisualKey, load func() theme.TooltipVisual) theme.TooltipVisual {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, key, load)
}

func tooltipVisualWithCache(ctx *core.Context, override *theme.TooltipOverride, cache *tooltipVisualCache) theme.TooltipVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := tooltipOverrideKey(override)
	key := tooltipVisualKey{overrideHash: hash, overrideSet: set}
	return cache.resolve(active, key, func() theme.TooltipVisual {
		return resolveTooltipVisual(active, override)
	})
}

func (p Popup) visualWithCache(ctx *core.Context, override *theme.TooltipOverride) theme.TooltipVisual {
	return tooltipVisualWithCache(ctx, override, p.visualCache)
}
