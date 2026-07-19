package badge

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Color selects the semantic color of a Badge.
type Color uint8

const (
	ColorDefault Color = iota
	ColorAccent
	ColorSuccess
	ColorWarning
	ColorDanger
)

// Variant selects the visual prominence of a Badge.
type Variant uint8

const (
	VariantPrimary Variant = iota
	VariantSecondary
	VariantSoft
)

// Size selects the minimum Badge dimensions and text size.
type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
	SizeLarge
)

// Placement selects the Badge corner relative to its anchor.
type Placement uint8

const (
	PlacementTopRight Placement = iota
	PlacementTopLeft
	PlacementBottomRight
	PlacementBottomLeft
)

// Widget displays a small status or count over an anchor component.
type Widget struct {
	theme     func(*theme.Theme)
	anchor    frame.Widget
	label     string
	content   frame.Widget
	alt       string
	color     Color
	variant   Variant
	size      Size
	placement Placement
}

// New creates a Badge. An empty label renders a dot indicator.
func New(anchor frame.Widget, label string) Widget {
	return Widget{anchor: anchor, label: label}
}

// Content replaces the text label with custom content.
func (b Widget) Content(content frame.Widget) Widget {
	b.content = content
	return b
}

// Alt sets the accessible label, especially for dot or icon badges.
func (b Widget) Alt(alt string) Widget {
	b.alt = alt
	return b
}

// Color sets the semantic color.
func (b Widget) Color(color Color) Widget {
	b.color = color
	return b
}

// Variant sets the visual prominence.
func (b Widget) Variant(variant Variant) Widget {
	b.variant = variant
	return b
}

// Size sets the size preset.
func (b Widget) Size(size Size) Widget {
	b.size = size
	return b
}

// Placement sets the corner relative to the anchor.
func (b Widget) Placement(placement Placement) Widget {
	b.placement = placement
	return b
}

func (b Widget) Theme(fn func(*theme.Theme)) Widget {
	b.theme = fn
	return b
}

func (b Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, b.theme); restore != nil {
		defer restore()
	}
	return b.layout(ctx, gtx)
}
