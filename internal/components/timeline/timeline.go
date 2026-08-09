package timeline

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/theme"
)

// Item is one event in a time line.
type Item struct {
	Key         string
	Title       frame.Widget
	Content     frame.Widget
	Icon        frame.Widget
	Color       Color
	CustomColor color.NRGBA
	Loading     bool
	Disabled    bool
	Placement   Placement
}

// Color is the semantic marker color used by an item.
type Color uint8

const (
	ColorBlue Color = iota
	ColorRed
	ColorGreen
	ColorGray
	ColorCustom
)

// Placement controls which side of the rail owns an item's title/content.
type Placement uint8

const (
	PlacementAuto Placement = iota
	PlacementStart
	PlacementEnd
)

// Tint sets a custom marker color.
func (i Item) Tint(value color.NRGBA) Item {
	i.Color, i.CustomColor = ColorCustom, value
	return i
}

// Mode controls which side contains titles in a vertical time line.
type Mode uint8

const (
	ModeStart Mode = iota
	ModeEnd
	ModeAlternate
)

// Orientation controls the direction of the time line.
type Orientation uint8

const (
	OrientationVertical Orientation = iota
	OrientationHorizontal
)

// Variant controls whether markers are filled or outlined.
type Variant uint8

const (
	VariantOutlined Variant = iota
	VariantFilled
)

// Widget renders a time line. It is immutable; fluent methods return copies.
type Widget struct {
	items          []Item
	mode           Mode
	orientation    Orientation
	variant        Variant
	reverse        bool
	pending        frame.Widget
	pendingIcon    frame.Widget
	pendingLoading bool
	disabled       bool
	titleWidth     unit.Dp
	titleSpan      float32
	gap            unit.Dp
}

type resolvedItem struct {
	Item
	placement Placement
}

func New(items []Item) Widget {
	return Widget{items: append([]Item(nil), items...), pendingLoading: true, titleSpan: 12}
}

func (w Widget) Mode(value Mode) Widget               { w.mode = value; return w }
func (w Widget) Orientation(value Orientation) Widget { w.orientation = value; return w }
func (w Widget) Variant(value Variant) Widget         { w.variant = value; return w }
func (w Widget) Reverse(value bool) Widget            { w.reverse = value; return w }
func (w Widget) Disabled(value bool) Widget           { w.disabled = value; return w }
func (w Widget) Gap(value int) Widget {
	if value < 0 {
		panic("flowui: timeline gap must not be negative")
	}
	w.gap = unit.Dp(value)
	return w
}

func (w Widget) TitleWidth(value int) Widget {
	if value < 0 {
		panic("flowui: timeline title width must not be negative")
	}
	w.titleWidth = unit.Dp(value)
	return w
}

// TitleSpan mirrors Ant's 24-column title span. Valid values are (0, 24].
func (w Widget) TitleSpan(value float32) Widget {
	if !(value > 0) || math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value > 24 {
		panic("flowui: timeline title span must be between 0 and 24")
	}
	w.titleSpan = value
	return w
}

func (w Widget) Pending(value frame.Widget) Widget     { w.pending = value; return w }
func (w Widget) PendingIcon(value frame.Widget) Widget { w.pendingIcon = value; return w }
func (w Widget) PendingLoading(value bool) Widget      { w.pendingLoading = value; return w }

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if w.disabled {
		gtx = gtx.Disabled()
	}
	items := append([]Item(nil), w.items...)
	if w.pending != nil {
		items = append(items, Item{
			Content:   w.pending,
			Icon:      w.pendingIcon,
			Loading:   w.pendingLoading,
			Color:     ColorGray,
			Disabled:  w.disabled,
			Placement: PlacementAuto,
		})
	}
	resolved := w.resolveItems(items)
	if len(resolved) == 0 {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Point{})}
	}
	if w.orientation == OrientationHorizontal {
		return w.layoutHorizontal(ctx, gtx, resolved)
	}
	if w.usesStructuredVertical(resolved) {
		return w.layoutVerticalStructured(ctx, gtx, resolved)
	}
	return w.layoutVerticalInline(ctx, gtx, resolved)
}

func (w Widget) resolveItems(items []Item) []resolvedItem {
	resolved := make([]resolvedItem, 0, len(items))
	for index, item := range items {
		if w.disabled {
			item.Disabled = true
		}
		resolved = append(resolved, resolvedItem{
			Item:      item,
			placement: w.resolvePlacement(item, index),
		})
	}
	if w.reverse {
		for left, right := 0, len(resolved)-1; left < right; left, right = left+1, right-1 {
			resolved[left], resolved[right] = resolved[right], resolved[left]
		}
	}
	return resolved
}

