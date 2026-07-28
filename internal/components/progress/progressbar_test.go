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
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
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

func TestProgressBarRatio(t *testing.T) {
	if got := ProgressBar("upload", 50).ratio(); got != 0.5 {
		t.Fatalf("ratio = %v, want 0.5", got)
	}
	if got := ProgressBar("upload", -10).ratio(); got != 0 {
		t.Fatalf("low ratio = %v, want 0", got)
	}
	if got := ProgressBar("upload", 140).ratio(); got != 1 {
		t.Fatalf("high ratio = %v, want 1", got)
	}
	if got := ProgressBar("upload", 75).Range(50, 100).ratio(); got != 0.5 {
		t.Fatalf("range ratio = %v, want 0.5", got)
	}
	if got := ProgressBar("upload", 75).Range(10, 10).ratio(); got != 0 {
		t.Fatalf("invalid range ratio = %v, want 0", got)
	}
	if got := ProgressBar("upload", math.NaN()).ratio(); got != 0 {
		t.Fatalf("nan ratio = %v, want 0", got)
	}
	if got := ProgressBar("upload", math.Inf(1)).ratio(); got != 0 {
		t.Fatalf("inf ratio = %v, want 0", got)
	}
}

func TestProgressBarOutputText(t *testing.T) {
	if got := ProgressBar("upload", 42).outputText(); got != "" {
		t.Fatalf("default output = %q, want empty", got)
	}
	if got := ProgressBar("upload", 42).ShowValue().outputText(); got != "42%" {
		t.Fatalf("percent output = %q, want 42%%", got)
	}
	if got := ProgressBar("upload", 42).ValueText("42 files").outputText(); got != "42 files" {
		t.Fatalf("custom output = %q, want 42 files", got)
	}
	if got := ProgressBar("upload", 42).ShowValue().Indeterminate().outputText(); got != "" {
		t.Fatalf("indeterminate output = %q, want empty", got)
	}
}

func TestProgressBarSemanticDescription(t *testing.T) {
	if got := ProgressBar("upload", 42).Label("Upload").semanticDescription(); got != "Upload 42%" {
		t.Fatalf("semantic description = %q, want Upload 42%%", got)
	}
	if got := ProgressBar("upload", 42).ValueText("42 files").semanticDescription(); got != "Progress 42 files" {
		t.Fatalf("custom semantic description = %q, want Progress 42 files", got)
	}
	if got := ProgressBar("upload", 0).Label("Sync").Indeterminate().semanticDescription(); got != "Sync in progress" {
		t.Fatalf("indeterminate semantic description = %q, want Sync in progress", got)
	}
}

func TestProgressBarStyleUsesThemePalette(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.Success = color.NRGBA{R: 1, G: 2, B: 3, A: 255}

	style := progressBarStyleFor(&theme, ProgressBarSuccess, false)

	if style.fill != theme.Palette.Success {
		t.Fatalf("success fill = %#v, want %#v", style.fill, theme.Palette.Success)
	}
}

func TestProgressBarDisabledStyleUsesDisabledOpacity(t *testing.T) {
	theme := DefaultTheme()
	theme.DisabledOpacity = 0.25

	style := progressBarStyleFor(&theme, ProgressBarDanger, true)

	if style.fill.A != byte(float32(theme.Palette.Danger.A)*0.25) {
		t.Fatalf("disabled fill alpha = %d, want %d", style.fill.A, byte(float32(theme.Palette.Danger.A)*0.25))
	}
}

func TestProgressBarSizeStyle(t *testing.T) {
	theme := DefaultTheme()
	if got := progressBarSizeStyleFor(&theme, ProgressBarSmall).height; got != theme.Components.ProgressBar.SmallHeight {
		t.Fatalf("small height = %v, want %v", got, theme.Components.ProgressBar.SmallHeight)
	}
	if got := progressBarSizeStyleFor(&theme, ProgressBarLarge).height; got != theme.Components.ProgressBar.LargeHeight {
		t.Fatalf("large height = %v, want %v", got, theme.Components.ProgressBar.LargeHeight)
	}
}

