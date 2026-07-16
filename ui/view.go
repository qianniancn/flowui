package ui

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// Widget is a FlowUI node that can lay itself out with Gio.
type Widget = frame.Widget

// WidgetFunc adapts a layout function to Widget.
type WidgetFunc func(ctx *Context, gtx layout.Context) layout.Dimensions

func (f WidgetFunc) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	return f(ctx, gtx)
}

// View renders a model to a widget tree. It must treat the model, including
// reference-backed fields such as slices, maps, and pointers, as read-only.
// Event callbacks report intent through Send so Update remains the only model
// mutation boundary.
type View[M any, Msg any] func(*Context, M, Send[Msg]) Widget
