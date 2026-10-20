package base_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/theme/buttonstyle"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/input"
)

type customProvider struct{ base.Provider }

var _ theme.ComponentVisualProvider = customProvider{}

func (customProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	config := buttonstyle.DefaultConfig()
	config.Radius = 0
	return buttonstyle.BuildButton(active, request, config)
}

func ExampleProvider() {
	active := base.New(theme.DefaultSeed())
	active.VisualProvider = customProvider{}
	visual := theme.ResolveVisual(&active, theme.ToolbarRequest{},
		theme.ComponentVisualProvider.Toolbar, theme.ValidateToolbarVisual)
	fmt.Println(theme.ValidateVisualProvider(active))
	fmt.Println(visual.Metrics.Gap == active.Components.Toolbar.Gap)
	// Output:
	// <nil>
	// true
}

type customToolbarProvider struct {
	customProvider
	gap unit.Dp
}

func (p customToolbarProvider) Toolbar(active *theme.Theme, request theme.ToolbarRequest) theme.ToolbarVisual {
	visual := p.Provider.Toolbar(active, request)
	visual.Metrics.Gap = p.gap
	return visual
}

func TestProviderUsesActiveThemeAndAllowsOverrides(t *testing.T) {
	provider := customProvider{}
	if _, immutable := any(provider).(theme.ImmutableVisualProvider); immutable {
		t.Fatal("embedding base implicitly promised immutable custom visuals")
	}
	first := theme.New(theme.DefaultSeed())
	first.Components.Toolbar.Gap = 7
	first.VisualProvider = provider
	second := theme.NewDark(theme.DefaultSeed())
	second.Components.Toolbar.Gap = 13
	second.VisualProvider = provider

	for _, active := range []*theme.Theme{&first, &second} {
		visual := theme.ResolveVisual(active, theme.ToolbarRequest{},
			theme.ComponentVisualProvider.Toolbar, theme.ValidateToolbarVisual)
		if visual.Metrics.Gap != active.Components.Toolbar.Gap {
			t.Fatalf("toolbar gap = %v, want active theme gap %v", visual.Metrics.Gap, active.Components.Toolbar.Gap)
		}
		theme.ResolveVisual(active, theme.ButtonRequest{},
			theme.ComponentVisualProvider.Button, theme.ValidateButtonVisual)
		theme.ResolveVisual(active, theme.InputRequest{},
			theme.ComponentVisualProvider.Input, theme.ValidateInputVisual)
		theme.ResolveVisual(active, theme.ToastRequest{},
			theme.ComponentVisualProvider.Toast, theme.ValidateToastVisual)
	}

	first.VisualProvider = customToolbarProvider{gap: 19}
	visual := theme.ResolveVisual(&first, theme.ToolbarRequest{},
		theme.ComponentVisualProvider.Toolbar, theme.ValidateToolbarVisual)
	if visual.Metrics.Gap != 19 {
		t.Fatalf("custom toolbar gap = %v, want 19", visual.Metrics.Gap)
	}
}

func TestButtonDisabledOverridesInteraction(t *testing.T) {
	for _, mode := range []struct {
		name string
		new  func() theme.Theme
	}{
		{name: "light", new: base.Light},
		{name: "dark", new: base.Dark},
	} {
		t.Run(mode.name, func(t *testing.T) {
			active := mode.new()
			for _, test := range []struct {
				name    string
				variant theme.ButtonVariant
			}{
				{name: "primary", variant: theme.ButtonPrimary},
				{name: "secondary", variant: theme.ButtonSecondary},
				{name: "tertiary", variant: theme.ButtonTertiary},
				{name: "ghost", variant: theme.ButtonGhost},
				{name: "outline", variant: theme.ButtonOutline},
				{name: "danger", variant: theme.ButtonDanger},
				{name: "danger-soft", variant: theme.ButtonDangerSoft},
				{name: "danger-outline", variant: theme.ButtonDangerOutline},
			} {
				t.Run(test.name, func(t *testing.T) {
					visual := base.Provider{}.Button(&active, theme.ButtonRequest{Variant: test.variant})
					for _, state := range []style.State{
						{Disabled: true}, {Disabled: true, Hovered: true}, {Disabled: true, Pressed: true},
					} {
						declaration := style.Cascade(state, visual.Defaults, visual.Variant, visual.Size)
						fg, _ := resolve.Color(declaration.Text.Color)
						brush, _ := resolve.Brush(declaration.Paint.Background)
						bg := brush.ColorAt(.5)
						sem := active.Semantic()
						if fg != sem.DisabledForeground || bg != sem.DisabledBackground {
							t.Errorf("state %+v: disabled colours = %v on %v", state, fg, bg)
						}
					}
				})
			}
		})
	}
}

