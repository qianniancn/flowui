// Package sliderstyle contains the shared Slider visual recipe.
// Slider behavior, pointer handling, and value state remain in widget/slider;
// this package owns the visual declaration and theme-token mapping.
package sliderstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildSlider builds the provider-owned Slider declaration and metrics.
func BuildSlider(active *theme.Theme, request theme.SliderRequest) theme.SliderVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Slider
	focusDuration := 100 * time.Millisecond
	draggingDuration := 150 * time.Millisecond
	if active.Motion.ControlDuration > 0 {
		focusDuration = active.Motion.ControlDuration
		draggingDuration = active.Motion.ControlDuration
	}
	return theme.SliderVisual{
		Defaults:         ruleSet(active, request.Orientation).Compile(),
		Override:         theme.SliderOverrideStyle(request),
		Metrics:          sliderMetrics(tokens),
		Motion:           active.Motion,
		FocusDuration:    focusDuration,
		DraggingDuration: draggingDuration,
	}
}

func sliderMetrics(tokens theme.SliderTheme) theme.SliderMetrics {
	return theme.SliderMetrics{
		TrackThickness:  tokens.TrackThickness,
		TrackRadius:     tokens.TrackRadius,
		EdgeInset:       tokens.EdgeInset,
		ThumbLength:     tokens.ThumbLength,
		ThumbExtra:      tokens.ThumbExtra,
		HeaderGap:       tokens.HeaderGap,
		VerticalGap:     tokens.VerticalGap,
		DraggingScale:   tokens.DraggingScale,
		FocusRingWidth:  tokens.FocusRingWidth,
		FocusRingOffset: tokens.FocusRingOffset,
	}
}

func ruleSet(active *theme.Theme, orientation theme.SliderOrientation) style.Recipe {
	tokens := active.Components.Slider
	enabled := colors(active, false)
	disabled := colors(active, true)

	thumbWidth := tokens.ThumbLength + tokens.ThumbExtra
	thumbHeight := tokens.TrackThickness
	thumbRadius := tokens.TrackRadius
	if tokens.ThumbSize > 0 {
		thumbWidth, thumbHeight = tokens.ThumbSize, tokens.ThumbSize
		thumbRadius = tokens.ThumbSize / 2
	}
	borderWidth := max(tokens.ThumbExtra/2, unit.Dp(1))

	trackBase := style.Style{}.
		Background(style.Solid(enabled.Track)).
		Radius(tokens.TrackRadius)
	fillBase := style.Style{}.Background(style.Solid(enabled.Fill))
	if orientation == theme.SliderVertical {
		fillBase = fillBase.RadiusBottomLeft(tokens.TrackRadius).RadiusBottomRight(tokens.TrackRadius)
		trackBase = trackBase.Width(tokens.TrackThickness).FillHeight()
		thumbWidth, thumbHeight = thumbHeight, thumbWidth
	} else {
		fillBase = fillBase.RadiusTopLeft(tokens.TrackRadius).RadiusBottomLeft(tokens.TrackRadius)
		trackBase = trackBase.FillWidth().Height(tokens.TrackThickness)
	}

	thumbBase := style.Style{}.
		Width(thumbWidth).
		Height(thumbHeight).
		Background(style.Solid(enabled.ThumbBackground)).
		BorderWidth(borderWidth).
		BorderColor(style.Solid(enabled.ThumbBorder)).
		Radius(thumbRadius).
		BoxShadow(0, 2, 0, 2, style.Solid(color.NRGBA{A: shadowAlpha(0x0a, enabled.ThumbBackground)})).
		BoxShadow(0, 1, 0, 1, style.Solid(color.NRGBA{A: shadowAlpha(0x0f, enabled.ThumbBackground)})).
		Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(enabled.Focus))

	thumbDisabled := style.Style{}.
		Background(style.Solid(disabled.ThumbBackground)).
		BorderColor(style.Solid(disabled.ThumbBorder)).
		BoxShadowNone().
		BoxShadow(0, 2, 0, 2, style.Solid(color.NRGBA{A: shadowAlpha(0x0a, disabled.ThumbBackground)})).
		BoxShadow(0, 1, 0, 1, style.Solid(color.NRGBA{A: shadowAlpha(0x0f, disabled.ThumbBackground)})).
		Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(disabled.Focus))

	sem := active.Semantic()
	labelBase := style.Style{}.
		TextColor(style.Solid(sem.Foreground)).
		FontSize(tokens.TextSize).
		FontWeight(int(font.Medium))
	disabledRow := func(set style.Style) []style.Rule {
		return []style.Rule{{When: style.Disabled, Set: set}}
	}
	return style.Recipe{
		style.PartTrack: {
			Base:  trackBase,
			Rules: disabledRow(style.Style{}.Background(style.Solid(disabled.Track))),
		},
		style.PartFill: {
			Base:  fillBase,
			Rules: disabledRow(style.Style{}.Background(style.Solid(disabled.Fill))),
		},
		style.PartThumb: {
			Base:  thumbBase,
			Rules: disabledRow(thumbDisabled),
		},
		style.PartLabel: {
			Base:  labelBase,
			Rules: disabledRow(style.Style{}.TextColor(style.Solid(active.DisabledColor(sem.Foreground)))),
		},
	}
}

type palette struct {
	Track, Fill, ThumbBorder, ThumbBackground, Focus color.NRGBA
}

func colors(active *theme.Theme, disabled bool) palette {
	sem := active.Semantic()
	value := palette{
		Track:           sem.SurfaceRaised,
		Fill:            sem.Tone(theme.TonePrimary).Fill,
		ThumbBorder:     sem.Tone(theme.TonePrimary).Fill,
		ThumbBackground: sem.Tone(theme.TonePrimary).OnFill,
		Focus:           sem.Ring,
	}
	if disabled {
		value.Track = active.DisabledColor(value.Track)
		value.Fill = active.DisabledColor(value.Fill)
		value.ThumbBorder = active.DisabledColor(value.ThumbBorder)
		value.ThumbBackground = active.DisabledColor(value.ThumbBackground)
		value.Focus = color.NRGBA{}
	}
	return value
}

func shadowAlpha(alpha uint8, background color.NRGBA) uint8 {
	return uint8(uint16(alpha) * uint16(background.A) / 0xff)
}
