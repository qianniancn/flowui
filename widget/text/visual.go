package text

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type textVisualKey struct {
	Foreground color.NRGBA
}

type textVisualCache struct {
	cache computed.ProviderCache[textVisualKey, theme.TextVisual]
}

func (t Widget) visualWithCache(ctx *core.Context) theme.TextVisual {
	active := core.ActiveTheme(ctx)
	request := theme.TextRequest{Foreground: ctx.ForegroundColor()}
	if t.visualCache == nil {
		return resolveTextVisual(active, request)
	}
	return t.visualCache.cache.Resolve(active, textVisualKey{Foreground: request.Foreground}, func() theme.TextVisual {
		return resolveTextVisual(active, request)
	})
}

func resolveTextVisual(active *theme.Theme, request theme.TextRequest) theme.TextVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Text, theme.ValidateTextVisual)
}
