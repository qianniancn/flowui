package badge

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type badgeResolvedStyle struct {
	root  style.ResolvedStyle
	label style.ResolvedStyle
}

type badgeVisualKey struct {
	color   Color
	variant Variant
	size    Size
}

type badgeVisualCache struct {
	cache computed.ProviderCache[badgeVisualKey, theme.BadgeVisual]
}

// BadgeVisualCacheStats exposes provider cache behavior for benchmarks and
// regression tests without exposing cache implementation details.
type BadgeVisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

func (c *badgeVisualCache) Stats() BadgeVisualCacheStats {
	if c == nil {
		return BadgeVisualCacheStats{}
	}
	stats := c.cache.Stats()
	return BadgeVisualCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func (b Widget) visualWithCache(ctx *core.Context, override *theme.BadgeOverride) theme.BadgeVisual {
	active := core.ActiveTheme(ctx)
	request := theme.BadgeRequest{
		Color: b.color, Variant: b.variant, Size: b.size,
		Override: override,
	}
	if override != nil || b.visualCache == nil {
		return resolveBadgeVisual(active, request)
	}
	key := badgeVisualKey{
		color: b.color, variant: b.variant, size: b.size,
	}
	return b.visualCache.cache.Resolve(active, key, func() theme.BadgeVisual {
		return resolveBadgeVisual(active, request)
	})
}

func resolveBadgeVisual(active *theme.Theme, request theme.BadgeRequest) theme.BadgeVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Badge, theme.ValidateBadgeVisual)
}

func (b Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual theme.BadgeVisual) badgeResolvedStyle {
	state := style.State{}
	root := computed.ResolveStatic(ctx, state, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	label := computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if !computed.HasTransitions(root, label) {
		return badgeResolvedStyle{root: root, label: label}
	}
	key := core.ClaimKey(ctx, core.KindStyle, "badge")
	return badgeResolvedStyle{
		root:  computed.ApplyTransitions(ctx, gtx, key, root),
		label: computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label),
	}
}
