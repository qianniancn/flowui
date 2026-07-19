package progress

import (
	"fmt"
	"math"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type ProgressBarWidget struct {
	theme         func(*theme.Theme)
	key           string
	value         float64
	minValue      float64
	maxValue      float64
	label         string
	semanticLabel string
	valueText     string
	hasValueText  bool
	showValue     bool
	indeterminate bool
	color         ProgressBarColor
	size          ProgressBarSize
	disabled      bool
}

type ProgressBarColor int

const (
	ProgressBarAccent ProgressBarColor = iota
	ProgressBarDefault
	ProgressBarSuccess
	ProgressBarWarning
	ProgressBarDanger
)

type ProgressBarSize int

const (
	ProgressBarMedium ProgressBarSize = iota
	ProgressBarSmall
	ProgressBarLarge
)

const (
	progressBarValueDuration         = 300 * time.Millisecond
	progressBarIndeterminatePeriod   = 1500 * time.Millisecond
	progressBarIndeterminateFillRate = 0.4
)

func ProgressBar(key string, value float64) ProgressBarWidget {
	return ProgressBarWidget{
		key:      key,
		value:    value,
		maxValue: 100,
	}
}

func (p ProgressBarWidget) Label(label string) ProgressBarWidget {
	p.label = label
	return p
}

func (p ProgressBarWidget) ShowValue() ProgressBarWidget {
	p.showValue = true
	return p
}

func (p ProgressBarWidget) ValueText(text string) ProgressBarWidget {
	p.valueText = text
	p.hasValueText = true
	p.showValue = true
	return p
}

func (p ProgressBarWidget) Range(minValue, maxValue float64) ProgressBarWidget {
	p.minValue = minValue
	p.maxValue = maxValue
	return p
}

func (p ProgressBarWidget) Indeterminate() ProgressBarWidget {
	p.indeterminate = true
	return p
}

func (p ProgressBarWidget) Color(color ProgressBarColor) ProgressBarWidget {
	p.color = color
	return p
}

func (p ProgressBarWidget) Size(size ProgressBarSize) ProgressBarWidget {
	p.size = size
	return p
}

func (p ProgressBarWidget) Disabled(disabled bool) ProgressBarWidget {
	p.disabled = disabled
	return p
}

func (p ProgressBarWidget) Theme(fn func(*theme.Theme)) ProgressBarWidget {
	p.theme = fn
	return p
}

func (p ProgressBarWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, p.theme); restore != nil {
		defer restore()
	}
	state := progressBarStateFor(ctx, p.key)
	style := progressBarStyleFor(frame.ActiveTheme(ctx), p.color, p.disabled)
	sizeStyle := progressBarSizeStyleFor(frame.ActiveTheme(ctx), p.size)
	progress := state.progress(gtx, p.ratio(), p.indeterminate, frame.ActiveTheme(ctx).Motion)

	macro := op.Record(gtx.Ops)
	dims := p.layout(ctx, gtx, style, sizeStyle, progress)
	call := macro.Stop()

	clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	p.addSemantic(gtx)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (p ProgressBarWidget) ratio() float32 {
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

func (p ProgressBarWidget) outputText() string {
	if p.hasValueText {
		return p.valueText
	}
	if !p.showValue || p.indeterminate {
		return ""
	}
	return fmt.Sprintf("%.0f%%", p.ratio()*100)
}

func (p ProgressBarWidget) addSemantic(gtx layout.Context) {
	semantic.EnabledOp(!p.disabled).Add(gtx.Ops)
	if description := p.semanticDescription(); description != "" {
		semantic.DescriptionOp(description).Add(gtx.Ops)
	}
}

func (p ProgressBarWidget) semanticDescription() string {
	label := p.semanticLabel
	if label == "" {
		label = p.label
	}
	if label == "" {
		label = "Progress"
	}
	if p.indeterminate {
		return label + " in progress"
	}
	value := p.outputText()
	if value == "" {
		value = fmt.Sprintf("%.0f%%", p.ratio()*100)
	}
	return label + " " + value
}
