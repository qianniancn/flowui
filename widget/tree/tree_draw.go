package tree

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/theme"
)

type treeGuideSegment struct {
	from   image.Point
	to     image.Point
	extend bool
}

func drawTreeGuides(gtx layout.Context, entry flatItem, metrics theme.TreeMetrics, expanded, connectors bool, guideStyle GuideStyle, visual theme.TreeVisual) {
	rowHeight := gtx.Constraints.Max.Y
	segments := treeGuideSegments(
		entry,
		gtx.Dp(metrics.RowPaddingX),
		gtx.Dp(metrics.Indent),
		gtx.Dp(metrics.ChevronSlotSize),
		gtx.Dp(metrics.ContentGap),
		rowHeight,
		expanded,
		connectors,
	)
	if len(segments) == 0 || visual.Colors.Guides.A == 0 || visual.Guides == nil {
		return
	}
	providerSegments := make([]theme.TreeGuideSegment, 0, len(segments))
	for _, segment := range segments {
		providerSegments = append(providerSegments, theme.TreeGuideSegment{From: segment.from, To: segment.to, Extend: segment.extend})
	}
	style := theme.TreeGuideSolid
	if guideStyle == GuideDashed {
		style = theme.TreeGuideDashed
	}
	visual.Guides(gtx, theme.TreeGuidesRequest{Segments: providerSegments, Style: style, Color: visual.Colors.Guides}, metrics.Gap)
}

func treeGuideSegments(entry flatItem, rowPadding, indent, toggleSlot, contentGap, rowHeight int, expanded, connectors bool) []treeGuideSegment {
	if indent <= 0 || toggleSlot <= 0 || rowHeight <= 0 {
		return nil
	}
	centerY := rowHeight / 2
	guideX := func(depth int) int {
		return rowPadding + depth*indent + toggleSlot/2
	}
	toggleCenterX := func(depth int) int {
		return guideX(depth)
	}
	segments := make([]treeGuideSegment, 0, entry.depth+1)
	for level := 0; level < entry.depth-1 && level+1 < len(entry.ancestorsLast); level++ {
		if !entry.ancestorsLast[level+1] {
			x := guideX(level)
			segments = append(segments, treeGuideSegment{from: image.Pt(x, 0), to: image.Pt(x, rowHeight), extend: true})
		}
	}
	if entry.depth > 0 {
		x := guideX(entry.depth - 1)
		y2 := rowHeight
		if entry.isLast && connectors {
			y2 = centerY
		}
		segments = append(segments, treeGuideSegment{from: image.Pt(x, 0), to: image.Pt(x, y2), extend: !entry.isLast})
		if connectors {
			endX := toggleCenterX(entry.depth)
			if !treeItemExpandable(entry.item) {
				endX = rowPadding + entry.depth*indent + toggleSlot + contentGap
			}
			segments = append(segments, treeGuideSegment{from: image.Pt(x, centerY), to: image.Pt(endX, centerY)})
		}
	}
	if expanded && len(entry.item.Children) > 0 {
		x := guideX(entry.depth)
		y1 := min(centerY+toggleSlot/2, rowHeight)
		segments = append(segments, treeGuideSegment{from: image.Pt(x, y1), to: image.Pt(x, rowHeight), extend: true})
	}
	return segments
}

func treeGuideDashes(start, end, dash, period int) [][2]int {
	if start >= end || dash <= 0 || period <= 0 {
		return nil
	}
	parts := make([][2]int, 0, (end-start+period-1)/period)
	for position := start; position < end; position += period {
		parts = append(parts, [2]int{position, min(position+dash, end)})
	}
	return parts
}

func treeRootRadius(gtx layout.Context, metrics theme.TreeMetrics, variant Variant, size image.Point) int {
	if variant != VariantSurface {
		return 0
	}
	return min(max(gtx.Dp(metrics.SurfaceRadius), 0), min(size.X, size.Y)/2)
}

func treeDropPosition(value DropPosition) theme.TreeDropPosition {
	switch value {
	case DropInside:
		return theme.TreeDropInside
	case DropAfter:
		return theme.TreeDropAfter
	default:
		return theme.TreeDropBefore
	}
}

func treeConnectorToggleBars(size, thickness int) (image.Rectangle, image.Rectangle) {
	margin := max(size/4, 1)
	center := size / 2
	horizontal := image.Rect(margin, center-thickness/2, size-margin, center+(thickness+1)/2)
	vertical := image.Rect(center-thickness/2, margin, center+(thickness+1)/2, size-margin)
	return horizontal, vertical
}
