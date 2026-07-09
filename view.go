package flowui

import flowview "github.com/qianniancn/FlowUI/view"

// Widget is a FlowUI node that can lay itself out with Gio.
type Widget = flowview.Widget[Context]

// View renders a model to a widget tree.
type View[M any, Msg any] func(*Context, M, Send[Msg]) Widget
