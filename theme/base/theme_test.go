package base_test

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/checkbox"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/selectfield"
	"github.com/qianniancn/flowui/widget/slider"
)

func TestDefaultModes(t *testing.T) {
	for _, test := range []struct {
		name  string
		new   func() theme.Theme
		light bool
	}{
		{"light", base.Light, true},
		{"dark", base.Dark, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			active := test.new()
			if err := theme.ValidateVisualProvider(active); err != nil {
				t.Fatal(err)
			}
			semantic := active.Semantic()
			bg, fg := semantic.Background, semantic.Foreground
			contrast := int(bg.R) + int(bg.G) + int(bg.B) - int(fg.R) - int(fg.G) - int(fg.B)
			if contrast == 0 || (contrast > 0) != test.light {
				t.Fatalf("%s mode has background %v and foreground %v", test.name, bg, fg)
			}
		})
	}
}

func TestNew(t *testing.T) {
	for _, test := range []struct {
		name     string
		newTheme func(theme.Seed) theme.Theme
	}{
		{"light", base.New},
		{"dark", base.NewDark},
	} {
		t.Run(test.name, func(t *testing.T) {
			active := test.newTheme(theme.DefaultSeed())
			if err := theme.ValidateVisualProvider(active); err != nil {
				t.Fatal(err)
			}
			theme.ResolveVisual(&active, theme.ButtonRequest{},
				theme.ComponentVisualProvider.Button, theme.ValidateButtonVisual)
			if _, immutable := active.VisualProvider.(theme.ImmutableVisualProvider); immutable {
				t.Fatal("base provider implicitly promised immutable custom visuals")
			}
		})
	}
}

func TestFilledLabelsStayLegibleInAllStates(t *testing.T) {
	themes := []theme.Theme{base.Light(), base.Dark()}
	for _, primary := range []color.NRGBA{
		{R: 240, G: 230, B: 180, A: 255},
		{R: 10, G: 30, B: 50, A: 255},
		{R: 117, G: 117, B: 117, A: 255},
	} {
		seed := theme.Seed{Primary: primary}
		themes = append(themes, base.New(seed), base.NewDark(seed))
	}
	for _, active := range themes {
		for _, tone := range []theme.Tone{
			theme.TonePrimary, theme.ToneSecondary, theme.ToneDanger, theme.ToneWarning, theme.ToneSuccess,
		} {
			pair := active.Semantic().Tone(tone)
			for _, fill := range []color.NRGBA{pair.Fill, pair.Hover, pair.Pressed} {
				// Solid labels use white on the red and green fills, which
				// clears the large-text threshold rather than the 4.5
				// normal-text one.
				if ratio := contrastRatio(fill, pair.OnFill); ratio < 3 {
					t.Errorf("tone %v label %v on %v: contrast %.2f, want >= 3", tone, pair.OnFill, fill, ratio)
				}
			}
			for _, fill := range []color.NRGBA{pair.Soft, pair.SoftHover} {
				if ratio := contrastRatio(fill, pair.OnSoft); ratio < 4.5 {
					t.Errorf("tone %v soft label %v on %v: contrast %.2f, want >= 4.5", tone, pair.OnSoft, fill, ratio)
				}
			}
		}
		for _, test := range []struct {
			variant theme.ButtonVariant
			preset  style.Style
		}{
			{theme.ButtonPrimary, active.PrimaryAction()},
			{theme.ButtonDanger, active.DangerAction()},
		} {
			visual := base.Provider{}.Button(&active, theme.ButtonRequest{Variant: test.variant})
			for _, state := range []style.State{{}, {Hovered: true}, {Pressed: true}} {
				for _, declaration := range []style.ResolvedStyle{
					style.Cascade(state, visual.Defaults, visual.Variant, visual.Size),
					style.Cascade(state, test.preset),
				} {
					fg, _ := resolve.Color(declaration.Text.Color)
					brush, _ := resolve.Brush(declaration.Paint.Background)
					bg := brush.ColorAt(.5)
					if ratio := contrastRatio(fg, bg); ratio < 3 {
						t.Errorf("button %v state %+v: contrast %.2f, want >= 3", test.variant, state, ratio)
					}
				}
			}
		}
	}
}

type checkboxColorProbe struct {
	base.Provider
	foreground color.NRGBA
	called     bool
}

func (p *checkboxColorProbe) Checkbox(active *theme.Theme, request theme.CheckboxRequest) theme.CheckboxVisual {
	visual := p.Provider.Checkbox(active, request)
	render := visual.IndicatorRenderer
	visual.IndicatorRenderer = func(
		gtx giolayout.Context,
		request theme.CheckboxIndicatorRenderRequest,
	) giolayout.Dimensions {
		p.foreground, p.called = request.Foreground, true
		return render(gtx, request)
	}
	return visual
}

