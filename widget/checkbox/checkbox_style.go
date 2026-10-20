package checkbox

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type checkboxResolvedStyle struct {
	visual            theme.CheckboxVisual
	indicator         checkboxIndicatorStyle
	label             style.ResolvedStyle
	description       style.ResolvedStyle
	metrics           theme.CheckboxMetrics
	colors            theme.CheckboxColors
	motion            theme.MotionTheme
	selectionDuration time.Duration
	focusDuration     time.Duration
	override          style.Style
}

// VisualCache retains provider-resolved visuals for compound widgets that
// draw many checkboxes without owning a Checkbox Widget state object.
type VisualCache struct {
	cache computed.ProviderCacheMap[Variant, theme.CheckboxVisual]
}

// VisualCacheStats makes provider-cache behavior measurable in compound
// widget benchmarks without exposing declaration internals.
type VisualCacheStats = computed.ProviderCacheStats

// Stats reports a snapshot of cache hits, misses, and theme invalidations.
func (c *VisualCache) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	return c.cache.Stats()
}

// Resolve returns the cached visual for the active theme and variant.
func (c *VisualCache) Resolve(activeTheme *theme.Theme, variant Variant) theme.CheckboxVisual {
	if c == nil {
		return resolveCheckboxVisual(activeTheme, variant, nil)
	}
	return c.cache.Resolve(activeTheme, variant, func() theme.CheckboxVisual {
		return resolveCheckboxVisual(activeTheme, variant, nil)
	})
}

type checkboxIndicatorStyle struct {
	off style.ResolvedStyle
	on  style.ResolvedStyle
}

func (c Widget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	override *theme.CheckboxOverride,
	caches ...*VisualCache,
) checkboxResolvedStyle {
	activeTheme := core.ActiveTheme(ctx)
	visual := checkboxVisualCached(activeTheme, c.variant, override, caches)
	defaults := visual.Defaults
	offState := state
	offState.Checked, offState.Selected, offState.Indeterminate = false, false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	off := computed.ResolvePartStatic(ctx, style.PartIndicator, offState, defaults, visual.Variant, style.Style{}, visual.Override)
	on := computed.ResolvePartStatic(ctx, style.PartIndicator, onState, defaults, visual.Variant, style.Style{}, visual.Override)
	label := computed.ResolvePartStatic(ctx, style.PartLabel, state, defaults, visual.Variant, style.Style{}, visual.Override)
	description := computed.ResolvePartStatic(ctx, style.PartDescription, state, defaults, visual.Variant, style.Style{}, visual.Override)
	offOwner := core.DerivedKey(ctx, key, "checkbox-indicator-off")
	onOwner := core.DerivedKey(ctx, key, "checkbox-indicator-on")
	off = computed.ApplyPartTransitions(ctx, gtx, offOwner, style.PartIndicator, off)
	on = computed.ApplyPartTransitions(ctx, gtx, onOwner, style.PartIndicator, on)
	label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label)
	description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, description)

	return checkboxResolvedStyle{
		visual:    visual,
		indicator: checkboxIndicatorStyle{off: off, on: on},
		label:     label, description: description,
		metrics: visual.Metrics, colors: visual.Colors, motion: visual.Motion,
		selectionDuration: visual.SelectionDuration, focusDuration: visual.FocusDuration,
		override: visual.Override,
	}
}

func checkboxVisualCached(
	activeTheme *theme.Theme,
	variant Variant,
	override *theme.CheckboxOverride,
	caches []*VisualCache,
) theme.CheckboxVisual {
	if override == nil && len(caches) > 0 && caches[0] != nil {
		return caches[0].Resolve(activeTheme, variant)
	}
	return resolveCheckboxVisual(activeTheme, variant, override)
}

func resolveCheckboxVisual(activeTheme *theme.Theme, variant Variant, override *theme.CheckboxOverride) theme.CheckboxVisual {
	return theme.ResolveVisual(activeTheme, theme.CheckboxRequest{Variant: variant, Override: override},
		theme.ComponentVisualProvider.Checkbox, theme.ValidateCheckboxVisual)
}
