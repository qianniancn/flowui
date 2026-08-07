package table

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/checkbox"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/components/spinner"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
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
	if t.striped {
		style.body = activeTheme.Palette.Surface
	}

	padding := 0
	if t.variant == VariantPrimary {
		padding = gtx.Dp(tokens.RootPadding)
	}
	viewportWidth := max(gtx.Constraints.Max.X-padding*2, 0)
	columns := t.resolveColumns(ctx, gtx, stateValue, viewportWidth)
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
	contentDims, contentPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutHorizontal(ctx, innerGtx, stateValue, columns, style)
	})
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
	radius := tableRootRadius(gtx, activeTheme, size, t.variant, t.usesUnifiedFrame())
	drawTableRoot(gtx, size, radius, style.root)
	root := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	contentPlacement.PlaceOffset(image.Pt(padding, 0))
	contentPlacement.ClipTo(image.Rectangle{Max: image.Pt(viewportWidth, contentDims.Size.Y)})
	offset := op.Offset(image.Pt(padding, 0)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	if t.footer != nil {
		footerOffset := op.Offset(image.Pt(0, contentDims.Size.Y)).Push(gtx.Ops)
		footerCall.Add(gtx.Ops)
		footerOffset.Pop()
	}
	offset.Pop()
	root.Pop()
	if t.bordered {
		width := max(gtx.Dp(tokens.SeparatorWidth), 1)
		drawTableBorder(gtx, size, radius, width, style.border)
	}
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutHorizontal(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	stateValue.horizontal.Axis = layout.Horizontal
	stateValue.horizontal.Gap = 0
	stateValue.horizontal.Alignment = layout.Start
	stateValue.horizontal.ScrollAnyAxis = false
	return layoutui.LayoutTrackedScrollbarWithVisualOutset(ctx, gtx, &stateValue.horizontal, &stateValue.horizontalBar, &stateValue.horizontalVisualOutset, 1, t.disabled, false, func(gtx layout.Context, _ int) layout.Dimensions {
		gtx.Constraints.Min.X = columns.width
		gtx.Constraints.Max.X = columns.width
		return t.layoutContent(ctx, gtx, stateValue, columns, style)
	})
}

func (t Widget) layoutContent(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Table
	headerHeightToken := tokens.HeaderHeight
	if t.headerHeight > 0 {
		headerHeightToken = unit.Dp(t.headerHeight)
	}
	headerHeight := min(gtx.Dp(headerHeightToken), gtx.Constraints.Max.Y)
	headerGtx := gtx
	headerGtx.Constraints = layout.Exact(image.Pt(columns.width, headerHeight))
	headerDims := t.layoutHeader(ctx, headerGtx, stateValue, columns, style)

	bodyGtx := gtx
	bodyGtx.Constraints.Min.Y = 0
	bodyGtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-headerDims.Size.Y, 0)
	bodyOffset := op.Offset(image.Pt(0, headerDims.Size.Y)).Push(gtx.Ops)
	bodyDims, bodyPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutBody(ctx, bodyGtx, stateValue, columns, style)
	})
	bodyOffset.Pop()
	bodyPlacement.PlaceOffset(image.Pt(0, headerDims.Size.Y))
	size := image.Pt(columns.width, headerDims.Size.Y+bodyDims.Size.Y)
	t.drawColumnResizeGuide(ctx, gtx, stateValue, columns, size)
	return layout.Dimensions{Size: size}
}

func (t Widget) drawColumnResizeGuide(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, size image.Point) {
	tokens := frame.ActiveTheme(ctx).Components.Table
	width := max(gtx.Dp(tokens.ColumnResizerWidth), 1)
	x := columns.selection
	for index, column := range t.columns {
		x += columns.widths[index]
		if column.Resizable && stateValue.column(column.Key).resize.dragging {
			drawTableColumnResizer(gtx, x, size.Y, size.Y, width, color.NRGBA{}, frame.ActiveTheme(ctx).Palette.Accent, true, 0)
		}
	}
}

