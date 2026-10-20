package layout_test

import (
	"image"
	"testing"

	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
)

func TestLayoutWidgetsImplementUIWidget(t *testing.T) {
	child := ui.WidgetFunc(func(_ *ui.Context, _ giolayout.Context) giolayout.Dimensions {
		return giolayout.Dimensions{}
	})

	values := []ui.Widget{
		layout.Box(child),
		layout.Center(child),
		layout.Column(child).Gap(4).AlignMiddle().Spacing(layout.SpaceBetween),
		layout.Row(child).Gap(4).AlignBaseline().Spacing(layout.SpaceAround),
		layout.Wrap(child).Gap(4).AlignEnd(),
		layout.Grid(2, child).Gap(4),
		layout.AutoGrid(120, child),
		layout.Expanded(child),
		layout.Flexible(1, child),
		layout.Spacer(8, 8),
		layout.Divider(),
		layout.Divider().Vertical(),
		layout.Stack(layout.Stacked(child), layout.Overlay(child)),
		layout.Key("stable", child),
		layout.Scroll(child),
		layout.Scrollbar(child),
		layout.SplitPane(child, child),
		layout.AspectRatio(1, child),
	}

	for index, value := range values {
		if value == nil {
			t.Fatalf("layout %d returned a nil ui.Widget", index)
		}
	}
}

func TestLayoutHelpersAcceptPublicWidgetAndContextAliases(t *testing.T) {
	// Assigning the public function types verifies that callers only need
	// ui.Context/ui.Widget aliases; no internal package is required.
	var _ layout.VisualOverflowDraw = func(*ui.Context, giolayout.Context, image.Rectangle) {}
	var _ func(*ui.Context, giolayout.Context, ui.Widget, layout.VisualOverflowDraw) giolayout.Dimensions = layout.VisualOverflow
	var _ func(*ui.Context, giolayout.Context, ui.Widget, unit.Dp, unit.Dp, unit.Dp, unit.Dp) giolayout.Dimensions = layout.VisualOutset
	var _ func(*ui.Context, ...ui.Widget) = layout.PrepareFieldAssociations
}
