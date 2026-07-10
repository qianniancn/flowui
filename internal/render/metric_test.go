package render

import (
	"math"
	"testing"

	"gioui.org/layout"
	"gioui.org/unit"
)

func TestDpFloatPreservesFractionalPixelsAtScale(t *testing.T) {
	gtx := layout.Context{Metric: unit.Metric{PxPerDp: 1.5}}
	got := DpFloat(gtx, unit.Dp(1.8))
	const want = float32(2.7)
	if math.Abs(float64(got-want)) > 1e-6 {
		t.Fatalf("DpFloat() = %v, want %v", got, want)
	}
	if got == float32(gtx.Dp(unit.Dp(1.8))) {
		t.Fatalf("DpFloat() unexpectedly rounded to integer pixels: %v", got)
	}
}

func TestDpFloatUsesIdentityScaleForZeroMetric(t *testing.T) {
	gtx := layout.Context{}
	got := DpFloat(gtx, unit.Dp(1.8))
	const want = float32(1.8)
	if math.Abs(float64(got-want)) > 1e-6 {
		t.Fatalf("DpFloat() = %v, want %v", got, want)
	}
}
