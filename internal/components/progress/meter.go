package progress

import (
	"fmt"
	"math"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

type MeterWidget struct {
	key            string
	value          float64
	minValue       float64
	maxValue       float64
	label          string
	alt            string
	valueText      string
	hasValueText   bool
	showValue      bool
	valueFormatter func(float64) string
	valueContent   frame.Widget
	color          MeterColor
	size           MeterSize
	disabled       bool
}

type MeterColor uint8

const (
	MeterAccent MeterColor = iota
	MeterDefault
	MeterSuccess
	MeterWarning
	MeterDanger
)

type MeterSize uint8

const (
	MeterMedium MeterSize = iota
	MeterSmall
	MeterLarge
)

const stateSlotMeter = "meter"

func Meter(key string, value float64) MeterWidget {
	return MeterWidget{key: key, value: value, maxValue: 100}
}

func (m MeterWidget) Label(label string) MeterWidget {
	m.label = label
	return m
}

func (m MeterWidget) Alt(alt string) MeterWidget {
	m.alt = alt
	return m
}

func (m MeterWidget) ShowValue() MeterWidget {
	m.showValue = true
	return m
}

func (m MeterWidget) ValueText(text string) MeterWidget {
	m.valueText = text
	m.hasValueText = true
	m.showValue = true
	return m
}

func (m MeterWidget) ValueFormatter(formatter func(float64) string) MeterWidget {
	m.valueFormatter = formatter
	m.showValue = true
	return m
}

func (m MeterWidget) ValueContent(content frame.Widget) MeterWidget {
	m.valueContent = content
	m.showValue = true
	return m
}

func (m MeterWidget) Range(minValue, maxValue float64) MeterWidget {
	m.minValue = minValue
	m.maxValue = maxValue
	return m
}

func (m MeterWidget) Color(color MeterColor) MeterWidget {
	m.color = color
	return m
}

func (m MeterWidget) Size(size MeterSize) MeterWidget {
	m.size = size
	return m
}

func (m MeterWidget) Disabled(disabled bool) MeterWidget {
	m.disabled = disabled
	return m
}

func (m MeterWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	meterState := meterStateFor(ctx, m.key)
	style := meterStyleFor(frame.ActiveTheme(ctx), m.color, m.disabled)
	sizeStyle := meterSizeStyleFor(frame.ActiveTheme(ctx), m.size)
	progress := meterState.progress(gtx, m.ratio(), false)
	output := m.outputText()

	macro := op.Record(gtx.Ops)
	dims := m.layout(ctx, gtx, style, sizeStyle, progress, output)
	call := macro.Stop()
	root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	semantic.EnabledOp(!m.disabled).Add(gtx.Ops)
	semantic.DescriptionOp(m.semanticDescription(output)).Add(gtx.Ops)
	call.Add(gtx.Ops)
	root.Pop()
	return dims
}

func meterStateFor(ctx *frame.Context, key string) *progressBarState {
	key = frame.ClaimKey(ctx, state.KindMeter, key)
	return frame.UseState[progressBarState](ctx, key, stateSlotMeter)
}

func (m MeterWidget) ratio() float32 {
	if m.maxValue <= m.minValue {
		return 0
	}
	ratio := (m.value - m.minValue) / (m.maxValue - m.minValue)
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		return 0
	}
	return float32(min(max(ratio, 0), 1))
}

func (m MeterWidget) outputText() string {
	if m.hasValueText {
		return m.valueText
	}
	if m.valueFormatter != nil {
		return m.valueFormatter(m.value)
	}
	return m.defaultOutputText()
}

func (m MeterWidget) defaultOutputText() string {
	return fmt.Sprintf("%.0f%%", m.ratio()*100)
}

func (m MeterWidget) semanticDescription(output string) string {
	label := m.alt
	if label == "" {
		label = m.label
	}
	if label == "" {
		label = "Meter"
	}
	if output == "" {
		output = m.defaultOutputText()
	}
	return label + " " + output
}
