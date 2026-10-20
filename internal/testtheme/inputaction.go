package testtheme

import (
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) InputAction(_ *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	return theme.InputActionVisual{
		Declaration: style.Join(
			style.Style{}.
				MinWidth(24).
				Height(24).
				PaddingX(4).
				TextColor(style.Solid(muted)).
				Cursor(pointer.CursorPointer).
				On(style.Disabled, style.Style{}.Cursor(pointer.CursorDefault)),
			theme.InputActionOverrideStyle(request),
		),
		Metrics: theme.InputActionMetrics{SlotPadding: 4},
	}
}
