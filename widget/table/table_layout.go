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
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/checkbox"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/spinner"
)

type tableColumns struct {
	selection int
	widths    []int
	width     int
}

func (t Widget) layout(ctx *core.Context, gtx layout.Context, stateValue *tableState, visual theme.TableVisual) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	metrics, colors := visual.Metrics, visual.Colors

	padding := 0
	if t.variant == VariantPrimary {
		padding = gtx.Dp(metrics.RootPadding)
	}
	viewportWidth := max(gtx.Constraints.Max.X-padding*2, 0)
	columns := t.resolveColumns(ctx, gtx, stateValue, viewportWidth, visual.Metrics)
	contentMaxHeight := gtx.Constraints.Max.Y
	if !t.flow {
		maxHeight := metrics.MaxHeight
		if t.maxHeight > 0 {
			maxHeight = unit.Dp(t.maxHeight)
		}
		if maxHeight > 0 {
			contentMaxHeight = min(contentMaxHeight, gtx.Dp(maxHeight))
		}
	}

	macro := op.Record(gtx.Ops)
	innerGtx := gtx
	innerGtx.Constraints.Min.X = viewportWidth
	innerGtx.Constraints.Max.X = viewportWidth
	innerGtx.Constraints.Min.Y = 0
	innerGtx.Constraints.Max.Y = max(contentMaxHeight, 0)
	contentDims, contentPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutHorizontal(ctx, innerGtx, stateValue, columns, visual)
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
		footerBackground := colors.Root
		if footerBackground.A == 0 {
			footerBackground = ctx.BackgroundColor()
		}
		restore := core.PushColors(ctx, colors.Foreground, footerBackground)
		footerDims := layout.Inset{Top: metrics.FooterPaddingY, Right: metrics.FooterPaddingX, Bottom: metrics.FooterPaddingY, Left: metrics.FooterPaddingX}.Layout(footerGtx, func(gtx layout.Context) layout.Dimensions {
			return t.footer.Layout(ctx, gtx)
		})
		restore()
		footerHeight = footerDims.Size.Y
		footerCall = footerMacro.Stop()
	}

	size := image.Pt(gtx.Constraints.Max.X, contentDims.Size.Y+footerHeight+padding)
	size = gtx.Constraints.Constrain(size)
	radius := tableVisualRootRadius(metrics, t.variant, t.usesUnifiedFrame())
	visual.Root(gtx, size, radius, colors.Root)
	root := clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(radius)).Push(gtx.Ops)
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
		visual.Border(gtx, size, radius, metrics.SeparatorWidth, colors.Border)
	}
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutHorizontal(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, visual theme.TableVisual) layout.Dimensions {
	if t.flow {
		// A flow table owns no viewport, so it cannot scroll sideways either;
		// content wider than the container simply overflows.
		gtx.Constraints.Min.X = columns.width
		gtx.Constraints.Max.X = columns.width
		return t.layoutContent(ctx, gtx, stateValue, columns, visual)
	}
	stateValue.horizontal.Axis = layout.Horizontal
	stateValue.horizontal.Gap = 0
	stateValue.horizontal.Alignment = layout.Start
	stateValue.horizontal.ScrollAnyAxis = false
	return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &stateValue.horizontal, &stateValue.horizontalBar, &stateValue.horizontalVisualOutset, 1, t.disabled, false, func(gtx layout.Context, _ int) layout.Dimensions {
		gtx.Constraints.Min.X = columns.width
		gtx.Constraints.Max.X = columns.width
		return t.layoutContent(ctx, gtx, stateValue, columns, visual)
	})
}

