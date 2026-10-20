package chip

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type chipResolvedStyle struct {
	root          style.ResolvedStyle
	label         style.ResolvedStyle
	icon          style.ResolvedStyle
	labelPaddingX unit.Dp
	contentGap    unit.Dp
}

type chipVisualKey struct {
	color   Color
	variant Variant
	size    Size
}

type chipVisualCache struct {
	cache computed.ProviderCache[chipVisualKey, theme.ChipVisual]
}

type ChipVisualCacheStats = computed.ProviderCacheStats

func (c *chipVisualCache) Stats() ChipVisualCacheStats {
	if c == nil {
		return ChipVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (c Widget) visualWithCache(ctx *core.Context, override *theme.ChipOverride) theme.ChipVisual {
	active := core.ActiveTheme(ctx)
	request := theme.ChipRequest{Color: c.color, Variant: c.variant, Size: c.size, Override: override}
	if override != nil || c.visualCache == nil {
		return resolveChipVisual(active, request)
	}
	key := chipVisualKey{color: c.color, variant: c.variant, size: c.size}
	return c.visualCache.cache.Resolve(active, key, func() theme.ChipVisual {
		return resolveChipVisual(active, request)
	})
}

func resolveChipVisual(active *theme.Theme, request theme.ChipRequest) theme.ChipVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Chip, theme.ValidateChipVisual)
}

func (c Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual theme.ChipVisual) chipResolvedStyle {
	state := style.State{}
	resolved := chipResolvedStyle{
		root:          computed.ResolveStatic(ctx, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		label:         computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		icon:          computed.ResolvePartStatic(ctx, style.PartIcon, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		labelPaddingX: visual.Metrics.LabelPaddingX,
		contentGap:    visual.Metrics.ContentGap,
	}
	if !computed.HasTransitions(resolved.root, resolved.label, resolved.icon) {
		return resolved
	}
	key := core.ClaimKey(ctx, core.KindStyle, "chip")
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, resolved.label)
	resolved.icon = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIcon, resolved.icon)
	return resolved
}
