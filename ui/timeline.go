package ui

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/components/timeline"
)

// TimeLineWidget renders an Ant Design-inspired event time line.
type TimeLineWidget = timeline.Widget

// TimeLineItem describes one event. Title and Content accept any FlowUI
// widget, allowing labels, rich layouts, or application-specific content.
type TimeLineItem = timeline.Item

type TimeLineColor = timeline.Color
type TimeLinePlacement = timeline.Placement
type TimeLineMode = timeline.Mode
type TimeLineOrientation = timeline.Orientation
type TimeLineVariant = timeline.Variant

const (
	TimeLineBlue   = timeline.ColorBlue
	TimeLineRed    = timeline.ColorRed
	TimeLineGreen  = timeline.ColorGreen
	TimeLineGray   = timeline.ColorGray
	TimeLineCustom = timeline.ColorCustom

	TimeLinePlacementAuto  = timeline.PlacementAuto
	TimeLinePlacementStart = timeline.PlacementStart
	TimeLinePlacementEnd   = timeline.PlacementEnd

	TimeLineStart     = timeline.ModeStart
	TimeLineEnd       = timeline.ModeEnd
	TimeLineAlternate = timeline.ModeAlternate

	TimeLineVertical   = timeline.OrientationVertical
	TimeLineHorizontal = timeline.OrientationHorizontal

	TimeLineOutlined = timeline.VariantOutlined
	TimeLineFilled   = timeline.VariantFilled
)

// TimeLine creates a vertical time line by default. The name intentionally
// differs from ui.Timeline, which is the existing animation timeline API.
func TimeLine(items []TimeLineItem) TimeLineWidget { return timeline.New(items) }

// CustomTimeLineItem is a convenience constructor for custom marker colors.
func CustomTimeLineItem(title, content Widget, marker color.NRGBA) TimeLineItem {
	return TimeLineItem{Title: title, Content: content, Color: TimeLineCustom, CustomColor: marker}
}

// TimeLineItemTint returns a copy of item with a custom marker color.
func TimeLineItemTint(item TimeLineItem, marker color.NRGBA) TimeLineItem { return item.Tint(marker) }
