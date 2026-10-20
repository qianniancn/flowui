package chip

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	textui "github.com/qianniancn/flowui/widget/text"
)

// Color selects the semantic color of a Chip.
type Color = theme.ChipColor

const (
	ColorDefault = theme.ChipColorDefault
	ColorAccent  = theme.ChipColorAccent
	ColorSuccess = theme.ChipColorSuccess
	ColorWarning = theme.ChipColorWarning
	ColorDanger  = theme.ChipColorDanger
)

// Variant selects the visual prominence of a Chip.
type Variant = theme.ChipVariant

const (
	VariantSecondary = theme.ChipVariantSecondary
	VariantPrimary   = theme.ChipVariantPrimary
	VariantTertiary  = theme.ChipVariantTertiary
	VariantSoft      = theme.ChipVariantSoft
)

// Size selects the height and horizontal padding of a Chip.
type Size = theme.ChipSize

const (
	SizeMedium = theme.ChipSizeMedium
	SizeSmall  = theme.ChipSizeSmall
	SizeLarge  = theme.ChipSizeLarge
)

// Widget presents compact, non-interactive metadata or status information.
type Widget struct {
	override *theme.ChipOverride
	label    string
	// Keep the label widget alive across frames so its text Provider cache is
	// reused instead of being rebuilt for every Chip layout.
	labelWidget  *textui.Widget
	color        Color
	variant      Variant
	size         Size
	startContent core.Widget
	endContent   core.Widget
	visualCache  *chipVisualCache
}

// New creates a Chip with the default color and secondary variant.
func New(label string) Widget {
	text := textui.New(label)
	return Widget{label: label, labelWidget: &text, visualCache: &chipVisualCache{}}
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
func (c Widget) StartContent(content core.Widget) Widget {
	c.startContent = content
	return c
}

// EndContent places an icon or other compact content after the label.
func (c Widget) EndContent(content core.Widget) Widget {
	c.endContent = content
	return c
}

func (c Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ChipOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
