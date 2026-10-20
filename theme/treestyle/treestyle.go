// Package treestyle contains the theme-owned visual recipe for Tree.
// Tree behavior stays in widget/tree; this package owns geometry, typography,
// palette, and drawing strategies.
package treestyle

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

// BuildTree returns the complete visual plan for the active theme.
func BuildTree(active *theme.Theme, request theme.TreeRequest) theme.TreeVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Tree
	if request.Size == theme.TreeSizeSmall {
		tokens.Padding = tokens.SmallPadding
		tokens.Gap = tokens.SmallGap
		tokens.RowHeight = tokens.SmallRowHeight
		tokens.DescriptionRowHeight = tokens.SmallDescriptionRowHeight
		tokens.RowRadius = tokens.SmallRowRadius
		tokens.RowPaddingX = tokens.SmallRowPaddingX
		tokens.RowPaddingY = tokens.SmallRowPaddingY
		tokens.Indent = tokens.SmallIndent
		tokens.ChevronSlotSize = tokens.SmallChevronSlotSize
		tokens.ChevronIconSize = tokens.SmallChevronIconSize
		tokens.ContentGap = tokens.SmallContentGap
		tokens.ItemTextSize = tokens.SmallItemTextSize
		tokens.ItemDescriptionSize = tokens.SmallItemDescriptionSize
	}
	semantic := active.Semantic()
	rootBackground := color.NRGBA{}
	rootForeground := semantic.Foreground
	if request.Variant == theme.TreeVariantSurface {
		rootBackground = semantic.Surface
		rootForeground = semantic.SurfaceForeground
	}
	selectedForeground := semantic.Tone(theme.TonePrimary).OnSoft
	selectedDescription := selectedForeground
	selectedDescription.A = byte(float32(selectedDescription.A)*0.78 + 0.5)
	itemVisual := func(request theme.TreeItemRequest) theme.TreeItemVisual {
		value := theme.TreeItemVisual{
			Foreground: semantic.Foreground, Description: semantic.MutedForeground,
			Chevron: semantic.MutedForeground, Focus: semantic.Ring, Opacity: 1,
		}
		if request.Hovered {
			value.Background = semantic.SurfaceTertiary
		}
		if request.Selected {
			value.Background = semantic.Tone(theme.TonePrimary).Soft
			value.Foreground = selectedForeground
			value.Description = selectedDescription
			value.Chevron = selectedForeground
			if request.Hovered {
				value.Background = semantic.Tone(theme.TonePrimary).SoftHover
			}
		}
		if request.Disabled {
			value.Opacity = active.DisabledOpacityValue()
			value.Focus = color.NRGBA{}
		}
		return value
	}
	return theme.TreeVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(rootBackground)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.TreeOverrideStyle(request),
		Metrics: theme.TreeMetrics{
			Padding: tokens.Padding, Gap: tokens.Gap, MaxHeight: tokens.MaxHeight,
			RowHeight: tokens.RowHeight, DescriptionRowHeight: tokens.DescriptionRowHeight,
			RowRadius: tokens.RowRadius, RowPaddingX: tokens.RowPaddingX, RowPaddingY: tokens.RowPaddingY,
			Indent: tokens.Indent, ChevronSlotSize: tokens.ChevronSlotSize,
			ChevronIconSize: tokens.ChevronIconSize, ContentGap: tokens.ContentGap,
			ItemTextSize: tokens.ItemTextSize, ItemDescriptionSize: tokens.ItemDescriptionSize, ItemTextWeight: font.Medium,
			FocusRingWidth: tokens.FocusRingWidth, SurfaceRadius: tokens.SurfaceRadius,
			DragPreviewOffset: tokens.DragPreviewOffset, DragPreviewMaxWidth: tokens.DragPreviewMaxWidth,
			DragPreviewPaddingX: tokens.DragPreviewPaddingX, DragPreviewPaddingY: tokens.DragPreviewPaddingY,
			DragPreviewRadius: tokens.DragPreviewRadius,
		},
		Colors: theme.TreeColors{
			RootBackground: rootBackground, RootForeground: rootForeground,
			RowHover: semantic.SurfaceTertiary, RowSelected: semantic.Tone(theme.TonePrimary).Soft,
			RowSelectedHover: semantic.Tone(theme.TonePrimary).SoftHover,
			Foreground:       semantic.Foreground, Description: semantic.MutedForeground,
			SelectedForeground: selectedForeground, SelectedDescription: selectedDescription,
			Chevron: semantic.MutedForeground, Focus: semantic.Ring, Guides: semantic.MutedForeground,
			DropIndicator: semantic.Tone(theme.TonePrimary).Fill,
			Retry:         semantic.Tone(theme.ToneDanger).Fill, ToggleBackground: semantic.SurfaceTertiary,
			DragPreviewBackground: semantic.Overlay, DragPreviewForeground: semantic.OverlayForeground,
			EditorSelection: semantic.Selection,
		},
		DisabledOpacity: active.DisabledOpacityValue(), Motion: active.Motion,
		Item: itemVisual,
		EditorStyle: func(request theme.TreeItemRequest) theme.TreeEditorStyle {
			return theme.TreeEditorStyle{
				TextSize:   tokens.ItemTextSize,
				Foreground: itemVisual(request).Foreground,
				Selection:  semantic.Selection,
				FontWeight: font.Medium,
			}
		},
		Root: func(gtx layout.Context, size image.Point, radius unit.Dp, col color.NRGBA, withShadow bool) {
			drawRoot(active, gtx, size, radius, col, withShadow)
		},
		Row: func(gtx layout.Context, size image.Point, value theme.TreeItemVisual, focus float32, radius unit.Dp) {
			drawRow(gtx, size, value, focus, radius, tokens.FocusRingWidth)
		},
		DropIndicator: func(gtx layout.Context, size image.Point, position theme.TreeDropPosition, depth int, metrics theme.TreeMetrics, col color.NRGBA) {
			drawDropIndicator(gtx, size, position, depth, metrics, col)
		},
		Toggle: func(gtx layout.Context, size image.Point, metrics theme.TreeMetrics, expansion float32, connectors bool, foreground, background color.NRGBA) {
			drawToggle(gtx, size, metrics, expansion, connectors, foreground, background)
		},
		Retry: func(gtx layout.Context, size image.Point, metrics theme.TreeMetrics, col color.NRGBA) {
			drawRetry(gtx, size, metrics, col)
		},
		Guides: func(gtx layout.Context, request theme.TreeGuidesRequest, gap unit.Dp) {
			drawGuides(gtx, request, gap)
		},
		Text: func(gtx layout.Context, request theme.TreeTextRequest) layout.Dimensions {
			label := material.Label(theme.MaterialOf(active), request.Size, request.Text)
			label.Color = request.Color
			label.Font.Weight = request.Weight
			label.MaxLines = 1
			label.Truncator = "..."
			return label.Layout(gtx)
		},
		DragPreview: func(gtx layout.Context, request theme.TreeDragPreviewRequest, metrics theme.TreeMetrics) layout.Dimensions {
			return dragPreview(active, gtx, request.Label, metrics)
		},
	}
}

