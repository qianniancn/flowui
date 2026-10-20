package alert

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type alertVisualKey struct {
	status      Status
	override    uint64
	overrideSet bool
}

type alertVisualCache struct {
	cache computed.ProviderCache[alertVisualKey, theme.AlertVisual]
}

func (a Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.AlertOverride) layout.Dimensions {
	visual := a.visualWithCache(ctx, override)
	return a.layoutWithVisual(ctx, gtx, visual)
}

func (a Widget) visualWithCache(ctx *core.Context, override *theme.AlertOverride) theme.AlertVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := alertOverrideKey(override)
	key := alertVisualKey{status: a.status, override: hash, overrideSet: set}
	load := func() theme.AlertVisual {
		return theme.ResolveVisual(active, theme.AlertRequest{Status: a.status, Override: override},
			theme.ComponentVisualProvider.Alert, theme.ValidateAlertVisual)
	}
	if a.visualCache == nil {
		return load()
	}
	return a.visualCache.cache.Resolve(active, key, load)
}

func alertOverrideKey(override *theme.AlertOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

type alertResolvedStyle struct {
	root        style.ResolvedStyle
	title       style.ResolvedStyle
	description style.ResolvedStyle
	content     style.ResolvedStyle
	indicator   style.ResolvedStyle
}

func (a Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual theme.AlertVisual) alertResolvedStyle {
	state := style.State{}
	resolved := alertResolvedStyle{
		root:        computed.ResolveStatic(ctx, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		title:       computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Title, style.Style{}, style.Style{}, visual.Override),
		description: computed.ResolvePartStatic(ctx, style.PartDescription, state, visual.Description, style.Style{}, style.Style{}, visual.Override),
		content:     computed.ResolvePartStatic(ctx, style.PartContent, state, visual.Content, style.Style{}, style.Style{}, visual.Override),
		indicator:   computed.ResolvePartStatic(ctx, style.PartIndicator, state, visual.Indicator, style.Style{}, style.Style{}, visual.Override),
	}
	if !computed.HasTransitions(resolved.root, resolved.title, resolved.description, resolved.content, resolved.indicator) {
		return resolved
	}
	key := core.ClaimKey(ctx, core.KindStyle, "alert")
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.title = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, resolved.title)
	resolved.description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, resolved.description)
	resolved.content = computed.ApplyPartTransitions(ctx, gtx, key, style.PartContent, resolved.content)
	resolved.indicator = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIndicator, resolved.indicator)
	return resolved
}
