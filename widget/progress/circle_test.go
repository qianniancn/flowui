package progress

import (
	"image"
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

func TestCircleRatioAndSemantics(t *testing.T) {
	if got := Circle("upload", 50).ratio(); got != .5 {
		t.Fatalf("ratio = %v, want .5", got)
	}
	if got := Circle("upload", 75).Range(50, 100).ratio(); got != .5 {
		t.Fatalf("custom range ratio = %v, want .5", got)
	}
	if got := Circle("upload", math.NaN()).ratio(); got != 0 {
		t.Fatalf("NaN ratio = %v, want 0", got)
	}
	if got := Circle("upload", 42).Label("Upload").semanticDescription(); got != "Upload 42%" {
		t.Fatalf("description = %q", got)
	}
	if got := Circle("upload", 42).ValueText("42 files").semanticDescription(); got != "Progress 42 files" {
		t.Fatalf("custom description = %q", got)
	}
	if got := Circle("upload", 42).ValueText("").semanticDescription(); got != "Progress 42%" {
		t.Fatalf("empty custom description = %q", got)
	}
	if got := Circle("upload", 0).Label("Sync").Indeterminate().semanticDescription(); got != "Sync in progress" {
		t.Fatalf("indeterminate description = %q", got)
	}
}

func TestIndeterminateCircleRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := *testtheme.Behavior()
		themeValue.Motion.Enabled = enabled
		ctx := core.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		Circle("sync", 0).Indeterminate().Layout(ctx, layout.Context{
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

func TestCircleLayout(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(100, 100)}, Ops: &ops, Now: time.Unix(1, 0)}
	dims := Circle("upload", 60).Layout(newContext(nil), gtx)
	if dims.Size != image.Pt(28, 28) {
		t.Fatalf("ProgressCircle size = %v, want (28,28)", dims.Size)
	}
}

func TestCircleInstanceOverrideAppliesPartOpacity(t *testing.T) {
	declaration := style.Style{}.
		Part(style.PartFill, style.Style{}.Opacity(.5))
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageAuto)
	resolved := Circle("upload", 50).Style(declaration).resolveStyle(
		ctx,
		testLayoutContext(),
		"upload",
		&progressCircleVisualCache{},
		&theme.ProgressOverride{Declaration: declaration},
	)
	if resolved.fill.A != 128 {
		t.Fatalf("circle fill alpha = %d, want 128", resolved.fill.A)
	}
}
