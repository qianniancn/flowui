package uitest_test

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/FlowUI/uitest"
)

func TestHarnessRoutesPointerAndKeyboardClicks(t *testing.T) {
	harness := uitest.New(image.Pt(240, 120))
	button := new(testButton)

	harness.Frame(button)
	harness.Click(f32.Pt(20, 16))
	harness.Frame(button)
	if button.clicks != 1 {
		t.Fatalf("pointer clicks = %d, want 1", button.clicks)
	}

	harness.Router().Source().Execute(key.FocusCmd{Tag: &button.state})
	if !harness.Router().Source().Focused(&button.state) {
		t.Fatal("button did not gain keyboard focus")
	}
	harness.Key(key.NameReturn, 0)
	harness.Frame(button)
	if button.clicks != 2 {
		t.Fatalf("clicks after keyboard activation = %d, want 2", button.clicks)
	}
	if len(harness.Router().AppendSemantics(nil)) == 0 {
		t.Fatal("completed frame did not expose semantics")
	}
}

func TestHarnessAdvancesTweenAndLaysOutPortal(t *testing.T) {
	harness := uitest.New(image.Pt(240, 120))
	var progress float32
	portalLayouts := 0
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		progress = ui.Tween("progress", 1).
			Initial(0).
			Duration(time.Second).
			Easing(ui.EaseLinear).
			Value(ctx, gtx)
		return ui.Portal("probe", true, ui.Spacer(20, 10), func(image.Rectangle, bool) ui.Widget {
			return ui.WidgetFunc(func(*ui.Context, layout.Context) layout.Dimensions {
				portalLayouts++
				return layout.Dimensions{}
			})
		}).Layout(ctx, gtx)
	})

	harness.Frame(root)
	if progress != 0 || portalLayouts != 1 {
		t.Fatalf("first frame = progress %v, portal layouts %d; want 0, 1", progress, portalLayouts)
	}
	harness.Advance(500 * time.Millisecond)
	harness.Frame(root)
	if math.Abs(float64(progress-.5)) > .001 || portalLayouts != 2 {
		t.Fatalf("second frame = progress %v, portal layouts %d; want .5, 2", progress, portalLayouts)
	}
}

func TestHarnessConfigurationAndResize(t *testing.T) {
	if language := uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 100)}).Context().Language(); language != ui.LanguageEnglish {
		t.Fatalf("default language = %q, want English", language)
	}

	theme := ui.DarkTheme()
	theme.Palette.Accent = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	harness := uitest.NewWithConfig(uitest.Config{
		Size:     image.Pt(200, 100),
		Theme:    &theme,
		Language: ui.LanguageChinese,
		Now:      time.Unix(10, 0),
	})
	if harness.Context().Language() != ui.LanguageChinese {
		t.Fatalf("language = %q, want Chinese", harness.Context().Language())
	}
	if harness.Context().Theme().Palette.Background != theme.Palette.Background {
		t.Fatal("harness did not use the configured theme")
	}
	if harness.Context().Theme().Material.Palette.ContrastBg != theme.Palette.Accent {
		t.Fatal("harness did not synchronize the Gio material theme")
	}

	var now time.Time
	probe := ui.WidgetFunc(func(_ *ui.Context, gtx layout.Context) layout.Dimensions {
		now = gtx.Now
		return layout.Dimensions{Size: gtx.Constraints.Min}
	})
	harness.Resize(image.Pt(320, 180))
	dimensions := harness.Frame(probe)
	if dimensions.Size != image.Pt(320, 180) || now != time.Unix(10, 0) {
		t.Fatalf("resized frame = dimensions %v, time %v", dimensions.Size, now)
	}
	harness.Advance(time.Second)
	harness.Frame(probe)
	if now != time.Unix(11, 0) {
		t.Fatalf("time = %v, want 11 seconds", now)
	}
}

func TestHarnessRejectsInvalidTimeAndSize(t *testing.T) {
	mustPanic(t, func() { uitest.New(image.Point{}) })
	harness := uitest.New(image.Pt(100, 100))
	mustPanic(t, func() { harness.Resize(image.Pt(-1, 100)) })
	mustPanic(t, func() { harness.Advance(-time.Second) })
}

type testButton struct {
	state  widget.Clickable
	clicks int
}

func (button *testButton) Layout(_ *ui.Context, gtx layout.Context) layout.Dimensions {
	for button.state.Clicked(gtx) {
		button.clicks++
	}
	return button.state.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp("Test button").Add(gtx.Ops)
		return layout.Dimensions{Size: image.Pt(80, 32)}
	})
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
