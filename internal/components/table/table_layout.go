package table

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/checkbox"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

type tableColumns struct {
	selection int
	widths    []int
	width     int
}

func (t Widget) layout(ctx *frame.Context, gtx layout.Context, stateValue *tableState) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Table
	style := tableStyleFor(activeTheme, t.variant)

	padding := 0
	if t.variant == VariantPrimary {
		padding = gtx.Dp(tokens.RootPadding)
	}
	viewportWidth := max(gtx.Constraints.Max.X-padding*2, 0)
	columns := t.resolveColumns(ctx, gtx, viewportWidth)
	contentMaxHeight := gtx.Constraints.Max.Y
	maxHeight := tokens.MaxHeight
	if t.maxHeight > 0 {
		maxHeight = unit.Dp(t.maxHeight)
	}
	if maxHeight > 0 {
		contentMaxHeight = min(contentMaxHeight, gtx.Dp(maxHeight))
	}

	macro := op.Record(gtx.Ops)
	innerGtx := gtx
	innerGtx.Constraints.Min.X = viewportWidth
	innerGtx.Constraints.Max.X = viewportWidth
	innerGtx.Constraints.Min.Y = 0
	innerGtx.Constraints.Max.Y = max(contentMaxHeight, 0)
	contentDims := t.layoutHorizontal(ctx, innerGtx, stateValue, columns, style)
	call := macro.Stop()

	footerHeight := 0
	var footerCall op.CallOp
	if t.footer != nil {
		footerMacro := op.Record(gtx.Ops)
		footerGtx := gtx
		footerGtx.Constraints.Min.X = viewportWidth
		footerGtx.Constraints.Max.X = viewportWidth
		footerGtx.Constraints.Min.Y = 0
		footerGtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-contentDims.Size.Y-padding, 0)
		footerBackground := style.root
		if footerBackground.A == 0 {
			footerBackground = ctx.BackgroundColor()
		}
		restore := frame.PushColors(ctx, style.foreground, footerBackground)
		footerDims := layout.Inset{Top: tokens.FooterPaddingY, Right: tokens.FooterPaddingX, Bottom: tokens.FooterPaddingY, Left: tokens.FooterPaddingX}.Layout(footerGtx, func(gtx layout.Context) layout.Dimensions {
			return t.footer.Layout(ctx, gtx)
		})
		restore()
		footerHeight = footerDims.Size.Y
		footerCall = footerMacro.Stop()
	}

	size := image.Pt(gtx.Constraints.Max.X, contentDims.Size.Y+footerHeight+padding)
	size = gtx.Constraints.Constrain(size)
	radius := tableRootRadius(gtx, activeTheme, size, t.variant)
	drawTableRoot(gtx, size, radius, style.root)
	root := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	offset := op.Offset(image.Pt(padding, 0)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	if t.footer != nil {
		footerOffset := op.Offset(image.Pt(0, contentDims.Size.Y)).Push(gtx.Ops)
		footerCall.Add(gtx.Ops)
		footerOffset.Pop()
	}
	offset.Pop()
	root.Pop()
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutHorizontal(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	stateValue.horizontal.Axis = layout.Horizontal
	stateValue.horizontal.Gap = 0
	stateValue.horizontal.Alignment = layout.Start
	stateValue.horizontal.ScrollAnyAxis = false
	return stateValue.horizontal.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		gtx.Constraints.Min.X = columns.width
		gtx.Constraints.Max.X = columns.width
		return t.layoutContent(ctx, gtx, stateValue, columns, style)
	})
}

func (t Widget) layoutContent(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Table
	headerHeight := min(gtx.Dp(tokens.HeaderHeight), gtx.Constraints.Max.Y)
	headerGtx := gtx
	headerGtx.Constraints = layout.Exact(image.Pt(columns.width, headerHeight))
	headerDims := t.layoutHeader(ctx, headerGtx, stateValue, columns, style)

	bodyGtx := gtx
	bodyGtx.Constraints.Min.Y = 0
	bodyGtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-headerDims.Size.Y, 0)
	bodyOffset := op.Offset(image.Pt(0, headerDims.Size.Y)).Push(gtx.Ops)
	bodyDims := t.layoutBody(ctx, bodyGtx, stateValue, columns, style)
	bodyOffset.Pop()
	return layout.Dimensions{Size: image.Pt(columns.width, headerDims.Size.Y+bodyDims.Size.Y)}
}

