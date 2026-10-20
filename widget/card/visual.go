package card

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type cardVisualCache struct {
	cache computed.ProviderCache[Variant, theme.CardVisual]
}

func (c Widget) visualWithCache(ctx *core.Context, override *theme.CardOverride) theme.CardVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	request := theme.CardRequest{Variant: c.variant, Override: override}
	if override != nil || c.visualCache == nil {
		return resolveCardVisual(active, request)
	}
	return c.visualCache.cache.Resolve(active, c.variant, func() theme.CardVisual {
		return resolveCardVisual(active, request)
	})
}

func resolveCardVisual(active *theme.Theme, request theme.CardRequest) theme.CardVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Card, theme.ValidateCardVisual)
}
