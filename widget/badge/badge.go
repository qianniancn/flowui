package badge

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	textui "github.com/qianniancn/flowui/widget/text"
)

// Color selects the semantic color of a Badge.
type Color = theme.BadgeColor

const (
	ColorDefault = theme.BadgeColorDefault
	ColorAccent  = theme.BadgeColorAccent
	ColorSuccess = theme.BadgeColorSuccess
	ColorWarning = theme.BadgeColorWarning
	ColorDanger  = theme.BadgeColorDanger
)

// Variant selects the visual prominence of a Badge.
type Variant = theme.BadgeVariant

const (
	VariantPrimary   = theme.BadgeVariantPrimary
	VariantSecondary = theme.BadgeVariantSecondary
	VariantSoft      = theme.BadgeVariantSoft
)

// Size selects the minimum Badge dimensions and text size.
type Size = theme.BadgeSize

const (
	SizeMedium = theme.BadgeSizeMedium
	SizeSmall  = theme.BadgeSizeSmall
	SizeLarge  = theme.BadgeSizeLarge
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
	override *theme.BadgeOverride
	anchor   core.Widget
	label    string
	// Keep the label widget alive across frames so its text Provider cache is
	// reused instead of being rebuilt for every Badge layout.
	labelWidget *textui.Widget
	content     core.Widget
	alt         string
	color       Color
	variant     Variant
	size        Size
	placement   Placement
	visualCache *badgeVisualCache
}

// New creates a Badge. An empty label renders a dot indicator.
func New(anchor core.Widget, label string) Widget {
	value := Widget{anchor: anchor, label: label, visualCache: &badgeVisualCache{}}
	if label != "" {
		text := textui.New(label)
		value.labelWidget = &text
	}
	return value
}

// Content replaces the text label with custom content.
func (b Widget) Content(content core.Widget) Widget {
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

func (b Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return b.layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.BadgeOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
