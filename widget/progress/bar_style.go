package progress

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type progressBarResolvedStyle struct {
	root     style.ResolvedStyle
	track    style.ResolvedStyle
	fill     style.ResolvedStyle
	label    style.ResolvedStyle
	metrics  theme.ProgressBarMetrics
	motion   theme.MotionTheme
	override style.Style
}

type progressDeclarationKey struct {
	color BarColor
	size  BarSize
}

type progressBarVisualCache struct {
	cache computed.ProviderCache[progressDeclarationKey, theme.ProgressBarVisual]
}

type progressCircleVisualCache struct {
	cache computed.ProviderCache[progressDeclarationKey, theme.ProgressCircleVisual]
}

// ProgressVisualCacheStats makes provider cache behavior observable to
// component benchmarks without exposing the internal cache implementation.
type ProgressVisualCacheStats struct {
	Hits, Misses, ThemeInvalidations, KeyInvalidations uint64
}

func (c *progressBarVisualCache) resolve(active *theme.Theme, key progressDeclarationKey) theme.ProgressBarVisual {
	return c.cache.Resolve(active, key, func() theme.ProgressBarVisual {
		return resolveProgressBarProvider(active, key, nil)
	})
}

func (c *progressCircleVisualCache) resolve(active *theme.Theme, key progressDeclarationKey) theme.ProgressCircleVisual {
	return c.cache.Resolve(active, key, func() theme.ProgressCircleVisual {
		return resolveProgressCircleProvider(active, key, nil)
	})
}

func resolveProgressBarProvider(active *theme.Theme, key progressDeclarationKey, override *theme.ProgressOverride) theme.ProgressBarVisual {
	return theme.ResolveVisual(active, theme.ProgressBarRequest{Color: key.color, Size: key.size, Override: override},
		theme.ComponentVisualProvider.ProgressBar, theme.ValidateProgressBarVisual)
}

func resolveProgressCircleProvider(active *theme.Theme, key progressDeclarationKey, override *theme.ProgressOverride) theme.ProgressCircleVisual {
	return theme.ResolveVisual(active, theme.ProgressCircleRequest{Color: key.color, Size: key.size, Override: override},
		theme.ComponentVisualProvider.ProgressCircle, theme.ValidateProgressCircleVisual)
}

func (c *progressBarVisualCache) Stats() ProgressVisualCacheStats {
	if c == nil {
		return ProgressVisualCacheStats{}
	}
	stats := c.cache.Stats()
	return ProgressVisualCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func (c *progressCircleVisualCache) Stats() ProgressVisualCacheStats {
	if c == nil {
		return ProgressVisualCacheStats{}
	}
	stats := c.cache.Stats()
	return ProgressVisualCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func (p BarWidget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, cache *progressBarVisualCache, overrides ...*theme.ProgressOverride) progressBarResolvedStyle {
	active := core.ActiveTheme(ctx)
	declarationKey := progressDeclarationKey{color: p.color, size: p.size}
	var visual theme.ProgressBarVisual
	if len(overrides) == 0 || overrides[0] == nil {
		visual = cache.resolve(active, declarationKey)
	} else {
		visual = resolveProgressBarProvider(active, declarationKey, overrides[0])
	}
	state := style.State{Disabled: p.disabled}
	parts := [...]style.Part{style.PartRoot, style.PartTrack, style.PartFill, style.PartLabel}
	var resolved [len(parts)]style.ResolvedStyle
	computed.ResolvePartsInto(ctx, gtx, key, state, parts[:], resolved[:], visual.Defaults, visual.Variant, visual.Size, visual.Override)
	return progressBarResolvedStyle{
		root: resolved[0], track: resolved[1], fill: resolved[2], label: resolved[3],
		metrics: visual.Metrics, motion: visual.Motion, override: visual.Override,
	}
}
