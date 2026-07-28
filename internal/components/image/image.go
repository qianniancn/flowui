package imageview

import (
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	flowlayout "github.com/qianniancn/flowui/internal/layout"
)

// Fit controls how an image is scaled inside its layout bounds.
type Fit uint8

const (
	FitScaleDown Fit = iota
	FitContain
	FitCover
	FitFill
	FitUnscaled
)

// Widget displays a reusable Gio image operation.
type Widget struct {
	source     paint.ImageOp
	fit        Fit
	position   flowlayout.Align
	width      unit.Dp
	height     unit.Dp
	radius     unit.Dp
	opacity    float32
	alt        string
	hasWidth   bool
	hasHeight  bool
	hasOpacity bool
}

// New creates an Image that preserves its intrinsic size and scales down when needed.
func New(source paint.ImageOp) Widget {
	return Widget{source: source, position: flowlayout.AlignCenter}
}

// Fit sets the image scaling mode.
func (i Widget) Fit(fit Fit) Widget {
	i.fit = fit
	return i
}

// Position sets the image alignment inside its bounds.
func (i Widget) Position(position flowlayout.Align) Widget {
	i.position = position
	return i
}

// Width sets a fixed width in dp.
func (i Widget) Width(dp int) Widget {
	i.width = unit.Dp(max(dp, 0))
	i.hasWidth = true
	return i
}

// Height sets a fixed height in dp.
func (i Widget) Height(dp int) Widget {
	i.height = unit.Dp(max(dp, 0))
	i.hasHeight = true
	return i
}

// Radius sets the clipping radius in dp.
func (i Widget) Radius(dp int) Widget {
	i.radius = unit.Dp(max(dp, 0))
	return i
}

// Opacity sets image opacity in the range 0 to 1.
func (i Widget) Opacity(opacity float32) Widget {
	i.opacity = min(max(opacity, 0), 1)
	i.hasOpacity = true
	return i
}

// Alt sets the accessible description.
func (i Widget) Alt(alt string) Widget {
	i.alt = alt
	return i
}

func (i Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return i.layout(ctx, gtx)
}
