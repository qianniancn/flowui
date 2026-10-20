package layout

import "github.com/qianniancn/flowui/internal/core"

// Context is the per-frame context passed to FlowUI widgets.
//
// It aliases the same context exposed by ui.Context so custom layout code can
// stay within the public API contract.
type Context = core.Context

// Widget is a composable FlowUI node.
//
// It aliases the same widget contract exposed by ui.Widget.
type Widget = core.Widget

// WidgetFunc adapts a layout function to Widget.
type WidgetFunc = core.WidgetFunc
