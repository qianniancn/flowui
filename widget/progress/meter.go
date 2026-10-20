package progress

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// MeterWidget keeps meter semantics while reusing the linear progress control.
type MeterWidget struct {
	bar            BarWidget
	alt            string
	valueFormatter func(float64) string
}

type MeterColor = BarColor

const (
	MeterAccent  = BarAccent
	MeterDefault = BarDefault
	MeterSuccess = BarSuccess
	MeterWarning = BarWarning
	MeterDanger  = BarDanger
)

type MeterSize = BarSize

const (
	MeterMedium = BarMedium
	MeterSmall  = BarSmall
	MeterLarge  = BarLarge
)

func Meter(key string, value float64) MeterWidget {
	return MeterWidget{bar: Bar(key, value)}
}

func (m MeterWidget) Label(label string) MeterWidget {
	m.bar = m.bar.Label(label)
	return m
}

func (m MeterWidget) Alt(alt string) MeterWidget {
	m.alt = alt
	return m
}

func (m MeterWidget) ShowValue() MeterWidget {
	m.bar = m.bar.ShowValue()
	return m
}

func (m MeterWidget) ValueText(text string) MeterWidget {
	m.bar = m.bar.ValueText(text)
	return m
}

func (m MeterWidget) ValueFormatter(formatter func(float64) string) MeterWidget {
	m.valueFormatter = formatter
	m.bar.showValue = true
	return m
}

func (m MeterWidget) Range(minValue, maxValue float64) MeterWidget {
	m.bar = m.bar.Range(minValue, maxValue)
	return m
}

func (m MeterWidget) Color(value MeterColor) MeterWidget {
	m.bar = m.bar.Color(value)
	return m
}

// Fill overrides the meter's fill colour with a concrete colour, for series
// colours that are not one of the semantic MeterColor values.
func (m MeterWidget) Fill(value color.NRGBA) MeterWidget {
	m.bar = m.bar.Fill(value)
	return m
}

func (m MeterWidget) Size(size MeterSize) MeterWidget {
	m.bar = m.bar.Size(size)
	return m
}

func (m MeterWidget) Disabled(disabled bool) MeterWidget {
	m.bar = m.bar.Disabled(disabled)
	return m
}

func (m MeterWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return m.progressBar().Layout(ctx, gtx)
}

func (m MeterWidget) progressBar() BarWidget {
	bar := m.bar
	bar.semanticLabel = m.alt
	if bar.semanticLabel == "" && bar.label == "" {
		bar.semanticKey = i18n.MeterDefault
	}
	if !bar.hasValueText && m.valueFormatter != nil {
		bar = bar.ValueText(m.valueFormatter(bar.value))
	}
	return bar
}

// Style merges instance declarations into the meter.
func (m MeterWidget) Style(values ...style.Style) MeterWidget {
	m.bar = m.bar.Style(values...)
	return m
}
