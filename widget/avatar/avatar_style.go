package avatar

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type avatarResolvedStyle struct{ root, label, icon style.ResolvedStyle }

type avatarVisualKey struct {
	color   Color
	variant Variant
	size    Size
}

type avatarVisualCache struct {
	cache computed.ProviderCache[avatarVisualKey, theme.AvatarVisual]
}

type AvatarVisualCacheStats = computed.ProviderCacheStats

func (c *avatarVisualCache) Stats() AvatarVisualCacheStats {
	if c == nil {
		return AvatarVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (a Widget) visualWithCache(ctx *core.Context, override *theme.AvatarOverride) theme.AvatarVisual {
	active := core.ActiveTheme(ctx)
	request := theme.AvatarRequest{
		Color: a.color, Variant: a.variant, Size: a.size, Override: override,
	}
	if override != nil || a.visualCache == nil {
		return resolveAvatarVisual(active, request)
	}
	key := avatarVisualKey{color: a.color, variant: a.variant, size: a.size}
	return a.visualCache.cache.Resolve(active, key, func() theme.AvatarVisual {
		return resolveAvatarVisual(active, request)
	})
}

func resolveAvatarVisual(active *theme.Theme, request theme.AvatarRequest) theme.AvatarVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Avatar, theme.ValidateAvatarVisual)
}

func (a Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual theme.AvatarVisual) avatarResolvedStyle {
	state := style.State{}
	resolved := avatarResolvedStyle{
		root:  computed.ResolveStatic(ctx, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		label: computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		icon:  computed.ResolvePartStatic(ctx, style.PartIcon, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
	}
	if !computed.HasTransitions(resolved.root, resolved.label, resolved.icon) {
		return resolved
	}
	key := core.ClaimKey(ctx, core.KindStyle, "avatar")
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, resolved.label)
	resolved.icon = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIcon, resolved.icon)
	return resolved
}
