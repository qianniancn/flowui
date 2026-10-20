package ui

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
)

// Widget is a FlowUI node that can lay itself out with Gio.
type Widget = core.Widget

// Measurable is an optional side-effect-free measurement contract for custom
// widgets used in intrinsic sizing passes. Implementations must not inspect
// or consume frame input.
type Measurable = core.Measurable

// WidgetFunc adapts a layout function to Widget.
type WidgetFunc = core.WidgetFunc

// MeasureWidget measures a widget without allowing it to consume frame input.
func MeasureWidget(ctx *Context, gtx layout.Context, value Widget) layout.Dimensions {
	return core.MeasureWidget(ctx, gtx, value)
}

// View renders a model to a widget tree. The model arrives by value, so View
// cannot rebind it, but a struct copy still shares the backing array of a slice,
// the buckets of a map, and the target of a pointer: View must treat those as
// read-only too. Event callbacks report intent through Send, which keeps Update
// the only model mutation boundary.
//
// View may call Send, but each call queues a message and schedules another
// frame; dispatching unconditionally during render is a render loop. Dispatch
// from an event callback instead, or from Update.
//
// The contract deliberately mixes schools: the model is read-only here (Elm's
// view), Update mutates in place (not an Elm pure reducer), and messages are
// dispatched imperatively rather than returned as a description of the tree
// (TCA/Redux). See "The View Contract" in docs/architecture.md for why.
type View[M any, Msg any] func(*Context, M, Send[Msg]) Widget
