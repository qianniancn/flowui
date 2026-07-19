package surface

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
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
	theme         func(*theme.Theme)
	child         frame.Widget
	variant       SurfaceVariant
	radius        unit.Dp
	shadow        bool
	background    render.Brush
	foreground    color.NRGBA
	hasBackground bool
	hasForeground bool
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

func (s SurfaceWidget) Theme(fn func(*theme.Theme)) SurfaceWidget {
	s.theme = fn
	return s
}

func (s SurfaceWidget) Background(brush render.Brush) SurfaceWidget {
	s.background = brush
	s.hasBackground = true
	return s
}

func (s SurfaceWidget) Foreground(col color.NRGBA) SurfaceWidget {
	s.foreground = col
	s.hasForeground = true
	return s
}

func (s SurfaceWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, s.theme); restore != nil {
		defer restore()
	}
	style := surfaceStyleFor(frame.ActiveTheme(ctx), s.variant)
	if s.hasForeground {
		style.foreground = s.foreground
	}
	return s.layout(ctx, gtx, style)
}
