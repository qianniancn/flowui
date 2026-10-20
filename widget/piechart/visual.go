package piechart

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type pieChartVisualCache struct {
	cache computed.ProviderCache[struct{}, *theme.PieChartVisual]
}

// PieChartVisualCacheStats exposes provider cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type PieChartVisualCacheStats = computed.ProviderCacheStats

func (c *pieChartVisualCache) Stats() PieChartVisualCacheStats {
	if c == nil {
		return PieChartVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.PieChartOverride) *theme.PieChartVisual {
	if override != nil || w.visualCache == nil {
		return pieChartVisualForTheme(active, override)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() *theme.PieChartVisual {
		return pieChartVisualForTheme(active, nil)
	})
}

func pieChartVisualForTheme(active *theme.Theme, override *theme.PieChartOverride) *theme.PieChartVisual {
	visual := theme.ResolveVisual(active, theme.PieChartRequest{Override: override},
		theme.ComponentVisualProvider.PieChart, theme.ValidatePieChartVisual)
	return &visual
}