func TestInputFieldStateFace(t *testing.T) {
	for _, test := range []struct {
		name  string
		theme theme.Theme
	}{{"light", base.Light()}, {"dark", base.Dark()}} {
		t.Run(test.name, func(t *testing.T) {
			active := test.theme
			sem := active.Semantic()
			visual := base.Provider{}.Input(&active, theme.InputRequest{})
			resolved := func(state style.State) style.ResolvedStyle {
				return visual.Defaults.Resolve(state)
			}
			colorOf := func(source style.ColorSource) color.NRGBA {
				value, ok := resolve.Color(source)
				if !ok {
					t.Fatalf("field colour %v is not solid", source)
				}
				return value
			}
			paintColor := func(source style.PaintSource) color.NRGBA {
				brush, ok := resolve.Brush(source)
				if !ok {
					t.Fatalf("field paint %v is empty", source)
				}
				return brush.ColorAt(.5)
			}
			stroke := func(state style.State) (unit.Dp, color.NRGBA, style.ResolvedStyle) {
				value := resolved(state)
				if value.Paint == nil || value.Paint.Border == nil || value.Paint.Border.Width == nil {
					t.Fatalf("state %+v has no field stroke", state)
				}
				return *value.Paint.Border.Width, colorOf(value.Paint.Border.Color), value
			}

			restingWidth, restingColor, _ := stroke(style.State{})
			if restingWidth != active.Components.Input.BorderWidth {
				t.Fatalf("resting width = %v, want %v", restingWidth, active.Components.Input.BorderWidth)
			}
			if restingColor != sem.Border {
				t.Fatalf("resting colour = %v, want border %v", restingColor, sem.Border)
			}

			// A bordered field darkens its stroke on hover; only focus and
			// validation change the stroke colour.
			hoverWidth, hoverColor, _ := stroke(style.State{Hovered: true})
			if hoverWidth != restingWidth || hoverColor != theme.Mix(sem.Border, sem.Foreground, 18) {
				t.Fatalf("hover stroke = %v/%v, want darkened border", hoverWidth, hoverColor)
			}

			// Focus draws an inset stroke in the ring colour plus the soft glow
			// requested by the theme tokens; it never paints an outline outside
			// the field.
			focusWidth, focusColor, focus := stroke(style.State{Focused: true})
			if focus.Paint.Outline != nil {
				t.Fatal("focused field paints an outline outside its box")
			}
			if len(focus.Paint.Shadows) == 0 {
				t.Fatal("focused field has no glow")
			}
			for _, shadow := range focus.Paint.Shadows {
				if !shadow.Transient {
					t.Fatal("focus glow must not reserve layout space")
				}
			}
			if focusWidth != active.Components.Input.FocusRingWidth {
				t.Fatalf("focus width = %v, want ring width %v", focusWidth, active.Components.Input.FocusRingWidth)
			}
			if focusColor != sem.Ring {
				t.Fatalf("focus colour = %v, want ring %v", focusColor, sem.Ring)
			}

			// Validation keeps the resting stroke width and turns the stroke itself
			// into the tone colour; focus adopts the focus ring token.
			invalidWidth, invalidColor, invalid := stroke(style.State{Invalid: true})
			if invalid.Paint.Outline != nil {
				t.Fatal("invalid field paints an outline outside its box")
			}
			if invalidWidth != max(active.Components.Input.BorderWidth, 1) {
				t.Fatalf("invalid width = %v, want the resting width", invalidWidth)
			}
			if want := sem.Tone(theme.ToneDanger).Fill; invalidColor != want {
				t.Fatalf("invalid colour = %v, want danger %v", invalidColor, want)
			}
			invalidFocusWidth, invalidFocusColor, _ := stroke(style.State{Invalid: true, Focused: true})
			if invalidFocusWidth != active.Components.Input.FocusRingWidth || invalidFocusColor != invalidColor {
				t.Fatalf("invalid focus stroke = %v/%v, want %v/%v",
					invalidFocusWidth, invalidFocusColor, active.Components.Input.FocusRingWidth, invalidColor)
			}

			// Read-only and disabled fields share the disabled face so both read as
			// non-editable instead of looking like an ordinary field.
			for _, state := range []style.State{{ReadOnly: true}, {Disabled: true}} {
				value := resolved(state)
				if got := paintColor(value.Paint.Background); got != sem.DisabledBackground {
					t.Fatalf("state %+v background = %v, want disabled %v", state, got, sem.DisabledBackground)
				}
				if got := colorOf(value.Text.Color); got != sem.DisabledForeground {
					t.Fatalf("state %+v text = %v, want disabled %v", state, got, sem.DisabledForeground)
				}
			}
		})
	}
}

func TestSelectPanelUsesMenuElevation(t *testing.T) {
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		visual := base.Provider{}.Select(&active, theme.SelectRequest{})
		panel := style.CascadePart(style.State{Open: true}, theme.SelectPartPanel, visual.Defaults, visual.Override)
		if panel.Paint == nil || panel.Paint.Border == nil || panel.Paint.Border.Width == nil {
			t.Fatal("select panel has no border")
		}
		if got := *panel.Paint.Border.Width; got != active.Components.Select.PanelBorderWidth {
			t.Fatalf("select panel border = %v, want %v", got, active.Components.Select.PanelBorderWidth)
		}
		if len(panel.Paint.Shadows) == 0 || panel.Paint.Shadows[0].Profile == nil ||
			*panel.Paint.Shadows[0].Profile != style.ShadowMenu {
			t.Fatal("select panel does not use the menu shadow profile")
		}
	}
}

