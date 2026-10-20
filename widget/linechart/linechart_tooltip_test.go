package linechart

import (
	"image"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/uitest"
)

// A chart tooltip must not feed back into the clipping viewport around the
// chart. An inline panel joins the chart's ops, so its shadow is reported as
// visual overflow; the scroll viewport then reserves that outset, which moved
// the chart out from under a stationary pointer, dropped the hover, hid the
// tooltip, restored the layout, and repeated. The hover below must therefore
// keep both the chart size and the tooltip steady.
func TestLineChartTooltipKeepsScrollItemStable(t *testing.T) {
	activeTheme := base.Light()
	h := uitest.NewWithConfig(uitest.Config{Size: image.Pt(640, 420), Theme: &activeTheme})
	chart := New("chart", []Series{
		Values("visits", "Visits", []float64{28, 42, 36, 58, 64, 52, 78}),
	}).Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).Height(300)

	var size image.Point
	measured := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		dims := chart.Layout(ctx, gtx)
		size = dims.Size
		return dims
	})
	root := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Scroll(
			layoutui.Box(measured).Style(style.New().Padding(24)),
		).Layout(ctx, gtx)
	})
	h.Frame(root)
	h.Advance(time.Second)
	h.Frame(root)
	settled := size

	h.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(320, 200)})
	sizes := map[image.Point]int{}
	tooltips := 0
	for range 12 {
		h.Frame(root)
		sizes[size]++
		for _, label := range uitest.SemanticLabels(h.Semantics()) {
			if strings.HasPrefix(label, "Visits ") {
				tooltips++
			}
		}
		h.Advance(32 * time.Millisecond)
	}

	if len(sizes) != 1 {
		t.Fatalf("chart size changed while its tooltip was hovered: %v, settled at %v", sizes, settled)
	}
	if tooltips == 0 {
		t.Fatal("hovering the chart did not lay out a tooltip")
	}
}
