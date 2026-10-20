package timeline

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Item is one event in a timeline.
type Item struct {
	Key         string
	Title       core.Widget
	Content     core.Widget
	Icon        core.Widget
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

// Placement controls which side of the rail owns an Item's title/content.
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

// Mode controls which side contains titles in a vertical timeline.
type Mode uint8

const (
	ModeStart Mode = iota
	ModeEnd
	ModeAlternate
)

// Orientation controls the direction of the timeline.
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

// timelineDefaultMaxWidthDp is the neutral width used when the parent leaves
// the horizontal constraint open. It is a layout fallback, not theme styling.
const timelineDefaultMaxWidthDp = 320

// Widget renders a timeline. It is immutable; fluent methods return copies.
type Widget struct {
	override       *theme.TimelineOverride
	items          []Item
	mode           Mode
	orientation    Orientation
	variant        Variant
	reverse        bool
	pending        core.Widget
	pendingIcon    core.Widget
	pendingLoading bool
	disabled       bool
	titleWidth     unit.Dp
	titleSpan      float32
	gap            unit.Dp
	visualCache    *timelineVisualCache
}

type resolvedItem struct {
	Item
	placement Placement
}

func New(items []Item) Widget {
	return Widget{
		items:          append([]Item(nil), items...),
		pendingLoading: true,
		titleSpan:      12,
		visualCache:    &timelineVisualCache{},
	}
}

func (w Widget) Mode(value Mode) Widget {
	w.mode = value
	return w
}
func (w Widget) Orientation(value Orientation) Widget {
	w.orientation = value
	return w
}
func (w Widget) Variant(value Variant) Widget {
	w.variant = value
	return w
}
func (w Widget) Reverse(value bool) Widget  { w.reverse = value; return w }
func (w Widget) Disabled(value bool) Widget { w.disabled = value; return w }
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

func (w Widget) Pending(value core.Widget) Widget {
	w.pending = value
	return w
}
func (w Widget) PendingIcon(value core.Widget) Widget {
	w.pendingIcon = value
	return w
}
func (w Widget) PendingLoading(value bool) Widget { w.pendingLoading = value; return w }

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.TimelineOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	if w.visualCache == nil {
		w.visualCache = &timelineVisualCache{}
	}
	active := core.ActiveTheme(ctx)
	visual := w.resolveVisual(active, override)
	resolved := computed.ResolveStatic(ctx, style.State{Disabled: w.disabled}, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	host := &w.visualCache.visualHost
	host.value = w
	host.visual = visual
	return layoutui.Resolved(ctx, gtx, resolved, host)
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, visual theme.TimelineVisual) layout.Dimensions {
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
		return w.layoutHorizontal(ctx, gtx, resolved, visual)
	}
	if w.usesStructuredVertical(resolved) {
		return w.layoutVerticalStructured(ctx, gtx, resolved, visual)
	}
	return w.layoutVerticalInline(ctx, gtx, resolved, visual)
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

func (w Widget) layoutVerticalStructured(ctx *core.Context, gtx layout.Context, items []resolvedItem, visual theme.TimelineVisual) layout.Dimensions {
	metrics := visual.Metrics
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		maxWidth = gtx.Dp(timelineDefaultMaxWidthDp)
	}
	gap := max(gtx.Dp(metrics.SectionGap), gtx.Dp(metrics.DotSize))
	if w.gap > 0 {
		gap = max(gtx.Dp(w.gap), gtx.Dp(metrics.DotSize))
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
		height := w.layoutVerticalStructuredItem(ctx, itemGtx, item, index, len(items), headerWidth, contentWidth, railCenter, gap, y, visual)
		y += height
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(maxWidth, y))}
}

