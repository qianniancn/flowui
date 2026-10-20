package statusbar

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type statusBarVisualCache struct {
	cache computed.ProviderCache[struct{}, theme.StatusBarVisual]
}

func (w Widget) visualWithCache(ctx *core.Context, override *theme.StatusBarOverride) theme.StatusBarVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	if override != nil || w.visualCache == nil {
		return resolveStatusBarVisual(active, theme.StatusBarRequest{Override: override})
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() theme.StatusBarVisual {
		return resolveStatusBarVisual(active, theme.StatusBarRequest{})
	})
}

func resolveStatusBarVisual(active *theme.Theme, request theme.StatusBarRequest) theme.StatusBarVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.StatusBar, theme.ValidateStatusBarVisual)
}