func TestComboPanelUsesMenuElevation(t *testing.T) {
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		visual := base.Provider{}.Combo(&active, theme.ComboRequest{})
		panel := style.CascadePart(style.State{Open: true}, theme.ComboPartPanel, visual.Defaults, visual.Override)
		if panel.Paint == nil || panel.Paint.Border == nil || panel.Paint.Border.Width == nil {
			t.Fatal("combo panel has no border")
		}
		if got := *panel.Paint.Border.Width; got != active.Components.ComboBox.PanelBorderWidth {
			t.Fatalf("combo panel border = %v, want %v", got, active.Components.ComboBox.PanelBorderWidth)
		}
		if len(panel.Paint.Shadows) == 0 || panel.Paint.Shadows[0].Profile == nil ||
			*panel.Paint.Shadows[0].Profile != style.ShadowMenu {
			t.Fatal("combo panel does not use the menu shadow profile")
		}
	}
}

func TestDatePickerPanelUsesMenuElevation(t *testing.T) {
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		visual := base.Provider{}.DatePicker(&active, theme.DatePickerRequest{})
		panel := style.CascadePart(style.State{}, theme.DatePickerPartCalendarPanel, visual.Calendar)
		if panel.Paint == nil || panel.Paint.Border == nil || panel.Paint.Border.Width == nil {
			t.Fatal("date picker panel has no border")
		}
		if got := *panel.Paint.Border.Width; got != active.Components.DatePicker.PopoverBorderWidth {
			t.Fatalf("date picker panel border = %v, want %v", got, active.Components.DatePicker.PopoverBorderWidth)
		}
		if len(panel.Paint.Shadows) == 0 || panel.Paint.Shadows[0].Profile == nil ||
			*panel.Paint.Shadows[0].Profile != style.ShadowMenu {
			t.Fatal("date picker panel does not use the menu shadow profile")
		}
	}
}

func TestDatePickerOutsideCellIsMuted(t *testing.T) {
	active := base.Light()
	visual := base.Provider{}.DatePicker(&active, theme.DatePickerRequest{})
	current := visual.Cell(theme.DatePickerCellRequest{}).Resolve(style.State{})
	outside := visual.Cell(theme.DatePickerCellRequest{Outside: true}).Resolve(style.State{})
	currentColor, ok := resolve.Color(current.Text.Color)
	if !ok {
		t.Fatal("current-month cell has no text colour")
	}
	outsideColor, ok := resolve.Color(outside.Text.Color)
	if !ok {
		t.Fatal("outside-month cell has no text colour")
	}
	if outsideColor != active.Components.DatePicker.OutsideTextColor ||
		currentColor != active.Components.DatePicker.CurrentTextColor {
		t.Fatalf("day colours = current %v / outside %v, want configured values", currentColor, outsideColor)
	}
	if theme.ContrastRatio(outsideColor, active.Semantic().Background) >=
		theme.ContrastRatio(currentColor, active.Semantic().Background) {
		t.Fatalf("outside-month text = %v, want a muted colour than %v", outsideColor, currentColor)
	}
}

func TestDatePickerSelectedCellUsesWhiteOnPrimary(t *testing.T) {
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for _, active := range []theme.Theme{base.Light(), base.Dark()} {
		visual := base.Provider{}.DatePicker(&active, theme.DatePickerRequest{})
		cell := visual.Cell(theme.DatePickerCellRequest{
			State: style.State{Selected: true},
		}).Resolve(style.State{Selected: true})
		foreground, ok := resolve.Color(cell.Text.Color)
		if !ok {
			t.Fatal("selected cell has no text colour")
		}
		if foreground != white {
			t.Fatalf("selected date text = %v, want white", foreground)
		}
		primary := active.Semantic().Tone(theme.TonePrimary)
		if ratio := theme.ContrastRatio(foreground, primary.Fill); ratio < 4.5 {
			t.Fatalf("selected date contrast = %.2f, want at least 4.5", ratio)
		}
	}
}

func TestControlSizeTiers(t *testing.T) {
	for _, test := range []struct {
		name       string
		buttonSize button.Size
		inputSize  input.Size
		height     int
	}{
		{name: "small", buttonSize: button.Small, inputSize: input.Small, height: 24},
		{name: "medium", buttonSize: button.Medium, inputSize: input.Medium, height: 32},
		{name: "large", buttonSize: button.Large, inputSize: input.Large, height: 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, active := range []theme.Theme{base.Light(), base.Dark()} {
				harness := uitest.New(image.Pt(300, 100), active)
				for _, control := range []ui.Widget{
					button.New("保存 Save").Size(test.buttonSize),
					input.New("名称 Name").Size(test.inputSize),
				} {
					harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
						gtx.Constraints.Min = image.Point{}
						dims := control.Layout(ctx, gtx)
						if dims.Size.Y != test.height {
							t.Errorf("%T height = %d, want %d", control, dims.Size.Y, test.height)
						}
						return dims
					}))
				}
			}
		})
	}
}
