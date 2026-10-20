package testtheme

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Table(active *theme.Theme, request theme.TableRequest) theme.TableVisual {
	tokens := active.Components.Table
	return theme.TableVisual{
		Metrics: theme.TableMetrics{
			RootPadding: tokens.RootPadding, RootRadius: tokens.RootRadius, HeaderRadius: tokens.HeaderRadius,
			BodyRadius: tokens.BodyRadius, HeaderHeight: tokens.HeaderHeight, RowMinHeight: tokens.RowMinHeight,
			EmptyHeight: tokens.EmptyHeight, MaxHeight: tokens.MaxHeight, MinColumnWidth: tokens.MinColumnWidth,
			CellPaddingX: tokens.CellPaddingX, CellPaddingY: tokens.CellPaddingY, HeaderTextSize: tokens.HeaderTextSize, HeaderTextWeight: font.Medium,
			CellTextSize: tokens.CellTextSize, SeparatorWidth: tokens.SeparatorWidth, ColumnSeparatorHeight: tokens.ColumnSeparatorHeight,
			ColumnResizerHitSize: tokens.ColumnResizerHitSize, ColumnResizerWidth: tokens.ColumnResizerWidth,
			ColumnResizeStep: tokens.ColumnResizeStep, FocusRingWidth: tokens.FocusRingWidth, FocusRadius: tokens.FocusRadius,
			SelectionColumnWidth: tokens.SelectionColumnWidth, SortIconSize: tokens.SortIconSize, SortGap: tokens.SortGap,
			FooterPaddingX: tokens.FooterPaddingX, FooterPaddingY: tokens.FooterPaddingY, LoadMoreHeight: tokens.LoadMoreHeight,
		},
		Colors: theme.TableColors{Root: background, Header: background, Body: background, Foreground: foreground, Muted: muted, Focus: focus, Stripe: background},
		Motion: active.Motion, Override: theme.TableOverrideStyle(request),
		Row: func(row theme.TableRowRequest) theme.TableRowVisual {
			value := theme.TableRowVisual{Background: row.Stripe, Foreground: foreground, Focus: focus, Opacity: 1}
			if row.Disabled {
				value.Opacity = .5
				value.Focus = color.NRGBA{}
			}
			return value
		},
		Root:      func(layout.Context, image.Point, unit.Dp, color.NRGBA) {},
		Border:    func(layout.Context, image.Point, unit.Dp, unit.Dp, color.NRGBA) {},
		Header:    func(layout.Context, image.Point, unit.Dp, color.NRGBA, color.NRGBA) {},
		Body:      func(layout.Context, image.Point, unit.Dp, color.NRGBA) {},
		RowRender: func(layout.Context, image.Point, theme.TableRowVisual, color.NRGBA, bool, float32) {},
		HeaderSep: func(layout.Context, int, int, color.NRGBA, bool) {},
		RowSep:    func(layout.Context, int, []int, int, color.NRGBA) {},
		Resizer:   func(layout.Context, int, int, int, int, color.NRGBA, color.NRGBA, bool, float32) {},
		CellFocus: func(layout.Context, image.Point, float32, color.NRGBA) {},
		Sort:      func(layout.Context, image.Point, theme.TableSortDirection, color.NRGBA) {},
		Text: func(gtx layout.Context, request theme.TableTextRequest) layout.Dimensions {
			label := material.Label(theme.MaterialOf(active), request.Size, request.Text)
			label.Color, label.Font.Weight, label.MaxLines, label.Truncator = request.Color, request.Weight, 1, "..."
			return label.Layout(gtx)
		},
	}
}
