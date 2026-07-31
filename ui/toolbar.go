package ui

import "github.com/qianniancn/flowui/internal/components/toolbar"

type ToolbarWidget = toolbar.Widget

// ToolbarOrientation controls the direction of a toolbar.
type ToolbarOrientation = toolbar.Orientation

type ToolbarSeparatorWidget = toolbar.SeparatorWidget

const (
	ToolbarHorizontal = toolbar.Horizontal
	ToolbarVertical   = toolbar.Vertical
)

// Toolbar lays out command and control widgets in a themed bar.
func Toolbar(children ...Widget) ToolbarWidget {
	return toolbar.New(children...)
}

// ToolbarSeparator returns a separator for a toolbar.
func ToolbarSeparator() ToolbarSeparatorWidget {
	return toolbar.Separator()
}
