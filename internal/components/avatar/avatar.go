package avatar

import (
	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

// Color selects the semantic color of an Avatar fallback.
type Color uint8

const (
	ColorDefault Color = iota
	ColorAccent
	ColorSuccess
	ColorWarning
	ColorDanger
)

// Variant selects the visual treatment of an Avatar fallback.
type Variant uint8

const (
	VariantDefault Variant = iota
	VariantSoft
)

// Size selects the Avatar diameter.
type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
	SizeLarge
)

// Widget displays a profile image or fallback content.
type Widget struct {
	fallbackText string
	fallback     frame.Widget
	image        paint.ImageOp
	alt          string
	color        Color
	variant      Variant
	size         Size
	customStyle  flowstyle.Style
}

// New creates an Avatar with optional text fallback content.
func New(fallbackText string) Widget {
	return Widget{fallbackText: fallbackText}
}

// Image sets a reusable Gio image operation. An empty image shows the fallback.
func (a Widget) Image(value paint.ImageOp) Widget {
	a.image = value
	return a
}

// Alt sets the accessible description for the image.
func (a Widget) Alt(alt string) Widget {
	a.alt = alt
	return a
}

// Fallback replaces the text fallback with custom content.
func (a Widget) Fallback(content frame.Widget) Widget {
	a.fallback = content
	return a
}

// Color sets the semantic fallback color.
func (a Widget) Color(color Color) Widget {
	a.color = color
	return a
}

// Variant sets the fallback visual treatment.
func (a Widget) Variant(variant Variant) Widget {
	a.variant = variant
	return a
}

// Size sets the Avatar size preset.
func (a Widget) Size(size Size) Widget {
	a.size = size
	return a
}

func (a Widget) Style(value flowstyle.Style) Widget {
	a.customStyle = value
	return a
}

func (a Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return a.layout(ctx, gtx)
}
