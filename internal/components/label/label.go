package label

import (
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

// LabelWidget identifies and describes a form field.
type LabelWidget struct {
	text        string
	forKey      string
	required    bool
	disabled    bool
	invalid     bool
	customStyle flowstyle.Style
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

func (l LabelWidget) Style(value flowstyle.Style) LabelWidget {
	l.customStyle = value
	return l
}

func (l LabelWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	disabled := l.disabled || !gtx.Enabled()
	styleState := flowstyle.StyleState{Disabled: disabled, Invalid: l.invalid}
	if l.forKey == "" {
		resolved := l.resolveLayoutStyle(ctx, gtx, styleState)
		return layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return l.layoutContent(ctx, gtx, resolved, labelRequiredColor(frame.ActiveTheme(ctx), disabled))
		}))
	}

	fieldKey := l.registerFieldAssociation(ctx)
	key, clickable := frame.DerivedClickableWithKey(ctx, l.forKey, "label")
	presses := state.SnapshotPresses(clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			frame.RequestFieldFocusVisible(ctx, fieldKey, presses.ClickFocusVisible(clickable.History()))
		}
	}
	focused := gtx.Focused(clickable)
	styleState.Hovered = clickable.Hovered()
	styleState.Pressed = clickable.Pressed()
	styleState.Focused = focused
	styleState.FocusVisible = frame.FocusVisible(ctx, clickable, focused)
	resolved := l.resolveStyle(ctx, gtx, key, styleState)
	child := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return l.layoutContent(ctx, gtx, resolved, labelRequiredColor(frame.ActiveTheme(ctx), disabled))
	})
	return layoutui.LayoutInteractiveResolved(ctx, gtx, resolved, child, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return clickable.Layout(gtx, visual)
	})
}

func (l LabelWidget) resolveLayoutStyle(ctx *frame.Context, gtx layout.Context, styleState flowstyle.StyleState) flowstyle.ResolvedStyle {
	resolved := l.resolveStyleStatic(ctx, styleState)
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	key := frame.ClaimKey(ctx, state.KindStyle, "label")
	return styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
}

func (l LabelWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) flowstyle.ResolvedStyle {
	return styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		labelDefaultDeclaration(frame.ActiveTheme(ctx), ctx.ForegroundColor()),
		labelStateDeclaration(frame.ActiveTheme(ctx), ctx.ForegroundColor(), state),
		flowstyle.Style{},
		l.customStyle,
	)
}

func (l LabelWidget) resolveStyleStatic(ctx *frame.Context, state flowstyle.StyleState) flowstyle.ResolvedStyle {
	return styleruntime.ResolveStatic(
		ctx,
		state,
		labelDefaultDeclaration(frame.ActiveTheme(ctx), ctx.ForegroundColor()),
		labelStateDeclaration(frame.ActiveTheme(ctx), ctx.ForegroundColor(), state),
		flowstyle.Style{},
		l.customStyle,
	)
}

func (l LabelWidget) registerFieldAssociation(ctx *frame.Context) string {
	fieldKey := frame.FullKey(ctx, l.forKey)
	frame.RegisterFieldLabel(ctx, fieldKey, l.text)
	return fieldKey
}

// PrepareFieldAssociation registers the label before a container chooses its
// child layout order.
func (l LabelWidget) PrepareFieldAssociation(ctx *frame.Context) bool {
	if l.forKey == "" {
		return false
	}
	frame.PrepareFieldLabel(ctx, frame.FullKey(ctx, l.forKey), l.text)
	return true
}
