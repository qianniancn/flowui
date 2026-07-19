package progress

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// MeterWidget keeps meter semantics while reusing the linear progress control.
type MeterWidget struct {
	bar            ProgressBarWidget
	alt            string
	valueFormatter func(float64) string
}

type MeterColor = ProgressBarColor

const (
	MeterAccent  = ProgressBarAccent
	MeterDefault = ProgressBarDefault
	MeterSuccess = ProgressBarSuccess
	MeterWarning = ProgressBarWarning
	MeterDanger  = ProgressBarDanger
)

type MeterSize = ProgressBarSize

const (
	MeterMedium = ProgressBarMedium
	MeterSmall  = ProgressBarSmall
	MeterLarge  = ProgressBarLarge
)

func Meter(key string, value float64) MeterWidget {
	return MeterWidget{bar: ProgressBar(key, value)}
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

func (m MeterWidget) Size(size MeterSize) MeterWidget {
	m.bar = m.bar.Size(size)
	return m
}

func (m MeterWidget) Disabled(disabled bool) MeterWidget {
	m.bar = m.bar.Disabled(disabled)
	return m
}

func (m MeterWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return m.progressBar().Layout(ctx, gtx)
}

func (m MeterWidget) progressBar() ProgressBarWidget {
	bar := m.bar
	bar.semanticLabel = m.alt
	if bar.semanticLabel == "" && bar.label == "" {
		bar.semanticLabel = "Meter"
	}
	if !bar.hasValueText && m.valueFormatter != nil {
		bar = bar.ValueText(m.valueFormatter(bar.value))
	}
	return bar
}
