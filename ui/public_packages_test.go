package ui_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	giolayout "gioui.org/layout"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

func TestPublicPackageComposition(t *testing.T) {
	custom := ui.WidgetFunc(func(_ *ui.Context, _ giolayout.Context) giolayout.Dimensions {
		return giolayout.Dimensions{}
	})

	root := flowlayout.Column(
		(button.New("Save")).Style((style.Style{}).Radius(unit.Dp(8))),
		flowlayout.Row(
			flowlayout.Expanded(custom),
			text.New("Ready"),
		),
	)
	var _ ui.Widget = root

	var _ window.Option = window.CenterOnStart()
	var _ window.Option = window.Title("FlowUI")

	brush := draw.LinearGradient(
		draw.GradientStop{Offset: 0, Color: color.NRGBA{R: 0xff, A: 0xff}},
		draw.GradientStop{Offset: 1, Color: color.NRGBA{B: 0xff, A: 0xff}},
	)
	if got := brush.ColorAt(0.5); got.A != 0xff {
		t.Fatalf("draw brush alpha = %d, want 255", got.A)
	}

	_ = (style.Style{}).Padding(unit.Dp(8))
}

func TestPointerHelpersExposeGioContracts(t *testing.T) {
	var _ func(giolayout.Context, event.Tag, image.Rectangle, pointer.Cursor) = ui.AddPointerArea
	var _ func(giolayout.Context, event.Tag, pointer.Kind) (pointer.Event, bool) = ui.NextPointerEvent
	var _ func(pointer.Event) bool = ui.IsPrimaryPointerPress
	var _ func(giolayout.Context, event.Tag, pointer.Event) = ui.GrabPointer
}

type publicCustomState struct{}

type publicMeasuredWidget struct{}

func (publicMeasuredWidget) Measure(_ *ui.Context, _ giolayout.Context) giolayout.Dimensions {
	return giolayout.Dimensions{}
}

func TestCustomWidgetContractsArePublic(t *testing.T) {
	var _ func(*ui.Context, string) *publicCustomState = ui.UseState[publicCustomState]
	var _ func(*ui.Context, string, func() *publicCustomState) *publicCustomState = ui.UseStateWith[publicCustomState]
	var _ func(*ui.Context, giolayout.Context, ui.Widget) giolayout.Dimensions = ui.MeasureWidget
	var _ ui.Measurable = publicMeasuredWidget{}
	var _ ui.Widget = ui.WidgetFunc(func(_ *ui.Context, _ giolayout.Context) giolayout.Dimensions {
		return giolayout.Dimensions{}
	})
	// Custom widgets that lay out mixed-style text need the frame's shaper;
	// component tests in examples/showcase exercise the returned value.
	var _ func(*ui.Context) *giotext.Shaper = (*ui.Context).TextShaper
}

func TestEffectHelpersExposeTypedContracts(t *testing.T) {
	var _ func(...ui.Cmd[int]) ui.Cmd[int] = ui.Batch[int]
	var _ func(ui.Subscription[int], func(int) string) ui.Subscription[string] = ui.MapSubscription[int, string]
}

func TestMVUProgramUsesPublicWidgetContract(t *testing.T) {
	type model struct{ count int }
	type message int

	program := ui.NewProgram(model{}, func(value *model, msg message) ui.Cmd[message] {
		value.count += int(msg)
		return nil
	}, func(_ *ui.Context, value model, _ ui.Send[message]) ui.Widget {
		return text.New(fmt.Sprintf("count=%d", value.count))
	})

	initial, cmd := program.Init()
	if cmd != nil || initial.count != 0 {
		t.Fatalf("program init = %#v, %v", initial, cmd)
	}
	if update := program.Update(&initial, 2); update != nil || initial.count != 2 {
		t.Fatalf("program update = count %d, cmd=%v", initial.count, update)
	}
	if program.View(nil, initial, func(message) {}) == nil {
		t.Fatal("program view returned a nil widget")
	}
}
