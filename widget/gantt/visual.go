package gantt

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type ganttVisualCache struct {
	cache computed.ProviderCache[struct{}, theme.GanttVisual]
}

// GanttVisualCacheStats exposes provider cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type GanttVisualCacheStats = computed.ProviderCacheStats

func (c *ganttVisualCache) Stats() GanttVisualCacheStats {
	if c == nil {
		return GanttVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.GanttOverride) theme.GanttVisual {
	load := func(value *theme.GanttOverride) theme.GanttVisual {
		return theme.ResolveVisual(active, theme.GanttRequest{Override: value},
			theme.ComponentVisualProvider.Gantt, theme.ValidateGanttVisual)
	}
	if override != nil || w.visualCache == nil {
		return load(override)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() theme.GanttVisual {
		return load(nil)
	})
}
