package tabs

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/theme"
)

const (
	tabsColorDuration     = 150 * time.Millisecond
	tabsIndicatorDuration = 250 * time.Millisecond
)

func (s *tabsIndicatorState) transition(gtx layout.Context, key string, orientation Orientation, target image.Rectangle, motions ...theme.MotionTheme) image.Rectangle {
	return s.transitionWithDuration(gtx, key, orientation, target, tabsIndicatorDuration, motions...)
}

func (s *tabsItemState) selectionProgress(gtx layout.Context, selected bool, motions ...theme.MotionTheme) float32 {
	return s.selectionProgressWithDuration(gtx, selected, tabsColorDuration, motions...)
}

// tabsSizeStyleFor is a behavior-test measurement adapter. Production layout
// receives the provider-owned TabsMetrics directly.
func tabsSizeStyleFor(active *theme.Theme, size Size) tabsSizeStyle {
	tokens := active.Components.Tabs
	return tabsSizeStyleForVisual(theme.TabsMetrics{
		TabHeight: tokens.TabHeight, SmallTabHeight: tokens.SmallTabHeight, LargeTabHeight: tokens.LargeTabHeight,
		TabPaddingX: tokens.TabPaddingX, SmallTabPaddingX: tokens.SmallTabPaddingX, LargeTabPaddingX: tokens.LargeTabPaddingX,
	}, size)
}
