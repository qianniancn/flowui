package progress

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestProgressCircleRatioAndSemantics(t *testing.T) {
	if got := ProgressCircle("upload", 50).ratio(); got != .5 {
		t.Fatalf("ratio = %v, want .5", got)
	}
	if got := ProgressCircle("upload", 75).Range(50, 100).ratio(); got != .5 {
		t.Fatalf("custom range ratio = %v, want .5", got)
	}
	if got := ProgressCircle("upload", math.NaN()).ratio(); got != 0 {
		t.Fatalf("NaN ratio = %v, want 0", got)
	}
	if got := ProgressCircle("upload", 42).Label("Upload").semanticDescription(); got != "Upload 42%" {
		t.Fatalf("description = %q", got)
	}
	if got := ProgressCircle("upload", 42).ValueText("42 files").semanticDescription(); got != "Progress 42 files" {
		t.Fatalf("custom description = %q", got)
	}
	if got := ProgressCircle("upload", 42).ValueText("").semanticDescription(); got != "Progress 42%" {
		t.Fatalf("empty custom description = %q", got)
	}
	if got := ProgressCircle("upload", 0).Label("Sync").Indeterminate().semanticDescription(); got != "Sync in progress" {
		t.Fatalf("indeterminate description = %q", got)
	}
}

func TestProgressCircleMatchesHeroUISizes(t *testing.T) {
	activeTheme := DefaultTheme()
	tokens := activeTheme.Components.ProgressCircle
	if tokens.SmallSize != 20 || tokens.MediumSize != 28 || tokens.LargeSize != 36 || tokens.StrokeRatio != 4.0/36.0 {
		t.Fatalf("ProgressCircle theme = %+v", tokens)
	}
	geometry, ok := resolveProgressCircleGeometry(36, tokens.StrokeRatio)
	if !ok || geometry.outerRadius != 18 || geometry.innerRadius != 14 || geometry.outerRadius-geometry.innerRadius != 4 {
		t.Fatalf("ProgressCircle geometry = %+v/%v", geometry, ok)
	}
}

func TestProgressCircleStyleUsesPaletteAndDisabledOpacity(t *testing.T) {
	activeTheme := DefaultTheme()
	activeTheme.Palette.Success = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	style := progressCircleStyleFor(&activeTheme, ProgressCircleSuccess, false)
	if style.track != activeTheme.Palette.DefaultColor() || style.fill != activeTheme.Palette.Success {
		t.Fatalf("ProgressCircle style = %+v", style)
	}
	activeTheme.DisabledOpacity = .25
	style = progressCircleStyleFor(&activeTheme, ProgressCircleDanger, true)
	if style.fill.A != byte(float32(activeTheme.Palette.Danger.A)*.25) {
		t.Fatalf("disabled fill alpha = %d", style.fill.A)
	}
}

func TestProgressCirclePhase(t *testing.T) {
	start := time.Unix(1, 0)
	if got := progressCirclePhase(time.Time{}, progressCircleSpinDuration); got != 0 {
		t.Fatalf("zero phase = %v", got)
	}
	first := progressCirclePhase(start, progressCircleSpinDuration)
	middle := progressCirclePhase(start.Add(progressCircleSpinDuration/2), progressCircleSpinDuration)
	if first == middle || math.Abs(float64(middle-first)-math.Pi) > .001 {
		t.Fatalf("phases = %v/%v", first, middle)
	}
	if got := progressCirclePhase(start, 0); got != 0 {
		t.Fatalf("disabled motion phase = %v, want 0", got)
	}
}

func TestIndeterminateProgressCircleRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := theme.DefaultTheme()
		themeValue.Motion.Enabled = enabled
		ctx := frame.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		ProgressCircle("sync", 0).Indeterminate().Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Source:      router.Source(),
			Ops:         &ops,
			Now:         time.Unix(1, 0),
		})
		router.Frame(&ops)
		_, wake := router.WakeupTime()
		return wake
	}
	if !wakes(true) {
		t.Fatal("enabled indeterminate progress circle did not request redraw")
	}
	if wakes(false) {
		t.Fatal("indeterminate progress circle requested redraw with motion disabled")
	}
}

func TestProgressCircleLayout(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(100, 100)}, Ops: &ops, Now: time.Unix(1, 0)}
	dims := ProgressCircle("upload", 60).Layout(newContext(nil), gtx)
	if dims.Size != image.Pt(28, 28) {
		t.Fatalf("ProgressCircle size = %v, want (28,28)", dims.Size)
	}
}
