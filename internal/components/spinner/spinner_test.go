package spinner

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

func TestSpinnerColorsUseThemePalette(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.Foreground = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	activeTheme.Palette.Success = color.NRGBA{R: 4, G: 5, B: 6, A: 255}
	activeTheme.Palette.Warning = color.NRGBA{R: 7, G: 8, B: 9, A: 255}
	activeTheme.Palette.Danger = color.NRGBA{R: 10, G: 11, B: 12, A: 255}

	tests := []struct {
		color SpinnerColor
		want  color.NRGBA
	}{
		{SpinnerAccent, activeTheme.Palette.Accent},
		{SpinnerCurrent, activeTheme.Palette.Foreground},
		{SpinnerSuccess, activeTheme.Palette.Success},
		{SpinnerWarning, activeTheme.Palette.Warning},
		{SpinnerDanger, activeTheme.Palette.Danger},
	}
	for _, test := range tests {
		if got := spinnerStyleFor(&activeTheme, test.color).color; got != test.want {
			t.Fatalf("spinner color %d = %#v, want %#v", test.color, got, test.want)
		}
	}
}

func TestSpinnerHeroUISizes(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tests := []struct {
		size SpinnerSize
		want int
	}{
		{SpinnerSmall, 16},
		{SpinnerMedium, 24},
		{SpinnerLarge, 32},
		{SpinnerExtraLarge, 40},
	}
	for _, test := range tests {
		style := spinnerSizeStyleFor(&activeTheme, test.size)
		if got := int(style.diameter); got != test.want {
			t.Fatalf("spinner size %d = %d, want %d", test.size, got, test.want)
		}
	}
}

func TestSpinnerHeroUIGeometry(t *testing.T) {
	geometry, ok := resolveSpinnerGeometry(24, .125, .0625)
	if !ok {
		t.Fatal("HeroUI spinner geometry was rejected")
	}
	if geometry.center.X != 12 || geometry.center.Y != 12 {
		t.Fatalf("spinner center = %v, want (12,12)", geometry.center)
	}
	if geometry.strokeWidth != 3 || geometry.radius != 9 {
		t.Fatalf("spinner stroke=%v radius=%v, want stroke=3 radius=9", geometry.strokeWidth, geometry.radius)
	}
	if outer := geometry.radius + geometry.strokeWidth/2; outer != 10.5 {
		t.Fatalf("spinner outer radius = %v, want 10.5", outer)
	}
}

func TestSpinnerHeroUIArcGapAndGradients(t *testing.T) {
	gap := (spinnerArcs[1].startAngle - spinnerArcs[0].startAngle) * 180 / math.Pi
	if math.Abs(float64(gap-51.5)) > .01 {
		t.Fatalf("spinner arc gap = %v degrees, want 51.5", gap)
	}
	if spinnerArcs[0].startAlpha != 1 || spinnerArcs[0].endAlpha != .55 {
		t.Fatalf("first arc alpha = %v to %v", spinnerArcs[0].startAlpha, spinnerArcs[0].endAlpha)
	}
	if spinnerArcs[1].startAlpha != 0 || spinnerArcs[1].endAlpha != .55 {
		t.Fatalf("second arc alpha = %v to %v", spinnerArcs[1].startAlpha, spinnerArcs[1].endAlpha)
	}
	if spinnerArcs[1].gradientStart <= spinnerArcs[0].gradientStart {
		t.Fatal("transparent arc gradient should begin below the leading arc gradient")
	}
}

func TestSpinnerGeometryFallsBackFromInvalidRatios(t *testing.T) {
	geometry, ok := resolveSpinnerGeometry(24, 0, -1)
	if !ok || geometry.strokeWidth != 3 || geometry.radius != 9 {
		t.Fatalf("fallback geometry = %#v, ok=%v", geometry, ok)
	}
	if _, ok := resolveSpinnerGeometry(0, .125, .0625); ok {
		t.Fatal("zero diameter geometry should be rejected")
	}
	geometry, ok = resolveSpinnerGeometry(24, .4, .2)
	if !ok || geometry.strokeWidth != 3 || geometry.radius != 9 {
		t.Fatalf("incompatible ratio fallback = %#v, ok=%v", geometry, ok)
	}
}

func TestSpinnerLayoutUsesSelectedSize(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	var ops op.Ops
	dims := Spinner().Size(SpinnerLarge).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	})
	if dims.Size != image.Pt(32, 32) {
		t.Fatalf("spinner dimensions = %v, want (32,32)", dims.Size)
	}
}

func TestSpinnerPhaseUsesHeroUIPeriod(t *testing.T) {
	start := time.Unix(1, 0)
	if got := spinnerPhase(time.Time{}, spinnerPeriod); got != 0 {
		t.Fatalf("zero phase = %v, want 0", got)
	}
	if got, want := spinnerPhase(start.Add(spinnerPeriod/4), spinnerPeriod), spinnerPhase(start, spinnerPeriod)+float32(math.Pi/2); math.Abs(float64(got-want)) > 0.0001 {
		t.Fatalf("quarter phase = %v, want %v", got, want)
	}
	if got := spinnerPhase(start.Add(spinnerPeriod), spinnerPeriod); math.Abs(float64(got-spinnerPhase(start, spinnerPeriod))) > 0.0001 {
		t.Fatalf("period phase = %v, want %v", got, spinnerPhase(start, spinnerPeriod))
	}
	if got := spinnerPhase(start, 0); got != 0 {
		t.Fatalf("disabled motion phase = %v, want 0", got)
	}
}

func TestSpinnerRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := theme.DefaultTheme()
		themeValue.Motion.Enabled = enabled
		ctx := frame.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		Spinner().Layout(ctx, layout.Context{
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
		t.Fatal("enabled spinner did not request redraw")
	}
	if wakes(false) {
		t.Fatal("spinner requested redraw with motion disabled")
	}
}

func TestSpinnerAlphaPreservesColor(t *testing.T) {
	base := color.NRGBA{R: 10, G: 20, B: 30, A: 200}
	got := spinnerAlpha(base, .55)
	if got.R != base.R || got.G != base.G || got.B != base.B || got.A != 110 {
		t.Fatalf("spinner alpha = %#v, want RGB preserved and alpha 110", got)
	}
}
