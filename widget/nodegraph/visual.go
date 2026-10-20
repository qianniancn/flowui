package nodegraph

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type nodeGraphVisualCache struct {
	cache          computed.ProviderCache[struct{}, theme.NodeGraphVisual]
	visualHost     nodeGraphVisualHost
	minimapScratch nodeGraphMinimapScratch
}

// NodeGraphVisualCacheStats exposes provider cache counters for diagnostics
// and benchmark assertions without exposing the cache implementation.
type NodeGraphVisualCacheStats = computed.ProviderCacheStats

func (c *nodeGraphVisualCache) Stats() NodeGraphVisualCacheStats {
	if c == nil {
		return NodeGraphVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.NodeGraphOverride) theme.NodeGraphVisual {
	load := func(value *theme.NodeGraphOverride) theme.NodeGraphVisual {
		return theme.ResolveVisual(active, theme.NodeGraphRequest{Override: value},
			theme.ComponentVisualProvider.NodeGraph, theme.ValidateNodeGraphVisual)
	}
	if override != nil || w.visualCache == nil {
		return load(override)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() theme.NodeGraphVisual {
		return load(nil)
	})
}
