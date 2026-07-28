package layoutui

import (
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func TestSplitPaneSeparatesRootTrackAndIndicatorStyles(t *testing.T) {
	root := flowstyle.RGB(0x010000)
	track := flowstyle.RGB(0x000200)
	indicator := flowstyle.RGB(0x000003)
	gradient := flowstyle.LinearGradient(
		flowstyle.ColorStop(0, track),
		flowstyle.ColorStop(1, indicator),
	)
	custom := flowstyle.Style{}.
		Background(root).
		Part(flowstyle.PartTrack, flowstyle.Style{}.
			Width(3).
			Background(gradient).
			BorderWidth(1).
			Radius(2).
			BoxShadow(1, 2, 3, 4, flowstyle.RGBA(0x01020380)).
			Opacity(.5).
			Translate(2, 3).
			Cursor(pointer.CursorGrab)).
		Part(flowstyle.PartIndicator, flowstyle.Style{}.
			Width(5).
			Height(20).
			Background(indicator).
			Radius(2).
			Scale(.9, .8))

	ctx := newContext(nil)
	gtx := testLayoutContext()
	resolvedRoot, resolved := SplitPane("workspace", Spacer(1, 1), Spacer(1, 1)).Style(custom).resolveStyle(ctx, gtx, "workspace", flowstyle.StyleState{})
	if resolvedRoot.Paint == nil || resolvedRoot.Paint.Background != root {
		t.Fatalf("root paint = %#v", resolvedRoot.Paint)
	}
	if resolved.dividerWidth != 3 || resolved.activeWidth != 5 || resolved.handleLength != 20 || resolved.cursor != pointer.CursorGrab {
		t.Fatalf("parts = %#v", resolved)
	}
	if resolved.track.Paint == nil || resolved.track.Paint.Opacity == nil || *resolved.track.Paint.Opacity != .5 || resolved.track.Paint.Border == nil || *resolved.track.Paint.Border.Width != 1 || *resolved.track.Paint.Radius != 2 || len(resolved.track.Paint.Shadows) != 1 {
		t.Fatalf("track paint = %#v", resolved.track.Paint)
	}
	if _, ok := resolved.track.Paint.Background.(flowstyle.StyleGradient); !ok {
		t.Fatalf("track background = %T, want gradient", resolved.track.Paint.Background)
	}
	if resolved.track.Trans == nil || *resolved.track.Trans.TranslateX != 2 || *resolved.track.Trans.TranslateY != 3 {
		t.Fatalf("track transform = %#v", resolved.track.Trans)
	}
	if resolved.indicator.Paint == nil || resolved.indicator.Paint.Background != indicator || *resolved.indicator.Paint.Radius != 2 {
		t.Fatalf("indicator paint = %#v", resolved.indicator.Paint)
	}
	if resolved.indicator.Trans == nil || *resolved.indicator.Trans.ScaleX != .9 || *resolved.indicator.Trans.ScaleY != .8 {
		t.Fatalf("indicator transform = %#v", resolved.indicator.Trans)
	}
}

func TestSplitPaneHorizontalLayout(t *testing.T) {
	first := new(splitPaneProbe)
	second := new(splitPaneProbe)
	dims := SplitPane("workspace", first, second).
		DefaultRatio(.25).
		Layout(newContext(nil), splitPaneTestContext(image.Pt(400, 200)))

	if dims.Size != image.Pt(400, 200) {
		t.Fatalf("split pane size = %v, want (400,200)", dims.Size)
	}
	if first.constraints != layout.Exact(image.Pt(100, 200)) {
		t.Fatalf("first constraints = %#v, want 100x200", first.constraints)
	}
	if second.constraints != layout.Exact(image.Pt(299, 200)) {
		t.Fatalf("second constraints = %#v, want 299x200", second.constraints)
	}
}

func TestSplitPaneVerticalLayoutAndMinimums(t *testing.T) {
	first := new(splitPaneProbe)
	second := new(splitPaneProbe)
	SplitPane("workspace", first, second).
		Vertical().
		DefaultRatio(.1).
		MinFirst(100).
		MinSecond(120).
		Layout(newContext(nil), splitPaneTestContext(image.Pt(200, 300)))

	if first.constraints != layout.Exact(image.Pt(200, 100)) {
		t.Fatalf("first constraints = %#v, want 200x100", first.constraints)
	}
	if second.constraints != layout.Exact(image.Pt(200, 199)) {
		t.Fatalf("second constraints = %#v, want 200x199", second.constraints)
	}
}

func TestSplitPaneControlledRatio(t *testing.T) {
	first := new(splitPaneProbe)
	SplitPane("workspace", first, new(splitPaneProbe)).
		Ratio(.75).
		Layout(newContext(nil), splitPaneTestContext(image.Pt(401, 100)))
	if first.constraints.Max.X != 300 {
		t.Fatalf("controlled first width = %d, want 300", first.constraints.Max.X)
	}
}

func TestSplitPaneConflictingMinimumsShareAvailableSpace(t *testing.T) {
	first := new(splitPaneProbe)
	second := new(splitPaneProbe)
	SplitPane("workspace", first, second).
		MinFirst(80).
		MinSecond(80).
		Layout(newContext(nil), splitPaneTestContext(image.Pt(100, 60)))

	if first.constraints.Max.X != 50 || second.constraints.Max.X != 49 {
		t.Fatalf("conflicting minimum widths = %d/%d, want 50/49", first.constraints.Max.X, second.constraints.Max.X)
	}
}

func TestSplitPaneKeyboardResize(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var changed float32
	value := SplitPane("workspace", new(splitPaneProbe), new(splitPaneProbe)).
		DefaultRatio(.5).
		OnRatioChange(func(ratio float32) { changed = ratio })
	start := time.Unix(1, 0)
	layoutSplitPaneFrame(ctx, router, value, start, image.Pt(300, 100))
	state := testComponentState[splitPaneState](ctx, "workspace", stateSlotSplitPane)
	if state == nil {
		t.Fatal("missing split pane state")
	}
	router.Source().Execute(key.FocusCmd{Tag: state})
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutSplitPaneFrame(ctx, router, value, start.Add(time.Millisecond), image.Pt(300, 100))

	if changed <= .5 {
		t.Fatalf("keyboard ratio = %v, want greater than .5", changed)
	}
}

func TestSplitPanePointerDragAndCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var changed float32
	value := SplitPane("workspace", new(splitPaneProbe), new(splitPaneProbe)).
		DefaultRatio(.5).
		OnRatioChange(func(ratio float32) { changed = ratio })
	start := time.Unix(1, 0)
	layoutSplitPaneFrame(ctx, router, value, start, image.Pt(300, 100))

	center := f32.Pt(150, 50)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: center})
	layoutSplitPaneFrame(ctx, router, value, start.Add(time.Millisecond), image.Pt(300, 100))
	if cursor := router.Cursor(); cursor != pointer.CursorColResize {
		t.Fatalf("split pane cursor = %v, want col-resize", cursor)
	}

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: center})
	layoutSplitPaneFrame(ctx, router, value, start.Add(2*time.Millisecond), image.Pt(300, 100))
	router.Queue(
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(210, 50)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(210, 50)},
	)
	layoutSplitPaneFrame(ctx, router, value, start.Add(3*time.Millisecond), image.Pt(300, 100))

	if math.Abs(float64(changed-.7)) > .02 {
		t.Fatalf("drag ratio = %v, want approximately .7", changed)
	}
	state := testComponentState[splitPaneState](ctx, "workspace", stateSlotSplitPane)
	if state.dragging {
		t.Fatal("split pane remained in dragging state after release")
	}
}