func TestProgressBarSeparatesRootTrackAndIndicatorStyles(t *testing.T) {
	rootColor := color.NRGBA{R: 1, A: 0xff}
	trackColor := color.NRGBA{G: 2, A: 0xff}
	fillColor := color.NRGBA{B: 3, A: 0xff}
	custom := flowstyle.Style{}.
		Height(60).
		Background(flowstyle.SolidColor{Color: rootColor}).
		Part(flowstyle.PartTrack, flowstyle.Style{}.Height(4).Background(flowstyle.SolidColor{Color: trackColor})).
		Part(flowstyle.PartFill, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: fillColor}))

	resolved := ProgressBar("upload", 50).Style(custom).resolveStyle(newContext(nil), testLayoutContext(), "upload")
	if resolved.root.Box == nil || resolved.root.Box.Height == nil || *resolved.root.Box.Height != 60 {
		t.Fatalf("root height = %#v", resolved.root.Box)
	}
	if got := resolved.root.Paint.Background.(flowstyle.SolidColor).Color; got != rootColor {
		t.Fatalf("root background = %#v", got)
	}
	if resolved.track.Box == nil || resolved.track.Box.Height == nil || *resolved.track.Box.Height != 4 {
		t.Fatalf("track box = %#v", resolved.track.Box)
	}
	if got := resolved.track.Paint.Background.(flowstyle.SolidColor).Color; got != trackColor {
		t.Fatalf("track background = %#v", got)
	}
	if got := resolved.fill.Paint.Background.(flowstyle.SolidColor).Color; got != fillColor {
		t.Fatalf("fill background = %#v", got)
	}
}

func TestProgressBarAnimation(t *testing.T) {
	state := new(progressBarState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.progress(gtx, 0, false); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	if got := state.progress(gtx, 1, false); got != 0 {
		t.Fatalf("animation start = %v, want 0", got)
	}

	gtx.Now = start.Add(progressBarValueDuration / 2)
	mid := state.progress(gtx, 1, false)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("animation midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(progressBarValueDuration)
	if got := state.progress(gtx, 1, false); got != 1 {
		t.Fatalf("animation end = %v, want 1", got)
	}
}

func TestProgressBarIndeterminateOffset(t *testing.T) {
	width := 40
	start := time.Unix(1, 0)
	if got := progressBarIndeterminateOffset(time.Time{}, width, progressBarIndeterminatePeriod); got != -width {
		t.Fatalf("zero time offset = %d, want %d", got, -width)
	}
	first := progressBarIndeterminateOffset(start, width, progressBarIndeterminatePeriod)
	mid := progressBarIndeterminateOffset(start.Add(progressBarIndeterminatePeriod/2), width, progressBarIndeterminatePeriod)
	if first == mid {
		t.Fatal("indeterminate offset did not advance")
	}
	if got := progressBarIndeterminatePosition(start, width, 0); got != 0 {
		t.Fatalf("static indeterminate offset = %d, want 0", got)
	}
}

func TestIndeterminateProgressBarRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := theme.DefaultTheme()
		themeValue.Motion.Enabled = enabled
		ctx := frame.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		ProgressBar("sync", 0).Indeterminate().Layout(ctx, layout.Context{
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

func TestProgressBarLayoutTrackOnly(t *testing.T) {
	dims := ProgressBar("upload", 50).Layout(newContext(nil), progressBarTestContext())
	if dims.Size != image.Pt(300, 8) {
		t.Fatalf("track-only size = %v, want (300,8)", dims.Size)
	}
}

func TestProgressBarLayoutWithHeader(t *testing.T) {
	dims := ProgressBar("upload", 50).
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

func TestProgressBarLabelPartUsesCommonBoxRenderer(t *testing.T) {
	base := ProgressBar("upload", 50).
		Label("Upload").
		Layout(newContext(nil), progressBarTestContext())
	styled := ProgressBar("upload", 50).
		Label("Upload").
		Style(flowstyle.Style{}.Part(flowstyle.PartLabel, flowstyle.Style{}.PaddingY(7))).
		Layout(newContext(nil), progressBarTestContext())
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled label height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

func TestProgressBarTrackPartUsesCommonBoxRenderer(t *testing.T) {
	base := ProgressBar("upload", 50).Layout(newContext(nil), progressBarTestContext())
	styled := ProgressBar("upload", 50).
		Style(flowstyle.Style{}.Part(flowstyle.PartTrack, flowstyle.Style{}.MarginY(7))).
		Layout(newContext(nil), progressBarTestContext())
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled track height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

func progressBarTestContext() layout.Context {
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 120)},
		Ops:         &ops,
	}
}
