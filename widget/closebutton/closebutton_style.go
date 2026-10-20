package closebutton

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type closeButtonResolvedStyle struct {
	root style.ResolvedStyle
	icon style.ResolvedStyle
}

type closeButtonVisualCache struct {
	cache computed.ProviderCache[struct{}, theme.CloseButtonVisual]
}

type CloseButtonVisualCacheStats = computed.ProviderCacheStats

func (c *closeButtonVisualCache) Stats() CloseButtonVisualCacheStats {
	if c == nil {
		return CloseButtonVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (b Widget) visualWithCache(ctx *core.Context, override *theme.CloseButtonOverride) theme.CloseButtonVisual {
	active := core.ActiveTheme(ctx)
	request := theme.CloseButtonRequest{Override: override}
	if override != nil || b.visualCache == nil {
		return resolveCloseButtonVisual(active, request)
	}
	return b.visualCache.cache.Resolve(active, struct{}{}, func() theme.CloseButtonVisual {
		return resolveCloseButtonVisual(active, request)
	})
}

func resolveCloseButtonVisual(active *theme.Theme, request theme.CloseButtonRequest) theme.CloseButtonVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.CloseButton, theme.ValidateCloseButtonVisual)
}

func (b Widget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	visual theme.CloseButtonVisual,
) closeButtonResolvedStyle {
	return closeButtonResolvedStyle{
		root: computed.Resolve(ctx, gtx, key, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		icon: computed.ResolvePart(ctx, gtx, key, style.PartIcon, state, visual.Icon, style.Style{}, style.Style{}, visual.Override),
	}
}
