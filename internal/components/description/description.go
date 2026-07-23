package description

import (
	"gioui.org/layout"
	"gioui.org/text"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	componenttext "github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
)

// DescriptionWidget provides supporting text for a form field or control.
type DescriptionWidget struct {
	text        string
	forKey      string
	disabled    bool
	customStyle flowstyle.Style
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

func (d DescriptionWidget) Style(value flowstyle.Style) DescriptionWidget {
	d.customStyle = value
	return d
}

func (d DescriptionWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if d.forKey != "" {
		d.registerFieldAssociation(ctx)
	}
	state := flowstyle.StyleState{Disabled: d.disabled || !gtx.Enabled()}
	resolved := styleruntime.ResolveStatic(
		ctx,
		state,
		descriptionDefaultDeclaration(frame.ActiveTheme(ctx)),
		descriptionStateDeclaration(frame.ActiveTheme(ctx), state),
		flowstyle.Style{},
		d.customStyle,
	)
	if len(resolved.Transitions) != 0 {
		var key string
		if d.forKey != "" {
			key = frame.ClaimDerivedKey(ctx, stateutil.KindStyle, d.forKey, "description")
		} else {
			key = frame.ClaimKey(ctx, stateutil.KindStyle, "description")
		}
		resolved = styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return componenttext.New(d.text).
			Wrap(text.WrapHeuristically).
			Style(flowstyle.TextDeclaration(resolved.Text)).
			Layout(ctx, gtx)
	}))
}

func (d DescriptionWidget) registerFieldAssociation(ctx *frame.Context) {
	frame.RegisterFieldDescription(ctx, frame.FullKey(ctx, d.forKey), d.text)
}

// PrepareFieldAssociation registers the description before a container
// chooses its child layout order.
func (d DescriptionWidget) PrepareFieldAssociation(ctx *frame.Context) bool {
	if d.forKey == "" {
		return false
	}
	frame.PrepareFieldDescription(ctx, frame.FullKey(ctx, d.forKey), d.text)
	return true
}
