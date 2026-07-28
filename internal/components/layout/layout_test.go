package layoutui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

const (
	stateSlotList   = "list"
	stateSlotScroll = "scroll"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

type enabledProbeWidget struct {
	enabled bool
}

type styledBoxProbe struct {
	constraints layout.Constraints
	foreground  color.NRGBA
	background  color.NRGBA
}

func (w *enabledProbeWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.enabled = gtx.Enabled()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func (w *styledBoxProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	w.foreground = ctx.ForegroundColor()
	w.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func testComponentState[T any](ctx *frame.Context, key, slot string) *T {
	value, _ := frame.PeekState[T](ctx, key, slot)
	return value
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

func TestBoxAppliesWidthAndPadding(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := Box(child).Style(flowstyle.Style{}.Width(100).Padding(10)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size.X != 100 {
		t.Fatalf("width = %d, want 100", dims.Size.X)
	}
	if child.constraints.Min.X != 80 || child.constraints.Max.X != 80 {
		t.Fatalf("child width constraints = %v, want exact 80", child.constraints)
	}
}

func TestBoxAllowsExplicitZeroSize(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).Style(flowstyle.Style{}.Width(0).Height(0)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	})

	if dims.Size != (image.Point{}) || child.constraints != layout.Exact(image.Point{}) {
		t.Fatalf("zero-sized Box = %v/%v, want zero size and constraints", dims.Size, child.constraints)
	}
}

func TestInteractiveBoxHandlesClickAndDisabledState(t *testing.T) {
	for _, test := range []struct {
		key      string
		disabled bool
		want     int
	}{
		{key: "enabled", want: 1},
		{key: "disabled", disabled: true},
	} {
		ctx := newContext(nil)
		clickable := new(widget.Clickable)
		clickable.Click()
		frame.UseStateWith(ctx, test.key, "clickable", func() *widget.Clickable { return clickable })
		clicks := 0
		dims := Box(nil).
			Key(test.key).
			Disabled(test.disabled).
			Style(flowstyle.Style{}.
				Width(80).
				Height(30).
				When(flowstyle.Disabled, flowstyle.Style{}.Width(90)),
			).
			OnClick(func() { clicks++ }).
			Layout(ctx, testLayoutContext())

		wantWidth := 80
		if test.disabled {
			wantWidth = 90
		}
		if clicks != test.want || dims.Size != image.Pt(wantWidth, 30) {
			t.Fatalf("%s clicks/size = %d/%v, want %d/(%d,30)", test.key, clicks, dims.Size, test.want, wantWidth)
		}
	}
}

func TestInteractiveBoxSemantics(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 40)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	Box(nil).
		Key("save").
		Label("Save changes").
		Disabled(true).
		Style(flowstyle.Style{}.Width(80).Height(30)).
		OnClick(nil).
		Layout(newContext(nil), gtx)
	router.Frame(&ops)
	node, ok := semanticButton(router.AppendSemantics(nil))
	if !ok || node.Desc.Label != "Save changes" || !node.Desc.Disabled {
		t.Fatalf("interactive Box semantics = %#v", node.Desc)
	}
}

func semanticButton(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			return node, true
		}
		if child, ok := semanticButton(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func TestBoxStyleCascadesScopeAndInstance(t *testing.T) {
	ctx := newContext(nil)
	restore := frame.PushStyle(ctx, flowstyle.Style{}.
		PaddingX(12).
		TextColor(flowstyle.RGB(0x112233)),
	)
	defer restore()
	probe := new(styledBoxProbe)
	instanceForeground := flowstyle.RGB(0x445566)
	instanceBackground := flowstyle.RGB(0x778899)

	// Scope supplies horizontal padding; instance owns size, vertical padding, and colors.
	dims := Box(probe).
		Style(flowstyle.Style{}.
			Width(120).
			PaddingTop(4).
			PaddingBottom(4).
			PaddingLeft(20). // overrides scope PaddingX on the left only
			TextColor(instanceForeground).
			Background(instanceBackground),
		).
		Layout(ctx, layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 100)}, Ops: new(op.Ops)})

	// Padding: L=20 (instance), R=12 (scope), T/B=4 (instance) → child width 120-32=88.
	wantConstraints := layout.Constraints{Min: image.Pt(88, 0), Max: image.Pt(88, 92)}
	if dims.Size.X != 120 || probe.constraints != wantConstraints {
		t.Fatalf("box size/child constraints = %v/%v", dims.Size, probe.constraints)
	}
	if probe.foreground != instanceForeground.Color || probe.background != instanceBackground.Color {
		t.Fatalf("box colors = %#v/%#v", probe.foreground, probe.background)
	}
}

