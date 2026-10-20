package animation_test

import (
	"image"
	"image/color"
	"testing"
	"time"

	giolayout "gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
)

func TestAnimationLayoutUsesPublicWidgetContract(t *testing.T) {
	var _ func(string, ui.Widget) animation.AnimateLayoutWidget = animation.AnimateLayout
}

func TestAnimationMotionThemeMatchesStyleContract(t *testing.T) {
	var motion theme.MotionTheme
	var _ animation.MotionTheme = motion
}

func TestAnimationHelpersAndBuilders(t *testing.T) {
	if got := animation.Progress(50*time.Millisecond, 100*time.Millisecond); got != 0.5 {
		t.Fatalf("progress = %v, want 0.5", got)
	}
	if got := animation.LerpFloat(0, 10, 0.25); got != 2.5 {
		t.Fatalf("lerp float = %v, want 2.5", got)
	}
	if got := animation.LerpRect(image.Rect(0, 0, 10, 10), image.Rect(10, 20, 30, 40), 0.5); got != image.Rect(5, 10, 20, 25) {
		t.Fatalf("lerp rect = %v", got)
	}
	if got := animation.LerpColor(color.NRGBA{R: 0xff, A: 0xff}, color.NRGBA{B: 0xff, A: 0xff}, 0.5); got.R != 0x80 || got.B != 0x80 {
		t.Fatalf("lerp color = %#v", got)
	}
	if got := animation.EaseLinear(0.25); got != 0.25 {
		t.Fatalf("linear easing = %v, want 0.25", got)
	}

	var easing animation.Easing = animation.EaseCubicOut
	_ = easing
	_ = animation.Tween("opacity", 1).
		Initial(0).
		Duration(200 * time.Millisecond).
		Easing(animation.EaseCubicOut).
		Revision(2).
		Spring(animation.SpringSnappy()).
		Disabled(false)
	_ = animation.AnimateRect("bounds", image.Rect(0, 0, 10, 10)).Duration(time.Second).Easing(animation.EaseLinear)
	_ = animation.AnimateLayout("layout", nil).Spring(animation.DefaultSpring()).Disabled(true)
	_ = animation.Timeline("timeline").Keyframe(0, 0).Keyframe(1, 1).Duration(time.Second).Loop(true).Playing(true)
}

func TestComponentLocalTransitionsExposeObservableState(t *testing.T) {
	now := time.Unix(0, 0)
	var ops op.Ops
	gtx := giolayout.Context{Now: now, Ops: &ops}

	var floatTransition animation.FloatTransition
	floatTransition.Initialize(0, now)
	if got := floatTransition.Value(gtx, 1, 0, animation.EaseLinear); got != 1 || !floatTransition.Ready() {
		t.Fatalf("float transition = %v, ready=%v", got, floatTransition.Ready())
	}
	if floatTransition.Target() != 1 || !floatTransition.Increasing() {
		t.Fatal("float transition target metadata is not observable")
	}
	floatTransition.Reset()
	if floatTransition.Ready() {
		t.Fatal("reset transition remained ready")
	}

	var colorTransition animation.ColorTransition
	colorTransition.Set(color.NRGBA{}, color.NRGBA{G: 0xff, A: 0xff}, now)
	if got := colorTransition.Value(gtx, color.NRGBA{G: 0xff, A: 0xff}, 0, animation.EaseLinear); got.G != 0xff {
		t.Fatalf("color transition = %#v", got)
	}
	colorTransition.Reset()
	if colorTransition.Current() != (color.NRGBA{}) {
		t.Fatal("color transition reset did not clear current value")
	}

}