func (w Widget) layoutVerticalStructuredItem(ctx *core.Context, gtx layout.Context, item resolvedItem, index, count, headerWidth, contentWidth, railCenter, gap, y int, visual theme.TimelineVisual) int {
	metrics := visual.Metrics
	colors := visual.Colors
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
	header := recordStackColor(ctx, constrained(gtx, headerWidth), item.Title, nil, colors.Title, colors.Content, colors.Background, 0, headerAlign)
	content := recordStackColor(ctx, constrained(gtx, contentWidth), nil, item.Content, colors.Title, colors.Content, colors.Background, 0, contentAlign)
	height := max(max(header.dims.Size.Y, content.dims.Size.Y), gtx.Dp(metrics.DotSize))
	markerCenterY := markerCenter(gtx.Dp(metrics.DotSize), header.firstHeight, content.secondHeight)
	itemHeight := height + gtx.Dp(metrics.ItemPaddingBottom)
	if index == count-1 {
		itemHeight = height
	}

	local := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
	icon := measureMarkerIcon(ctx, gtx, item, visual)
	visual.Rail(gtx, theme.TimelineRailRequest{
		Vertical:      true,
		Center:        railCenter,
		Start:         markerCenterY,
		End:           itemHeight,
		MarkerX:       railCenter,
		MarkerY:       markerCenterY,
		Tail:          index < count-1,
		Process:       item.Loading,
		DotSize:       gtx.Dp(metrics.DotSize),
		TailWidth:     gtx.Dp(metrics.TailWidth),
		BorderWidth:   gtx.Dp(metrics.DotBorderWidth),
		DashLength:    gtx.Dp(metrics.ProcessDashLength),
		DashGap:       gtx.Dp(metrics.ProcessDashGap),
		DotBackground: colors.DotBackground,
		TailColor:     colors.Tail,
		MarkerColor:   markerColor(colors, item.Item),
		CustomIcon:    icon.call != (op.CallOp{}),
		Filled:        w.variant == VariantFilled,
	})
	placeCall(gtx.Ops, header.call, image.Pt(headerX+alignedX(headerWidth, header.dims.Size.X, headerAlign), 0))
	placeCall(gtx.Ops, content.call, image.Pt(contentX+alignedX(contentRegionWidth, content.dims.Size.X, contentAlign), 0))
	placeCall(gtx.Ops, icon.call, image.Pt(railCenter-icon.dims.Size.X/2, max(markerCenterY-icon.dims.Size.Y/2, 0)))
	local.Pop()
	return itemHeight
}

func (w Widget) layoutVerticalInline(ctx *core.Context, gtx layout.Context, items []resolvedItem, visual theme.TimelineVisual) layout.Dimensions {
	metrics := visual.Metrics
	colors := visual.Colors
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		maxWidth = gtx.Dp(timelineDefaultMaxWidthDp)
	}
	inset := max(gtx.Dp(metrics.InlineInset), gtx.Dp(metrics.DotSize))
	gap := max(gtx.Dp(metrics.SectionGap), gtx.Dp(metrics.DotSize))
	if w.gap > 0 {
		gap = max(gtx.Dp(w.gap), gtx.Dp(metrics.DotSize))
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
		block := recordStackColor(ctx, constrained(gtx, contentWidth), item.Title, item.Content, colors.Title, colors.Content, colors.Background, gtx.Dp(metrics.TitleContentGap), contentAlign)
		height := max(block.dims.Size.Y, gtx.Dp(metrics.DotSize))
		markerCenterY := markerCenter(gtx.Dp(metrics.DotSize), block.firstHeight, block.secondHeight)
		itemHeight := height + gtx.Dp(metrics.ItemPaddingBottom)
		if index == len(items)-1 {
			itemHeight = height
		}
		local := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		icon := measureMarkerIcon(ctx, gtx, item, visual)
		visual.Rail(gtx, theme.TimelineRailRequest{
			Vertical:      true,
			Center:        railCenter,
			Start:         markerCenterY,
			End:           itemHeight,
			MarkerX:       railCenter,
			MarkerY:       markerCenterY,
			Tail:          index < len(items)-1,
			Process:       item.Loading,
			DotSize:       gtx.Dp(metrics.DotSize),
			TailWidth:     gtx.Dp(metrics.TailWidth),
			BorderWidth:   gtx.Dp(metrics.DotBorderWidth),
			DashLength:    gtx.Dp(metrics.ProcessDashLength),
			DashGap:       gtx.Dp(metrics.ProcessDashGap),
			DotBackground: colors.DotBackground,
			TailColor:     colors.Tail,
			MarkerColor:   markerColor(colors, item.Item),
			CustomIcon:    icon.call != (op.CallOp{}),
			Filled:        w.variant == VariantFilled,
		})
		placeCall(gtx.Ops, block.call, image.Pt(contentX+alignedX(contentWidth, block.dims.Size.X, contentAlign), 0))
		placeCall(gtx.Ops, icon.call, image.Pt(railCenter-icon.dims.Size.X/2, max(markerCenterY-icon.dims.Size.Y/2, 0)))
		local.Pop()
		y += itemHeight
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(maxWidth, y))}
}

