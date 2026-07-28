package slider

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type sliderStyle struct {
	track flowstyle.ResolvedStyle
	fill  flowstyle.ResolvedStyle
	thumb flowstyle.ResolvedStyle
	label flowstyle.ResolvedStyle
}

type sliderColors struct {
	track           color.NRGBA
	fill            color.NRGBA
	thumbBorder     color.NRGBA
	thumbBackground color.NRGBA
	focus           color.NRGBA
}

func sliderColorsFor(activeTheme *theme.Theme, disabled bool) sliderColors {
	colors := sliderColors{
		track:           activeTheme.Palette.SurfaceRaised,
		fill:            activeTheme.Palette.Accent,
		thumbBorder:     activeTheme.Palette.Accent,
		thumbBackground: activeTheme.Palette.AccentForeground,
		focus:           activeTheme.Palette.Focus,
	}
	if disabled {
		colors.track = activeTheme.DisabledColor(colors.track)
		colors.fill = activeTheme.DisabledColor(colors.fill)
		colors.thumbBorder = activeTheme.DisabledColor(colors.thumbBorder)
		colors.thumbBackground = activeTheme.DisabledColor(colors.thumbBackground)
		colors.focus = color.NRGBA{}
	}
	return colors
}

func (s SliderWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) sliderStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := sliderStyleDeclaration(activeTheme, sliderColorsFor(activeTheme, state.Disabled), state.Disabled, s.orientation)
	track := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartTrack, state, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	fill := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartFill, state, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	thumb := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartThumb, state, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	label := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartLabel, state, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	return sliderStyle{track: track, fill: fill, thumb: thumb, label: label}
}

func sliderStyleDeclaration(activeTheme *theme.Theme, colors sliderColors, disabled bool, orientation SliderOrientation) flowstyle.Style {
	tokens := activeTheme.Components.Slider
	label := activeTheme.Palette.Foreground
	if disabled {
		label = activeTheme.DisabledColor(label)
	}
	track := flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: colors.track}).
		Radius(tokens.TrackRadius)
	fill := flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: colors.fill})
	if orientation == SliderVertical {
		fill = fill.RadiusBottomLeft(tokens.TrackRadius).RadiusBottomRight(tokens.TrackRadius)
	} else {
		fill = fill.RadiusTopLeft(tokens.TrackRadius).RadiusBottomLeft(tokens.TrackRadius)
	}
	thumbWidth := tokens.ThumbLength + tokens.ThumbExtra
	thumbHeight := tokens.TrackThickness
	if orientation == SliderVertical {
		track = track.Width(tokens.TrackThickness).FillHeight()
		thumbWidth, thumbHeight = thumbHeight, thumbWidth
	} else {
		track = track.FillWidth().Height(tokens.TrackThickness)
	}
	borderWidth := max(tokens.ThumbExtra/2, unit.Dp(1))
	shadowAlpha := func(alpha uint8) uint8 {
		return uint8(uint16(alpha) * uint16(colors.thumbBackground.A) / 0xff)
	}
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, track).
		Part(flowstyle.PartFill, fill).
		Part(flowstyle.PartThumb, flowstyle.Style{}.
			Width(thumbWidth).
			Height(thumbHeight).
			Background(flowstyle.SolidColor{Color: colors.thumbBackground}).
			BorderWidth(borderWidth).
			BorderColor(flowstyle.SolidColor{Color: colors.thumbBorder}).
			Radius(tokens.TrackRadius).
			BoxShadow(0, 2, 0, 2, flowstyle.SolidColor{Color: color.NRGBA{A: shadowAlpha(0x0a)}}).
			BoxShadow(0, 1, 0, 1, flowstyle.SolidColor{Color: color.NRGBA{A: shadowAlpha(0x0f)}}).
			Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, flowstyle.SolidColor{Color: colors.focus})).
		Part(flowstyle.PartLabel, flowstyle.Style{}.
			TextColor(flowstyle.SolidColor{Color: label}).
			FontSize(tokens.TextSize).
			FontWeight(int(font.Medium)))

}