func TestSplitPaneDisabledIgnoresPointer(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	changed := false
	value := SplitPane("workspace", new(splitPaneProbe), new(splitPaneProbe)).
		Disabled(true).
		OnRatioChange(func(float32) { changed = true })
	start := time.Unix(1, 0)
	layoutSplitPaneFrame(ctx, router, value, start, image.Pt(300, 100))
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(150, 50)})
	layoutSplitPaneFrame(ctx, router, value, start.Add(time.Millisecond), image.Pt(300, 100))

	if changed {
		t.Fatal("disabled split pane changed ratio")
	}
	if cursor := router.Cursor(); cursor == pointer.CursorColResize {
		t.Fatal("disabled split pane exposed resize cursor")
	}
}

func TestSplitPaneTheme(t *testing.T) {
	tokens := frame.ActiveTheme(newContext(nil)).Components.SplitPane
	if tokens.DividerWidth != 1 || tokens.HitSize != 12 || tokens.ActiveWidth != 2 || tokens.HandleLength != 32 {
		t.Fatalf("split pane theme = %#v", tokens)
	}
}

type splitPaneProbe struct {
	constraints layout.Constraints
}

func (p *splitPaneProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func splitPaneTestContext(size image.Point) layout.Context {
	return layout.Context{Constraints: layout.Exact(size), Ops: new(op.Ops)}
}

func layoutSplitPaneFrame(ctx *frame.Context, router *input.Router, value SplitPaneWidget, now time.Time, size image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(size), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, size)
	dims := value.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
