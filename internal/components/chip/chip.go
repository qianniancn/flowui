package chip

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Color selects the semantic color of a Chip.
type Color uint8

const (
	ColorDefault Color = iota
	ColorAccent
	ColorSuccess
	ColorWarning
	ColorDanger
)

// Variant selects the visual prominence of a Chip.
type Variant uint8

const (
	VariantSecondary Variant = iota
	VariantPrimary
	VariantTertiary
	VariantSoft
)

// Size selects the height and horizontal padding of a Chip.
type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
	SizeLarge
)

// Widget presents compact, non-interactive metadata or status information.
type Widget struct {
	theme        func(*theme.Theme)
	label        string
	color        Color
	variant      Variant
	size         Size
	startContent frame.Widget
	endContent   frame.Widget
}

// New creates a Chip with HeroUI's default color and secondary variant.
func New(label string) Widget {
	return Widget{label: label}
}

// Color sets the semantic color.
func (c Widget) Color(color Color) Widget {
	c.color = color
	return c
}

// Variant sets the visual prominence.
func (c Widget) Variant(variant Variant) Widget {
	c.variant = variant
	return c
}

// Size sets the size preset.
func (c Widget) Size(size Size) Widget {
	c.size = size
	return c
}

// StartContent places an icon or other compact content before the label.
func (c Widget) StartContent(content frame.Widget) Widget {
	c.startContent = content
	return c
}

// EndContent places an icon or other compact content after the label.
func (c Widget) EndContent(content frame.Widget) Widget {
	c.endContent = content
	return c
}

func (c Widget) Theme(fn func(*theme.Theme)) Widget {
	c.theme = fn
	return c
}

func (c Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, c.theme); restore != nil {
		defer restore()
	}
	return c.layout(ctx, gtx)
}
