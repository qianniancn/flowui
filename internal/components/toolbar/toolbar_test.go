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
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestToolbarOptionsUseValueSemantics(t *testing.T) {
	base := New(probe{size: image.Pt(20, 10)})
	styled := base.
		Orientation(Vertical).
		Attached(true).
		Disabled(true).
		LoopFocus(false).
		Alt("Editor tools").
		Style(flowstyle.Style{}.Radius(4))
	if base.orientation != Horizontal || base.attached || base.disabled || !base.loopFocus || base.alt != "" {
		t.Fatalf("base Toolbar mutated: %#v", base)
	}
	if styled.orientation != Vertical || !styled.attached || !styled.disabled || styled.loopFocus || styled.alt != "Editor tools" {
		t.Fatalf("styled Toolbar options = %#v", styled)
	}
	if styled.customStyle.Resolve(flowstyle.StyleState{}).Paint == nil || Separator().Style(flowstyle.Style{}.Width(2)).customStyle.Resolve(flowstyle.StyleState{}).Box == nil {
		t.Fatal("Toolbar styles were not retained")
	}
}

func TestToolbarLayoutMatchesHeroUISpacing(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
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
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	toolbar := New(
		button.Button("first", probe{size: image.Pt(8, 8)}).Size(button.ButtonSmall).IconOnly(),
		button.Button("disabled", probe{size: image.Pt(8, 8)}).Size(button.ButtonSmall).IconOnly().Disabled(true),
		button.Button("last", probe{size: image.Pt(8, 8)}).Size(button.ButtonSmall).IconOnly(),
	)
	now := time.Unix(4, 0)
	layoutToolbarFrame(ctx, router, toolbar, now)
	first, _ := frame.PeekState[widget.Clickable](ctx, "first", "clickable")
	last, _ := frame.PeekState[widget.Clickable](ctx, "last", "clickable")
	if first == nil || last == nil {
		t.Fatal("Toolbar button focus targets were not retained")
	}

	router.Source().Execute(key.FocusCmd{Tag: first})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(2*time.Millisecond))
	if !router.Source().Focused(last) {
		t.Fatal("Right Arrow did not skip the disabled Toolbar button")
	}
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(3*time.Millisecond))
	if !router.Source().Focused(first) {
		t.Fatal("Toolbar focus did not loop")
	}
}

func TestToolbarThemeAndSeparatorOrientation(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Toolbar
	if tokens.Gap != 8 || tokens.Padding != 4 || tokens.Radius != 24 || tokens.SeparatorLength != 20 || tokens.SeparatorWidth != 1 {
		t.Fatalf("Toolbar theme = %#v", tokens)
	}
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	if dims := layoutToolbarFrame(ctx, new(input.Router), New(Separator()), time.Unix(5, 0)); dims.Size != image.Pt(1, 20) {
		t.Fatalf("horizontal Toolbar separator = %v, want (1,20)", dims.Size)
	}
	if dims := layoutToolbarFrame(ctx, new(input.Router), New(Separator()).Orientation(Vertical), time.Unix(6, 0)); dims.Size != image.Pt(40, 1) {
		t.Fatalf("vertical Toolbar separator = %v, want (40,1)", dims.Size)
	}
}

func TestToolbarOwnsTransitioningSeparatorIdentity(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	transitioning := func() SeparatorWidget {
		return Separator().Style(flowstyle.Style{}.
			Opacity(1).
			Transition(flowstyle.PropOpacity, time.Second))
	}
	layoutToolbarFrame(ctx, new(input.Router), New(transitioning(), transitioning()), time.Unix(7, 0))
	if got := frame.StateLen(ctx); got != 2 {
		t.Fatalf("separator transition states = %d, want 2", got)
	}
}

type probe struct {
	size image.Point
}

func (p probe) Layout(*frame.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: p.size}
}

func layoutToolbarFrame(ctx *frame.Context, router *input.Router, toolbar Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(400, 240)
	return layoutToolbarFrameWithConstraints(ctx, router, toolbar, now, layout.Constraints{Max: viewport})
}

func layoutToolbarFrameWithConstraints(ctx *frame.Context, router *input.Router, toolbar Widget, now time.Time, constraints layout.Constraints) layout.Dimensions {
	viewport := image.Pt(400, 240)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: constraints,
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := toolbar.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
