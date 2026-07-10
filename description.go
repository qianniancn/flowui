package flowui

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
)

// DescriptionWidget provides supporting text for a form field or control.
type DescriptionWidget struct {
	text     string
	forKey   string
	disabled bool
}

// Description creates supporting text for a form field or control.
func Description(text string) DescriptionWidget {
	return DescriptionWidget{text: text}
}

// For associates the description with a keyed control.
func (d DescriptionWidget) For(key string) DescriptionWidget {
	d.forKey = key
	return d
}

func (d DescriptionWidget) Disabled(disabled bool) DescriptionWidget {
	d.disabled = disabled
	return d
}

func (d DescriptionWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	if d.forKey != "" {
		ctx.registerFieldDescription(ctx.fullKey(d.forKey), d.text)
	}
	label := material.Label(ctx.Theme.Material, ctx.Theme.Components.Description.TextSize, d.text)
	label.Color = descriptionStyleFor(ctx.Theme, d.disabled || !gtx.Enabled()).text
	label.WrapPolicy = text.WrapHeuristically
	return label.Layout(gtx)
}