func (t Widget) layoutContent(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, visual theme.TableVisual) layout.Dimensions {
	metrics := visual.Metrics
	headerHeightToken := metrics.HeaderHeight
	if t.headerHeight > 0 {
		headerHeightToken = unit.Dp(t.headerHeight)
	}
	headerHeight := min(gtx.Dp(headerHeightToken), gtx.Constraints.Max.Y)
	headerGtx := gtx
	headerGtx.Constraints = layout.Exact(image.Pt(columns.width, headerHeight))
	headerDims := t.layoutHeader(ctx, headerGtx, stateValue, columns, visual)

	bodyGtx := gtx
	bodyGtx.Constraints.Min.Y = 0
	bodyGtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-headerDims.Size.Y, 0)
	bodyOffset := op.Offset(image.Pt(0, headerDims.Size.Y)).Push(gtx.Ops)
	bodyDims, bodyPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutBody(ctx, bodyGtx, stateValue, columns, visual)
	})
	bodyOffset.Pop()
	bodyPlacement.PlaceOffset(image.Pt(0, headerDims.Size.Y))
	size := image.Pt(columns.width, headerDims.Size.Y+bodyDims.Size.Y)
	t.drawColumnResizeGuide(ctx, gtx, stateValue, columns, size, visual)
	return layout.Dimensions{Size: size}
}

func (t Widget) drawColumnResizeGuide(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, size image.Point, visual theme.TableVisual) {
	width := max(gtx.Dp(visual.Metrics.ColumnResizerWidth), 1)
	x := columns.selection
	for index, column := range t.columns {
		x += columns.widths[index]
		if column.Resizable && stateValue.column(column.Key).resize.dragging {
			visual.Resizer(gtx, x, size.Y, size.Y, width, color.NRGBA{}, visual.Colors.ResizeAccent, true, 0)
		}
	}
}

