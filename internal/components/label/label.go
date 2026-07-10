package label

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// LabelWidget identifies and describes a form field.
type LabelWidget struct {
	text     string
	forKey   string
	required bool
	disabled bool
	invalid  bool
}

// Label creates a form label.
func Label(text string) LabelWidget {
	return LabelWidget{text: text}
}

// For associates the label with a keyed form control.
func (l LabelWidget) For(key string) LabelWidget {
	l.forKey = key
	return l
}

func (l LabelWidget) Required(required bool) LabelWidget {
	l.required = required
	return l
}

func (l LabelWidget) Disabled(disabled bool) LabelWidget {
	l.disabled = disabled
	return l
}

func (l LabelWidget) Invalid(invalid bool) LabelWidget {
	l.invalid = invalid
	return l
}

func (l LabelWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	disabled := l.disabled || !gtx.Enabled()
	content := func(gtx layout.Context) layout.Dimensions {
		return l.layoutContent(ctx, gtx, labelStyleFor(frame.ActiveTheme(ctx), ctx.ForegroundColor(), disabled, l.invalid))
	}
	if l.forKey == "" {
		return content(gtx)
	}

	fieldKey := l.registerFieldAssociation(ctx)
	_, clickable := frame.DerivedClickableWithKey(ctx, l.forKey, "label")
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			frame.RequestFieldFocus(ctx, fieldKey)
		}
	}
	return clickable.Layout(gtx, content)
}

func (l LabelWidget) registerFieldAssociation(ctx *frame.Context) string {
	fieldKey := frame.FullKey(ctx, l.forKey)
	frame.RegisterFieldLabel(ctx, fieldKey, l.text)
	return fieldKey
}

func (l LabelWidget) prepareFieldAssociation(ctx *frame.Context) {
	frame.PrepareFieldLabel(ctx, frame.FullKey(ctx, l.forKey), l.text)
}

// PrepareFieldAssociation registers a LabelWidget before its layout order is
// evaluated by an internal container.
func PrepareFieldAssociation(ctx *frame.Context, widget frame.Widget) bool {
	switch label := widget.(type) {
	case LabelWidget:
		if label.forKey == "" {
			return false
		}
		label.prepareFieldAssociation(ctx)
		return true
	case *LabelWidget:
		if label == nil || label.forKey == "" {
			return false
		}
		label.prepareFieldAssociation(ctx)
		return true
	default:
		return false
	}
}
