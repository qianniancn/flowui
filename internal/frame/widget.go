package frame

import (
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

// Widget is a FlowUI node that lays itself out with the current frame context.
type Widget interface {
	Layout(ctx *Context, gtx layout.Context) layout.Dimensions
}

// Measurable is an optional side-effect-free measurement contract. Components
// that implement it can be measured without registering paint or input ops;
// Measure implementations must not inspect or consume frame input.
type Measurable interface {
	Measure(ctx *Context, gtx layout.Context) layout.Dimensions
}

// MeasureWidget measures a widget without allowing it to consume frame input.
// Measurable widgets use their dedicated method; other widgets receive a
// disabled context with an empty input source as a conservative fallback.
func MeasureWidget(ctx *Context, gtx layout.Context, value Widget) layout.Dimensions {
	if value == nil {
		return layout.Dimensions{}
	}
	measureGtx := gtx
	measureGtx.Ops = new(op.Ops)
	if measurable, ok := value.(Measurable); ok {
		return measurable.Measure(ctx, measureGtx)
	}
	measureGtx.Source = input.Source{}
	restoreHidden := PushHiddenLayout(ctx)
	defer restoreHidden()
	restoreMeasurement := PushMeasurement(ctx)
	defer restoreMeasurement()
	return value.Layout(ctx, measureGtx.Disabled())
}

// WidgetFunc adapts a function to Widget.
type WidgetFunc func(ctx *Context, gtx layout.Context) layout.Dimensions

func (f WidgetFunc) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	return f(ctx, gtx)
}