func TestBoxStyleJoinsMultipleStyleCalls(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).
		Style(flowstyle.Style{}.Width(100).Padding(4)).
		Style(flowstyle.Style{}.Width(120).PaddingLeft(20)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 300)},
			Ops:         new(op.Ops),
		})
	if dims.Size.X != 120 {
		t.Fatalf("joined width = %d, want 120", dims.Size.X)
	}
	// Padding all 4 then left 20 → child max width 120-24=96.
	if child.constraints.Max.X != 96 {
		t.Fatalf("joined padding child max X = %d, want 96", child.constraints.Max.X)
	}
}

func TestBoxConditionalStyleTransition(t *testing.T) {
	ctx := newContext(nil)
	start := time.Unix(1, 0)
	from := flowstyle.RGB(0x102030).Color
	to := flowstyle.RGB(0x8090a0).Color
	layoutAt := func(now time.Time, active bool) color.NRGBA {
		probe := new(styledBoxProbe)
		frame.BeginFrame(ctx)
		Box(probe).
			Key("status").
			Style(flowstyle.Style{}.Background(flowstyle.SolidColor{Color: from}).
				Transition(flowstyle.PropBackgroundColor, 100*time.Millisecond).
				WhenIf(active, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: to}))).
			Layout(ctx, layout.Context{
				Constraints: layout.Exact(image.Pt(20, 20)),
				Ops:         new(op.Ops),
				Now:         now,
			})
		frame.EndFrame(ctx)
		return probe.background
	}

	if got := layoutAt(start, false); got != from {
		t.Fatalf("initial background = %#v, want %#v", got, from)
	}
	if got := layoutAt(start, true); got != from {
		t.Fatalf("transition start = %#v, want %#v", got, from)
	}
	middle := layoutAt(start.Add(50*time.Millisecond), true)
	if middle == from || middle == to {
		t.Fatalf("transition midpoint = %#v", middle)
	}
	if got := layoutAt(start.Add(100*time.Millisecond), true); got != to {
		t.Fatalf("transition end = %#v, want %#v", got, to)
	}
}

func TestBoxStyleTransformMovesOverlayPlacement(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Constraints{Max: image.Pt(100, 80)}, Box(&overlayProbeWidget{
		key:    "styled-box",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}).Style(flowstyle.Style{}.Translate(10, 5)))
	if want := image.Rect(10, 5, 30, 15); got != want {
		t.Fatalf("transformed anchor = %v, want %v", got, want)
	}
}

func TestBoxStyleTranslateUsesDp(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 80)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Metric:      unit.Metric{PxPerDp: 2, PxPerSp: 2},
		Ops:         new(op.Ops),
	}
	var got image.Rectangle
	probe := &overlayProbeWidget{
		key:    "scaled-translate",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
		got:    &got,
	}

	frame.BeginFrameWithViewport(ctx, viewport)
	Box(probe).Style(flowstyle.Style{}.Translate(10, 5)).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	if want := image.Rect(20, 10, 40, 20); got != want {
		t.Fatalf("translated anchor = %v, want %v", got, want)
	}
}

