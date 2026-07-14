package ui

import "github.com/qianniancn/FlowUI/internal/components/layout"

type ViewSize = layoutui.ViewSize
type AdaptiveWidget = layoutui.AdaptiveWidget
type AspectRatioWidget = layoutui.AspectRatioWidget
type BoxWidget = layoutui.BoxWidget
type Overflow = layoutui.Overflow
type Align = layoutui.Align
type SpacerWidget = layoutui.SpacerWidget
type DividerWidget = layoutui.DividerWidget
type SeparatorWidget = layoutui.SeparatorWidget
type CenterWidget = layoutui.CenterWidget
type ColumnWidget = layoutui.ColumnWidget
type RowWidget = layoutui.RowWidget
type FlexWidget = layoutui.FlexWidget
type GridWidget = layoutui.GridWidget
type KeyWidget = layoutui.KeyWidget
type ListWidget = layoutui.ListWidget
type ScrollWidget = layoutui.ScrollWidget
type ScrollbarWidget = layoutui.ScrollbarWidget
type SplitPaneWidget = layoutui.SplitPaneWidget
type SplitPaneOrientation = layoutui.SplitPaneOrientation
type StackWidget = layoutui.StackWidget
type StackLayer = layoutui.StackLayer
type WrapWidget = layoutui.WrapWidget

const (
	OverflowVisible = layoutui.OverflowVisible
	OverflowHidden  = layoutui.OverflowHidden

	AlignTopStart    = layoutui.AlignTopStart
	AlignTop         = layoutui.AlignTop
	AlignTopEnd      = layoutui.AlignTopEnd
	AlignStart       = layoutui.AlignStart
	AlignCenter      = layoutui.AlignCenter
	AlignEnd         = layoutui.AlignEnd
	AlignBottomStart = layoutui.AlignBottomStart
	AlignBottom      = layoutui.AlignBottom
	AlignBottomEnd   = layoutui.AlignBottomEnd

	SplitPaneHorizontal = layoutui.SplitPaneHorizontal
	SplitPaneVertical   = layoutui.SplitPaneVertical
)

func Adaptive(view func(ViewSize) Widget) AdaptiveWidget {
	return layoutui.Adaptive(view)
}

func AspectRatio(ratio float32, child Widget) AspectRatioWidget {
	return layoutui.AspectRatio(ratio, child)
}

func Box(child Widget) BoxWidget {
	return layoutui.Box(child)
}

func Spacer(width, height int) SpacerWidget {
	return layoutui.Spacer(width, height)
}

func Divider() DividerWidget {
	return layoutui.Divider()
}

func Separator() SeparatorWidget {
	return layoutui.Separator()
}

func Center(child Widget) CenterWidget {
	return layoutui.Center(child)
}

func Column(children ...Widget) ColumnWidget {
	return layoutui.Column(children...)
}

func Row(children ...Widget) RowWidget {
	return layoutui.Row(children...)
}

func Expanded(child Widget) FlexWidget {
	return layoutui.Expanded(child)
}

func Flexible(weight float32, child Widget) FlexWidget {
	return layoutui.Flexible(weight, child)
}

func Grid(columns int, children ...Widget) GridWidget {
	return layoutui.Grid(columns, children...)
}

func AutoGrid(minColumnWidth int, children ...Widget) GridWidget {
	return layoutui.AutoGrid(minColumnWidth, children...)
}

func Key(key string, child Widget) KeyWidget {
	return layoutui.Key(key, child)
}

func List(key string, count int, item func(int) Widget) ListWidget {
	return layoutui.List(key, count, item)
}

func Scroll(key string, child Widget) ScrollWidget {
	return layoutui.Scroll(key, child)
}

func Scrollbar(key string, child Widget) ScrollbarWidget {
	return layoutui.Scrollbar(key, child)
}

func SplitPane(key string, first, second Widget) SplitPaneWidget {
	return layoutui.SplitPane(key, first, second)
}

func Stack(layers ...StackLayer) StackWidget {
	return layoutui.Stack(layers...)
}

func Stacked(child Widget) StackLayer {
	return layoutui.Stacked(child)
}

func Overlay(child Widget) StackLayer {
	return layoutui.Overlay(child)
}

func Wrap(children ...Widget) WrapWidget {
	return layoutui.Wrap(children...)
}