func (t Widget) layoutHeader(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, visual theme.TableVisual) layout.Dimensions {
	size := gtx.Constraints.Max
	radius := tableVisualHeaderRadius(visual.Metrics, t.variant, t.usesUnifiedFrame())
	headerSeparator := visual.Colors.HeaderSeparator
	if !t.showsGridLines() {
		headerSeparator = color.NRGBA{}
	}
	visual.Header(gtx, size, radius, visual.Colors.Header, headerSeparator)
	headerBackground := visual.Colors.Header
	if headerBackground.A == 0 {
		headerBackground = ctx.BackgroundColor()
	}
	restore := core.PushColors(ctx, visual.Colors.Muted, headerBackground)
	defer restore()
	x := 0
	if columns.selection > 0 {
		cellGtx := gtx
		cellGtx.Constraints = layout.Exact(image.Pt(columns.selection, size.Y))
		if t.selectionMode == SelectionMultiple {
			all, some := t.selectionSummary()
			motion := visual.Motion
			selection := stateValue.selectAllSelection.Progress(gtx, all || some, motion.ControlDuration, motion)
			presses := core.ActivePresses(stateValue.selectAll.History())
			core.FocusOnPress(ctx, &stateValue.selectAll, stateValue.selectAll.History(), presses)
			stateValue.selectAll.Layout(cellGtx, func(gtx layout.Context) layout.Dimensions {
				semantic.CheckBox.Add(gtx.Ops)
				semantic.LabelOp(ctx.Text(i18n.TableSelectAllRows)).Add(gtx.Ops)
				semantic.SelectedOp(all).Add(gtx.Ops)
				focusVisible := core.FocusVisible(ctx, &stateValue.selectAll, gtx.Focused(&stateValue.selectAll))
				focus := stateValue.selectAllFocus.Opacity(gtx, focusVisible && !t.disabled, motion)
				checkbox.DrawControl(ctx, gtx, checkbox.ControlOptions{
					Variant: checkbox.Primary, Selection: selection,
					Indeterminate: some, Hovered: stateValue.selectAll.Hovered(),
					Pressed: stateValue.selectAll.Pressed(), Focused: focus, Disabled: t.disabled,
					VisualCache: &stateValue.checkboxVisual,
				})
				return layout.Dimensions{Size: cellGtx.Constraints.Max}
			})
		}
		x += columns.selection
		if len(t.columns) > 0 && t.showsGridLines() {
			visual.HeaderSep(gtx, x, size.Y, visual.Colors.ColumnSeparator, t.showsFullGrid())
		}
	}
	for index, column := range t.columns {
		width := columns.widths[index]
		cellGtx := gtx
		cellGtx.Constraints = layout.Exact(image.Pt(width, size.Y))
		offset := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
		t.layoutHeaderCell(ctx, cellGtx, stateValue, column, visual)
		offset.Pop()
		x += width
		if index < len(t.columns)-1 && !column.Resizable && t.showsGridLines() {
			visual.HeaderSep(gtx, x, size.Y, visual.Colors.ColumnSeparator, t.showsFullGrid())
		}
	}
	t.layoutColumnResizers(ctx, gtx, stateValue, columns, visual, size)
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutColumnResizers(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, visual theme.TableVisual, size image.Point) {
	metrics := visual.Metrics
	hitSize := max(gtx.Dp(metrics.ColumnResizerHitSize), 1)
	activeWidth := max(gtx.Dp(metrics.ColumnResizerWidth), 1)
	baseHeight := max(gtx.Dp(metrics.ColumnSeparatorHeight), 1)
	if t.showsFullGrid() {
		baseHeight = size.Y
	}
	baseColor := visual.Colors.ColumnSeparator
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
		focusVisible := core.FocusVisible(ctx, resize, gtx.Focused(resize))
		focus := resize.focus.Opacity(gtx, focusVisible && enabled, visual.Motion)
		visual.Resizer(gtx, x, size.Y, baseHeight, activeWidth, baseColor, visual.Colors.ResizeAccent, false, focus)

		hit := image.Rect(max(x-hitSize/2, 0), 0, min(x+(hitSize+1)/2, size.X), size.Y)
		if hit.Empty() {
			continue
		}
		clipped := clip.Rect(hit).Push(gtx.Ops)
		semantic.LabelOp(ctx.Textf(i18n.TableResizeColumn, column.Label)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		if enabled {
			pointer.CursorColResize.Add(gtx.Ops)
			event.Op(gtx.Ops, resize)
		}
		clipped.Pop()
	}
}

func (t Widget) layoutHeaderCell(ctx *core.Context, gtx layout.Context, stateValue *tableState, column Column, visual theme.TableVisual) layout.Dimensions {
	metrics := visual.Metrics
	content := func(gtx layout.Context, child layout.Widget) layout.Dimensions {
		return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = 0
			return layout.Inset{Left: metrics.CellPaddingX, Right: metrics.CellPaddingX}.Layout(gtx, child)
		})
	}
	if !column.Sortable {
		return content(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutHeaderContent(ctx, gtx, column, visual)
		})
	}
	columnState := stateValue.column(column.Key)
	presses := core.ActivePresses(columnState.clickable.History())
	core.FocusOnPress(ctx, &columnState.clickable, columnState.clickable.History(), presses)
	return columnState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(column.Label).Add(gtx.Ops)
		semantic.DescriptionOp(ctx.Text(i18n.TableSortColumn)).Add(gtx.Ops)
		foreground := visual.Colors.Muted
		if columnState.clickable.Hovered() || t.sort.Column == column.Key {
			foreground = visual.Colors.Foreground
		}
		focused := core.FocusVisible(ctx, &columnState.clickable, gtx.Focused(&columnState.clickable))
		focus := columnState.focus.Opacity(gtx, focused && !t.disabled, visual.Motion)
		visual.CellFocus(gtx, gtx.Constraints.Max, focus, visual.Colors.Focus)
		return content(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutSortableHeaderContent(ctx, gtx, column, foreground, visual)
		})
	})
}

func (t Widget) layoutHeaderContent(ctx *core.Context, gtx layout.Context, column Column, visual theme.TableVisual) layout.Dimensions {
	child := column.Header
	if child == nil {
		child = tableCellText{
			text: column.Label, size: visual.Metrics.HeaderTextSize,
			color: visual.Colors.Muted, weight: visual.Metrics.HeaderTextWeight, textRenderer: visual.Text,
		}
	}
	return alignWidget(ctx, gtx, column.Align, child)
}