func (w Widget) resolvePlacement(item Item, index int) Placement {
	if item.Placement == PlacementStart || item.Placement == PlacementEnd {
		return item.Placement
	}
	switch w.mode {
	case ModeEnd:
		return PlacementEnd
	case ModeAlternate:
		if index%2 == 0 {
			return PlacementStart
		}
		return PlacementEnd
	default:
		return PlacementStart
	}
}

func (w Widget) usesStructuredVertical(items []resolvedItem) bool {
	if w.mode == ModeAlternate {
		return true
	}
	for _, item := range items {
		if item.Title != nil {
			return true
		}
	}
	return false
}

func (w Widget) layoutVerticalStructured(ctx *frame.Context, gtx layout.Context, items []resolvedItem) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.TimeLine
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		maxWidth = gtx.Dp(320)
	}
	gap := max(gtx.Dp(tokens.SectionGap), gtx.Dp(tokens.DotSize))
	if w.gap > 0 {
		gap = max(gtx.Dp(w.gap), gtx.Dp(tokens.DotSize))
	}
	span := w.resolveTitleSpanPx(gtx, maxWidth)
	y := 0
	for index, item := range items {
		headerWidth := max(span-gap/2, 0)
		contentWidth := max(maxWidth-(span+gap/2), 0)
		railCenter := span
		if item.placement == PlacementEnd {
			headerWidth = max((maxWidth-span)-gap/2, 0)
			contentWidth = max(span-gap/2, 0)
			railCenter = maxWidth - span
		}
		itemGtx := gtx
		itemGtx.Constraints.Min = image.Point{}
		itemGtx.Constraints.Max = image.Pt(maxWidth, gtx.Constraints.Max.Y)
		height := w.layoutVerticalStructuredItem(ctx, itemGtx, item, index, len(items), headerWidth, contentWidth, railCenter, gap, y)
		y += height
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(maxWidth, y))}
}

func (w Widget) layoutVerticalStructuredItem(ctx *frame.Context, gtx layout.Context, item resolvedItem, index, count, headerWidth, contentWidth, railCenter, gap, y int) int {
	tokens := frame.ActiveTheme(ctx).Components.TimeLine
	headerAlign, contentAlign := alignEnd, alignStart
	contentX := railCenter + gap/2
	headerX := 0
	contentRegionWidth := contentWidth
	if item.placement == PlacementEnd {
		headerAlign, contentAlign = alignStart, alignEnd
		headerX = railCenter + gap/2
		contentX = 0
		contentRegionWidth = contentWidth
	}
	header := recordStackColor(ctx, constrained(gtx, headerWidth), item.Title, nil, tokens.TitleColor, tokens.ContentColor, 0, headerAlign)
	content := recordStackColor(ctx, constrained(gtx, contentWidth), nil, item.Content, tokens.TitleColor, tokens.ContentColor, 0, contentAlign)
	height := max(max(header.dims.Size.Y, content.dims.Size.Y), gtx.Dp(tokens.DotSize))
	markerCenterY := markerCenter(gtx.Dp(tokens.DotSize), header.firstHeight, content.secondHeight)
	itemHeight := height + gtx.Dp(tokens.ItemPaddingBottom)
	if index == count-1 {
		itemHeight = height
	}

	local := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
	icon := measureMarkerIcon(ctx, gtx, item)
	drawVerticalRail(gtx.Ops, railCenter, itemHeight, markerCenterY, index < count-1, item.Loading, gtx.Dp(tokens.DotSize), gtx.Dp(tokens.TailWidth), gtx.Dp(tokens.DotBorderWidth), gtx.Dp(tokens.ProcessDashLength), gtx.Dp(tokens.ProcessDashGap), tokens.DotBackground, tokens.TailColor, markerColor(frame.ActiveTheme(ctx), item.Item), icon.call != (op.CallOp{}), w.variant)
	placeCall(gtx.Ops, header.call, image.Pt(headerX+alignedX(headerWidth, header.dims.Size.X, headerAlign), 0))
	placeCall(gtx.Ops, content.call, image.Pt(contentX+alignedX(contentRegionWidth, content.dims.Size.X, contentAlign), 0))
	placeCall(gtx.Ops, icon.call, image.Pt(railCenter-icon.dims.Size.X/2, max(markerCenterY-icon.dims.Size.Y/2, 0)))
	local.Pop()
	return itemHeight
}