func TestBoxClampsExplicitSizeToParentConstraints(t *testing.T) {
	for _, test := range []struct {
		name        string
		requested   image.Point
		constraints layout.Constraints
		want        image.Point
	}{
		{name: "maximum", requested: image.Pt(120, 120), constraints: layout.Constraints{Min: image.Pt(40, 50), Max: image.Pt(80, 90)}, want: image.Pt(80, 90)},
		{name: "minimum", requested: image.Pt(20, 20), constraints: layout.Constraints{Min: image.Pt(40, 50), Max: image.Pt(80, 90)}, want: image.Pt(40, 50)},
	} {
		t.Run(test.name, func(t *testing.T) {
			child := &constraintWidget{}
			Box(child).Style(flowstyle.Style{}.Width(unit.Dp(test.requested.X)).Height(unit.Dp(test.requested.Y))).Layout(newContext(nil), layout.Context{
				Constraints: test.constraints,
				Ops:         new(op.Ops),
			})
			if child.constraints != layout.Exact(test.want) {
				t.Fatalf("child constraints = %v, want exact %v", child.constraints, test.want)
			}
		})
	}
}

func TestBoxAppliesMinMaxConstraints(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(flowstyle.Style{}.MinWidth(80).MaxWidth(120).MinHeight(40).MaxHeight(90)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if child.constraints.Min.X != 80 || child.constraints.Max.X != 120 {
		t.Fatalf("width constraints = %v, want min 80 max 120", child.constraints)
	}
	if child.constraints.Min.Y != 40 || child.constraints.Max.Y != 90 {
		t.Fatalf("height constraints = %v, want min 40 max 90", child.constraints)
	}
}

func TestBoxFill(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(flowstyle.Style{}.FillWidth().FillHeight()).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if child.constraints.Min != image.Pt(300, 200) || child.constraints.Max != image.Pt(300, 200) {
		t.Fatalf("constraints = %v, want exact (300,200)", child.constraints)
	}
}

func TestBoxStyleFill(t *testing.T) {
	child := &constraintWidget{}
	Box(child).Style(flowstyle.Style{}.FillWidth().FillHeight()).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         new(op.Ops),
	})

	if child.constraints != layout.Exact(image.Pt(300, 200)) {
		t.Fatalf("constraints = %v, want exact (300,200)", child.constraints)
	}
}

func TestBoxStyleAspectRatioResolvesFromWidth(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).Style(flowstyle.Style{}.Width(80).AspectRatio(2)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(80, 40) || child.constraints != layout.Exact(image.Pt(80, 40)) {
		t.Fatalf("aspect box/constraints = %v/%v, want (80,40)", dims.Size, child.constraints)
	}
}

func TestBoxStyleAspectRatioShrinksFixedAxisToFit(t *testing.T) {
	dims := Box(nil).Style(flowstyle.Style{}.Width(80).AspectRatio(1)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 50)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(50, 50) {
		t.Fatalf("aspect box = %v, want (50,50)", dims.Size)
	}
}

func TestBoxStyleCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	box := Box(Spacer(40, 20)).Style(flowstyle.Style{}.Cursor(pointer.CursorPointer))
	layoutFrame := func() {
		var ops op.Ops
		box.Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Source:      router.Source(),
			Ops:         &ops,
		})
		router.Frame(&ops)
	}

	layoutFrame()
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(10, 10)})
	layoutFrame()
	if got := router.Cursor(); got != pointer.CursorPointer {
		t.Fatalf("cursor = %v, want %v", got, pointer.CursorPointer)
	}
}

func TestBoxPaddingUsesBorderBoxMinimum(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).Style(flowstyle.Style{}.MinHeight(36).PaddingY(6)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         new(op.Ops),
	})

	if dims.Size.Y != 36 || child.constraints.Min.Y != 24 {
		t.Fatalf("box/child height = %d/%v, want total 36 and child minimum 24", dims.Size.Y, child.constraints)
	}
}

