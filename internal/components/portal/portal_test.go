package portal

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

type fixedWidget image.Point

func (w fixedWidget) Layout(*frame.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Point(w)}
}

func TestPortalResolvesAnchorAndRootContent(t *testing.T) {
	ctx, gtx := portalTestContext(image.Pt(200, 120))
	var gotAnchor image.Rectangle
	var gotInteractive bool
	portal := New("inspector", true, fixedWidget(image.Pt(20, 10)), func(anchor image.Rectangle, interactive bool) frame.Widget {
		gotAnchor = anchor
		gotInteractive = interactive
		return fixedWidget(image.Pt(40, 30))
	})

	_, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return portal.Layout(ctx, gtx)
	})
	placement.PlaceOffset(image.Pt(15, 25))
	frame.LayoutOverlays(ctx, gtx)

	if gotAnchor != image.Rect(15, 25, 35, 35) || !gotInteractive {
		t.Fatalf("Portal anchor = %v interactive %v", gotAnchor, gotInteractive)
	}
	if !frame.HasTopOverlay(ctx) {
		t.Fatal("visible Portal was not registered with the root host")
	}
}

func TestPassivePortalDoesNotOwnRootInput(t *testing.T) {
	ctx, gtx := portalTestContext(image.Pt(100, 80))
	interactive := true
	New("decoration", true, nil, func(_ image.Rectangle, boolValue bool) frame.Widget {
		interactive = boolValue
		return fixedWidget(image.Pt(10, 10))
	}).Passive(true).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	if interactive || frame.HasTopOverlay(ctx) {
		t.Fatalf("passive Portal = interactive %v top %v", interactive, frame.HasTopOverlay(ctx))
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
			New("disabled", true, nil, func(image.Rectangle, bool) frame.Widget {
				return frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
					enabled = gtx.Enabled()
					return layout.Dimensions{}
				})
			}).Disabled(test.explicit).Layout(ctx, gtx)
			frame.LayoutOverlays(ctx, gtx)
			if enabled {
				t.Fatal("Portal content remained enabled")
			}
		})
	}
}

func TestPortalRejectsNilRootContent(t *testing.T) {
	ctx, gtx := portalTestContext(image.Pt(100, 80))
	New("nil-content", true, nil, func(image.Rectangle, bool) frame.Widget { return nil }).Layout(ctx, gtx)
	defer func() {
		if recover() == nil {
			t.Fatal("Portal accepted nil root content")
		}
	}()
	frame.LayoutOverlays(ctx, gtx)
}

func portalTestContext(viewport image.Point) (*frame.Context, layout.Context) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	frame.BeginFrameWithViewport(ctx, viewport)
	router := new(input.Router)
	return ctx, layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}
}
