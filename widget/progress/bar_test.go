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
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return *testtheme.Behavior()
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

func TestBarRatio(t *testing.T) {
	if got := Bar("upload", 50).ratio(); got != 0.5 {
		t.Fatalf("ratio = %v, want 0.5", got)
	}
	if got := Bar("upload", -10).ratio(); got != 0 {
		t.Fatalf("low ratio = %v, want 0", got)
	}
	if got := Bar("upload", 140).ratio(); got != 1 {
		t.Fatalf("high ratio = %v, want 1", got)
	}
	if got := Bar("upload", 75).Range(50, 100).ratio(); got != 0.5 {
		t.Fatalf("range ratio = %v, want 0.5", got)
	}
	if got := Bar("upload", 75).Range(10, 10).ratio(); got != 0 {
		t.Fatalf("invalid range ratio = %v, want 0", got)
	}
	if got := Bar("upload", math.NaN()).ratio(); got != 0 {
		t.Fatalf("nan ratio = %v, want 0", got)
	}
	if got := Bar("upload", math.Inf(1)).ratio(); got != 0 {
		t.Fatalf("inf ratio = %v, want 0", got)
	}
}

func TestBarOutputText(t *testing.T) {
	if got := Bar("upload", 42).outputText(); got != "" {
		t.Fatalf("default output = %q, want empty", got)
	}
	if got := Bar("upload", 42).ShowValue().outputText(); got != "42%" {
		t.Fatalf("percent output = %q, want 42%%", got)
	}
	if got := Bar("upload", 42).ValueText("42 files").outputText(); got != "42 files" {
		t.Fatalf("custom output = %q, want 42 files", got)
	}
	if got := Bar("upload", 42).ShowValue().Indeterminate().outputText(); got != "" {
		t.Fatalf("indeterminate output = %q, want empty", got)
	}
}

func TestBarSemanticDescription(t *testing.T) {
	if got := Bar("upload", 42).Label("Upload").semanticDescription(); got != "Upload 42%" {
		t.Fatalf("semantic description = %q, want Upload 42%%", got)
	}
	if got := Bar("upload", 42).ValueText("42 files").semanticDescription(); got != "Progress 42 files" {
		t.Fatalf("custom semantic description = %q, want Progress 42 files", got)
	}
	if got := Bar("upload", 0).Label("Sync").Indeterminate().semanticDescription(); got != "Sync in progress" {
		t.Fatalf("indeterminate semantic description = %q, want Sync in progress", got)
	}
}

// TestBarSizeStyle reads the size layer off the widget's own seam rather than
// calling BarSizeDeclaration directly (§1.9 (e)). The size layer has no rule table
// -- nothing about the geometry answers to a state -- so it is the one layer where
// the seam call and the declaration are the same function, and going through the
// seam still buys the thing that matters: the test fails if the seam stops
// installing it.
func TestBarSizeStyle(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	small := Bar("upload", 50).Size(theme.ProgressBarSmall).resolveStyle(ctx, testLayoutContext(), "upload", &progressBarVisualCache{}).track
	large := Bar("upload", 50).Size(theme.ProgressBarLarge).resolveStyle(ctx, testLayoutContext(), "upload", &progressBarVisualCache{}).track
	if got := small.Box.Height; got == nil || *got != activeTheme.Components.ProgressBar.SmallHeight {
		t.Fatalf("small height = %v, want %v", got, activeTheme.Components.ProgressBar.SmallHeight)
	}
	if got := large.Box.Height; got == nil || *got != activeTheme.Components.ProgressBar.LargeHeight {
		t.Fatalf("large height = %v, want %v", got, activeTheme.Components.ProgressBar.LargeHeight)
	}
}

func TestBarSeparatesRootTrackAndIndicatorStyles(t *testing.T) {
	rootColor := color.NRGBA{R: 1, A: 0xff}
	trackColor := color.NRGBA{G: 2, A: 0xff}
	fillColor := color.NRGBA{B: 3, A: 0xff}
	custom := style.Style{}.
		Height(60).
		Background(style.SolidColor{Color: rootColor}).
		Part(style.PartTrack, style.Style{}.Height(4).Background(style.SolidColor{Color: trackColor})).
		Part(style.PartFill, style.Style{}.Background(style.SolidColor{Color: fillColor}))

	resolved := Bar("upload", 50).Style(custom).resolveStyle(newContext(nil), testLayoutContext(), "upload", &progressBarVisualCache{}, &theme.ProgressOverride{Declaration: custom})
	if resolved.root.Box == nil || resolved.root.Box.Height == nil || *resolved.root.Box.Height != 60 {
		t.Fatalf("root height = %#v", resolved.root.Box)
	}
	if got := resolved.root.Paint.Background.(style.SolidColor).Color; got != rootColor {
		t.Fatalf("root background = %#v", got)
	}
	if resolved.track.Box == nil || resolved.track.Box.Height == nil || *resolved.track.Box.Height != 4 {
		t.Fatalf("track box = %#v", resolved.track.Box)
	}
	if got := resolved.track.Paint.Background.(style.SolidColor).Color; got != trackColor {
		t.Fatalf("track background = %#v", got)
	}
	if got := resolved.fill.Paint.Background.(style.SolidColor).Color; got != fillColor {
		t.Fatalf("fill background = %#v", got)
	}
}

