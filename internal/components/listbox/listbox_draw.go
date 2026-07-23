package listbox

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawListBoxIndicator(gtx layout.Context, theme *theme.Theme, size image.Point, style listBoxItemStyle) {
	optionrow.DrawCheck(gtx, size, style.selected, theme.Components.ListBox.ItemIndicatorInset, theme.Components.ListBox.ItemIndicatorStroke, style.indicator)
}