func TestBoxAppliesSidePadding(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(flowstyle.Style{}.PaddingTop(1).PaddingRight(2).PaddingBottom(3).PaddingLeft(4)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         &ops,
		})

	if child.constraints.Max != image.Pt(94, 96) {
		t.Fatalf("child max constraints = %v, want (94,96)", child.constraints.Max)
	}
}

func TestBoxAppliesMarginOutsideSize(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := Box(child).Style(flowstyle.Style{}.Width(100).Height(50).Margin(10)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(120, 70) {
		t.Fatalf("box size = %v, want (120,70)", dims.Size)
	}
	if child.constraints.Min != image.Pt(100, 50) || child.constraints.Max != image.Pt(100, 50) {
		t.Fatalf("child constraints = %v, want exact (100,50)", child.constraints)
	}
}

func TestBoxAppliesSideMargin(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(flowstyle.Style{}.MarginLeft(4).MarginRight(2).MarginTop(1).MarginBottom(3)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	if child.constraints.Max != image.Pt(94, 96) {
		t.Fatalf("child max constraints = %v, want (94,96)", child.constraints.Max)
	}
}

func TestBoxAlignClearsChildMinimum(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(flowstyle.Style{}.FillWidth().FillHeight()).Align(AlignCenter).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if child.constraints.Min != (image.Point{}) {
		t.Fatalf("child min constraints = %v, want zero", child.constraints.Min)
	}
	if child.constraints.Max != image.Pt(300, 200) {
		t.Fatalf("child max constraints = %v, want (300,200)", child.constraints.Max)
	}
}

func TestBoxClipConstrainsReportedSize(t *testing.T) {
	var ops op.Ops

	dims := Box(overflowWidget{}).Style(flowstyle.Style{}.Width(100).Height(50).Overflow(flowstyle.OverflowHidden)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(100, 50) {
		t.Fatalf("clipped size = %v, want (100,50)", dims.Size)
	}
}

func TestBoxOverflowStyleDoesNotClip(t *testing.T) {
	// OverflowVisible is the default Style overflow; ensure an explicit value still layouts.
	dims := Box(Spacer(10, 10)).
		Style(flowstyle.Style{}.Width(20).Height(20).Overflow(flowstyle.OverflowVisible)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         new(op.Ops),
		})
	if dims.Size != image.Pt(20, 20) {
		t.Fatalf("overflow-visible box = %v, want (20,20)", dims.Size)
	}
}

func TestBoxOverflowHiddenStyleClips(t *testing.T) {
	dims := Box(overflowWidget{}).
		Style(flowstyle.Style{}.Width(100).Height(50).Overflow(flowstyle.OverflowHidden)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(200, 200)},
			Ops:         new(op.Ops),
		})
	if dims.Size != image.Pt(100, 50) {
		t.Fatalf("overflow-hidden box = %v, want (100,50)", dims.Size)
	}
}

func TestRowColumnAlignment(t *testing.T) {
	if Row().AlignMiddle().align != layout.Middle {
		t.Fatal("row alignment was not middle")
	}
	if Row().AlignBaseline().align != layout.Baseline {
		t.Fatal("row alignment was not baseline")
	}
	if Column().AlignEnd().align != layout.End {
		t.Fatal("column alignment was not end")
	}
}

func TestRowSpacingBetweenUsesRemainingMainAxisSpace(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 20)
	var first, second image.Rectangle
	firstProbe := &overlayProbeWidget{key: "first", size: image.Pt(10, 10), anchor: image.Rect(0, 0, 10, 10), got: &first}
	secondProbe := &overlayProbeWidget{key: "second", size: image.Pt(10, 10), anchor: image.Rect(0, 0, 10, 10), got: &second}
	gtx := layout.Context{Constraints: layout.Exact(viewport), Ops: new(op.Ops)}
	frame.BeginFrameWithViewport(ctx, viewport)
	Row(firstProbe, secondProbe).Spacing(SpaceBetween).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	if first != image.Rect(0, 0, 10, 10) || second != image.Rect(90, 0, 100, 10) {
		t.Fatalf("spacing positions = %v/%v, want (0,0)-(10,10) and (90,0)-(100,10)", first, second)
	}
}

