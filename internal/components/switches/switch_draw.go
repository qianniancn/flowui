package switches

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type switchDrawResult struct {
	dims      layout.Dimensions
	trackRect image.Rectangle
	thumbRect image.Rectangle
}

func switchRects(gtx layout.Context, theme *theme.Theme, style switchStyle, size switchSizeStyle) switchDrawResult {
	focusSpace := max(gtx.Dp(theme.Components.Switch.FocusSpace), 1)
	trackWidthDp, trackHeightDp := switchTrackDpSize(style, size)
	thumbWidthDp, thumbHeightDp := switchThumbDpSize(style, size)
	trackWidth := min(gtx.Dp(trackWidthDp), max(gtx.Constraints.Max.X-focusSpace*2, 0))
	trackHeight := min(gtx.Dp(trackHeightDp), max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := image.Pt(trackWidth+focusSpace*2, trackHeight+focusSpace*2)
	dims := layout.Dimensions{Size: gtx.Constraints.Constrain(bounds)}
	if trackWidth <= 0 || trackHeight <= 0 {
		return switchDrawResult{dims: dims}
	}

	trackOrigin := image.Pt((dims.Size.X-trackWidth)/2, (dims.Size.Y-trackHeight)/2)
	track := image.Rectangle{
		Min: trackOrigin,
		Max: trackOrigin.Add(image.Pt(trackWidth, trackHeight)),
	}
	thumbWidth := min(gtx.Dp(thumbWidthDp), trackWidth)
	thumbHeight := min(gtx.Dp(thumbHeightDp), trackHeight)
	padding := max((trackHeight-thumbHeight)/2, 0)
	travel := max(trackWidth-padding*2-thumbWidth, 0)
	thumbX := track.Min.X + padding + int(float32(travel)*style.selected+0.5)
	thumbY := track.Min.Y + (trackHeight-thumbHeight)/2
	thumb := image.Rectangle{
		Min: image.Pt(thumbX, thumbY),
		Max: image.Pt(thumbX+thumbWidth, thumbY+thumbHeight),
	}
	return switchDrawResult{
		dims:      dims,
		trackRect: track,
		thumbRect: thumb,
	}
}

func switchTrackDpSize(style switchStyle, fallback switchSizeStyle) (unit.Dp, unit.Dp) {
	return switchPartDpSize(style.trackOff, style.trackOn, fallback.trackWidth, fallback.trackHeight)
}

func switchThumbDpSize(style switchStyle, fallback switchSizeStyle) (unit.Dp, unit.Dp) {
	return switchPartDpSize(style.thumbOff, style.thumbOn, fallback.thumbWidth, fallback.thumbHeight)
}

func switchPartDpSize(first, second flowstyle.ResolvedStyle, fallbackWidth, fallbackHeight unit.Dp) (unit.Dp, unit.Dp) {
	var width, height unit.Dp
	var hasWidth, hasHeight bool
	for _, value := range [...]flowstyle.ResolvedStyle{first, second} {
		if value.Box == nil {
			continue
		}
		if value.Box.Width != nil {
			width = max(width, *value.Box.Width)
			hasWidth = true
		}
		if value.Box.Height != nil {
			height = max(height, *value.Box.Height)
			hasHeight = true
		}
	}
	if !hasWidth {
		width = fallbackWidth
	}
	if !hasHeight {
		height = fallbackHeight
	}
	return width, height
}
