package layoutui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/host"
	flowlayout "github.com/qianniancn/flowui/internal/layout"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

// BoxWidget is the single-box host (implemented in package host).
type BoxWidget = host.BoxWidget

// Box returns a host for child. Geometry comes only from Style.
func Box(child frame.Widget) BoxWidget {
	return host.Box(child)
}

// LayoutStyled resolves then lays out through the host.
func LayoutStyled(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState, value flowstyle.Style, child frame.Widget) layout.Dimensions {
	return host.LayoutStyled(ctx, gtx, key, state, value, child)
}

// LayoutResolved lays out a pre-resolved style through the host.
func LayoutResolved(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle, child frame.Widget) layout.Dimensions {
	return host.LayoutResolved(ctx, gtx, resolved, child)
}

// LayoutInteractiveResolved is the interactive host entry.
func LayoutInteractiveResolved(
	ctx *frame.Context,
	gtx layout.Context,
	resolved flowstyle.ResolvedStyle,
	child frame.Widget,
	hostFn func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return host.LayoutInteractiveResolved(ctx, gtx, resolved, child, hostFn)
}

type Overflow = flowlayout.Overflow

const (
	OverflowVisible = flowlayout.OverflowVisible
	OverflowHidden  = flowlayout.OverflowHidden
)

type Align = flowlayout.Align

const (
	AlignTopStart    = flowlayout.AlignTopStart
	AlignTop         = flowlayout.AlignTop
	AlignTopEnd      = flowlayout.AlignTopEnd
	AlignStart       = flowlayout.AlignStart
	AlignCenter      = flowlayout.AlignCenter
	AlignEnd         = flowlayout.AlignEnd
	AlignBottomStart = flowlayout.AlignBottomStart
	AlignBottom      = flowlayout.AlignBottom
	AlignBottomEnd   = flowlayout.AlignBottomEnd
)

// SpacerWidget occupies fixed space.
type SpacerWidget struct {
	width  unit.Dp
	height unit.Dp
}

// Spacer returns a fixed-size empty widget.
func Spacer(width, height int) SpacerWidget {
	return SpacerWidget{width: unit.Dp(width), height: unit.Dp(height)}
}

// Layout implements frame.Widget.
func (s SpacerWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Spacer{Width: s.width, Height: s.height}.Layout(gtx)
}

func init() {
	host.SetAssociationPreparer(prepareFieldAssociations)
}