func (t Widget) layoutHeader(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	size := gtx.Constraints.Max
	radius := tableHeaderRadius(gtx, frame.ActiveTheme(ctx), size, t.variant)
	drawTableHeader(gtx, size, radius, style.header)
	headerBackground := style.header
	if headerBackground.A == 0 {
		headerBackground = ctx.BackgroundColor()
	}
	restore := frame.PushColors(ctx, style.muted, headerBackground)
	defer restore()
	x := 0
	if columns.selection > 0 {
		cellGtx := gtx
		cellGtx.Constraints = layout.Exact(image.Pt(columns.selection, size.Y))
		if t.selectionMode == SelectionMultiple {
			all, some := t.selectionSummary()
			presses := state.ActivePresses(stateValue.selectAll.History())
			frame.FocusOnPress(ctx, &stateValue.selectAll, stateValue.selectAll.History(), presses)
			stateValue.selectAll.Layout(cellGtx, func(gtx layout.Context) layout.Dimensions {
				semantic.CheckBox.Add(gtx.Ops)
				semantic.LabelOp("Select all rows").Add(gtx.Ops)
				semantic.SelectedOp(all).Add(gtx.Ops)
				focusVisible := stateValue.selectAllFocus.Visible(gtx.Focused(&stateValue.selectAll), stateValue.selectAll.History())
				focus := stateValue.selectAllFocus.Opacity(gtx, focusVisible && !t.disabled)
				checkbox.DrawControl(ctx, gtx, checkbox.ControlOptions{
					Variant: checkbox.CheckboxPrimary, Selection: boolFloat(all || some),
					Indeterminate: some, Hovered: stateValue.selectAll.Hovered(),
					Pressed: stateValue.selectAll.Pressed(), Focused: focus, Disabled: t.disabled,
				})
				return layout.Dimensions{Size: cellGtx.Constraints.Max}
			})
		}
		x += columns.selection
		if len(t.columns) > 0 {
			drawTableHeaderSeparator(gtx, frame.ActiveTheme(ctx), x, size.Y, style.separator)
		}
	}
	for index, column := range t.columns {
		width := columns.widths[index]
		cellGtx := gtx
		cellGtx.Constraints = layout.Exact(image.Pt(width, size.Y))
		offset := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
		t.layoutHeaderCell(ctx, cellGtx, stateValue, column, style)
		offset.Pop()
		x += width
		if index < len(t.columns)-1 {
			drawTableHeaderSeparator(gtx, frame.ActiveTheme(ctx), x, size.Y, style.separator)
		}
	}
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutHeaderCell(ctx *frame.Context, gtx layout.Context, stateValue *tableState, column Column, style tableStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Table
	content := func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: tokens.CellPaddingX, Right: tokens.CellPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutHeaderContent(ctx, gtx, column, style)
		})
	}
	if !column.Sortable {
		return content(gtx)
	}
	columnState := stateValue.column(column.Key)
	presses := state.ActivePresses(columnState.clickable.History())
	frame.FocusOnPress(ctx, &columnState.clickable, columnState.clickable.History(), presses)
	return columnState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(column.Label).Add(gtx.Ops)
		semantic.DescriptionOp("Sort column").Add(gtx.Ops)
		foreground := style.muted
		if columnState.clickable.Hovered() || t.sort.Column == column.Key {
			foreground = style.foreground
		}
		focused := columnState.focus.Visible(gtx.Focused(&columnState.clickable), columnState.clickable.History())
		focus := columnState.focus.Opacity(gtx, focused && !t.disabled)
		drawTableCellFocus(gtx, frame.ActiveTheme(ctx), gtx.Constraints.Max, focus, style.focus)
		return layout.Inset{Left: tokens.CellPaddingX, Right: tokens.CellPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutSortableHeaderContent(ctx, gtx, column, foreground)
		})
	})
}

func (t Widget) layoutHeaderContent(ctx *frame.Context, gtx layout.Context, column Column, style tableStyle) layout.Dimensions {
	child := column.Header
	if child == nil {
		child = tableCellText{
			text: column.Label, size: frame.ActiveTheme(ctx).Components.Table.HeaderTextSize,
			color: style.muted, weight: font.Medium,
		}
	}
	return alignWidget(ctx, gtx, column.Align, child)
}

func (t Widget) layoutSortableHeaderContent(ctx *frame.Context, gtx layout.Context, column Column, foreground color.NRGBA) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Table
	label := column.Header
	if label == nil {
		label = tableCellText{text: column.Label, size: tokens.HeaderTextSize, color: foreground, weight: font.Medium}
	}
	children := []layout.FlexChild{
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return alignWidget(ctx, gtx, column.Align, label)
		}),
	}
	if t.sort.Column == column.Key {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Dp(tokens.SortIconSize), gtx.Dp(tokens.SortIconSize))
			drawTableSortIndicator(gtx, frame.ActiveTheme(ctx), size, t.sort.Direction, foreground)
			return layout.Dimensions{Size: size}
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(tokens.SortGap)}.Layout(gtx, children...)
}

