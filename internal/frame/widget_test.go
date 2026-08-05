package frame

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/locale"
)

type measurableProbe struct {
	layoutCalls  int
	measureCalls int
	enabled      bool
}

func (p *measurableProbe) Layout(_ *Context, gtx layout.Context) layout.Dimensions {
	p.layoutCalls++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(24, 12))}
}

func (p *measurableProbe) Measure(_ *Context, gtx layout.Context) layout.Dimensions {
	p.measureCalls++
	p.enabled = gtx.Enabled()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(48, 20))}
}

type inputProbe struct {
	consumed bool
}

func (p *inputProbe) Layout(_ *Context, gtx layout.Context) layout.Dimensions {
	_, p.consumed = gtx.Event(key.Filter{Name: key.NameReturn})
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(24, 12))}
}

func TestMeasureWidgetPrefersSideEffectFreeMeasurement(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}
	probe := new(measurableProbe)
	dims := MeasureWidget(ctx, gtx, probe)
	if dims.Size != image.Pt(48, 20) || probe.measureCalls != 1 || probe.layoutCalls != 0 || !probe.enabled {
		t.Fatalf("measurement = %v, calls layout=%d measure=%d", dims.Size, probe.layoutCalls, probe.measureCalls)
	}
}

func TestMeasureWidgetFallbackDoesNotConsumeInput(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	router.Queue(key.Event{Name: key.NameReturn, State: key.Press})
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}
	probe := new(inputProbe)
	MeasureWidget(ctx, gtx, probe)
	if probe.consumed {
		t.Fatal("fallback measurement consumed a keyboard event")
	}
}