func (w Widget) layoutVerticalInline(ctx *frame.Context, gtx layout.Context, items []resolvedItem) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.TimeLine
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		maxWidth = gtx.Dp(320)
	}
	inset := max(gtx.Dp(tokens.InlineInset), gtx.Dp(tokens.DotSize))
	gap := max(gtx.Dp(tokens.SectionGap), gtx.Dp(tokens.DotSize))
	if w.gap > 0 {
		gap = max(gtx.Dp(w.gap), gtx.Dp(tokens.DotSize))
	}
	y := 0
	for index, item := range items {
		contentAlign := alignStart
		contentX := inset + gap/2
		railCenter := inset
		contentWidth := max(maxWidth-contentX, 0)
		if item.placement == PlacementEnd {
			contentAlign = alignEnd
			railCenter = maxWidth - inset
			contentX = 0
			contentWidth = max(railCenter-gap/2, 0)
		}
		block := recordStackColor(ctx, constrained(gtx, contentWidth), item.Title, item.Content, tokens.TitleColor, tokens.ContentColor, gtx.Dp(tokens.TitleContentGap), contentAlign)
		height := max(block.dims.Size.Y, gtx.Dp(tokens.DotSize))
		markerCenterY := markerCenter(gtx.Dp(tokens.DotSize), block.firstHeight, block.secondHeight)
		itemHeight := height + gtx.Dp(tokens.ItemPaddingBottom)
		if index == len(items)-1 {
			itemHeight = height
		}
		local := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		icon := measureMarkerIcon(ctx, gtx, item)
		drawVerticalRail(gtx.Ops, railCenter, itemHeight, markerCenterY, index < len(items)-1, item.Loading, gtx.Dp(tokens.DotSize), gtx.Dp(tokens.TailWidth), gtx.Dp(tokens.DotBorderWidth), gtx.Dp(tokens.ProcessDashLength), gtx.Dp(tokens.ProcessDashGap), tokens.DotBackground, tokens.TailColor, markerColor(frame.ActiveTheme(ctx), item.Item), icon.call != (op.CallOp{}), w.variant)
		placeCall(gtx.Ops, block.call, image.Pt(contentX+alignedX(contentWidth, block.dims.Size.X, contentAlign), 0))
		placeCall(gtx.Ops, icon.call, image.Pt(railCenter-icon.dims.Size.X/2, max(markerCenterY-icon.dims.Size.Y/2, 0)))
		local.Pop()
		y += itemHeight
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(maxWidth, y))}
}

func (w Widget) layoutHorizontal(ctx *frame.Context, gtx layout.Context, items []resolvedItem) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.TimeLine
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		maxWidth = gtx.Dp(320)
	}
	columnGap := gtx.Dp(tokens.HorizontalItemGap)
	columnWidth := max((maxWidth-columnGap*max(len(items)-1, 0))/len(items), 1)
	topGap := gtx.Dp(tokens.HorizontalTitleGap)
	bottomGap := gtx.Dp(tokens.HorizontalContentGap)
	if w.gap > 0 {
		topGap = gtx.Dp(w.gap)
		bottomGap = gtx.Dp(w.gap)
	}
	topBlocks := make([]measured, len(items))
	bottomBlocks := make([]measured, len(items))
	icons := make([]measured, len(items))
	topHeight, bottomHeight := 0, 0
	for index, item := range items {
		itemGtx := constrained(gtx, columnWidth)
		title := measureAndRecordColor(ctx, itemGtx, item.Title, tokens.TitleColor)
		content := measureAndRecordColor(ctx, itemGtx, item.Content, tokens.ContentColor)
		if item.placement == PlacementEnd {
			topBlocks[index] = content
			bottomBlocks[index] = title
		} else {
			topBlocks[index] = title
			bottomBlocks[index] = content
		}
		icons[index] = measureMarkerIcon(ctx, gtx, item)
		topHeight = max(topHeight, topBlocks[index].dims.Size.Y)
		bottomHeight = max(bottomHeight, bottomBlocks[index].dims.Size.Y)
	}
	dotSize := gtx.Dp(tokens.DotSize)
	railY := topHeight + topGap + dotSize/2
	bottomY := railY + dotSize/2 + bottomGap
	totalHeight := max(bottomY+bottomHeight, dotSize)
	for index, item := range items {
		x := index * (columnWidth + columnGap)
		icon := icons[index]
		tailEnd := x + columnWidth + columnGap
		if index == len(items)-1 {
			tailEnd = x + columnWidth
		}
		drawHorizontalRail(gtx.Ops, x+dotSize/2, tailEnd, railY, index < len(items)-1, item.Loading, dotSize, gtx.Dp(tokens.TailWidth), gtx.Dp(tokens.DotBorderWidth), gtx.Dp(tokens.ProcessDashLength), gtx.Dp(tokens.ProcessDashGap), tokens.DotBackground, tokens.TailColor, markerColor(frame.ActiveTheme(ctx), item.Item), icon.call != (op.CallOp{}), w.variant)
		placeCall(gtx.Ops, topBlocks[index].call, image.Pt(x+alignedX(columnWidth, topBlocks[index].dims.Size.X, alignCenter), topHeight-topBlocks[index].dims.Size.Y))
		placeCall(gtx.Ops, bottomBlocks[index].call, image.Pt(x+alignedX(columnWidth, bottomBlocks[index].dims.Size.X, alignCenter), bottomY))
		placeCall(gtx.Ops, icon.call, image.Pt(x+dotSize/2-icon.dims.Size.X/2, railY-icon.dims.Size.Y/2))
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(maxWidth, totalHeight))}
}