func TestFlexSpacingModes(t *testing.T) {
	for _, test := range []struct {
		name        string
		spacing     FlexSpacing
		wantStart   int
		wantBetween int
	}{
		{name: "end", spacing: SpaceEnd},
		{name: "start", spacing: SpaceStart, wantStart: 10},
		{name: "sides", spacing: SpaceSides, wantStart: 5},
		{name: "around", spacing: SpaceAround, wantStart: 2, wantBetween: 5},
		{name: "between", spacing: SpaceBetween, wantBetween: 10},
		{name: "evenly", spacing: SpaceEvenly, wantStart: 3, wantBetween: 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			start, between := flexSpacing(test.spacing, 10, 2)
			if start != test.wantStart || between != test.wantBetween {
				t.Fatalf("spacing = %d/%d, want %d/%d", start, between, test.wantStart, test.wantBetween)
			}
		})
	}
}

func TestCenterPropagatesOverlayPosition(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Exact(image.Pt(100, 80)), Center(&overlayProbeWidget{
		key:    "center",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}))
	want := image.Rect(40, 35, 60, 45)
	if got != want {
		t.Fatalf("center anchor = %v, want %v", got, want)
	}
}

func TestFlexPropagatesOverlayPositions(t *testing.T) {
	t.Run("row alignment", func(t *testing.T) {
		got := resolveOverlayAnchor(t, layout.Exact(image.Pt(100, 40)), Row(
			Spacer(10, 10),
			&overlayProbeWidget{key: "row", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 20, 10)},
		).Gap(5).AlignEnd())
		want := image.Rect(15, 30, 35, 40)
		if got != want {
			t.Fatalf("row anchor = %v, want %v", got, want)
		}
	})

	t.Run("column flexible child", func(t *testing.T) {
		got := resolveOverlayAnchor(t, layout.Exact(image.Pt(60, 80)), Column(
			Spacer(10, 10),
			Flexible(1, &overlayProbeWidget{key: "column", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 10, 10)}),
		).Gap(5).AlignMiddle())
		want := image.Rect(20, 15, 30, 25)
		if got != want {
			t.Fatalf("column anchor = %v, want %v", got, want)
		}
	})
}

func TestFlexOverlayPositionMatchesIgnoredNegativeGap(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Exact(image.Pt(100, 30)), Row(
		Spacer(10, 10),
		&overlayProbeWidget{key: "negative-gap", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 20, 10)},
	).Gap(-5))
	want := image.Rect(10, 0, 30, 10)
	if got != want {
		t.Fatalf("negative-gap anchor = %v, want %v", got, want)
	}
}

func TestLayoutItemsPropagatesOverlayPositions(t *testing.T) {
	tests := []struct {
		name        string
		horizontal  bool
		constraints layout.Constraints
		want        image.Rectangle
	}{
		{
			name:        "vertical second item",
			constraints: layout.Constraints{Max: image.Pt(100, 100)},
			want:        image.Rect(0, 15, 20, 25),
		},
		{
			name:        "horizontal wrapped item",
			horizontal:  true,
			constraints: layout.Constraints{Max: image.Pt(70, 100)},
			want:        image.Rect(0, 15, 20, 25),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			frame.BeginFrameWithViewport(ctx, image.Pt(100, 100))
			gtx := layout.Context{Constraints: test.constraints, Ops: new(op.Ops)}
			probe := &overlayProbeWidget{key: "items", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 20, 10)}
			var got image.Rectangle
			probe.got = &got
			children := []layout.Widget{
				func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(60, 10)} },
				func(gtx layout.Context) layout.Dimensions { return probe.Layout(ctx, gtx) },
			}
			LayoutItems(ctx, gtx, test.horizontal, 10, 5, children)
			frame.LayoutOverlays(ctx, gtx)
			frame.EndFrame(ctx)
			if got != test.want {
				t.Fatalf("item anchor = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLayoutItemsClampsNegativeGaps(t *testing.T) {
	dims := LayoutItems(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	}, true, -10, -10, []layout.Widget{
		func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(20, 10)} },
		func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(20, 10)} },
	})

	if dims.Size != image.Pt(40, 10) {
		t.Fatalf("items size = %v, want (40,10)", dims.Size)
	}
}