func (t Widget) layoutHeader(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle) layout.Dimensions {
	size := gtx.Constraints.Max
	radius := tableHeaderRadius(gtx, frame.ActiveTheme(ctx), size, t.variant, t.usesUnifiedFrame())
	headerSeparator := style.headerSeparator
	if !t.showsGridLines() {
		headerSeparator = color.NRGBA{}
	}
	drawTableHeader(gtx, frame.ActiveTheme(ctx), size, radius, style.header, headerSeparator)
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
			motion := frame.ActiveTheme(ctx).Motion
			selection := stateValue.selectAllSelection.Progress(gtx, all || some, motion)
			presses := state.ActivePresses(stateValue.selectAll.History())
			frame.FocusOnPress(ctx, &stateValue.selectAll, stateValue.selectAll.History(), presses)
			stateValue.selectAll.Layout(cellGtx, func(gtx layout.Context) layout.Dimensions {
				semantic.CheckBox.Add(gtx.Ops)
				semantic.LabelOp("Select all rows").Add(gtx.Ops)
				semantic.SelectedOp(all).Add(gtx.Ops)
				focusVisible := frame.FocusVisible(ctx, &stateValue.selectAll, gtx.Focused(&stateValue.selectAll))
				focus := stateValue.selectAllFocus.Opacity(gtx, focusVisible && !t.disabled, motion)
				checkbox.DrawControl(ctx, gtx, checkbox.ControlOptions{
					Variant: checkbox.CheckboxPrimary, Selection: selection,
					Indeterminate: some, Hovered: stateValue.selectAll.Hovered(),
					Pressed: stateValue.selectAll.Pressed(), Focused: focus, Disabled: t.disabled,
				})
				return layout.Dimensions{Size: cellGtx.Constraints.Max}
			})
		}
		x += columns.selection
		if len(t.columns) > 0 && t.showsGridLines() {
			drawTableHeaderSeparator(gtx, frame.ActiveTheme(ctx), x, size.Y, style.columnSeparator, t.showsFullGrid())
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
		if index < len(t.columns)-1 && !column.Resizable && t.showsGridLines() {
			drawTableHeaderSeparator(gtx, frame.ActiveTheme(ctx), x, size.Y, style.columnSeparator, t.showsFullGrid())
		}
	}
	t.layoutColumnResizers(ctx, gtx, stateValue, columns, style, size)
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutColumnResizers(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle, size image.Point) {
	tokens := frame.ActiveTheme(ctx).Components.Table
	hitSize := max(gtx.Dp(tokens.ColumnResizerHitSize), 1)
	activeWidth := max(gtx.Dp(tokens.ColumnResizerWidth), 1)
	baseHeight := max(gtx.Dp(tokens.ColumnSeparatorHeight), 1)
	if t.showsFullGrid() {
		baseHeight = size.Y
	}
	baseColor := style.columnSeparator
	if !t.showsGridLines() {
		baseColor = color.NRGBA{}
	}
	enabled := gtx.Enabled() && !t.disabled
	x := columns.selection
	for index, column := range t.columns {
		x += columns.widths[index]
		if !column.Resizable {
			continue
		}
		resize := &stateValue.column(column.Key).resize
		focusVisible := frame.FocusVisible(ctx, resize, gtx.Focused(resize))
		focus := resize.focus.Opacity(gtx, focusVisible && enabled, frame.ActiveTheme(ctx).Motion)
		drawTableColumnResizer(gtx, x, size.Y, baseHeight, activeWidth, baseColor, frame.ActiveTheme(ctx).Palette.Accent, false, focus)

		hit := image.Rect(max(x-hitSize/2, 0), 0, min(x+(hitSize+1)/2, size.X), size.Y)
		if hit.Empty() {
			continue
		}
		clipped := clip.Rect(hit).Push(gtx.Ops)
		semantic.LabelOp("Resize " + column.Label + " column").Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		if enabled {
			pointer.CursorColResize.Add(gtx.Ops)
			event.Op(gtx.Ops, resize)
		}
		clipped.Pop()
	}
}

