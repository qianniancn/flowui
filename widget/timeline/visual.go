package timeline

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type timelineVisualCache struct {
	cache      computed.ProviderCache[struct{}, theme.TimelineVisual]
	visualHost timelineVisualHost
}

type timelineVisualHost struct {
	value  Widget
	visual theme.TimelineVisual
}

func (v *timelineVisualHost) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return v.value.layout(ctx, gtx, v.visual)
}

// TimelineVisualCacheStats exposes provider cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type TimelineVisualCacheStats = computed.ProviderCacheStats

func (c *timelineVisualCache) Stats() TimelineVisualCacheStats {
	if c == nil {
		return TimelineVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.TimelineOverride) theme.TimelineVisual {
	load := func(value *theme.TimelineOverride) theme.TimelineVisual {
		return theme.ResolveVisual(active, theme.TimelineRequest{Override: value},
			theme.ComponentVisualProvider.Timeline, theme.ValidateTimelineVisual)
	}
	if override != nil || w.visualCache == nil {
		return load(override)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() theme.TimelineVisual {
		return load(nil)
	})
}
