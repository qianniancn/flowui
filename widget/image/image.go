package image

import (
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
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

// ImageWidget displays a reusable Gio image operation.
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
	hasRadius  bool
	override   *theme.ImageOverride
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

// Radius sets the clipping radius in dp. Without it the provider's default
// applies.
func (i Widget) Radius(dp int) Widget {
	i.radius = unit.Dp(max(dp, 0))
	i.hasRadius = true
	return i
}

// Style merges instance declarations into the image. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (i Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if i.override == nil {
		i.override = &theme.ImageOverride{}
	}
	i.override.Declaration = style.Join(i.override.Declaration, declaration)
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

func (i Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return i.layout(ctx, gtx)
}
