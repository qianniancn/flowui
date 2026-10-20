package linechart

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type lineChartVisualCache struct {
	cache computed.ProviderCache[struct{}, *theme.LineChartVisual]
}

// LineChartVisualCacheStats exposes provider cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type LineChartVisualCacheStats = computed.ProviderCacheStats

func (c *lineChartVisualCache) Stats() LineChartVisualCacheStats {
	if c == nil {
		return LineChartVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.LineChartOverride) *theme.LineChartVisual {
	if override != nil || w.visualCache == nil {
		return lineChartVisualForTheme(active, override)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() *theme.LineChartVisual {
		return lineChartVisualForTheme(active, nil)
	})
}

func lineChartVisualForTheme(active *theme.Theme, override *theme.LineChartOverride) *theme.LineChartVisual {
	visual := theme.ResolveVisual(active, theme.LineChartRequest{Override: override},
		theme.ComponentVisualProvider.LineChart, theme.ValidateLineChartVisual)
	return &visual
}