func (t Widget) layoutHeaderCell(ctx *frame.Context, gtx layout.Context, stateValue *tableState, column Column, style tableStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Table
	content := func(gtx layout.Context, child layout.Widget) layout.Dimensions {
		return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = 0
			return layout.Inset{Left: tokens.CellPaddingX, Right: tokens.CellPaddingX}.Layout(gtx, child)
		})
	}
	if !column.Sortable {
		return content(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutHeaderContent(ctx, gtx, column, style)
		})
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
		focused := frame.FocusVisible(ctx, &columnState.clickable, gtx.Focused(&columnState.clickable))
		focus := columnState.focus.Opacity(gtx, focused && !t.disabled, frame.ActiveTheme(ctx).Motion)
		drawTableCellFocus(gtx, frame.ActiveTheme(ctx), gtx.Constraints.Max, focus, style.focus)
		return content(gtx, func(gtx layout.Context) layout.Dimensions {
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
	rowCount := t.count()
	showLoader := t.hasMore || t.loadingMore
	if rowCount == 0 && !showLoader {
		contentDims = t.layoutEmpty(ctx, gtx, columns.width, style)
	} else {
		stateValue.vertical.Axis = layout.Vertical
		stateValue.vertical.Gap = 0
		stateValue.vertical.Alignment = layout.Start
		stateValue.vertical.ScrollAnyAxis = false
		itemCount := rowCount
		if showLoader {
			itemCount++
		}
		contentDims = layoutui.LayoutTrackedScrollbarWithVisualOutset(ctx, gtx, &stateValue.vertical, &stateValue.verticalBar, &stateValue.verticalVisualOutset, itemCount, t.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
			if index == rowCount {
				return t.layoutLoadMore(ctx, gtx, columns.width)
			}
			row := t.row(index)
			return t.layoutRow(ctx, gtx, stateValue, columns, style, row, index, index == rowCount-1 && !showLoader)
		})
		stateValue.updateLoadMore(rowCount, t.hasMore, t.loadingMore, stateValue.vertical.Position.First+stateValue.vertical.Position.Count >= itemCount, t.onLoadMore)
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

func (t Widget) layoutLoadMore(ctx *frame.Context, gtx layout.Context, width int) layout.Dimensions {
	height := 1
	if t.loadingMore {
		height = max(gtx.Dp(frame.ActiveTheme(ctx).Components.Table.LoadMoreHeight), 1)
	}
	size := image.Pt(width, min(height, gtx.Constraints.Max.Y))
	loaderGtx := gtx
	loaderGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(loaderGtx, func(gtx layout.Context) layout.Dimensions {
		if !t.loadingMore {
			return layout.Dimensions{}
		}
		if t.loadMoreContent != nil {
			return t.loadMoreContent.Layout(ctx, gtx)
		}
		return spinner.Spinner().Size(spinner.SpinnerMedium).Label("Loading more rows").Layout(ctx, gtx)
	})
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
	call         op.CallOp
	dims         layout.Dimensions
	placement    frame.OverlayPlacement
	interactive  bool
	focusTargets []event.Tag
}

type tableRowTrigger func(*frame.Context, layout.Context) layout.Dimensions

func (trigger tableRowTrigger) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return trigger(ctx, gtx)
}

func (t Widget) layoutRow(ctx *frame.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, style tableStyle, row Row, index int, last bool) layout.Dimensions {
	rowState := stateValue.rowAt(row.Key, index)
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
	trigger := tableRowTrigger(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return rowState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.LabelOp(t.rowLabel(row)).Add(gtx.Ops)
			semantic.SelectedOp(selected).Add(gtx.Ops)
			semantic.EnabledOp(!disabled).Add(gtx.Ops)
			stripe := color.NRGBA{}
			if t.striped && index%2 == 1 {
				stripe = frame.ActiveTheme(ctx).Components.Table.StripeBackground
			}
			rowStyle := tableRowStyleFor(frame.ActiveTheme(ctx), t.variant, stripe, selected, rowState.clickable.Hovered() && !disabled, rowState.clickable.Pressed() && !disabled, disabled)
			motion := frame.ActiveTheme(ctx).Motion
			rowStyle.background = rowState.background.Value(animGtx, rowStyle.background, tableColorDuration, animation.EaseSmoothstep, motion)
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
			rowHeight := t.rowMinHeight(gtx, ctx)
			if columns.selection > 0 {
				selection := rowState.selection.Progress(animGtx, selected, motion)
				recorded = append(recorded, t.recordSelectionCell(ctx, gtx, columns.selection, selection))
				rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
			}
			for index, cell := range row.Cells {
				recorded = append(recorded, t.recordCell(ctx, gtx, columns.widths[index], t.columns[index], cell, rowStyle))
				rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
			}
			rowHeight = min(rowHeight, gtx.Constraints.Max.Y)
			size := image.Pt(columns.width, rowHeight)
			focused := frame.FocusVisible(ctx, &rowState.clickable, gtx.Focused(&rowState.clickable))
			focus := rowState.focus.Opacity(animGtx, focused && !disabled, motion)
			opacity := paint.PushOpacity(gtx.Ops, rowStyle.opacity)
			showSeparator := t.variant == VariantSecondary || !last
			if t.gridLinesSet {
				showSeparator = t.gridLines && !last
			}
			drawTableRow(gtx, frame.ActiveTheme(ctx), size, rowStyle, style.rowSeparator, showSeparator, focus)
			if t.showsFullGrid() {
				drawTableRowSeparators(gtx, frame.ActiveTheme(ctx), columns, rowHeight, style.columnSeparator)
			}
			rowState.interactiveCells = rowState.interactiveCells[:0]
			rowState.focusTargets = rowState.focusTargets[:0]
			x := 0
			for _, cell := range recorded {
				y := max((rowHeight-cell.dims.Size.Y)/2, 0)
				if cell.interactive {
					rowState.interactiveCells = append(rowState.interactiveCells, image.Rect(x, 0, x+cell.dims.Size.X, rowHeight))
					rowState.focusTargets = append(rowState.focusTargets, cell.focusTargets...)
				}
				cell.placement.PlaceOffset(image.Pt(x, y))
				offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
				cell.call.Add(gtx.Ops)
				offset.Pop()
				x += cell.dims.Size.X
			}
			opacity.Pop()
			return layout.Dimensions{Size: size}
		})
	})
	if t.rowContextMenu == nil {
		return trigger.Layout(ctx, gtx)
	}
	owner := frame.FullKey(ctx, t.key)
	key := frame.DerivedKey(ctx, owner, "row-context-menu:"+row.Key)
	return menu.ContextMenu(key, trigger, t.rowContextMenu(row)).
		FocusTarget(&rowState.clickable).
		FocusTargets(rowState.focusTargets...).
		Disabled(disabled).
		Layout(ctx, gtx)
}

