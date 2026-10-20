package colorpicker

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type colorPickerRendererCounts struct {
	wheel        int
	area         int
	slider       int
	swatch       int
	swatchPicker int
	triggerLabel int
	triggerFocus int
	panel        int
}

type recordingColorPickerProvider struct {
	theme.ComponentVisualProvider
	counts *colorPickerRendererCounts
}

func (provider *recordingColorPickerProvider) ColorPicker(active *theme.Theme, request theme.ColorPickerRequest) theme.ColorPickerVisual {
	visual := provider.ComponentVisualProvider.ColorPicker(active, request)
	counts := provider.counts

	wheel := visual.Wheel.Render
	visual.Wheel.Render = func(gtx layout.Context, value theme.ColorWheelRenderRequest) {
		counts.wheel++
		wheel(gtx, value)
	}
	area := visual.Area.Render
	visual.Area.Render = func(gtx layout.Context, value theme.ColorAreaRenderRequest) {
		counts.area++
		area(gtx, value)
	}
	slider := visual.Slider.Render
	visual.Slider.Render = func(gtx layout.Context, value theme.ColorSliderRenderRequest) {
		counts.slider++
		slider(gtx, value)
	}
	swatch := visual.Swatch.Render
	visual.Swatch.Render = func(gtx layout.Context, value theme.ColorSwatchRenderRequest) {
		counts.swatch++
		swatch(gtx, value)
	}
	swatchPicker := visual.SwatchPicker.Render
	visual.SwatchPicker.Render = func(gtx layout.Context, value theme.ColorSwatchPickerItemRequest) {
		counts.swatchPicker++
		swatchPicker(gtx, value)
	}
	triggerLabel := visual.TriggerLabel
	visual.TriggerLabel = func(gtx layout.Context, value string) layout.Dimensions {
		counts.triggerLabel++
		return triggerLabel(gtx, value)
	}
	triggerFocus := visual.TriggerFocus
	visual.TriggerFocus = func(gtx layout.Context, size image.Point, opacity float32) {
		counts.triggerFocus++
		triggerFocus(gtx, size, opacity)
	}
	panel := visual.Panel
	visual.Panel = func(gtx layout.Context, size image.Point) {
		counts.panel++
		panel(gtx, size)
	}
	return visual
}

func TestColorPickerComponentTreeUsesProviderRenderersAndOverlay(t *testing.T) {
	active := *testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	counts := &colorPickerRendererCounts{}
	active.VisualProvider = &recordingColorPickerProvider{
		ComponentVisualProvider: base,
		counts:                  counts,
	}
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	router := new(input.Router)
	now := time.Unix(1, 0)
	value := color.NRGBA{R: 4, G: 133, B: 247, A: 180}

	layoutColorWidgetFrame(ctx, router, ColorWheel("renderer-wheel", value), now)
	layoutColorWidgetFrame(ctx, router, ColorSlider("renderer-slider", value, ColorChannelHue), now)
	layoutColorWidgetFrame(ctx, router, ColorSwatch(value), now)
	layoutColorWidgetFrame(ctx, router, ColorSwatchPicker("renderer-swatches", value, []color.NRGBA{value}), now)

	picker := New("renderer-picker", value).
		Label("Accent color").
		Alpha(true).
		Presets([]color.NRGBA{{R: 200, G: 100, B: 50, A: 255}})
	layoutColorPickerFrame(ctx, router, picker, now)
	state := colorPickerTestState(ctx, "renderer-picker")
	state.open = true
	layoutColorPickerFrame(ctx, router, picker, now.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, now.Add(colorPickerEnterDuration+time.Millisecond))

	if !core.HasVisibleOverlay(ctx) || !core.OverlayTopmost(ctx, core.OverlayLayerPopup, "renderer-picker") {
		t.Fatal("color picker did not register a visible topmost popup overlay")
	}
	if counts.wheel == 0 || counts.area == 0 || counts.slider < 2 || counts.swatch == 0 || counts.swatchPicker == 0 || counts.triggerLabel == 0 || counts.triggerFocus == 0 || counts.panel == 0 {
		t.Fatalf("provider renderer counts = %#v, want every ColorPicker renderer to be used", *counts)
	}

	state.open = false
	layoutColorPickerFrame(ctx, router, picker, now.Add(2*colorPickerEnterDuration+time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, now.Add(3*colorPickerEnterDuration+time.Millisecond))
	if core.HasVisibleOverlay(ctx) {
		t.Fatal("color picker overlay remained visible after close")
	}
}
