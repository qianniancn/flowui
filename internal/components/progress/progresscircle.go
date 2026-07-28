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
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	statepkg "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
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
	customStyle   flowstyle.Style
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

func (p ProgressCircleWidget) Style(value flowstyle.Style) ProgressCircleWidget {
	p.customStyle = value
	return p
}

func (p ProgressCircleWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	key := frame.ClaimKey(ctx, statepkg.KindProgressCircle, p.key)
	progressState := progressCircleStateFor(ctx, key)
	progress := progressState.progress(gtx, p.ratio(), p.indeterminate, frame.ActiveTheme(ctx).Motion)
	resolved := p.resolveStyle(ctx, gtx, key)
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		diameter := max(min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 1)
		size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
		iconSize := min(size.X, size.Y)
		if iconSize > 0 {
			offset := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
			period := time.Duration(0)
			if !p.disabled {
				period = theme.ResolveMotionDuration(activeTheme.Motion, progressCircleSpinDuration)
			}
			drawProgressCircle(gtx, iconSize, activeTheme.Components.ProgressCircle.StrokeRatio, resolved.visual, progress, p.indeterminate, period)
			offset.Pop()
		}
		clipped := clip.Rect{Max: size}.Push(gtx.Ops)
		semantic.EnabledOp(!p.disabled).Add(gtx.Ops)
		semantic.DescriptionOp(p.semanticDescription()).Add(gtx.Ops)
		clipped.Pop()
		return layout.Dimensions{Size: size}
	}))
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
	return frame.UseState[progressBarState](ctx, key, stateSlotProgressCircle)
}

type progressCircleStyle struct {
	track color.NRGBA
	fill  color.NRGBA
}

type progressCircleResolvedStyle struct {
	root   flowstyle.ResolvedStyle
	visual progressCircleStyle
}

func (p ProgressCircleWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string) progressCircleResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	state := flowstyle.StyleState{Disabled: p.disabled}
	defaults := progressCircleDefaultDeclaration(activeTheme)
	variant := progressCircleVariantDeclaration(activeTheme, p.color, p.disabled)
	size := progressCircleSizeDeclaration(activeTheme, p.size)
	root := styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults,
		variant,
		size,
		p.customStyle,
	)
	track := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartTrack, state, defaults, variant, size, p.customStyle)
	indicator := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartFill, state, defaults, variant, size, p.customStyle)
	result := progressCircleStyleFor(activeTheme, p.color, p.disabled)
	if track.Paint != nil {
		if brush, ok := styleruntime.Brush(track.Paint.Background); ok {
			result.track = brush.ColorAt(.5)
		}
		result.track = styleruntime.ApplyOpacity(result.track, track.Paint.Opacity)
	}
	if indicator.Paint != nil {
		if brush, ok := styleruntime.Brush(indicator.Paint.Background); ok {
			result.fill = brush.ColorAt(.5)
		}
		result.fill = styleruntime.ApplyOpacity(result.fill, indicator.Paint.Opacity)
	}
	return progressCircleResolvedStyle{root: root, visual: result}
}

func progressCircleDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: activeTheme.Palette.DefaultColor()}))

}

func progressCircleVariantDeclaration(activeTheme *theme.Theme, circleColor ProgressCircleColor, disabled bool) flowstyle.Style {
	resolved := progressCircleStyleFor(activeTheme, circleColor, disabled)
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: resolved.track})).
		Part(flowstyle.PartFill, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: resolved.fill}))

}

func progressCircleSizeDeclaration(activeTheme *theme.Theme, size ProgressCircleSize) flowstyle.Style {
	diameter := progressCircleSizeFor(activeTheme, size)
	return flowstyle.Style{}.Width(diameter).Height(diameter).AspectRatio(1)
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
