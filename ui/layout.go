package ui

import "github.com/qianniancn/flowui/internal/components/layout"

type AspectRatioWidget = layoutui.AspectRatioWidget

type BoxWidget = layoutui.BoxWidget

// Align describes the position of a child within its available space.
type Align = layoutui.Align

type SpacerWidget = layoutui.SpacerWidget

type DividerWidget = layoutui.DividerWidget

type SeparatorWidget = layoutui.SeparatorWidget

type CenterWidget = layoutui.CenterWidget

type ColumnWidget = layoutui.ColumnWidget

type RowWidget = layoutui.RowWidget

type FlexWidget = layoutui.FlexWidget

// FlexSpacing controls how free space is distributed in a flex layout.
type FlexSpacing = layoutui.FlexSpacing

type GridWidget = layoutui.GridWidget

type KeyWidget = layoutui.KeyWidget

type ListWidget = layoutui.ListWidget

type ScrollWidget = layoutui.ScrollWidget

type ScrollbarWidget = layoutui.ScrollbarWidget

type SplitPaneWidget = layoutui.SplitPaneWidget

// SplitPaneOrientation controls whether panes are arranged horizontally or vertically.
type SplitPaneOrientation = layoutui.SplitPaneOrientation

type StackWidget = layoutui.StackWidget

// StackLayer describes one child layer in a Stack.
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

	SpaceEnd     = layoutui.SpaceEnd
	SpaceStart   = layoutui.SpaceStart
	SpaceSides   = layoutui.SpaceSides
	SpaceAround  = layoutui.SpaceAround
	SpaceBetween = layoutui.SpaceBetween
	SpaceEvenly  = layoutui.SpaceEvenly

	SplitPaneHorizontal = layoutui.SplitPaneHorizontal
	SplitPaneVertical   = layoutui.SplitPaneVertical
)

// AspectRatio constrains child to the supplied width-to-height ratio.
func AspectRatio(ratio float32, child Widget) AspectRatioWidget {
	return layoutui.AspectRatio(ratio, child)
}

// Box lays out child inside a box that applies its resolved style.
func Box(child Widget) BoxWidget {
	return layoutui.Box(child)
}

// Spacer reserves width and height in dp without drawing content.
func Spacer(width, height int) SpacerWidget {
	return layoutui.Spacer(width, height)
}

// Divider creates a horizontal rule using the theme separator color.
func Divider() DividerWidget {
	return layoutui.Divider()
}

// Separator creates a vertical rule using the theme separator color.
func Separator() SeparatorWidget {
	return layoutui.Separator()
}

// Center centers child in the available constraints.
func Center(child Widget) CenterWidget {
	return layoutui.Center(child)
}

// Column lays out children from top to bottom.
func Column(children ...Widget) ColumnWidget {
	return layoutui.Column(children...)
}

// Row lays out children from left to right.
func Row(children ...Widget) RowWidget {
	return layoutui.Row(children...)
}

// Expanded makes child take the remaining space in a Row or Column.
func Expanded(child Widget) FlexWidget {
	return layoutui.Expanded(child)
}

// Flexible assigns child a proportional share of remaining flex space.
func Flexible(weight float32, child Widget) FlexWidget {
	return layoutui.Flexible(weight, child)
}

// Grid lays out children in a fixed number of columns.
func Grid(columns int, children ...Widget) GridWidget {
	return layoutui.Grid(columns, children...)
}

// AutoGrid chooses the column count from the minimum column width.
func AutoGrid(minColumnWidth int, children ...Widget) GridWidget {
	return layoutui.AutoGrid(minColumnWidth, children...)
}

// Key scopes child state and event identities under key.
func Key(key string, child Widget) KeyWidget {
	return layoutui.Key(key, child)
}

// List lazily lays out count items and retains their scroll state by key.
func List(key string, count int, item func(int) Widget) ListWidget {
	return layoutui.List(key, count, item)
}

// Scroll makes child scrollable without drawing a visible scrollbar.
func Scroll(key string, child Widget) ScrollWidget {
	return layoutui.Scroll(key, child)
}

// Scrollbar makes child scrollable and draws a theme scrollbar.
func Scrollbar(key string, child Widget) ScrollbarWidget {
	return layoutui.Scrollbar(key, child)
}

// SplitPane lays out first and second panes separated by a draggable divider.
func SplitPane(key string, first, second Widget) SplitPaneWidget {
	return layoutui.SplitPane(key, first, second)
}

// Stack creates a local positioning context and paints layers in declaration order.
func Stack(layers ...StackLayer) StackWidget {
	return layoutui.Stack(layers...)
}

// Stacked creates a relative layer that normally contributes to the Stack's size.
func Stacked(child Widget) StackLayer {
	return layoutui.Stacked(child)
}

// Overlay creates an absolute layer positioned within its Stack.
func Overlay(child Widget) StackLayer {
	return layoutui.Overlay(child)
}

// Wrap lays out children in lines that wrap to the available width.
func Wrap(children ...Widget) WrapWidget {
	return layoutui.Wrap(children...)
}