func (t Widget) layoutSortableHeaderContent(ctx *core.Context, gtx layout.Context, column Column, foreground color.NRGBA, visual theme.TableVisual) layout.Dimensions {
	metrics := visual.Metrics
	label := column.Header
	if label == nil {
		label = tableCellText{text: column.Label, size: metrics.HeaderTextSize, color: foreground, weight: metrics.HeaderTextWeight, textRenderer: visual.Text}
	}
	children := []layout.FlexChild{
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return alignWidget(ctx, gtx, column.Align, label)
		}),
	}
	if t.sort.Column == column.Key {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Dp(metrics.SortIconSize), gtx.Dp(metrics.SortIconSize))
			visual.Sort(gtx, size, theme.TableSortDirection(t.sort.Direction), foreground)
			return layout.Dimensions{Size: size}
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(metrics.SortGap)}.Layout(gtx, children...)
}

func (t Widget) layoutBody(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, visual theme.TableVisual) layout.Dimensions {
	bodyBackground := visual.Colors.Body
	if bodyBackground.A == 0 {
		bodyBackground = ctx.BackgroundColor()
	}
	restore := core.PushColors(ctx, visual.Colors.Foreground, bodyBackground)
	defer restore()
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	rowCount := t.count()
	showLoader := t.hasMore || t.loadingMore
	if rowCount == 0 && !showLoader {
		contentDims = t.layoutEmpty(ctx, gtx, columns.width, visual)
	} else {
		itemCount := rowCount
		if showLoader {
			itemCount++
		}
		if t.flow {
			// Flow mode stacks every row in the page scroll, so there is no
			// virtual window to advance and no scroll position to read.
			rows := make([]core.Widget, 0, itemCount)
			for index := range itemCount {
				rows = append(rows, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
					if index == rowCount {
						return t.layoutLoadMore(ctx, gtx, columns.width, visual.Metrics)
					}
					row := t.row(index)
					return t.layoutRow(ctx, gtx, stateValue, columns, visual, row, index, index == rowCount-1 && !showLoader)
				}))
			}
			contentDims = layoutui.Column(rows...).Layout(ctx, gtx)
		} else {
			stateValue.vertical.Axis = layout.Vertical
			stateValue.vertical.Gap = 0
			stateValue.vertical.Alignment = layout.Start
			stateValue.vertical.ScrollAnyAxis = false
			contentDims = layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &stateValue.vertical, &stateValue.verticalBar, &stateValue.verticalVisualOutset, itemCount, t.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
				if index == rowCount {
					return t.layoutLoadMore(ctx, gtx, columns.width, visual.Metrics)
				}
				row := t.row(index)
				return t.layoutRow(ctx, gtx, stateValue, columns, visual, row, index, index == rowCount-1 && !showLoader)
			})
			stateValue.updateLoadMore(rowCount, t.hasMore, t.loadingMore, stateValue.vertical.Position.First+stateValue.vertical.Position.Count >= itemCount, t.onLoadMore)
		}
	}
	call := macro.Stop()
	contentDims.Size.X = columns.width
	radius := tableVisualBodyRadius(visual.Metrics, t.variant)
	visual.Body(gtx, contentDims.Size, radius, visual.Colors.Body)
	bodyClip := clip.UniformRRect(image.Rectangle{Max: contentDims.Size}, gtx.Dp(radius)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	bodyClip.Pop()
	return contentDims
}

func (t Widget) layoutLoadMore(ctx *core.Context, gtx layout.Context, width int, metrics theme.TableMetrics) layout.Dimensions {
	height := 1
	if t.loadingMore {
		height = max(gtx.Dp(metrics.LoadMoreHeight), 1)
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
		return spinner.New().Size(spinner.Medium).Label(ctx.Text(i18n.TableLoadingMore)).Layout(ctx, gtx)
	})
}

func (t Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, width int, visual theme.TableVisual) layout.Dimensions {
	metrics := visual.Metrics
	height := min(gtx.Dp(metrics.EmptyHeight), gtx.Constraints.Max.Y)
	size := image.Pt(width, max(height, 0))
	emptyGtx := gtx
	emptyGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(emptyGtx, func(gtx layout.Context) layout.Dimensions {
		if t.emptyContent != nil {
			return t.emptyContent.Layout(ctx, gtx)
		}
		return visual.Text(gtx, theme.TableTextRequest{Text: t.resolvedEmptyText(ctx), Size: metrics.CellTextSize, Color: visual.Colors.Muted})
	})
}