func (w Widget) layoutHorizontal(ctx *core.Context, gtx layout.Context, items []resolvedItem, visual theme.TimelineVisual) layout.Dimensions {
	metrics := visual.Metrics
	colors := visual.Colors
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		maxWidth = gtx.Dp(timelineDefaultMaxWidthDp)
	}
	columnGap := gtx.Dp(metrics.HorizontalItemGap)
	columnWidth := max((maxWidth-columnGap*max(len(items)-1, 0))/len(items), 1)
	topGap := gtx.Dp(metrics.HorizontalTitleGap)
	bottomGap := gtx.Dp(metrics.HorizontalContentGap)
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
		title := measureAndRecordColor(ctx, itemGtx, item.Title, colors.Title, colors.Background)
		content := measureAndRecordColor(ctx, itemGtx, item.Content, colors.Content, colors.Background)
		if item.placement == PlacementEnd {
			topBlocks[index] = content
			bottomBlocks[index] = title
		} else {
			topBlocks[index] = title
			bottomBlocks[index] = content
		}
		icons[index] = measureMarkerIcon(ctx, gtx, item, visual)
		topHeight = max(topHeight, topBlocks[index].dims.Size.Y)
		bottomHeight = max(bottomHeight, bottomBlocks[index].dims.Size.Y)
	}
	dotSize := gtx.Dp(metrics.DotSize)
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
		visual.Rail(gtx, theme.TimelineRailRequest{
			Center:        railY,
			Start:         x + dotSize/2,
			End:           tailEnd,
			MarkerX:       x + dotSize/2,
			MarkerY:       railY,
			Tail:          index < len(items)-1,
			Process:       item.Loading,
			DotSize:       dotSize,
			TailWidth:     gtx.Dp(metrics.TailWidth),
			BorderWidth:   gtx.Dp(metrics.DotBorderWidth),
			DashLength:    gtx.Dp(metrics.ProcessDashLength),
			DashGap:       gtx.Dp(metrics.ProcessDashGap),
			DotBackground: colors.DotBackground,
			TailColor:     colors.Tail,
			MarkerColor:   markerColor(colors, item.Item),
			CustomIcon:    icon.call != (op.CallOp{}),
			Filled:        w.variant == VariantFilled,
		})
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

func recordStackColor(ctx *core.Context, gtx layout.Context, first, second core.Widget, firstColor, secondColor, background color.NRGBA, gap int, align blockAlign) measured {
	a := measureAndRecordColor(ctx, gtx, first, firstColor, background)
	b := measureAndRecordColor(ctx, gtx, second, secondColor, background)
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

func measureAndRecord(ctx *core.Context, gtx layout.Context, child core.Widget) measured {
	if child == nil {
		return measured{}
	}
	dims := core.MeasureWidget(ctx, gtx, child)
	macro := op.Record(gtx.Ops)
	child.Layout(ctx, gtx)
	return measured{dims: dims, call: macro.Stop(), firstHeight: dims.Size.Y}
}

func measureAndRecordColor(ctx *core.Context, gtx layout.Context, child core.Widget, foreground, background color.NRGBA) measured {
	if child == nil {
		return measured{}
	}
	restore := core.PushColors(ctx, foreground, background)
	defer restore()
	return measureAndRecord(ctx, gtx, child)
}

func measureMarkerIcon(ctx *core.Context, gtx layout.Context, item resolvedItem, visual theme.TimelineVisual) measured {
	if item.Icon == nil {
		return measured{}
	}
	iconSize := max(gtx.Dp(visual.Metrics.DotSize), 1)
	iconGtx := gtx
	iconGtx.Constraints.Min = image.Pt(iconSize, iconSize)
	iconGtx.Constraints.Max = image.Pt(iconSize, iconSize)
	return measureAndRecordColor(ctx, iconGtx, item.Icon, markerColor(visual.Colors, item.Item), visual.Colors.Background)
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

func markerColor(colors theme.TimelineColors, item Item) color.NRGBA {
	if item.Disabled {
		return colors.Disabled
	}
	if item.Loading {
		return colors.Primary
	}
	if item.Color == ColorCustom && item.CustomColor.A != 0 {
		return item.CustomColor
	}
	switch item.Color {
	case ColorRed:
		return colors.Error
	case ColorGreen:
		return colors.Success
	case ColorGray:
		return colors.Muted
	default:
		return colors.Primary
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

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.TimelineOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