func TestBarAnimation(t *testing.T) {
	state := new(progressBarState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start
	metrics := testProgressBarMetrics(t)
	duration := metrics.ValueDuration

	if got := state.progress(gtx, 0, false, duration, testtheme.Behavior().Motion); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	if got := state.progress(gtx, 1, false, duration, testtheme.Behavior().Motion); got != 0 {
		t.Fatalf("animation start = %v, want 0", got)
	}

	gtx.Now = start.Add(duration / 2)
	mid := state.progress(gtx, 1, false, duration, testtheme.Behavior().Motion)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("animation midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(duration)
	if got := state.progress(gtx, 1, false, duration, testtheme.Behavior().Motion); got != 1 {
		t.Fatalf("animation end = %v, want 1", got)
	}
}

func TestBarIndeterminateOffset(t *testing.T) {
	width := 40
	start := time.Unix(1, 0)
	period := testProgressBarMetrics(t).IndeterminatePeriod
	if got := progressBarIndeterminateOffset(time.Time{}, width, period); got != -width {
		t.Fatalf("zero time offset = %d, want %d", got, -width)
	}
	first := progressBarIndeterminateOffset(start, width, period)
	mid := progressBarIndeterminateOffset(start.Add(period/2), width, period)
	if first == mid {
		t.Fatal("indeterminate offset did not advance")
	}
	if got := progressBarIndeterminatePosition(start, width, 0); got != 0 {
		t.Fatalf("static indeterminate offset = %d, want 0", got)
	}
}

func testProgressBarMetrics(t *testing.T) theme.ProgressBarMetrics {
	t.Helper()
	active := testtheme.Behavior()
	provider, err := theme.ComponentProvider(active)
	if err != nil {
		t.Fatal(err)
	}
	return provider.ProgressBar(active, theme.ProgressBarRequest{}).Metrics
}

func TestIndeterminateBarRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := *testtheme.Behavior()
		themeValue.Motion.Enabled = enabled
		ctx := core.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		Bar("sync", 0).Indeterminate().Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(200, 100)},
			Source:      router.Source(),
			Ops:         &ops,
			Now:         time.Unix(1, 0),
		})
		router.Frame(&ops)
		_, wake := router.WakeupTime()
		return wake
	}
	if !wakes(true) {
		t.Fatal("enabled indeterminate progress bar did not request redraw")
	}
	if wakes(false) {
		t.Fatal("indeterminate progress bar requested redraw with motion disabled")
	}
}

func TestBarLayoutTrackOnly(t *testing.T) {
	dims := Bar("upload", 50).Layout(newContext(nil), progressBarTestContext())
	if dims.Size != image.Pt(300, 8) {
		t.Fatalf("track-only size = %v, want (300,8)", dims.Size)
	}
}

func TestBarLayoutWithHeader(t *testing.T) {
	dims := Bar("upload", 50).
		Label("Upload").
		ShowValue().
		Layout(newContext(nil), progressBarTestContext())
	if dims.Size.X != 300 {
		t.Fatalf("header width = %d, want 300", dims.Size.X)
	}
	if dims.Size.Y <= 8 {
		t.Fatalf("header height = %d, want greater than track height", dims.Size.Y)
	}
}

func TestBarLabelPartUsesCommonBoxRenderer(t *testing.T) {
	base := Bar("upload", 50).
		Label("Upload").
		Layout(newContext(nil), progressBarTestContext())
	styled := (Bar("upload", 50).Label("Upload")).Style(style.Style{}.Part(style.PartLabel, style.Style{}.PaddingY(7))).Layout(newContext(nil), progressBarTestContext())
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled label height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

func TestBarTrackPartUsesCommonBoxRenderer(t *testing.T) {
	base := Bar("upload", 50).Layout(newContext(nil), progressBarTestContext())
	styled := (Bar("upload", 50)).Style(style.Style{}.Part(style.PartTrack, style.Style{}.MarginY(7))).Layout(newContext(nil), progressBarTestContext())
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled track height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

// paletteTheme installs a theme's v1 palette as its colour source, so edits to
func progressBarTestContext() layout.Context {
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 120)},
		Ops:         &ops,
	}
}
