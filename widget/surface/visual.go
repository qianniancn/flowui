package surface

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type surfaceVisualCache struct {
	cache computed.ProviderCache[Variant, theme.SurfaceVisual]
}

func (s Widget) layoutWithOverride(
	ctx *core.Context,
	gtx layout.Context,
	override *theme.SurfaceOverride,
) layout.Dimensions {
	visual := s.visualWithCache(ctx, override)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		visual.Override,
	)
	if len(resolved.Transitions) != 0 {
		resolved = computed.ApplyTransitions(ctx, gtx, core.ClaimKey(ctx, core.KindStyle, "surface"), resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, s.child)
}

func (s Widget) visualWithCache(ctx *core.Context, override *theme.SurfaceOverride) theme.SurfaceVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	request := theme.SurfaceRequest{Variant: s.variant, Override: override}
	if override != nil || s.visualCache == nil {
		return resolveSurfaceVisual(active, request)
	}
	return s.visualCache.cache.Resolve(active, s.variant, func() theme.SurfaceVisual {
		return resolveSurfaceVisual(active, request)
	})
}

func resolveSurfaceVisual(active *theme.Theme, request theme.SurfaceRequest) theme.SurfaceVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Surface, theme.ValidateSurfaceVisual)
}