func TestCheckboxIndicatorColorInsideCard(t *testing.T) {
	custom := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}
	declaration := style.New().Part(style.PartIndicator, style.New().TextColor(style.Solid(custom)))
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		for _, layer := range []string{"default", "scope", "instance", "shared-control"} {
			probe := &checkboxColorProbe{}
			active.VisualProvider = probe
			harness := uitest.New(image.Pt(300, 100), active)
			var control ui.Widget = checkbox.New("Selected").Key("checkbox").Checked(true)
			want := active.Semantic().Tone(theme.TonePrimary).OnFill
			switch layer {
			case "scope":
				control = resolve.Scope(declaration, control)
				want = custom
			case "instance":
				control = (checkbox.New("Selected").Key("checkbox").Checked(true)).Style(declaration)
				want = custom
			case "shared-control":
				control = ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
					return checkbox.DrawControl(ctx, gtx, checkbox.ControlOptions{Selection: 1})
				})
			}
			harness.Frame(card.New(control))
			if !probe.called || probe.foreground != want {
				t.Errorf("%s indicator = %v (rendered=%v), want %v", layer, probe.foreground, probe.called, want)
			}
		}
	}
}

func contrastRatio(a, b color.NRGBA) float64 {
	luminance := func(c color.NRGBA) float64 {
		linear := func(channel uint8) float64 {
			v := float64(channel) / 255
			if v <= .04045 {
				return v / 12.92
			}
			return math.Pow((v+.055)/1.055, 2.4)
		}
		return .2126*linear(c.R) + .7152*linear(c.G) + .0722*linear(c.B)
	}
	x, y := luminance(a), luminance(b)
	return (max(x, y) + .05) / (min(x, y) + .05)
}

func TestDefaultControlAlignment(t *testing.T) {
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		harness := uitest.New(image.Pt(300, 100), active)
		for _, control := range []ui.Widget{
			button.New("Save"), input.New("FlowUI"),
			selectfield.New(nil).Key("select"), combobox.New(nil).Key("combo"),
			datepicker.New("date", time.Time{}),
		} {
			harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				dims := control.Layout(ctx, gtx)
				if dims.Size.Y != 32 {
					t.Errorf("%T height = %d, want 32dp", control, dims.Size.Y)
				}
				return dims
			}))
		}
	}
}

func TestDefaultSliderPointerAndKeyboard(t *testing.T) {
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		for _, vertical := range []bool{false, true} {
			harness := uitest.New(image.Pt(300, 300), active)
			value := 0.0
			root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				control := slider.New("slider", value).Step(5).
					OnChange(func(next float64) { value = next })
				if vertical {
					control = control.Vertical()
				}
				return control.Layout(ctx, gtx)
			})
			harness.Frame(root)
			position, arrow := f32.Pt(150, 8), key.NameRightArrow
			if vertical {
				position, arrow = f32.Pt(8, 150), key.NameUpArrow
			}
			harness.Click(position)
			harness.Frame(root)
			if value != 50 {
				t.Fatalf("vertical=%v pointer value = %v, want 50", vertical, value)
			}
			harness.Key(arrow, 0)
			harness.Frame(root)
			if value != 55 {
				t.Fatalf("vertical=%v keyboard value = %v, want 55", vertical, value)
			}
		}
	}
}

func TestDefaultSelectInteraction(t *testing.T) {
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		harness := uitest.New(image.Pt(300, 300), active)
		selected, open, disabled := "desktop", false, false
		root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return selectfield.New([]selectfield.SelectItem{
				{Key: "desktop", Label: "Desktop"}, {Key: "tools", Label: "Tools"},
			}).Key("platform").SelectedKey(selected).Open(open).Disabled(disabled).
				OnChange(func(value string) { selected = value }).
				OnOpenChange(func(value bool) { open = value }).Layout(ctx, gtx)
		})
		frame := func() {
			// Controlled open state arrives on the next frame, then its entrance
			// animation needs another frame before the options accept input.
			for range 3 {
				harness.Frame(root)
				harness.Advance(time.Second)
			}
		}
		frame()
		harness.Click(f32.Pt(30, 16))
		frame()
		if !open {
			t.Fatal("pointer did not open select")
		}
		harness.Key(key.NameDownArrow, 0)
		frame()
		harness.Key(key.NameReturn, 0)
		frame()
		if selected != "tools" || open {
			t.Fatalf("keyboard selection = %q, open=%v; want tools, false", selected, open)
		}
		harness.Click(f32.Pt(30, 16))
		frame()
		harness.Key(key.NameEscape, 0)
		frame()
		if open {
			t.Fatal("Escape did not close select")
		}
		disabled = true
		frame()
		harness.Click(f32.Pt(30, 16))
		frame()
		if open || selected != "tools" {
			t.Fatal("disabled select accepted input")
		}
	}
}