type measured struct {
	dims         layout.Dimensions
	call         op.CallOp
	firstHeight  int
	secondHeight int
}

type blockAlign uint8

const (
	alignStart blockAlign = iota
	alignCenter
	alignEnd
)

func constrained(gtx layout.Context, width int) layout.Context {
	child := gtx
	child.Constraints.Min = image.Point{}
	child.Constraints.Max.X = max(width, 0)
	return child
}

func recordStackColor(ctx *frame.Context, gtx layout.Context, first, second frame.Widget, firstColor, secondColor color.NRGBA, gap int, align blockAlign) measured {
	a := measureAndRecordColor(ctx, gtx, first, firstColor)
	b := measureAndRecordColor(ctx, gtx, second, secondColor)
	width := max(a.dims.Size.X, b.dims.Size.X)
	height := a.dims.Size.Y + b.dims.Size.Y
	if a.dims.Size.Y > 0 && b.dims.Size.Y > 0 {
		height += gap
	}
	macro := op.Record(gtx.Ops)
	placeCall(gtx.Ops, a.call, image.Pt(alignedX(width, a.dims.Size.X, align), 0))
	secondY := a.dims.Size.Y
	if a.dims.Size.Y > 0 && b.dims.Size.Y > 0 {
		secondY += gap
	}
	placeCall(gtx.Ops, b.call, image.Pt(alignedX(width, b.dims.Size.X, align), secondY))
	return measured{
		dims:         layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))},
		call:         macro.Stop(),
		firstHeight:  a.dims.Size.Y,
		secondHeight: b.dims.Size.Y,
	}
}

func measureAndRecord(ctx *frame.Context, gtx layout.Context, child frame.Widget) measured {
	if child == nil {
		return measured{}
	}
	dims := frame.MeasureWidget(ctx, gtx, child)
	macro := op.Record(gtx.Ops)
	child.Layout(ctx, gtx)
	return measured{dims: dims, call: macro.Stop(), firstHeight: dims.Size.Y}
}

func measureAndRecordColor(ctx *frame.Context, gtx layout.Context, child frame.Widget, foreground color.NRGBA) measured {
	if child == nil {
		return measured{}
	}
	restore := frame.PushColors(ctx, foreground, frame.ActiveTheme(ctx).Palette.Background)
	defer restore()
	return measureAndRecord(ctx, gtx, child)
}

func measureMarkerIcon(ctx *frame.Context, gtx layout.Context, item resolvedItem) measured {
	if item.Icon == nil {
		return measured{}
	}
	tokens := frame.ActiveTheme(ctx).Components.TimeLine
	iconSize := max(gtx.Dp(tokens.DotSize), 1)
	iconGtx := gtx
	iconGtx.Constraints.Min = image.Pt(iconSize, iconSize)
	iconGtx.Constraints.Max = image.Pt(iconSize, iconSize)
	return measureAndRecordColor(ctx, iconGtx, item.Icon, markerColor(frame.ActiveTheme(ctx), item.Item))
}

func placeCall(ops *op.Ops, call op.CallOp, point image.Point) {
	if call == (op.CallOp{}) {
		return
	}
	trans := op.Offset(point).Push(ops)
	call.Add(ops)
	trans.Pop()
}

func alignedX(totalWidth, childWidth int, align blockAlign) int {
	switch align {
	case alignCenter:
		return max((totalWidth-childWidth)/2, 0)
	case alignEnd:
		return max(totalWidth-childWidth, 0)
	default:
		return 0
	}
}

