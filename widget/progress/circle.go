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
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type CircleWidget struct {
	override      *theme.ProgressOverride
	key           string
	value         float64
	minValue      float64
	maxValue      float64
	label         string
	valueText     string
	hasValueText  bool
	indeterminate bool
	color         CircleColor
	size          CircleSize
	disabled      bool
}

type CircleColor = BarColor

const (
	CircleAccent  CircleColor = BarAccent
	CircleDefault CircleColor = BarDefault
	CircleSuccess CircleColor = BarSuccess
	CircleWarning CircleColor = BarWarning
	CircleDanger  CircleColor = BarDanger
)

type CircleSize = BarSize

const (
	CircleMedium CircleSize = BarMedium
	CircleSmall  CircleSize = BarSmall
	CircleLarge  CircleSize = BarLarge
)

const stateSlotProgressCircle = "progress-circle"

func Circle(key string, value float64) CircleWidget {
	return CircleWidget{key: key, value: value, maxValue: 100}
}

func (p CircleWidget) Label(label string) CircleWidget {
	p.label = label
	return p
}

func (p CircleWidget) ValueText(text string) CircleWidget {
	p.valueText = text
	p.hasValueText = true
	return p
}

func (p CircleWidget) Range(minValue, maxValue float64) CircleWidget {
	p.minValue = minValue
	p.maxValue = maxValue
	return p
}

func (p CircleWidget) Indeterminate() CircleWidget {
	p.indeterminate = true
	return p
}

func (p CircleWidget) Color(circleColor CircleColor) CircleWidget {
	p.color = circleColor
	return p
}

func (p CircleWidget) Size(size CircleSize) CircleWidget {
	p.size = size
	return p
}

func (p CircleWidget) Disabled(disabled bool) CircleWidget {
	p.disabled = disabled
	return p
}

func (p CircleWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithOverride(ctx, gtx, p.override)
}

func (p CircleWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ProgressOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	key := core.ClaimKey(ctx, core.KindProgressCircle, p.key)
	progressState := progressCircleStateFor(ctx, key)
	resolved := p.resolveStyle(ctx, gtx, key, &progressState.circleStyleCache, override)
	progress := progressState.progress(gtx, p.ratio(), p.indeterminate, resolved.metrics.ValueDuration, resolved.motion)
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		diameter := max(min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 1)
		size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
		iconSize := min(size.X, size.Y)
		if iconSize > 0 {
			offset := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
			period := time.Duration(0)
			if !p.disabled {
				period = theme.ResolveMotionDuration(resolved.motion, resolved.metrics.SpinDuration)
			}
			resolved.renderer(gtx, theme.ProgressCircleRenderRequest{
				Size:          gtx.Metric.PxToDp(iconSize),
				StrokeRatio:   resolved.metrics.StrokeRatio,
				Track:         resolved.track,
				Fill:          resolved.fill,
				Progress:      progress,
				Indeterminate: p.indeterminate,
				Period:        period,
			})
			if p.indeterminate && period > 0 {
				gtx.Execute(op.InvalidateCmd{})
			}
			offset.Pop()
		}
		clipped := clip.Rect{Max: size}.Push(gtx.Ops)
		semantic.EnabledOp(!p.disabled).Add(gtx.Ops)
		semantic.DescriptionOp(p.semanticDescriptionFor(ctx)).Add(gtx.Ops)
		clipped.Pop()
		return layout.Dimensions{Size: size}
	}))
}

func (p CircleWidget) ratio() float32 {
	return progressRatio(p.value, p.minValue, p.maxValue, p.indeterminate)
}

func (p CircleWidget) semanticDescription() string {
	return p.semanticDescriptionFor(nil)
}

func (p CircleWidget) semanticDescriptionFor(ctx *core.Context) string {
	label := p.label
	if label == "" {
		label = ctx.Text(i18n.ProgressDefault)
	}
	if p.indeterminate {
		return ctx.Textf(i18n.ProgressInProgress, label)
	}
	value := p.valueText
	if !p.hasValueText || value == "" {
		value = fmt.Sprintf("%.0f%%", p.ratio()*100)
	}
	return ctx.Textf(i18n.ProgressValue, label, value)
}

func progressCircleStateFor(ctx *core.Context, key string) *progressBarState {
	return core.UseState[progressBarState](ctx, key, stateSlotProgressCircle)
}

type progressCircleResolvedStyle struct {
	root     style.ResolvedStyle
	track    color.NRGBA
	fill     color.NRGBA
	renderer theme.ProgressCircleRenderer
	metrics  theme.ProgressCircleMetrics
	motion   theme.MotionTheme
	override style.Style
}

func (p CircleWidget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, cache *progressCircleVisualCache, override *theme.ProgressOverride) progressCircleResolvedStyle {
	state := style.State{Disabled: p.disabled}
	activeTheme := core.ActiveTheme(ctx)
	requestKey := progressDeclarationKey{color: p.color, size: p.size}
	var visual theme.ProgressCircleVisual
	if override == nil {
		visual = cache.resolve(activeTheme, requestKey)
	} else {
		visual = resolveProgressCircleProvider(activeTheme, requestKey, override)
	}
	parts := [...]style.Part{style.PartRoot, style.PartTrack, style.PartFill}
	var resolved [len(parts)]style.ResolvedStyle
	computed.ResolvePartsInto(ctx, gtx, key, state, parts[:], resolved[:], visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	if visual.Color == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	track, ok := visual.Color(resolved[1])
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	fill, ok := visual.Color(resolved[2])
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return progressCircleResolvedStyle{
		root: resolved[0], track: track, fill: fill, renderer: visual.Renderer,
		metrics: visual.Metrics, motion: visual.Motion, override: visual.Override,
	}
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w CircleWidget) Style(values ...style.Style) CircleWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ProgressOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
