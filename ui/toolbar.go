package ui

import "github.com/qianniancn/flowui/internal/components/toolbar"

type ToolbarWidget = toolbar.Widget
type ToolbarOrientation = toolbar.Orientation
type ToolbarSeparatorWidget = toolbar.SeparatorWidget

const (
	ToolbarHorizontal = toolbar.Horizontal
	ToolbarVertical   = toolbar.Vertical
)

func Toolbar(children ...Widget) ToolbarWidget {
	return toolbar.New(children...)
}

func ToolbarSeparator() ToolbarSeparatorWidget {
	return toolbar.Separator()
}
