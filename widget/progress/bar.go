package progress

import (
	"fmt"
	"image/color"
	"math"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type BarWidget struct {
	override      *theme.ProgressOverride
	key           string
	value         float64
	minValue      float64
	maxValue      float64
	label         string
	semanticLabel string
	semanticKey   i18n.Key
	valueText     string
	hasValueText  bool
	showValue     bool
	indeterminate bool
	color         BarColor
	size          BarSize
	disabled      bool
}

type BarColor = theme.ProgressBarColor

const (
	BarAccent  = theme.ProgressBarAccent
	BarDefault = theme.ProgressBarDefault
	BarSuccess = theme.ProgressBarSuccess
	BarWarning = theme.ProgressBarWarning
	BarDanger  = theme.ProgressBarDanger
)

type BarSize = theme.ProgressBarSize

const (
	BarMedium = theme.ProgressBarMedium
	BarSmall  = theme.ProgressBarSmall
	BarLarge  = theme.ProgressBarLarge
)

func Bar(key string, value float64) BarWidget {
	return BarWidget{
		key:      key,
		value:    value,
		maxValue: 100,
	}
}

func (p BarWidget) Label(label string) BarWidget {
	p.label = label
	return p
}

func (p BarWidget) ShowValue() BarWidget {
	p.showValue = true
	return p
}

func (p BarWidget) ValueText(text string) BarWidget {
	p.valueText = text
	p.hasValueText = true
	p.showValue = true
	return p
}

func (p BarWidget) Range(minValue, maxValue float64) BarWidget {
	p.minValue = minValue
	p.maxValue = maxValue
	return p
}

func (p BarWidget) Indeterminate() BarWidget {
	p.indeterminate = true
	return p
}

func (p BarWidget) Color(color BarColor) BarWidget {
	p.color = color
	return p
}

// Fill overrides the semantic fill colour with a concrete colour. It is useful
// for data-driven bars whose colour identifies a series rather than a state.
func (p BarWidget) Fill(value color.NRGBA) BarWidget {
	return p.Style(style.Style{}.Part(style.PartFill, style.Style{}.
		Background(style.Solid(value))))
}

func (p BarWidget) Size(size BarSize) BarWidget {
	p.size = size
	return p
}

func (p BarWidget) Disabled(disabled bool) BarWidget {
	p.disabled = disabled
	return p
}

func (p BarWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithOverride(ctx, gtx, p.override)
}

func (p BarWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ProgressOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	key := core.ClaimKey(ctx, core.KindProgressBar, p.key)
	progressState := progressBarStateFor(ctx, key)
	resolved := p.resolveStyle(ctx, gtx, key, &progressState.barStyleCache, override)
	progress := progressState.progress(gtx, p.ratio(), p.indeterminate, resolved.metrics.ValueDuration, resolved.motion)
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		dims := p.layout(ctx, gtx, resolved, progress)
		clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		p.addSemantic(ctx, gtx)
		clipStack.Pop()
		return dims
	}))
}

func (p BarWidget) ratio() float32 {
	return progressRatio(p.value, p.minValue, p.maxValue, p.indeterminate)
}

func progressRatio(value, minValue, maxValue float64, indeterminate bool) float32 {
	if indeterminate || maxValue <= minValue {
		return 0
	}
	ratio := (value - minValue) / (maxValue - minValue)
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		return 0
	}
	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return float32(ratio)
}

func (p BarWidget) outputText() string {
	if p.hasValueText {
		return p.valueText
	}
	if !p.showValue || p.indeterminate {
		return ""
	}
	return fmt.Sprintf("%.0f%%", p.ratio()*100)
}

func (p BarWidget) addSemantic(ctx *core.Context, gtx layout.Context) {
	semantic.EnabledOp(!p.disabled).Add(gtx.Ops)
	if description := p.semanticDescriptionFor(ctx); description != "" {
		semantic.DescriptionOp(description).Add(gtx.Ops)
	}
}

func (p BarWidget) semanticDescription() string {
	return p.semanticDescriptionFor(nil)
}

func (p BarWidget) semanticDescriptionFor(ctx *core.Context) string {
	label := p.semanticLabel
	if p.semanticKey != "" {
		label = ctx.Text(p.semanticKey)
	}
	if label == "" {
		if p.label != "" {
			label = p.label
		} else {
			label = ctx.Text(i18n.ProgressDefault)
		}
	}
	if p.indeterminate {
		return ctx.Textf(i18n.ProgressInProgress, label)
	}
	value := p.outputText()
	if value == "" {
		value = fmt.Sprintf("%.0f%%", p.ratio()*100)
	}
	return ctx.Textf(i18n.ProgressValue, label, value)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w BarWidget) Style(values ...style.Style) BarWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ProgressOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
