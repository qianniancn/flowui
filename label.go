package flowui

import "gioui.org/layout"

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

func (l LabelWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	disabled := l.disabled || !gtx.Enabled()
	content := func(gtx layout.Context) layout.Dimensions {
		return l.layoutContent(ctx, gtx, labelStyleFor(ctx.Theme, ctx.foregroundColor(), disabled, l.invalid))
	}
	if l.forKey == "" {
		return content(gtx)
	}

	fieldKey := ctx.fullKey(l.forKey)
	ctx.registerFieldLabel(fieldKey, l.text)
	_, clickable := ctx.clickableWithKey(l.forKey + ":label")
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			ctx.requestFieldFocus(fieldKey)
		}
	}
	return clickable.Layout(gtx, content)
}
