package tabs

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/theme"
)

// tabsSizeStyle is the resolved provider geometry used by neutral layout.
// It contains no colors or theme recipe rules.
type tabsSizeStyle struct {
	height   unit.Dp
	paddingX unit.Dp
}

func tabsSizeStyleForVisual(metrics theme.TabsMetrics, size Size) tabsSizeStyle {
	result := tabsSizeStyle{height: metrics.TabHeight, paddingX: metrics.TabPaddingX}
	switch size {
	case Small:
		result.height, result.paddingX = metrics.SmallTabHeight, metrics.SmallTabPaddingX
	case Large:
		result.height, result.paddingX = metrics.LargeTabHeight, metrics.LargeTabPaddingX
	}
	return result
}