func (t Widget) layoutBody(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	bodyBackground := style.body
	if bodyBackground.A == 0 {
		bodyBackground = ctx.BackgroundColor()
	}
	restore := frame.PushColors(ctx, style.foreground, bodyBackground)
	defer restore()
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	if len(t.rows) == 0 {
		contentDims = t.layoutEmpty(ctx, gtx, columns.width, style)
	} else {
		stateValue.vertical.Axis = layout.Vertical
		stateValue.vertical.Gap = 0
		stateValue.vertical.Alignment = layout.Start
		stateValue.vertical.ScrollAnyAxis = false
		contentDims = stateValue.vertical.Layout(gtx, len(t.rows), func(gtx layout.Context, index int) layout.Dimensions {
			return t.layoutRow(ctx, gtx, stateValue, columns, style, t.rows[index], index == len(t.rows)-1)
		})
	}
	call := macro.Stop()
	contentDims.Size.X = columns.width
	radius := tableBodyRadius(gtx, frame.ActiveTheme(ctx), contentDims.Size, t.variant)
	drawTableBody(gtx, contentDims.Size, radius, style.body)
	bodyClip := clip.UniformRRect(image.Rectangle{Max: contentDims.Size}, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	bodyClip.Pop()
	return contentDims
}

func (t Widget) layoutEmpty(ctx *frame.Context, gtx layout.Context, width int, style tableStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Table
	height := min(gtx.Dp(tokens.EmptyHeight), gtx.Constraints.Max.Y)
	size := image.Pt(width, max(height, 0))
	emptyGtx := gtx
	emptyGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(emptyGtx, func(gtx layout.Context) layout.Dimensions {
		if t.emptyContent != nil {
			return t.emptyContent.Layout(ctx, gtx)
		}
		return text.New(t.emptyText).Size(float32(tokens.CellTextSize)).Color(style.muted).Layout(ctx, gtx)
	})
}

type recordedCell struct {
	call op.CallOp
	dims layout.Dimensions
}

func (t Widget) layoutRow(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle, row Row, last bool) layout.Dimensions {
	rowState := stateValue.row(row.Key)
	disabled := t.rowDisabled(row)
	selected := t.isSelected(row.Key)
	animGtx := gtx
	if disabled {
		gtx = gtx.Disabled()
	} else {
		presses := state.ActivePresses(rowState.clickable.History())
		frame.FocusOnPress(ctx, &rowState.clickable, rowState.clickable.History(), presses)
	}
	gtx.Constraints.Min.X = columns.width
	gtx.Constraints.Max.X = columns.width
	return rowState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(t.rowLabel(row)).Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		rowStyle := tableRowStyleFor(frame.ActiveTheme(ctx), t.variant, selected, rowState.clickable.Hovered() && !disabled, rowState.clickable.Pressed() && !disabled, disabled)
		rowStyle.background = rowState.background.update(animGtx, rowStyle.background)
		background := rowStyle.background
		if background.A == 0 {
			background = style.body
			if background.A == 0 {
				background = ctx.BackgroundColor()
			}
		}
		restore := frame.PushColors(ctx, rowStyle.foreground, background)
		defer restore()

		recorded := make([]recordedCell, 0, len(row.Cells)+1)
		rowHeight := gtx.Dp(frame.ActiveTheme(ctx).Components.Table.RowMinHeight)
		if columns.selection > 0 {
			recorded = append(recorded, t.recordSelectionCell(ctx, gtx, columns.selection, selected))
			rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
		}
		for index, cell := range row.Cells {
			recorded = append(recorded, t.recordCell(ctx, gtx, columns.widths[index], t.columns[index], cell, rowStyle))
			rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
		}
		rowHeight = min(rowHeight, gtx.Constraints.Max.Y)
		size := image.Pt(columns.width, rowHeight)
		focused := rowState.focus.Visible(gtx.Focused(&rowState.clickable), rowState.clickable.History())
		focus := rowState.focus.Opacity(animGtx, focused && !disabled)
		opacity := paint.PushOpacity(gtx.Ops, rowStyle.opacity)
		showSeparator := t.variant == VariantSecondary || !last
		drawTableRow(gtx, frame.ActiveTheme(ctx), size, rowStyle, style.separator, showSeparator, focus)
		x := 0
		for _, cell := range recorded {
			y := max((rowHeight-cell.dims.Size.Y)/2, 0)
			offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
			cell.call.Add(gtx.Ops)
			offset.Pop()
			x += cell.dims.Size.X
		}
		opacity.Pop()
		return layout.Dimensions{Size: size}
	})
}

