package description

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
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

func (d DescriptionWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if d.forKey != "" {
		d.registerFieldAssociation(ctx)
	}
	label := material.Label(frame.ActiveTheme(ctx).Material, frame.ActiveTheme(ctx).Components.Description.TextSize, d.text)
	label.Color = descriptionStyleFor(frame.ActiveTheme(ctx), d.disabled || !gtx.Enabled()).text
	label.WrapPolicy = text.WrapHeuristically
	return label.Layout(gtx)
}

func (d DescriptionWidget) registerFieldAssociation(ctx *frame.Context) {
	frame.RegisterFieldDescription(ctx, frame.FullKey(ctx, d.forKey), d.text)
}

func (d DescriptionWidget) prepareFieldAssociation(ctx *frame.Context) {
	frame.PrepareFieldDescription(ctx, frame.FullKey(ctx, d.forKey), d.text)
}

// PrepareFieldAssociation registers a DescriptionWidget before its layout
// order is evaluated by an internal container.
func PrepareFieldAssociation(ctx *frame.Context, widget frame.Widget) bool {
	switch description := widget.(type) {
	case DescriptionWidget:
		if description.forKey == "" {
			return false
		}
		description.prepareFieldAssociation(ctx)
		return true
	case *DescriptionWidget:
		if description == nil || description.forKey == "" {
			return false
		}
		description.prepareFieldAssociation(ctx)
		return true
	default:
		return false
	}
}
