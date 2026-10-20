package spinner

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func TestSpinnerLayoutUsesSelectedSize(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
	var ops op.Ops
	dims := New().Size(Large).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	})
	if dims.Size != image.Pt(32, 32) {
		t.Fatalf("spinner dimensions = %v, want (32,32)", dims.Size)
	}
}

func TestSpinnerRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := *testtheme.Behavior()
		themeValue.Motion.Enabled = enabled
		ctx := core.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		New().Layout(ctx, layout.Context{
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
