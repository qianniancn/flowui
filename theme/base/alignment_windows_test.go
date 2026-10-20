//go:build windows

package base_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/selectfield"
)

// Check actual glyphs, since centering a font's line box alone can leave CJK
// text visibly above the control's center when a fallback face has a large gap.
func TestDefaultTextVerticalAlignment(t *testing.T) {
	for _, mode := range []struct {
		name   string
		active theme.Theme
	}{
		{"light", base.Light()}, {"dark", base.Dark()},
	} {
		for _, scale := range []float32{1, 1.5, 2} {
			for _, label := range []string{"保存更改", "Save"} {
				for _, control := range []struct {
					name   string
					widget ui.Widget
					color  color.NRGBA
				}{
					{"button", button.New(label).Key("button"), mode.active.Semantic().Tone(theme.TonePrimary).OnFill},
					{"input", input.New(label).Key("input"), mode.active.Semantic().Foreground},
					{"select", selectfield.New([]selectfield.SelectItem{
						{Key: "value", Label: label},
					}).Key("select").SelectedKey("value"), mode.active.Semantic().Foreground},
				} {
					t.Run(fmt.Sprintf("%s/%g/%s/%s", mode.name, scale, control.name, label), func(t *testing.T) {
						viewport := image.Pt(int(240*scale), int(56*scale))
						window, err := headless.NewWindow(viewport.X, viewport.Y)
						if err != nil {
							t.Skipf("headless renderer unavailable: %v", err)
						}
						defer window.Release()
						harness := uitest.NewWithConfig(uitest.Config{
							Size: viewport, Theme: &mode.active,
							Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale},
						})
						var ops *op.Ops
						dims := harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
							ops = gtx.Ops
							paint.Fill(ops, mode.active.Semantic().Background)
							gtx.Constraints.Min = image.Point{}
							return control.widget.Layout(ctx, gtx)
						}))
						if err := window.Frame(ops); err != nil {
							t.Fatal(err)
						}
						pixels := image.NewRGBA(image.Rectangle{Max: viewport})
						if err := window.Screenshot(pixels); err != nil {
							t.Fatal(err)
						}
						// Exclude field borders and the select's trailing arrow.
						area := image.Rect(int(6*scale), int(3*scale),
							min(dims.Size.X-int(6*scale), int(105*scale)), dims.Size.Y-int(3*scale))
						ink := textPixelBounds(pixels, area, control.color)
						if ink.Empty() {
							t.Fatal("no text pixels rendered")
						}
						top, bottom := ink.Min.Y, dims.Size.Y-ink.Max.Y
						if abs(top-bottom) > int(3*scale) {
							t.Errorf("text has %dpx above and %dpx below in a %dpx control", top, bottom, dims.Size.Y)
						}
					})
				}
			}
		}
	}
}

func textPixelBounds(pixels *image.RGBA, area image.Rectangle, foreground color.NRGBA) image.Rectangle {
	var bounds image.Rectangle
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			pixel := pixels.RGBAAt(x, y)
			difference := abs(int(pixel.R)-int(foreground.R)) +
				abs(int(pixel.G)-int(foreground.G)) + abs(int(pixel.B)-int(foreground.B))
			if pixel.A == 255 && difference <= 30 {
				bounds = bounds.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return bounds
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