func TestBoxPropagatesMarginPaddingAndAlignment(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Constraints{Max: image.Pt(120, 100)}, Box(&overlayProbeWidget{
		key:    "box",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}).Style(flowstyle.Style{}.Width(80).Height(60).MarginLeft(7).MarginTop(9).Padding(5)).
		Align(AlignBottomEnd))
	want := image.Rect(62, 54, 82, 64)
	if got != want {
		t.Fatalf("box anchor = %v, want %v", got, want)
	}
}

func TestBoxClipHidesOverlayAnchorOutsideBox(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(200, 120)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	called := false
	probe := &overlayProbeWidget{
		key:    "clipped-box",
		size:   image.Pt(100, 100),
		anchor: image.Rect(80, 80, 90, 90),
		capture: func(image.Rectangle) {
			called = true
		},
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	Box(probe).Style(flowstyle.Style{}.Width(50).Height(50).Overflow(flowstyle.OverflowHidden)).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)
	if called {
		t.Fatal("anchor outside a clipped Box produced an overlay")
	}
}

func TestExpandedUsesRemainingRowSpace(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Row(
		Spacer(50, 10),
		Expanded(child),
	).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 40)},
		Ops:         &ops,
	})

	if child.constraints.Min.X != 140 || child.constraints.Max.X != 140 {
		t.Fatalf("expanded width constraints = %v, want exact 140", child.constraints)
	}
}

func TestFlexibleOffersRemainingSpaceByWeight(t *testing.T) {
	left := &constraintWidget{}
	right := &constraintWidget{}
	var ops op.Ops

	Row(
		Flexible(1, left),
		Flexible(2, right),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 40)},
		Ops:         &ops,
	})

	if left.constraints.Min.X != 0 || left.constraints.Max.X != 100 {
		t.Fatalf("left constraints = %v, want loose maximum 100", left.constraints)
	}
	if right.constraints.Min.X != 0 || right.constraints.Max.X != 200 {
		t.Fatalf("right constraints = %v, want loose maximum 200", right.constraints)
	}
}

func TestFlexibleCanUseLessThanItsShare(t *testing.T) {
	dims := Row(Flexible(1, Spacer(20, 10))).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 40)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(20, 10) {
		t.Fatalf("loose Flexible size = %v, want (20,10)", dims.Size)
	}
}

func TestWrapMovesChildrenToNextLine(t *testing.T) {
	var ops op.Ops

	dims := Wrap(
		Spacer(60, 10),
		Spacer(60, 20),
		Spacer(60, 30),
	).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(130, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(130, 60) {
		t.Fatalf("wrap size = %v, want (130,60)", dims.Size)
	}
}

func TestWrapClampsNegativeGaps(t *testing.T) {
	dims := Wrap(Spacer(60, 10), Spacer(60, 10)).Gap(-10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 100)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(120, 10) {
		t.Fatalf("wrap size = %v, want (120,10)", dims.Size)
	}
}

func TestWrapAlignmentOffset(t *testing.T) {
	if Wrap().AlignEnd().align != layout.End {
		t.Fatal("wrap alignment was not end")
	}
	if got := alignmentOffset(layout.Middle, 11); got != 5 {
		t.Fatalf("middle offset = %d, want 5", got)
	}
	if got := alignmentOffset(layout.End, 11); got != 11 {
		t.Fatalf("end offset = %d, want 11", got)
	}
	if got := alignmentOffset(layout.Start, 11); got != 0 {
		t.Fatalf("start offset = %d, want 0", got)
	}
}

func TestWrapPropagatesWrappedAndAlignedPosition(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Constraints{Max: image.Pt(130, 100)}, Wrap(
		Spacer(60, 10),
		Spacer(60, 20),
		&overlayProbeWidget{key: "wrap", size: image.Pt(30, 10), anchor: image.Rect(0, 0, 30, 10)},
	).Gap(10).AlignEnd())
	want := image.Rect(100, 30, 130, 40)
	if got != want {
		t.Fatalf("wrap anchor = %v, want %v", got, want)
	}
}

