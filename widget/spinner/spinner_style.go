package spinner

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type spinnerResolvedStyle struct {
	root     style.ResolvedStyle
	color    color.NRGBA
	renderer theme.SpinnerRenderer
	metrics  theme.SpinnerMetrics
	motion   theme.MotionTheme
}

type spinnerVisualKey struct {
	color Color
	size  Size
}

type spinnerVisualCache struct {
	cache computed.ProviderCache[spinnerVisualKey, theme.SpinnerVisual]
}

// SpinnerVisualCacheStats exposes provider cache behavior for regression tests
// without exposing the cache implementation.
type SpinnerVisualCacheStats struct {
	Hits, Misses, ThemeInvalidations, KeyInvalidations uint64
}

func (c *spinnerVisualCache) Stats() SpinnerVisualCacheStats {
	if c == nil {
		return SpinnerVisualCacheStats{}
	}
	stats := c.cache.Stats()
	return SpinnerVisualCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func (s Widget) visualWithCache(ctx *core.Context, override *theme.SpinnerOverride) theme.SpinnerVisual {
	active := core.ActiveTheme(ctx)
	request := theme.SpinnerRequest{Color: s.color, Size: s.size, Override: override}
	if override != nil || s.visualCache == nil {
		return resolveSpinnerVisual(active, request)
	}
	key := spinnerVisualKey{color: s.color, size: s.size}
	return s.visualCache.cache.Resolve(active, key, func() theme.SpinnerVisual {
		return resolveSpinnerVisual(active, request)
	})
}

func resolveSpinnerVisual(active *theme.Theme, request theme.SpinnerRequest) theme.SpinnerVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Spinner, theme.ValidateSpinnerVisual)
}

func (s Widget) resolveStyle(ctx *core.Context, gtx layout.Context) spinnerResolvedStyle {
	return s.resolveStyleWithOverride(ctx, gtx, s.override)
}

func (s Widget) resolveStyleWithOverride(
	ctx *core.Context,
	gtx layout.Context,
	override *theme.SpinnerOverride,
) spinnerResolvedStyle {
	visual := s.visualWithCache(ctx, override)
	root := computed.ResolveStatic(ctx, style.State{}, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	indicator := computed.ResolvePartStatic(ctx, style.PartIndicator, style.State{}, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	if computed.HasTransitions(root, indicator) {
		key := core.ClaimKey(ctx, core.KindStyle, "spinner")
		root = computed.ApplyTransitions(ctx, gtx, key, root)
		indicator = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIndicator, indicator)
	}
	if visual.Color == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	resolvedColor, ok := visual.Color(indicator)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return spinnerResolvedStyle{
		root: root, color: resolvedColor, renderer: visual.Renderer,
		metrics: visual.Metrics, motion: visual.Motion,
	}
}