func clampRadius(gtx layout.Context, size image.Point, radius unit.Dp) int {
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func drawRoot(active *theme.Theme, gtx layout.Context, size image.Point, radiusDp unit.Dp, col color.NRGBA, shadow bool) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	radius := clampRadius(gtx, size, radiusDp)
	if shadow {
		sem := active.Semantic()
		draw.Shadow(gtx, image.Rectangle{Max: size}, draw.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp), draw.ThemeShadow(active.Shadows.Surface, sem.SurfaceShadow, 1))
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func drawRow(gtx layout.Context, size image.Point, value theme.TreeItemVisual, focus float32, radius, focusWidth unit.Dp) {
	rect := image.Rectangle{Max: size}
	rowRadius := min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
	if value.Background.A != 0 {
		paint.FillShape(gtx.Ops, value.Background, clip.UniformRRect(rect, rowRadius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(focusWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	col := value.Focus
	col.A = byte(float32(col.A)*focus + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(focusRect, max(rowRadius-inset, 0)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawDropIndicator(gtx layout.Context, size image.Point, position theme.TreeDropPosition, depth int, metrics theme.TreeMetrics, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	width := max(gtx.Dp(unit.Dp(1)), 1)
	if position == theme.TreeDropInside {
		inset := max(width/2+1, 1)
		rect := image.Rectangle{Max: size}.Inset(inset)
		if rect.Empty() {
			return
		}
		radius := min(max(gtx.Dp(metrics.RowRadius)-inset, 0), min(rect.Dx(), rect.Dy())/2)
		stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		stroke.Pop()
		return
	}
	y := 0
	if position == theme.TreeDropAfter {
		y = size.Y - width
	}
	x := max(gtx.Dp(metrics.RowPaddingX+unit.Dp(depth)*metrics.Indent), 0)
	rect := image.Rect(min(x, size.X), max(y, 0), size.X, min(max(y+width, 0), size.Y))
	if !rect.Empty() {
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, width/2).Op(gtx.Ops))
	}
}

func drawToggle(gtx layout.Context, size image.Point, metrics theme.TreeMetrics, expansion float32, connectors bool, foreground, background color.NRGBA) {
	if connectors {
		drawConnectorToggle(gtx, size, metrics, expansion, foreground, background)
		return
	}
	diameter := min(gtx.Dp(metrics.ChevronIconSize), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	position := image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	rotation := op.Affine(f32.AffineId().Rotate(center, expansion*float32(math.Pi/2))).Push(gtx.Ops)
	offset := op.Offset(position).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.ChevronRight, iconGtx, foreground)
	offset.Pop()
	rotation.Pop()
}

func drawRetry(gtx layout.Context, size image.Point, metrics theme.TreeMetrics, col color.NRGBA) {
	iconSize := min(gtx.Dp(metrics.ChevronIconSize), min(size.X, size.Y))
	if iconSize <= 0 {
		return
	}
	stack := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
	icon.Layout(lucide.RefreshCw, iconGtx, col)
	stack.Pop()
}

func drawConnectorToggle(gtx layout.Context, size image.Point, metrics theme.TreeMetrics, expansion float32, foreground, background color.NRGBA) {
	diameterDp := max(metrics.ChevronIconSize, unit.Dp(14))
	diameter := min(gtx.Dp(diameterDp), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	if background.A == 0 {
		background = foreground
		background.A = 32
	}
	position := image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)
	offset := op.Offset(position).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(image.Rect(0, 0, diameter, diameter), max(gtx.Dp(unit.Dp(3)), 1)).Op(gtx.Ops))
	thickness := max(gtx.Dp(unit.Dp(1)), 1)
	horizontal, vertical := connectorToggleBars(diameter, thickness)
	paint.FillShape(gtx.Ops, foreground, clip.Rect(horizontal).Op())
	if expansion < 1 {
		opacity := paint.PushOpacity(gtx.Ops, max(1-expansion, 0))
		paint.FillShape(gtx.Ops, foreground, clip.Rect(vertical).Op())
		opacity.Pop()
	}
	offset.Pop()
}

func connectorToggleBars(size, thickness int) (image.Rectangle, image.Rectangle) {
	margin := max(size/4, 1)
	center := size / 2
	horizontal := image.Rect(margin, center-thickness/2, size-margin, center+(thickness+1)/2)
	vertical := image.Rect(center-thickness/2, margin, center+(thickness+1)/2, size-margin)
	return horizontal, vertical
}

func drawGuides(gtx layout.Context, request theme.TreeGuidesRequest, gap unit.Dp) {
	if len(request.Segments) == 0 || request.Color.A == 0 {
		return
	}
	col := request.Color
	col.A = byte(float32(col.A)*0.5 + 0.5)
	width := max(gtx.Dp(unit.Dp(1)), 1)
	for _, segment := range request.Segments {
		to := segment.To
		if segment.Extend && segment.From.X == to.X {
			to.Y += gtx.Dp(gap)
		}
		drawGuideSegment(gtx, segment.From, to, width, request.Style, col)
	}
}

func drawGuideSegment(gtx layout.Context, from, to image.Point, width int, guideStyle theme.TreeGuideStyle, col color.NRGBA) {
	if guideStyle == theme.TreeGuideSolid {
		var rect image.Rectangle
		switch {
		case from.X == to.X && from.Y < to.Y:
			rect = image.Rect(from.X-width/2, from.Y, from.X+(width+1)/2, to.Y)
		case from.Y == to.Y && from.X < to.X:
			rect = image.Rect(from.X, from.Y-width/2, to.X, from.Y+(width+1)/2)
		default:
			return
		}
		paint.FillShape(gtx.Ops, col, clip.Rect(rect).Op())
		return
	}
	dash := max(gtx.Dp(unit.Dp(1)), 1)
	period := max(gtx.Dp(unit.Dp(3)), dash+1)
	if from.X == to.X && from.Y < to.Y {
		for _, part := range guideDashes(from.Y, to.Y, dash, period) {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(from.X-width/2, part[0], from.X+(width+1)/2, part[1])).Op())
		}
	} else if from.Y == to.Y && from.X < to.X {
		for _, part := range guideDashes(from.X, to.X, dash, period) {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(part[0], from.Y-width/2, part[1], from.Y+(width+1)/2)).Op())
		}
	}
}

func guideDashes(start, end, dash, period int) [][2]int {
	if start >= end || dash <= 0 || period <= 0 {
		return nil
	}
	parts := make([][2]int, 0, (end-start+period-1)/period)
	for position := start; position < end; position += period {
		parts = append(parts, [2]int{position, min(position+dash, end)})
	}
	return parts
}

func dragPreview(active *theme.Theme, gtx layout.Context, label string, metrics theme.TreeMetrics) layout.Dimensions {
	opacity := paint.PushOpacity(gtx.Ops, 0.5)
	defer opacity.Pop()
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(metrics.DragPreviewMaxWidth))
	gtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y, gtx.Dp(metrics.RowHeight+unit.Dp(8)))
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Top: metrics.DragPreviewPaddingY, Right: metrics.DragPreviewPaddingX, Bottom: metrics.DragPreviewPaddingY, Left: metrics.DragPreviewPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		text := material.Label(theme.MaterialOf(active), metrics.ItemTextSize, label)
		text.Color = active.Semantic().OverlayForeground
		text.Font.Weight = font.Medium
		text.MaxLines = 1
		text.Truncator = "..."
		return text.Layout(gtx)
	})
	content := macro.Stop()
	sem := active.Semantic()
	radius := min(max(gtx.Dp(metrics.DragPreviewRadius), 0), min(dims.Size.X, dims.Size.Y)/2)
	draw.Surface(gtx, image.Rectangle{Max: dims.Size}, radius, sem.Overlay, draw.ThemeShadow(active.Shadows.Overlay, sem.OverlayShadow, 1))
	content.Add(gtx.Ops)
	return dims
}
