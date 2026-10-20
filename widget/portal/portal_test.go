package portal

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

type portalFixedWidget image.Point

func (w portalFixedWidget) Layout(*core.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Point(w)}
}

func TestPortalResolvesAnchorAndRootContent(t *testing.T) {
	ctx, gtx := portalTestContext(image.Pt(200, 120))
	var gotAnchor image.Rectangle
	var gotInteractive bool
	portal := New("inspector", true, portalFixedWidget(image.Pt(20, 10)), func(anchor image.Rectangle, interactive bool) core.Widget {
		gotAnchor = anchor
		gotInteractive = interactive
		return portalFixedWidget(image.Pt(40, 30))
	})

	_, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return portal.Layout(ctx, gtx)
	})
	placement.PlaceOffset(image.Pt(15, 25))
	core.LayoutOverlays(ctx, gtx)

	if gotAnchor != image.Rect(15, 25, 35, 35) || !gotInteractive {
		t.Fatalf("Portal anchor = %v interactive %v", gotAnchor, gotInteractive)
	}
	if !core.HasTopOverlay(ctx) {
		t.Fatal("visible portal was not registered with the root host")
	}
}

func TestPassivePortalDoesNotOwnRootInput(t *testing.T) {
	ctx, gtx := portalTestContext(image.Pt(100, 80))
	interactive := true
	New("decoration", true, nil, func(_ image.Rectangle, boolValue bool) core.Widget {
		interactive = boolValue
		return portalFixedWidget(image.Pt(10, 10))
	}).Passive(true).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	if interactive || core.HasTopOverlay(ctx) {
		t.Fatalf("passive portal = interactive %v top %v", interactive, core.HasTopOverlay(ctx))
	}
}

func TestPortalDisablesRootContent(t *testing.T) {
	for _, test := range []struct {
		name     string
		parent   bool
		explicit bool
	}{
		{name: "parent", parent: true},
		{name: "explicit", explicit: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, gtx := portalTestContext(image.Pt(100, 80))
			if test.parent {
				gtx = gtx.Disabled()
			}
			enabled := true
			New("disabled", true, nil, func(image.Rectangle, bool) core.Widget {
				return core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
					enabled = gtx.Enabled()
					return layout.Dimensions{}
				})
			}).Disabled(test.explicit).Layout(ctx, gtx)
			core.LayoutOverlays(ctx, gtx)
			if enabled {
				t.Fatal("portal content remained enabled")
			}
		})
	}
}

func TestPortalRejectsNilRootContent(t *testing.T) {
	ctx, gtx := portalTestContext(image.Pt(100, 80))
	New("nil-content", true, nil, func(image.Rectangle, bool) core.Widget { return nil }).Layout(ctx, gtx)
	defer func() {
		if recover() == nil {
			t.Fatal("portal accepted nil root content")
		}
	}()
	core.LayoutOverlays(ctx, gtx)
}

func portalTestContext(viewport image.Point) (*core.Context, layout.Context) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	core.BeginFrameWithViewport(ctx, viewport)
	router := new(input.Router)
	return ctx, layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}
}