func (t Widget) recordSelectionCell(ctx *frame.Context, gtx layout.Context, width int, selection float32) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints = layout.Exact(image.Pt(width, t.rowMinHeight(gtx, ctx)))
	checkbox.DrawControl(ctx, cellGtx, checkbox.ControlOptions{
		Variant: checkbox.CheckboxSecondary, Selection: selection,
	})
	return recordedCell{call: macro.Stop(), dims: layout.Dimensions{Size: cellGtx.Constraints.Max}}
}

func (t Widget) rowMinHeight(gtx layout.Context, ctx *frame.Context) int {
	height := frame.ActiveTheme(ctx).Components.Table.RowMinHeight
	if t.rowHeight > 0 {
		height = unit.Dp(t.rowHeight)
	}
	return gtx.Dp(height)
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
	inset := layout.Inset{
		Top: frame.ActiveTheme(ctx).Components.Table.CellPaddingY, Right: frame.ActiveTheme(ctx).Components.Table.CellPaddingX,
		Bottom: frame.ActiveTheme(ctx).Components.Table.CellPaddingY, Left: frame.ActiveTheme(ctx).Components.Table.CellPaddingX,
	}
	collector := new(frame.FocusCollector)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		restore := frame.PushFocusCollector(ctx, collector)
		defer restore()
		return layoutui.LayoutTrackedInset(ctx, cellGtx, inset, func(gtx layout.Context) layout.Dimensions {
			return alignWidget(ctx, gtx, column.Align, content)
		})
	})
	dims.Size.X = width
	return recordedCell{call: macro.Stop(), dims: dims, placement: placement, interactive: cell.Interactive, focusTargets: collector.Targets}
}

