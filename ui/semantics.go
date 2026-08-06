package ui

import "github.com/qianniancn/flowui/internal/frame"

// SemanticRole identifies a FlowUI semantic role that Gio does not expose as
// a native class yet.
type SemanticRole = frame.SemanticRole

const (
	SemanticUnknown  = frame.SemanticUnknown
	SemanticTabList  = frame.SemanticTabList
	SemanticTab      = frame.SemanticTab
	SemanticTabPanel = frame.SemanticTabPanel
)

// SemanticNode is a framework-level semantic relationship registered by a
// component during the current frame.
type SemanticNode = frame.SemanticNode

// Semantics returns FlowUI's richer semantic registry for the current frame.
// Gio semantic operations remain available through its normal input router.
func Semantics(ctx *Context) []SemanticNode {
	return frame.Semantics(ctx)
}
