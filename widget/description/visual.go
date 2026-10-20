package description

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type descriptionVisualCache struct {
	cache computed.ProviderCache[struct{}, theme.DescriptionVisual]
}

func (d Widget) visualWithCache(
	ctx *core.Context,
	override *theme.DescriptionOverride,
) theme.DescriptionVisual {
	active := core.ActiveTheme(ctx)
	request := theme.DescriptionRequest{Override: override}
	if override != nil || d.visualCache == nil {
		return resolveDescriptionVisual(active, request)
	}
	return d.visualCache.cache.Resolve(active, struct{}{}, func() theme.DescriptionVisual {
		return resolveDescriptionVisual(active, request)
	})
}

func resolveDescriptionVisual(
	active *theme.Theme,
	request theme.DescriptionRequest,
) theme.DescriptionVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Description, theme.ValidateDescriptionVisual)
}