type recordedCell struct {
	call         op.CallOp
	dims         layout.Dimensions
	placement    core.OverlayPlacement
	interactive  bool
	focusTargets []event.Tag
}

type tableRowTrigger func(*core.Context, layout.Context) layout.Dimensions

func (trigger tableRowTrigger) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return trigger(ctx, gtx)
}

func (t Widget) layoutRow(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, visual theme.TableVisual, row Row, index int, last bool) layout.Dimensions {
	rowState := stateValue.rowAt(row.Key, index)
	disabled := t.rowDisabled(row)
	selected := t.isSelected(row.Key)
	animGtx := gtx
	if disabled {
		gtx = gtx.Disabled()
	} else {
		presses := core.ActivePresses(rowState.clickable.History())
		core.FocusOnPress(ctx, &rowState.clickable, rowState.clickable.History(), presses)
	}
	gtx.Constraints.Min.X = columns.width
	gtx.Constraints.Max.X = columns.width
	trigger := tableRowTrigger(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return rowState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.LabelOp(t.rowLabel(row)).Add(gtx.Ops)
			semantic.SelectedOp(selected).Add(gtx.Ops)
			semantic.EnabledOp(!disabled).Add(gtx.Ops)
			stripe := color.NRGBA{}
			if t.striped && index%2 == 1 {
				stripe = visual.Colors.Stripe
			}
			rowStyle := visual.Row(theme.TableRowRequest{Variant: tableThemeVariant(t.variant), Stripe: stripe, Selected: selected, Hovered: rowState.clickable.Hovered() && !disabled, Pressed: rowState.clickable.Pressed() && !disabled, Disabled: disabled})
			motion := visual.Motion
			rowStyle.Background = rowState.background.Value(
				animGtx,
				rowStyle.Background,
				motion.ControlDuration,
				animation.EasingFor(motion.Ease),
				motion,
			)
			background := rowStyle.Background
			if background.A == 0 {
				background = visual.Colors.Body
				if background.A == 0 {
					background = ctx.BackgroundColor()
				}
			}
			restore := core.PushColors(ctx, rowStyle.Foreground, background)
			defer restore()

			recorded := make([]recordedCell, 0, len(row.Cells)+1)
			rowHeight := t.rowMinHeight(gtx, ctx, visual.Metrics)
			if columns.selection > 0 {
				selection := rowState.selection.Progress(animGtx, selected, motion.ControlDuration, motion)
				recorded = append(recorded, t.recordSelectionCell(ctx, gtx, stateValue, columns.selection, selection, visual.Metrics))
				rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
			}
			for index, cell := range row.Cells {
				recorded = append(recorded, t.recordCell(ctx, gtx, columns.widths[index], t.columns[index], cell, rowStyle, visual))
				rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
			}
			rowHeight = min(rowHeight, gtx.Constraints.Max.Y)
			size := image.Pt(columns.width, rowHeight)
			focused := core.FocusVisible(ctx, &rowState.clickable, gtx.Focused(&rowState.clickable))
			focus := rowState.focus.Opacity(animGtx, focused && !disabled, motion)
			opacity := paint.PushOpacity(gtx.Ops, rowStyle.Opacity)
			showSeparator := t.variant == VariantSecondary || !last
			if t.gridLinesSet {
				showSeparator = t.gridLines && !last
			}
			visual.RowRender(gtx, size, rowStyle, visual.Colors.RowSeparator, showSeparator, focus)
			if t.showsFullGrid() {
				visual.RowSep(gtx, columns.selection, columns.widths, rowHeight, visual.Colors.ColumnSeparator)
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
	owner := core.FullKey(ctx, t.key)
	key := core.DerivedKey(ctx, owner, "row-context-menu:"+row.Key)
	return menu.ContextMenu(key, trigger, t.rowContextMenu(row)).
		FocusTarget(&rowState.clickable).
		FocusTargets(rowState.focusTargets...).
		Disabled(disabled).
		Layout(ctx, gtx)
}

func (t Widget) recordSelectionCell(ctx *core.Context, gtx layout.Context, stateValue *tableState, width int, selection float32, metrics theme.TableMetrics) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints = layout.Exact(image.Pt(width, t.rowMinHeight(gtx, ctx, metrics)))
	checkbox.DrawControl(ctx, cellGtx, checkbox.ControlOptions{
		Variant: checkbox.Secondary, Selection: selection,
		VisualCache: &stateValue.checkboxVisual,
	})
	return recordedCell{call: macro.Stop(), dims: layout.Dimensions{Size: cellGtx.Constraints.Max}}
}

func (t Widget) rowMinHeight(gtx layout.Context, ctx *core.Context, metrics theme.TableMetrics) int {
	height := metrics.RowMinHeight
	if t.rowHeight > 0 {
		height = unit.Dp(t.rowHeight)
	}
	return gtx.Dp(height)
}

func (t Widget) recordCell(ctx *core.Context, gtx layout.Context, width int, column Column, cell Cell, rowStyle theme.TableRowVisual, visual theme.TableVisual) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints.Min = image.Point{}
	cellGtx.Constraints.Max.X = width
	content := cell.Content
	if content == nil {
		content = tableCellText{text: cell.Text, size: visual.Metrics.CellTextSize, color: rowStyle.Foreground, textRenderer: visual.Text}
	}
	inset := layout.Inset{
		Top: visual.Metrics.CellPaddingY, Right: visual.Metrics.CellPaddingX,
		Bottom: visual.Metrics.CellPaddingY, Left: visual.Metrics.CellPaddingX,
	}
	collector := new(core.FocusCollector)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		restore := core.PushFocusCollector(ctx, collector)
		defer restore()
		return layoutui.TrackedInset(ctx, cellGtx, inset, func(gtx layout.Context) layout.Dimensions {
			return alignWidget(ctx, gtx, column.Align, content)
		})
	})
	dims.Size.X = width
	return recordedCell{call: macro.Stop(), dims: dims, placement: placement, interactive: cell.Interactive, focusTargets: collector.Targets}
}

