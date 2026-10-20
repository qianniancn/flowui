package avatar

import (
	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	textui "github.com/qianniancn/flowui/widget/text"
)

// Color selects the semantic color of an Avatar fallback.
type Color = theme.AvatarColor

const (
	ColorDefault = theme.AvatarColorDefault
	ColorAccent  = theme.AvatarColorAccent
	ColorSuccess = theme.AvatarColorSuccess
	ColorWarning = theme.AvatarColorWarning
	ColorDanger  = theme.AvatarColorDanger
)

// Variant selects the visual treatment of an Avatar fallback.
type Variant = theme.AvatarVariant

const (
	VariantDefault = theme.AvatarVariantDefault
	VariantSoft    = theme.AvatarVariantSoft
)

// Size selects the Avatar diameter.
type Size = theme.AvatarSize

const (
	SizeMedium = theme.AvatarSizeMedium
	SizeSmall  = theme.AvatarSizeSmall
	SizeLarge  = theme.AvatarSizeLarge
)

// Widget displays a profile image or fallback content.
type Widget struct {
	override     *theme.AvatarOverride
	fallbackText string
	// Keep the text widget alongside the Avatar so its Provider cache survives
	// each frame; constructing it in Layout would make every fallback cold.
	fallbackWidget *textui.Widget
	fallback       core.Widget
	image          paint.ImageOp
	alt            string
	color          Color
	variant        Variant
	size           Size
	visualCache    *avatarVisualCache
}

// New creates an Avatar with optional text fallback content.
func New(fallbackText string) Widget {
	value := Widget{fallbackText: fallbackText, visualCache: &avatarVisualCache{}}
	if fallbackText != "" {
		fallback := textui.New(fallbackText)
		value.fallbackWidget = &fallback
	}
	return value
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
func (a Widget) Fallback(content core.Widget) Widget {
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

func (a Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return a.layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.AvatarOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
