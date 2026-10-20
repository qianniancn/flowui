package candlestick

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type candlestickVisualCache struct {
	cache computed.ProviderCache[struct{}, *theme.CandlestickVisual]
}

// CandlestickVisualCacheStats exposes provider cache counters for diagnostics
// and benchmark assertions without exposing the cache implementation.
type CandlestickVisualCacheStats = computed.ProviderCacheStats

func (c *candlestickVisualCache) Stats() CandlestickVisualCacheStats {
	if c == nil {
		return CandlestickVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.CandlestickOverride) *theme.CandlestickVisual {
	if override != nil || w.visualCache == nil {
		return candlestickVisualForTheme(active, override)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() *theme.CandlestickVisual {
		return candlestickVisualForTheme(active, nil)
	})
}

func candlestickVisualForTheme(active *theme.Theme, override *theme.CandlestickOverride) *theme.CandlestickVisual {
	visual := theme.ResolveVisual(active, theme.CandlestickRequest{Override: override},
		theme.ComponentVisualProvider.Candlestick, theme.ValidateCandlestickVisual)
	return &visual
}
