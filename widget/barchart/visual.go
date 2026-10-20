package barchart

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type barChartVisualCache struct {
	cache computed.ProviderCache[struct{}, theme.BarChartVisual]
}

// BarChartVisualCacheStats exposes provider cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type BarChartVisualCacheStats = computed.ProviderCacheStats

func (c *barChartVisualCache) Stats() BarChartVisualCacheStats {
	if c == nil {
		return BarChartVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.BarChartOverride) theme.BarChartVisual {
	requestOverride := w.providerOverride(override)
	load := func(value *theme.BarChartOverride) theme.BarChartVisual {
		return theme.ResolveVisual(active, theme.BarChartRequest{Override: value},
			theme.ComponentVisualProvider.BarChart, theme.ValidateBarChartVisual)
	}
	if requestOverride != nil || w.visualCache == nil {
		return load(requestOverride)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() theme.BarChartVisual {
		return load(nil)
	})
}

func (w Widget) providerOverride(override *theme.BarChartOverride) *theme.BarChartOverride {
	return override
}
