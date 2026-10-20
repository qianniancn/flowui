package heatmap

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type heatmapVisualCache struct {
	cache computed.ProviderCache[struct{}, *theme.HeatmapVisual]
}

// HeatmapVisualCacheStats exposes provider cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type HeatmapVisualCacheStats = computed.ProviderCacheStats

func (c *heatmapVisualCache) Stats() HeatmapVisualCacheStats {
	if c == nil {
		return HeatmapVisualCacheStats{}
	}
	return c.cache.Stats()
}