func (t Widget) recordSelectionCell(ctx *frame.Context, gtx layout.Context, width int, selected bool) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints = layout.Exact(image.Pt(width, gtx.Dp(frame.ActiveTheme(ctx).Components.Table.RowMinHeight)))
	checkbox.DrawControl(ctx, cellGtx, checkbox.ControlOptions{
		Variant: checkbox.CheckboxSecondary, Selection: boolFloat(selected),
	})
	return recordedCell{call: macro.Stop(), dims: layout.Dimensions{Size: cellGtx.Constraints.Max}}
}

func boolFloat(value bool) float32 {
	if value {
		return 1
	}
	return 0
}

func (t Widget) recordCell(ctx *frame.Context, gtx layout.Context, width int, column Column, cell Cell, style tableRowStyle) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints.Min = image.Point{}
	cellGtx.Constraints.Max.X = width
	content := cell.Content
	if content == nil {
		content = tableCellText{text: cell.Text, size: frame.ActiveTheme(ctx).Components.Table.CellTextSize, color: style.foreground}
	}
	dims := layout.Inset{
		Top: frame.ActiveTheme(ctx).Components.Table.CellPaddingY, Right: frame.ActiveTheme(ctx).Components.Table.CellPaddingX,
		Bottom: frame.ActiveTheme(ctx).Components.Table.CellPaddingY, Left: frame.ActiveTheme(ctx).Components.Table.CellPaddingX,
	}.Layout(cellGtx, func(gtx layout.Context) layout.Dimensions {
		return alignWidget(ctx, gtx, column.Align, content)
	})
	dims.Size.X = width
	return recordedCell{call: macro.Stop(), dims: dims}
}

func alignWidget(ctx *frame.Context, gtx layout.Context, alignment Alignment, child frame.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	switch alignment {
	case AlignCenter:
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return child.Layout(ctx, gtx) })
	case AlignEnd:
		return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return child.Layout(ctx, gtx) })
	default:
		return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return child.Layout(ctx, gtx) })
	}
}

type tableCellText struct {
	text   string
	size   unit.Sp
	color  color.NRGBA
	weight font.Weight
}

func (t tableCellText) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	label := material.Label(frame.ActiveTheme(ctx).Material, t.size, t.text)
	label.Color = t.color
	label.Font.Weight = t.weight
	label.MaxLines = 1
	label.Truncator = "..."
	return label.Layout(gtx)
}

func (t Widget) resolveColumns(ctx *frame.Context, gtx layout.Context, available int) tableColumns {
	tokens := frame.ActiveTheme(ctx).Components.Table
	result := tableColumns{widths: make([]int, len(t.columns))}
	if t.showsSelectionIndicator() {
		result.selection = max(gtx.Dp(tokens.SelectionColumnWidth), 0)
	}
	minimum := result.selection
	flexWeight := float32(0)
	lastFlex := -1
	for index, column := range t.columns {
		minWidth := max(gtx.Dp(tokens.MinColumnWidth), 1)
		if column.MinWidth > 0 {
			minWidth = max(gtx.Dp(unit.Dp(column.MinWidth)), 1)
		}
		if column.Width > 0 {
			result.widths[index] = max(gtx.Dp(unit.Dp(column.Width)), minWidth)
		} else {
			result.widths[index] = minWidth
			weight := column.Weight
			if weight <= 0 {
				weight = 1
			}
			flexWeight += weight
			lastFlex = index
		}
		minimum += result.widths[index]
	}
	result.width = max(available, minimum)
	if t.minWidth > 0 {
		result.width = max(result.width, gtx.Dp(unit.Dp(t.minWidth)))
	}
	extra := result.width - minimum
	if extra <= 0 {
		return result
	}
	if lastFlex < 0 {
		result.widths[len(result.widths)-1] += extra
		return result
	}
	remaining := extra
	for index, column := range t.columns {
		if column.Width > 0 {
			continue
		}
		weight := column.Weight
		if weight <= 0 {
			weight = 1
		}
		share := int(float32(extra) * weight / flexWeight)
		if index == lastFlex {
			share = remaining
		}
		result.widths[index] += share
		remaining -= share
	}
	return result
}
