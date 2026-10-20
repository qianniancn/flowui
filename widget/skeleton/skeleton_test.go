package skeleton

import (
	"image"
	"testing"
	"time"

	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestAnimationAndSize(t *testing.T) {
	for _, tc := range []struct {
		name             string
		motion, animated bool
	}{
		{name: "pulse", motion: true, animated: true},
		{name: "static", motion: true},
		{name: "reduced motion", animated: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			active := base.Light()
			active.Motion.Enabled = tc.motion
			h := uitest.New(image.Pt(200, 100), active)
			root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return New().Height(40).Circle().Animated(tc.animated).Layout(ctx, gtx)
			})
			h.Frame(root)
			h.Router().WakeupTime()
			h.Advance(time.Second)
			dims := h.Frame(root)
			if dims.Size != image.Pt(40, 40) {
				t.Fatalf("circle size = %v, want (40,40)", dims.Size)
			}
			_, wake := h.Router().WakeupTime()
			if wake != (tc.motion && tc.animated) {
				t.Fatalf("requests frame = %v", wake)
			}
		})
	}
}

func TestSizeInStretchingParent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		widget Widget
		want   image.Point
	}{
		{name: "default width", widget: New(), want: image.Pt(200, 16)},
		{name: "fixed width", widget: New().Width(80), want: image.Pt(80, 16)},
		{name: "circle", widget: New().Circle().Height(40), want: image.Pt(40, 40)},
		{name: "oversized circle", widget: New().Circle().Height(300), want: image.Pt(100, 100)},
		{name: "oversized width", widget: New().Width(300), want: image.Pt(200, 16)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := uitest.New(image.Pt(200, 100), base.Light())
			if got := h.Frame(tc.widget).Size; got != tc.want {
				t.Fatalf("size = %v, want %v", got, tc.want)
			}
		})
	}
}
