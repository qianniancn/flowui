package surface

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type SurfaceVariant uint8

const (
	SurfaceDefault SurfaceVariant = iota
	SurfaceSecondary
	SurfaceTertiary
	SurfaceTransparent
)

// SurfaceWidget provides a semantic, theme-aware background for non-overlay content.
type SurfaceWidget struct {
	child   frame.Widget
	variant SurfaceVariant
	radius  unit.Dp
	shadow  bool
}

func Surface(child frame.Widget) SurfaceWidget {
	return SurfaceWidget{child: child}
}

func (s SurfaceWidget) Variant(variant SurfaceVariant) SurfaceWidget {
	s.variant = variant
	return s
}

func (s SurfaceWidget) Radius(dp int) SurfaceWidget {
	s.radius = unit.Dp(max(dp, 0))
	return s
}

func (s SurfaceWidget) Shadow(enabled bool) SurfaceWidget {
	s.shadow = enabled
	return s
}

func (s SurfaceWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return s.layout(ctx, gtx, surfaceStyleFor(frame.ActiveTheme(ctx), s.variant))
}
