package toolbar

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/button"
)

func TestToolbarLayoutSpacing(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
	horizontal := New(probe{size: image.Pt(20, 10)}, probe{size: image.Pt(30, 12)})
	if dims := layoutToolbarFrame(ctx, new(input.Router), horizontal, time.Unix(1, 0)); dims.Size != image.Pt(58, 12) {
		t.Fatalf("horizontal Toolbar size = %v, want (58,12)", dims.Size)
	}
	vertical := horizontal.Orientation(Vertical)
	if dims := layoutToolbarFrame(ctx, new(input.Router), vertical, time.Unix(2, 0)); dims.Size != image.Pt(30, 30) {
		t.Fatalf("vertical Toolbar size = %v, want (30,30)", dims.Size)
	}
	attached := horizontal.Attached(true)
	if dims := layoutToolbarFrame(ctx, new(input.Router), attached, time.Unix(3, 0)); dims.Size != image.Pt(66, 20) {
		t.Fatalf("attached Toolbar size = %v, want (66,20)", dims.Size)
	}
	if dims := layoutToolbarFrameWithConstraints(ctx, new(input.Router), attached, time.Unix(3, 0), layout.Exact(image.Pt(300, 100))); dims.Size != image.Pt(66, 20) {
		t.Fatalf("constrained attached Toolbar size = %v, want (66,20)", dims.Size)
	}
}

func TestToolbarKeyboardNavigation(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
	router := new(input.Router)
	toolbar := New(
		button.New("").Key("first").Content(probe{size: image.Pt(8, 8)}).Size(button.Small).IconOnly(),
		button.New("").Key("second").Content(probe{size: image.Pt(8, 8)}).Size(button.Small).IconOnly().Disabled(true),
		button.New("").Key("last").Content(probe{size: image.Pt(8, 8)}).Size(button.Small).IconOnly(),
	)
	now := time.Unix(4, 0)
	layoutToolbarFrame(ctx, router, toolbar, now)
	first, _ := core.PeekState[widget.Clickable](ctx, "first", "clickable")
	last, _ := core.PeekState[widget.Clickable](ctx, "last", "clickable")
	if first == nil || last == nil {
		t.Fatal("New button focus targets were not retained")
	}

	router.Source().Execute(key.FocusCmd{Tag: first})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(2*time.Millisecond))
	if !router.Source().Focused(last) {
		t.Fatal("Right Arrow did not skip the disabled New button")
	}
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(3*time.Millisecond))
	if !router.Source().Focused(first) {
		t.Fatal("New focus did not loop")
	}
}

func TestToolbarSeparatorOrientation(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
	if dims := layoutToolbarFrame(ctx, new(input.Router), New(Separator()), time.Unix(5, 0)); dims.Size != image.Pt(1, 20) {
		t.Fatalf("horizontal Toolbar separator = %v, want (1,20)", dims.Size)
	}
	if dims := layoutToolbarFrame(ctx, new(input.Router), New(Separator()).Orientation(Vertical), time.Unix(6, 0)); dims.Size != image.Pt(40, 1) {
		t.Fatalf("vertical Toolbar separator = %v, want (40,1)", dims.Size)
	}
}

func TestToolbarOwnsTransitioningSeparatorIdentity(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
	transitioning := func() SeparatorWidget {
		return Separator()
	}
	first := (transitioning()).Style(style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, time.Second))
	second := (transitioning()).Style(style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, time.Second))
	layoutToolbarFrame(ctx, new(input.Router), New(first, second), time.Unix(7, 0))
	if got := core.StateLen(ctx); got != 2 {
		t.Fatalf("Separator transition states = %d, want 2", got)
	}
}

type probe struct {
	size image.Point
}

func (p probe) Layout(*core.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: p.size}
}

func layoutToolbarFrame(ctx *core.Context, router *input.Router, toolbar Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(400, 240)
	return layoutToolbarFrameWithConstraints(ctx, router, toolbar, now, layout.Constraints{Max: viewport})
}

func layoutToolbarFrameWithConstraints(ctx *core.Context, router *input.Router, toolbar Widget, now time.Time, constraints layout.Constraints) layout.Dimensions {
	viewport := image.Pt(400, 240)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: constraints,
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	dims := toolbar.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