func alignWidget(ctx *core.Context, gtx layout.Context, alignment Alignment, child core.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	direction := layout.W
	switch alignment {
	case AlignCenter:
		direction = layout.Center
	case AlignEnd:
		direction = layout.E
	}
	return layoutui.TrackedDirection(ctx, gtx, direction, func(gtx layout.Context) layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
}

type tableCellText struct {
	text         string
	size         unit.Sp
	color        color.NRGBA
	weight       font.Weight
	textRenderer theme.TableTextRenderer
}

func (t tableCellText) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if t.textRenderer == nil {
		return layout.Dimensions{}
	}
	return t.textRenderer(gtx, theme.TableTextRequest{Text: t.text, Size: t.size, Color: t.color, Weight: t.weight})
}

func (t Widget) resolveColumns(ctx *core.Context, gtx layout.Context, stateValue *tableState, available int, metrics theme.TableMetrics) tableColumns {
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
		result.selection = max(gtx.Dp(metrics.SelectionColumnWidth), 0)
	}
	minimum := result.selection
	lastFlex := -1
	for index, column := range t.columns {
		minWidth := max(gtx.Dp(metrics.MinColumnWidth), 1)
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

func tableVisualRootRadius(metrics theme.TableMetrics, variant Variant, unified bool) unit.Dp {
	radius := metrics.RootRadius
	if variant == VariantSecondary {
		if !unified {
			return 0
		}
		radius = metrics.HeaderRadius
	}
	return radius
}

func tableVisualHeaderRadius(metrics theme.TableMetrics, variant Variant, unified bool) unit.Dp {
	if variant != VariantSecondary || unified {
		return 0
	}
	return metrics.HeaderRadius
}

func tableVisualBodyRadius(metrics theme.TableMetrics, variant Variant) unit.Dp {
	if variant != VariantPrimary {
		return 0
	}
	return metrics.BodyRadius
}
