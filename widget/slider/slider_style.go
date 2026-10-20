package slider

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type sliderStyle struct {
	track style.ResolvedStyle
	fill  style.ResolvedStyle
	thumb style.ResolvedStyle
	label style.ResolvedStyle
	// Keep the provider bundle intact so layout and motion consume one visual
	// source instead of copying theme-owned fields into widget state.
	visual theme.SliderVisual
}

// VisualCache retains provider-resolved Slider visuals for repeated layouts.
// Instance overrides bypass this cache because they are scoped to one widget.
type VisualCache struct {
	cache computed.ProviderCache[Orientation, theme.SliderVisual]
}

// VisualCacheStats exposes provider-cache behavior for benchmarks and tests.
type VisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

func (c *VisualCache) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	stats := c.cache.Stats()
	return VisualCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func (c *VisualCache) Resolve(activeTheme *theme.Theme, orientation Orientation) theme.SliderVisual {
	if c == nil {
		return resolveSliderProvider(activeTheme, orientation, nil)
	}
	return c.cache.Resolve(activeTheme, orientation, func() theme.SliderVisual {
		return resolveSliderProvider(activeTheme, orientation, nil)
	})
}

func resolveSliderProvider(activeTheme *theme.Theme, orientation Orientation, override *theme.SliderOverride) theme.SliderVisual {
	return theme.ResolveVisual(activeTheme, theme.SliderRequest{
		Orientation: theme.SliderOrientation(orientation),
		Override:    override,
	}, theme.ComponentVisualProvider.Slider, theme.ValidateSliderVisual)
}

func sliderVisualCached(activeTheme *theme.Theme, orientation Orientation, override *theme.SliderOverride, caches ...*VisualCache) theme.SliderVisual {
	if override == nil && len(caches) > 0 && caches[0] != nil {
		return caches[0].Resolve(activeTheme, orientation)
	}
	return resolveSliderProvider(activeTheme, orientation, override)
}

func (s Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, visuals ...theme.SliderVisual) sliderStyle {
	var visual theme.SliderVisual
	if len(visuals) > 0 {
		visual = visuals[0]
	} else {
		visual = sliderVisualCached(core.ActiveTheme(ctx), s.orientation, nil)
	}
	resolvePart := func(part style.Part) style.ResolvedStyle {
		return computed.ResolvePart(ctx, gtx, key, part, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	}
	return sliderStyle{
		track:  resolvePart(style.PartTrack),
		fill:   resolvePart(style.PartFill),
		thumb:  resolvePart(style.PartThumb),
		label:  resolvePart(style.PartLabel),
		visual: visual,
	}
}
