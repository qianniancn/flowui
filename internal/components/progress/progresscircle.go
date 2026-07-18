package progress

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	statepkg "github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type ProgressCircleWidget struct {
	key           string
	value         float64
	minValue      float64
	maxValue      float64
	label         string
	valueText     string
	hasValueText  bool
	indeterminate bool
	color         ProgressCircleColor
	size          ProgressCircleSize
	disabled      bool
}

type ProgressCircleColor = ProgressBarColor

const (
	ProgressCircleAccent  ProgressCircleColor = ProgressBarAccent
	ProgressCircleDefault ProgressCircleColor = ProgressBarDefault
	ProgressCircleSuccess ProgressCircleColor = ProgressBarSuccess
	ProgressCircleWarning ProgressCircleColor = ProgressBarWarning
	ProgressCircleDanger  ProgressCircleColor = ProgressBarDanger
)

type ProgressCircleSize = ProgressBarSize

const (
	ProgressCircleMedium ProgressCircleSize = ProgressBarMedium
	ProgressCircleSmall  ProgressCircleSize = ProgressBarSmall
	ProgressCircleLarge  ProgressCircleSize = ProgressBarLarge
)

const stateSlotProgressCircle = "progress-circle"

const progressCircleSpinDuration = time.Second

func ProgressCircle(key string, value float64) ProgressCircleWidget {
	return ProgressCircleWidget{key: key, value: value, maxValue: 100}
}

func (p ProgressCircleWidget) Label(label string) ProgressCircleWidget {
	p.label = label
	return p
}

func (p ProgressCircleWidget) ValueText(text string) ProgressCircleWidget {
	p.valueText = text
	p.hasValueText = true
	return p
}

func (p ProgressCircleWidget) Range(minValue, maxValue float64) ProgressCircleWidget {
	p.minValue = minValue
	p.maxValue = maxValue
	return p
}

func (p ProgressCircleWidget) Indeterminate() ProgressCircleWidget {
	p.indeterminate = true
	return p
}

func (p ProgressCircleWidget) Color(circleColor ProgressCircleColor) ProgressCircleWidget {
	p.color = circleColor
	return p
}

func (p ProgressCircleWidget) Size(size ProgressCircleSize) ProgressCircleWidget {
	p.size = size
	return p
}

func (p ProgressCircleWidget) Disabled(disabled bool) ProgressCircleWidget {
	p.disabled = disabled
	return p
}

func (p ProgressCircleWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	progressState := progressCircleStateFor(ctx, p.key)
	progress := progressState.progress(gtx, p.ratio(), p.indeterminate, frame.ActiveTheme(ctx).Motion)
	style := progressCircleStyleFor(activeTheme, p.color, p.disabled)
	diameter := max(gtx.Dp(progressCircleSizeFor(activeTheme, p.size)), 1)
	size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
	iconSize := min(diameter, size.X, size.Y)

	macro := op.Record(gtx.Ops)
	if iconSize > 0 {
		offset := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
		period := time.Duration(0)
		if !p.disabled {
			period = theme.ResolveMotionDuration(activeTheme.Motion, progressCircleSpinDuration)
		}
		drawProgressCircle(gtx, iconSize, activeTheme.Components.ProgressCircle.StrokeRatio, style, progress, p.indeterminate, period)
		offset.Pop()
	}
	call := macro.Stop()

	clipped := clip.Rect{Max: size}.Push(gtx.Ops)
	semantic.EnabledOp(!p.disabled).Add(gtx.Ops)
	semantic.DescriptionOp(p.semanticDescription()).Add(gtx.Ops)
	call.Add(gtx.Ops)
	clipped.Pop()
	return layout.Dimensions{Size: size}
}

func (p ProgressCircleWidget) ratio() float32 {
	return progressRatio(p.value, p.minValue, p.maxValue, p.indeterminate)
}

func (p ProgressCircleWidget) semanticDescription() string {
	label := p.label
	if label == "" {
		label = "Progress"
	}
	if p.indeterminate {
		return label + " in progress"
	}
	value := p.valueText
	if !p.hasValueText || value == "" {
		value = fmt.Sprintf("%.0f%%", p.ratio()*100)
	}
	return label + " " + value
}

func progressCircleStateFor(ctx *frame.Context, key string) *progressBarState {
	key = frame.ClaimKey(ctx, statepkg.KindProgressCircle, key)
	return frame.UseState[progressBarState](ctx, key, stateSlotProgressCircle)
}

type progressCircleStyle struct {
	track color.NRGBA
	fill  color.NRGBA
}

func progressCircleStyleFor(activeTheme *theme.Theme, circleColor ProgressCircleColor, disabled bool) progressCircleStyle {
	style := progressCircleStyle{track: activeTheme.Palette.DefaultColor(), fill: activeTheme.Palette.Accent}
	switch circleColor {
	case ProgressCircleDefault:
		style.fill = activeTheme.Palette.DefaultForegroundColor()
	case ProgressCircleSuccess:
		style.fill = activeTheme.Palette.Success
	case ProgressCircleWarning:
		style.fill = activeTheme.Palette.Warning
	case ProgressCircleDanger:
		style.fill = activeTheme.Palette.Danger
	}
	if disabled {
		style.track = activeTheme.DisabledColor(style.track)
		style.fill = activeTheme.DisabledColor(style.fill)
	}
	return style
}

func progressCircleSizeFor(activeTheme *theme.Theme, size ProgressCircleSize) unit.Dp {
	switch size {
	case ProgressCircleSmall:
		return activeTheme.Components.ProgressCircle.SmallSize
	case ProgressCircleLarge:
		return activeTheme.Components.ProgressCircle.LargeSize
	default:
		return activeTheme.Components.ProgressCircle.MediumSize
	}
}