func drawVerticalRail(ops *op.Ops, center, height, markerCenterY int, tail, process bool, dotSize, tailWidth, borderWidth, dashLength, dashGap int, dotBackground, tailColor, marker color.NRGBA, customIcon bool, variant Variant) {
	lineStart := markerCenterY + dotSize/2
	if tail {
		if process {
			paintVerticalDashes(ops, center, lineStart, height, max(tailWidth, 1), tailColor, max(dashLength, 2), max(dashGap, 2))
		} else {
			paint.FillShape(ops, tailColor, clip.Rect(image.Rect(center-max(tailWidth, 1)/2, lineStart, center+(max(tailWidth, 1)+1)/2, height)).Op())
		}
	}
	if !customIcon {
		drawMarker(ops, center, markerCenterY, dotSize, borderWidth, dotBackground, marker, variant)
	}
}

func drawHorizontalRail(ops *op.Ops, start, end, center int, tail, process bool, dotSize, tailWidth, borderWidth, dashLength, dashGap int, dotBackground, tailColor, marker color.NRGBA, customIcon bool, variant Variant) {
	lineStart := start + dotSize/2
	if tail {
		if process {
			paintHorizontalDashes(ops, lineStart, end, center, max(tailWidth, 1), tailColor, max(dashLength, 2), max(dashGap, 2))
		} else {
			paint.FillShape(ops, tailColor, clip.Rect(image.Rect(lineStart, center-max(tailWidth, 1)/2, end, center+(max(tailWidth, 1)+1)/2)).Op())
		}
	}
	if !customIcon {
		drawMarker(ops, start, center, dotSize, borderWidth, dotBackground, marker, variant)
	}
}

func paintVerticalDashes(ops *op.Ops, center, start, end, width int, value color.NRGBA, dashLength, dashGap int) {
	for y := start; y < end; y += dashLength + dashGap {
		dashEnd := min(y+dashLength, end)
		paint.FillShape(ops, value, clip.Rect(image.Rect(center-width/2, y, center+(width+1)/2, dashEnd)).Op())
	}
}

func paintHorizontalDashes(ops *op.Ops, start, end, center, width int, value color.NRGBA, dashLength, dashGap int) {
	for x := start; x < end; x += dashLength + dashGap {
		dashEnd := min(x+dashLength, end)
		paint.FillShape(ops, value, clip.Rect(image.Rect(x, center-width/2, dashEnd, center+(width+1)/2)).Op())
	}
}

func drawMarker(ops *op.Ops, centerX, centerY, size, borderWidth int, dotBackground, marker color.NRGBA, variant Variant) {
	size = max(size, 2)
	outer := image.Rect(centerX-size/2, centerY-size/2, centerX+(size+1)/2, centerY+(size+1)/2)
	paint.FillShape(ops, marker, clip.Ellipse{Min: outer.Min, Max: outer.Max}.Op(ops))
	if variant == VariantOutlined {
		inner := size - max(borderWidth*2, 2)
		if inner > 0 {
			innerRect := image.Rect(centerX-inner/2, centerY-inner/2, centerX+(inner+1)/2, centerY+(inner+1)/2)
			paint.FillShape(ops, dotBackground, clip.Ellipse{Min: innerRect.Min, Max: innerRect.Max}.Op(ops))
		}
	}
}

func markerColor(active *theme.Theme, item Item) color.NRGBA {
	if item.Disabled {
		return active.Components.TimeLine.DisabledColor
	}
	if item.Loading {
		return active.Components.TimeLine.PrimaryColor
	}
	if item.Color == ColorCustom && item.CustomColor.A != 0 {
		return item.CustomColor
	}
	switch item.Color {
	case ColorRed:
		return active.Components.TimeLine.ErrorColor
	case ColorGreen:
		return active.Components.TimeLine.SuccessColor
	case ColorGray:
		return active.Components.TimeLine.MutedColor
	default:
		return active.Components.TimeLine.PrimaryColor
	}
}

func (w Widget) resolveTitleSpanPx(gtx layout.Context, width int) int {
	if w.titleWidth > 0 {
		return min(max(gtx.Dp(w.titleWidth), 0), width)
	}
	span := w.titleSpan
	if !(span > 0) || math.IsNaN(float64(span)) || math.IsInf(float64(span), 0) {
		span = 12
	}
	return max(int(float32(width)*(span/24)), 0)
}

func markerCenter(dotSize, firstHeight, secondHeight int) int {
	anchor := firstHeight
	if anchor <= 0 {
		anchor = secondHeight
	}
	if anchor <= 0 {
		return dotSize / 2
	}
	return max(anchor/2, dotSize/2)
}