func alignWidget(ctx *frame.Context, gtx layout.Context, alignment Alignment, child frame.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	direction := layout.W
	switch alignment {
	case AlignCenter:
		direction = layout.Center
	case AlignEnd:
		direction = layout.E
	}
	return layoutui.LayoutTrackedDirection(ctx, gtx, direction, func(gtx layout.Context) layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
}

type tableCellText struct {
	text   string
	size   unit.Sp
	color  color.NRGBA
	weight font.Weight
}

func (t tableCellText) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	label := material.Label(frame.ActiveMaterial(ctx), t.size, t.text)
	label.Color = t.color
	label.Font.Weight = t.weight
	label.MaxLines = 1
	label.Truncator = "..."
	return label.Layout(gtx)
}

func (t Widget) resolveColumns(ctx *frame.Context, gtx layout.Context, stateValue *tableState, available int) tableColumns {
	tokens := frame.ActiveTheme(ctx).Components.Table
	result := tableColumns{widths: make([]int, len(t.columns))}
	fixed := make([]bool, len(t.columns))
	maxWidths := make([]int, len(t.columns))
	if stateValue != nil {
		defer func() {
			for index, column := range t.columns {
				if !column.Resizable {
					continue
				}
				configuredWidth := 0
				if column.Width > 0 {
					configuredWidth = gtx.Dp(unit.Dp(column.Width))
				}
				stateValue.column(column.Key).resize.sync(configuredWidth, result.widths[index])
			}
		}()
	}
	if t.showsSelectionIndicator() {
		result.selection = max(gtx.Dp(tokens.SelectionColumnWidth), 0)
	}
	minimum := result.selection
	lastFlex := -1
	for index, column := range t.columns {
		minWidth := max(gtx.Dp(tokens.MinColumnWidth), 1)
		if column.MinWidth > 0 {
			minWidth = max(gtx.Dp(unit.Dp(column.MinWidth)), 1)
		}
		configuredWidth := 0
		if column.Width > 0 {
			configuredWidth = gtx.Dp(unit.Dp(column.Width))
		}
		width := configuredWidth
		hasWidth := width > 0
		if column.Resizable && stateValue != nil {
			if resized, ok := stateValue.column(column.Key).resize.resolvedWidth(configuredWidth); ok {
				width, hasWidth = resized, true
			}
		}
		maxWidth := 0
		if column.MaxWidth > 0 {
			maxWidth = max(gtx.Dp(unit.Dp(column.MaxWidth)), minWidth)
		}
		maxWidths[index] = maxWidth
		if hasWidth {
			result.widths[index] = max(width, minWidth)
			if maxWidth > 0 {
				result.widths[index] = min(result.widths[index], maxWidth)
			}
			fixed[index] = true
		} else {
			result.widths[index] = minWidth
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
	remaining := extra
	for remaining > 0 && lastFlex >= 0 {
		weightRemaining := float32(0)
		lastCandidate := -1
		for index, column := range t.columns {
			if fixed[index] || maxWidths[index] > 0 && result.widths[index] >= maxWidths[index] {
				continue
			}
			weight := column.Weight
			if weight <= 0 {
				weight = 1
			}
			weightRemaining += weight
			lastCandidate = index
		}
		if lastCandidate < 0 {
			break
		}
		passRemaining := remaining
		distributed := 0
		for index, column := range t.columns {
			if fixed[index] || maxWidths[index] > 0 && result.widths[index] >= maxWidths[index] {
				continue
			}
			weight := column.Weight
			if weight <= 0 {
				weight = 1
			}
			share := passRemaining
			if index != lastCandidate {
				share = int(float32(passRemaining) * weight / weightRemaining)
			}
			if maxWidths[index] > 0 {
				share = min(share, maxWidths[index]-result.widths[index])
			}
			result.widths[index] += share
			passRemaining -= share
			remaining -= share
			distributed += share
			weightRemaining -= weight
		}
		if distributed == 0 {
			break
		}
	}
	return result
}
