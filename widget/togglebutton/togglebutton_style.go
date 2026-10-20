package togglebutton

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type toggleButtonResolvedStyle struct {
	root    style.ResolvedStyle
	content style.ResolvedStyle
}

// VisualCache retains provider-resolved toggle visuals for repeated layouts.
// The foreground is part of the key because inherited scope can change it
// without replacing the active theme. Instance overrides deliberately bypass
// this cache because they are scoped to one widget instance.
type VisualCache struct {
	cache computed.ProviderCacheMap[toggleVisualCacheKey, theme.ToggleVisual]
}

type toggleVisualCacheKey struct {
	variant    theme.ToggleButtonVariant
	size       theme.ToggleButtonSize
	iconOnly   bool
	foreground color.NRGBA
}

// VisualCacheStats exposes provider-cache behavior for benchmarks and
// regression tests without exposing cache internals.
type VisualCacheStats = computed.ProviderCacheStats

// Stats returns a snapshot of cache activity.
func (c *VisualCache) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	return c.cache.Stats()
}

// Resolve returns the cached visual for a provider request.
func (c *VisualCache) Resolve(activeTheme *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	if c == nil || request.Override != nil {
		return resolveToggleProvider(activeTheme, request)
	}
	key := toggleVisualCacheKey{
		variant:    request.Variant,
		size:       request.Size,
		iconOnly:   request.IconOnly,
		foreground: request.Foreground,
	}
	return c.cache.Resolve(activeTheme, key, func() theme.ToggleVisual {
		return resolveToggleProvider(activeTheme, request)
	})
}

func (b Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, override *theme.ToggleOverride, caches ...*VisualCache) toggleButtonResolvedStyle {
	visual := toggleVisualCached(core.ActiveTheme(ctx), b, ctx.ForegroundColor(), override, caches)
	part := style.PartLabel
	if b.iconOnly {
		part = style.PartIcon
	}
	return toggleButtonResolvedStyle{
		root:    computed.Resolve(ctx, gtx, key, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		content: computed.ResolvePart(ctx, gtx, key, part, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
	}
}

func resolveToggleProvider(activeTheme *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	return theme.ResolveVisual(activeTheme, request,
		theme.ComponentVisualProvider.ToggleButton, theme.ValidateToggleVisual)
}

func toggleVisualCached(activeTheme *theme.Theme, b Widget, foreground color.NRGBA, override *theme.ToggleOverride, caches []*VisualCache) theme.ToggleVisual {
	request := theme.ToggleRequest{
		Variant: theme.ToggleButtonVariant(b.variant), Size: theme.ToggleButtonSize(b.size),
		IconOnly: b.iconOnly, Foreground: foreground, Override: override,
	}
	if override == nil && len(caches) > 0 && caches[0] != nil {
		return caches[0].Resolve(activeTheme, request)
	}
	return resolveToggleProvider(activeTheme, request)
}
