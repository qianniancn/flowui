package radiogroup

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// VisualCache retains provider-resolved radio visuals for repeated layouts.
// Instance overrides bypass this cache because they belong to one widget.
type VisualCache struct {
	cache computed.ProviderCacheMap[Variant, theme.RadioGroupVisual]
}

// VisualCacheStats exposes provider-cache behavior to tests and benchmarks.
type VisualCacheStats = computed.ProviderCacheStats

// Stats returns a snapshot of cache activity.
func (c *VisualCache) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	return c.cache.Stats()
}

// Resolve returns the cached visual for the active theme and variant.
func (c *VisualCache) Resolve(activeTheme *theme.Theme, variant Variant) theme.RadioGroupVisual {
	if c == nil {
		return resolveRadioVisual(activeTheme, variant, nil)
	}
	return c.cache.Resolve(activeTheme, variant, func() theme.RadioGroupVisual {
		return resolveRadioVisual(activeTheme, variant, nil)
	})
}

type radioStyle struct {
	indicatorOff style.ResolvedStyle
	indicatorOn  style.ResolvedStyle
	label        style.ResolvedStyle
	description  style.ResolvedStyle
	selected     float32
	focus        float32
	pressed      bool
}

func (r Widget) resolveItemStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, visual theme.RadioGroupVisual) radioStyle {
	defaults := visual.Defaults
	offState := state
	offState.Checked, offState.Selected = false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	off := computed.ResolvePartStatic(ctx, style.PartIndicator, offState, defaults, style.Style{}, style.Style{}, visual.Override)
	on := computed.ResolvePartStatic(ctx, style.PartIndicator, onState, defaults, style.Style{}, style.Style{}, visual.Override)
	label := computed.ResolvePartStatic(ctx, style.PartLabel, state, defaults, style.Style{}, style.Style{}, visual.Override)
	description := computed.ResolvePartStatic(ctx, style.PartDescription, state, defaults, style.Style{}, style.Style{}, visual.Override)
	off = computed.ApplyPartTransitions(ctx, gtx, core.DerivedKey(ctx, key, "off"), style.PartIndicator, off)
	on = computed.ApplyPartTransitions(ctx, gtx, core.DerivedKey(ctx, key, "on"), style.PartIndicator, on)
	label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label)
	description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, description)
	return radioStyle{
		indicatorOff: off, indicatorOn: on,
		label: label, description: description,
		pressed: state.Pressed,
	}
}

func radioVisualCached(activeTheme *theme.Theme, variant Variant, override *theme.RadioGroupOverride, caches ...*VisualCache) theme.RadioGroupVisual {
	if override == nil && len(caches) > 0 && caches[0] != nil {
		return caches[0].Resolve(activeTheme, variant)
	}
	return resolveRadioVisual(activeTheme, variant, override)
}

func resolveRadioVisual(activeTheme *theme.Theme, variant Variant, override *theme.RadioGroupOverride) theme.RadioGroupVisual {
	return theme.ResolveVisual(activeTheme, theme.RadioGroupRequest{Variant: variant, Override: override},
		theme.ComponentVisualProvider.RadioGroup, theme.ValidateRadioGroupVisual)
}