func TestWrapPreservesZeroHeightRowBoundaries(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 20)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	var first, second image.Rectangle
	firstProbe := &overlayProbeWidget{key: "first-zero-row", size: image.Pt(90, 0), anchor: image.Rect(0, 0, 1, 1), got: &first}
	secondProbe := &overlayProbeWidget{key: "second-zero-row", size: image.Pt(30, 0), anchor: image.Rect(0, 0, 1, 1), got: &second}

	frame.BeginFrameWithViewport(ctx, viewport)
	Wrap(firstProbe, secondProbe).Gap(10).AlignEnd().Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	if first.Min.X != 0 || second.Min.X != 60 {
		t.Fatalf("zero-height row positions = %v/%v, want x=0/60", first, second)
	}
}

func TestFlexibleRejectsInvalidWeight(t *testing.T) {
	for _, weight := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1))} {
		t.Run(fmt.Sprintf("weight-%v", weight), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid flex weight panic")
				}
			}()
			Flexible(weight, Spacer(1, 1))
		})
	}
}

func TestSpacer(t *testing.T) {
	dims := Spacer(12, 8).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
	})
	if dims.Size != image.Pt(12, 8) {
		t.Fatalf("spacer size = %v, want (12,8)", dims.Size)
	}
}

type constraintWidget struct {
	constraints layout.Constraints
}

func (w *constraintWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

type overflowWidget struct{}

func (overflowWidget) Layout(_ *frame.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(500, 500)}
}

type overlayProbeWidget struct {
	key     string
	size    image.Point
	anchor  image.Rectangle
	got     *image.Rectangle
	capture func(image.Rectangle)
}

func (w *overlayProbeWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       w.key,
		Anchor:    w.anchor,
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			if w.got != nil {
				*w.got = anchor
			}
			if w.capture != nil {
				w.capture(anchor)
			}
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: w.size}
}

func resolveOverlayAnchor(t *testing.T, constraints layout.Constraints, widget frame.Widget) image.Rectangle {
	t.Helper()
	ctx := newContext(nil)
	frame.BeginFrameWithViewport(ctx, constraints.Max)
	gtx := layout.Context{Constraints: constraints, Ops: new(op.Ops)}
	var got image.Rectangle
	probe, ok := findOverlayProbe(widget)
	if !ok {
		t.Fatal("test widget does not contain an overlay probe")
	}
	probe.got = &got
	widget.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)
	return got
}

func findOverlayProbe(widget frame.Widget) (*overlayProbeWidget, bool) {
	switch widget := widget.(type) {
	case *overlayProbeWidget:
		return widget, true
	case CenterWidget:
		return findOverlayProbe(widget.child)
	case BoxWidget:
		return findOverlayProbe(widget.Child())
	case RowWidget:
		return findOverlayProbeIn(widget.children)
	case ColumnWidget:
		return findOverlayProbeIn(widget.children)
	case WrapWidget:
		return findOverlayProbeIn(widget.children)
	case FlexWidget:
		return findOverlayProbe(widget.child)
	default:
		return nil, false
	}
}

func findOverlayProbeIn(widgets []frame.Widget) (*overlayProbeWidget, bool) {
	for _, widget := range widgets {
		if probe, ok := findOverlayProbe(widget); ok {
			return probe, true
		}
	}
	return nil, false
}
