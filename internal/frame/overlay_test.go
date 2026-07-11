package frame

import (
	"image"
	"math"
	"reflect"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func TestLayoutOverlaysRecordsRootInputOps(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	router := new(input.Router)
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}
	clickable := new(widget.Clickable)
	RegisterOverlay(ctx, OverlayRequest{
		Key: "menu",
		Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			return clickable.Layout(gtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(20, 20)}
			})
		},
	})

	LayoutOverlays(ctx, gtx)
	router.Frame(gtx.Ops)
	router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  f32.Pt(10, 10),
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  f32.Pt(10, 10),
		},
	)
	eventGtx := gtx
	eventGtx.Ops = new(op.Ops)
	if !clickable.Clicked(eventGtx) {
		t.Fatal("overlay input operation was not attached to the root operation list")
	}
}

func TestPassiveOverlayDoesNotOwnOverlayInput(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var popupInteractive, tooltipInteractive bool
	RegisterOverlay(ctx, OverlayRequest{
		Key: "popup",
		Layout: func(_ layout.Context, _ image.Rectangle, interactive bool) layout.Dimensions {
			popupInteractive = interactive
			return layout.Dimensions{}
		},
	})
	RegisterOverlay(ctx, OverlayRequest{
		Key:     "tooltip",
		Passive: true,
		Layout: func(_ layout.Context, _ image.Rectangle, interactive bool) layout.Dimensions {
			tooltipInteractive = interactive
			return layout.Dimensions{}
		},
	})

	LayoutOverlays(ctx, gtx)
	if !popupInteractive || tooltipInteractive {
		t.Fatalf("interactive states = popup %v tooltip %v", popupInteractive, tooltipInteractive)
	}
	if !OverlayTopmost(ctx, OverlayLayerPopup, "popup") {
		t.Fatal("passive overlay replaced the active overlay owner")
	}
}

func TestOverlayAnchorAccumulatesPlacedTransforms(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(300, 200))
	var got image.Rectangle
	_, outer := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		_, inner := TrackOverlayPlacement(ctx, func() layout.Dimensions {
			RegisterOverlay(ctx, OverlayRequest{
				Key:       "select",
				Anchor:    image.Rect(1, 2, 11, 12),
				HasAnchor: true,
				Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
					got = anchor
					return layout.Dimensions{}
				},
			})
			return layout.Dimensions{Size: image.Pt(10, 10)}
		})
		inner.PlaceOffset(image.Pt(5, 7))
		return layout.Dimensions{Size: image.Pt(20, 20)}
	})
	outer.PlaceOffset(image.Pt(20, 30))

	LayoutOverlays(ctx, gtx)
	want := image.Rect(26, 39, 36, 49)
	if got != want {
		t.Fatalf("absolute anchor = %v, want %v", got, want)
	}
}

func TestOverlayAnchorUsesAffineBoundingBox(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var got image.Rectangle
	_, placement := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		RegisterOverlay(ctx, OverlayRequest{
			Key:       "scaled",
			Anchor:    image.Rect(0, 0, 10, 20),
			HasAnchor: true,
			Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
				got = anchor
				return layout.Dimensions{}
			},
		})
		return layout.Dimensions{}
	})
	placement.PlaceTransform(
		f32.AffineId().Offset(f32.Pt(10, 5)).Mul(
			f32.AffineId().Scale(f32.Point{}, f32.Pt(2, 2)),
		),
	)

	LayoutOverlays(ctx, gtx)
	if want := image.Rect(10, 5, 30, 45); got != want {
		t.Fatalf("affine anchor = %v, want %v", got, want)
	}
}

func TestOverlayPlacementClipControlsVisibilityWithoutChangingAnchor(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var got image.Rectangle
	_, placement := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		RegisterOverlay(ctx, OverlayRequest{
			Key:       "partially-visible",
			Anchor:    image.Rect(0, 35, 10, 45),
			HasAnchor: true,
			Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
				got = anchor
				return layout.Dimensions{}
			},
		})
		return layout.Dimensions{}
	})
	placement.PlaceOffset(image.Point{})
	placement.ClipTo(image.Rect(0, 0, 100, 40))

	LayoutOverlays(ctx, gtx)
	if want := image.Rect(0, 35, 10, 45); got != want {
		t.Fatalf("partially clipped anchor = %v, want full anchor %v", got, want)
	}

	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx.Ops.Reset()
	called := false
	_, hidden := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		RegisterOverlay(ctx, OverlayRequest{
			Key:       "hidden",
			Anchor:    image.Rect(0, 50, 10, 60),
			HasAnchor: true,
			Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
				called = true
				return layout.Dimensions{}
			},
		})
		return layout.Dimensions{}
	})
	hidden.PlaceOffset(image.Point{})
	hidden.ClipTo(image.Rect(0, 0, 100, 40))
	LayoutOverlays(ctx, gtx)
	if called {
		t.Fatal("fully clipped anchor was laid out")
	}
}

func TestOverlayPlacementRequiresClipAndViewportToShareVisibleAnchor(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	called := false
	_, placement := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		RegisterOverlay(ctx, OverlayRequest{
			Key:       "clipped-outside-viewport",
			Anchor:    image.Rect(-20, 10, 20, 30),
			HasAnchor: true,
			Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
				called = true
				return layout.Dimensions{}
			},
		})
		return layout.Dimensions{}
	})
	placement.PlaceOffset(image.Point{})
	placement.ClipTo(image.Rect(-20, 0, -10, 100))

	LayoutOverlays(ctx, gtx)
	if called {
		t.Fatal("anchor clipped to the area outside the viewport was laid out")
	}
}

func TestOverlayPlacementOpacityInheritsAcrossDynamicOverlays(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	RegisterOverlay(ctx, OverlayRequest{
		Key: "root",
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			return layout.Dimensions{}
		},
	})

	_, sibling := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		RegisterOverlay(ctx, OverlayRequest{
			Key: "sibling",
			Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
				return layout.Dimensions{}
			},
		})
		return layout.Dimensions{}
	})
	sibling.PlaceOffset(image.Point{})
	sibling.SetOpacity(.75)

	_, parent := TrackOverlayPlacement(ctx, func() layout.Dimensions {
		RegisterOverlay(ctx, OverlayRequest{
			Key: "parent",
			Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
				_, child := TrackOverlayPlacement(ctx, func() layout.Dimensions {
					RegisterOverlay(ctx, OverlayRequest{
						Key: "child",
						Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
							RegisterOverlay(ctx, OverlayRequest{
								Key: "grandchild",
								Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
									return layout.Dimensions{}
								},
							})
							return layout.Dimensions{}
						},
					})
					return layout.Dimensions{}
				})
				child.PlaceOffset(image.Point{})
				child.SetOpacity(.4)
				return layout.Dimensions{}
			},
		})
		return layout.Dimensions{}
	})
	parent.PlaceOffset(image.Point{})
	parent.SetOpacity(.5)

	LayoutOverlays(ctx, gtx)
	want := map[string]float32{
		"root":       1,
		"sibling":    .75,
		"parent":     .5,
		"child":      .2,
		"grandchild": .2,
	}
	if len(ctx.overlays.requests) != len(want) {
		t.Fatalf("overlay requests = %d, want %d", len(ctx.overlays.requests), len(want))
	}
	for _, request := range ctx.overlays.requests {
		_, ok, _, opacity := resolveOverlayAnchor(request)
		if !ok {
			t.Fatalf("overlay %q did not resolve", request.Key)
		}
		if expected, exists := want[request.Key]; !exists {
			t.Fatalf("unexpected overlay %q", request.Key)
		} else if math.Abs(float64(opacity-expected)) > 1e-6 {
			t.Errorf("overlay %q opacity = %v, want %v", request.Key, opacity, expected)
		}
	}
}

func TestOverlayPlacementOpacityClampsToUnitRange(t *testing.T) {
	tests := []struct {
		name    string
		opacity float32
		want    float32
	}{
		{name: "below zero", opacity: -1, want: 0},
		{name: "above one", opacity: 2, want: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, _ := overlayTestContext(image.Pt(100, 100))
			_, placement := TrackOverlayPlacement(ctx, func() layout.Dimensions {
				RegisterOverlay(ctx, OverlayRequest{
					Key: "clamped",
					Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
						return layout.Dimensions{}
					},
				})
				return layout.Dimensions{}
			})
			placement.PlaceOffset(image.Point{})
			placement.SetOpacity(test.opacity)

			_, ok, _, got := resolveOverlayAnchor(ctx.overlays.requests[0])
			if !ok {
				t.Fatal("overlay did not resolve")
			}
			if got != test.want {
				t.Fatalf("opacity = %v, want %v", got, test.want)
			}
		})
	}
}

func TestOverlayHostOrdersPopupBeforeModalAndTracksTopmost(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var first []string
	registerOrderProbe(ctx, "dialog", OverlayLayerModal, &first)
	registerOrderProbe(ctx, "menu", OverlayLayerPopup, &first)
	LayoutOverlays(ctx, gtx)
	if want := []string{"menu:true", "dialog:true"}; !reflect.DeepEqual(first, want) {
		t.Fatalf("first frame order = %v, want %v", first, want)
	}

	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx.Ops.Reset()
	var second []string
	registerOrderProbe(ctx, "dialog", OverlayLayerModal, &second)
	registerOrderProbe(ctx, "menu", OverlayLayerPopup, &second)
	LayoutOverlays(ctx, gtx)
	if want := []string{"menu:false", "dialog:true"}; !reflect.DeepEqual(second, want) {
		t.Fatalf("second frame order = %v, want %v", second, want)
	}
}

func TestOverlayFocusScopeTracksNearestTopmostTailAncestor(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	registerFocusScopeProbe(ctx, "child-one", false)
	LayoutOverlays(ctx, gtx)
	if !OverlayTopmost(ctx, OverlayLayerPopup, "child-one") {
		t.Fatal("first dynamic child was not the visual topmost overlay")
	}
	if !OverlayFocusScopeTopmost(ctx, OverlayLayerModal, "outer") || !OverlayFocusScopeBecameTopmost(ctx, OverlayLayerModal, "outer") {
		t.Fatal("outer modal did not become the initial focus scope")
	}

	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx.Ops.Reset()
	registerFocusScopeProbe(ctx, "child-two", false)
	LayoutOverlays(ctx, gtx)
	if !OverlayBecameTopmost(ctx, OverlayLayerPopup, "child-two") {
		t.Fatal("replacement child did not become visual topmost")
	}
	if !OverlayFocusScopeTopmost(ctx, OverlayLayerModal, "outer") {
		t.Fatal("outer modal lost its focus scope when its child changed")
	}
	if OverlayFocusScopeBecameTopmost(ctx, OverlayLayerModal, "outer") {
		t.Fatal("unchanged modal focus scope was reported as newly topmost")
	}

	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx.Ops.Reset()
	registerFocusScopeProbe(ctx, "inner", true)
	LayoutOverlays(ctx, gtx)
	if !OverlayFocusScopeTopmost(ctx, OverlayLayerModal, "inner") || !OverlayFocusScopeBecameTopmost(ctx, OverlayLayerModal, "inner") {
		t.Fatal("nested modal did not replace the outer focus scope")
	}
	if OverlayFocusScopeTopmost(ctx, OverlayLayerModal, "outer") {
		t.Fatal("outer modal remained the active focus scope below a nested modal")
	}
}

func TestOverlayHostOrdersSiblingPopupBelowNestedModal(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var order []string
	RegisterOverlay(ctx, OverlayRequest{
		Key:   "outer",
		Layer: OverlayLayerModal,
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			order = append(order, "outer")
			RegisterOverlay(ctx, OverlayRequest{
				Key:   "inner",
				Layer: OverlayLayerModal,
				Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
					order = append(order, "inner")
					return layout.Dimensions{}
				},
				Tail: func(layout.Context) {},
			})
			RegisterOverlay(ctx, OverlayRequest{
				Key:   "menu",
				Layer: OverlayLayerPopup,
				Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
					order = append(order, "menu")
					return layout.Dimensions{}
				},
			})
			return layout.Dimensions{}
		},
		Tail: func(layout.Context) {},
	})

	LayoutOverlays(ctx, gtx)
	if want := []string{"outer", "menu", "inner"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("dynamic sibling order = %v, want %v", order, want)
	}
	if !OverlayTopmost(ctx, OverlayLayerModal, "inner") {
		t.Fatal("nested modal was not the visual topmost overlay")
	}
	if !OverlayFocusScopeTopmost(ctx, OverlayLayerModal, "inner") {
		t.Fatal("nested modal did not own the topmost focus scope")
	}
}

func TestOverlayHostKeepsPopupRegisteredByNestedModalAboveIt(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var order []string
	RegisterOverlay(ctx, OverlayRequest{
		Key:   "outer",
		Layer: OverlayLayerModal,
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			order = append(order, "outer")
			RegisterOverlay(ctx, OverlayRequest{
				Key:   "inner",
				Layer: OverlayLayerModal,
				Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
					order = append(order, "inner")
					RegisterOverlay(ctx, OverlayRequest{
						Key:   "menu",
						Layer: OverlayLayerPopup,
						Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
							order = append(order, "menu")
							return layout.Dimensions{}
						},
					})
					return layout.Dimensions{}
				},
				Tail: func(layout.Context) {},
			})
			return layout.Dimensions{}
		},
		Tail: func(layout.Context) {},
	})

	LayoutOverlays(ctx, gtx)
	if want := []string{"outer", "inner", "menu"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("nested popup order = %v, want %v", order, want)
	}
	if !OverlayTopmost(ctx, OverlayLayerPopup, "menu") {
		t.Fatal("popup registered by the nested modal was not visually topmost")
	}
	if !OverlayFocusScopeTopmost(ctx, OverlayLayerModal, "inner") {
		t.Fatal("nested modal lost its focus scope to its popup")
	}
}

func registerFocusScopeProbe(ctx *Context, childKey string, childHasTail bool) {
	RegisterOverlay(ctx, OverlayRequest{
		Key:   "outer",
		Layer: OverlayLayerModal,
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			child := OverlayRequest{
				Key:   childKey,
				Layer: OverlayLayerPopup,
				Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
					return layout.Dimensions{}
				},
			}
			if childHasTail {
				child.Layer = OverlayLayerModal
				child.Tail = func(layout.Context) {}
			}
			RegisterOverlay(ctx, child)
			return layout.Dimensions{}
		},
		Tail: func(layout.Context) {},
	})
}

func TestOverlayEventOwnerTransfersOneFrameAfterCurrentTop(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	layoutFrame := func(keys ...string) map[string]bool {
		interactive := make(map[string]bool, len(keys))
		for _, key := range keys {
			key := key
			RegisterOverlay(ctx, OverlayRequest{
				Key: key,
				Layout: func(_ layout.Context, _ image.Rectangle, ownsEvents bool) layout.Dimensions {
					interactive[key] = ownsEvents
					return layout.Dimensions{}
				},
			})
		}
		LayoutOverlays(ctx, gtx)
		return interactive
	}
	beginNextFrame := func() {
		BeginFrameWithViewport(ctx, image.Pt(100, 100))
		gtx.Ops.Reset()
	}

	first := layoutFrame("a")
	if !first["a"] {
		t.Fatal("frame 1: first overlay did not receive initial event ownership")
	}
	if !OverlayTopmost(ctx, OverlayLayerPopup, "a") {
		t.Fatal("frame 1: a is not the current top")
	}
	if !OverlayBecameTopmost(ctx, OverlayLayerPopup, "a") {
		t.Fatal("frame 1: a did not report becoming current top")
	}

	beginNextFrame()
	second := layoutFrame("a", "b")
	if !second["a"] || second["b"] {
		t.Fatalf("frame 2 event owners = %v, want a=true b=false", second)
	}
	if OverlayTopmost(ctx, OverlayLayerPopup, "a") || !OverlayTopmost(ctx, OverlayLayerPopup, "b") {
		t.Fatal("frame 2: current top did not move from a to b")
	}
	if !OverlayInteractive(ctx, OverlayLayerPopup, "a") || OverlayInteractive(ctx, OverlayLayerPopup, "b") {
		t.Fatal("frame 2: event ownership moved before the next frame")
	}
	if !OverlayBecameTopmost(ctx, OverlayLayerPopup, "b") {
		t.Fatal("frame 2: b did not report becoming current top")
	}

	beginNextFrame()
	third := layoutFrame("a", "b")
	if third["a"] || !third["b"] {
		t.Fatalf("frame 3 event owners = %v, want a=false b=true", third)
	}
	if !OverlayTopmost(ctx, OverlayLayerPopup, "b") || !OverlayInteractive(ctx, OverlayLayerPopup, "b") {
		t.Fatal("frame 3: b is not both current top and event owner")
	}
	if OverlayBecameTopmost(ctx, OverlayLayerPopup, "b") {
		t.Fatal("frame 3: stable current top reported becoming top again")
	}
}

func TestOverlayHostKeepsNonTopPanelContextEnabledForNestedTriggers(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	registerOrderProbe(ctx, "menu", OverlayLayerPopup, new([]string))
	registerOrderProbe(ctx, "dialog", OverlayLayerModal, new([]string))
	LayoutOverlays(ctx, gtx)

	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx.Ops.Reset()
	enabled := make(map[string]bool)
	for _, request := range []OverlayRequest{
		{
			Key:   "menu",
			Layer: OverlayLayerPopup,
			Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
				enabled["menu"] = gtx.Enabled()
				return layout.Dimensions{}
			},
		},
		{
			Key:   "dialog",
			Layer: OverlayLayerModal,
			Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
				enabled["dialog"] = gtx.Enabled()
				return layout.Dimensions{}
			},
		},
	} {
		RegisterOverlay(ctx, request)
	}
	LayoutOverlays(ctx, gtx)
	if !enabled["menu"] {
		t.Fatal("non-topmost parent lost the event source needed by nested triggers")
	}
	if !enabled["dialog"] {
		t.Fatal("topmost modal received a disabled event source")
	}
}

func TestOverlayHostLayoutsDynamicallyRegisteredOverlay(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var order []string
	RegisterOverlay(ctx, OverlayRequest{
		Key: "parent",
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			order = append(order, "parent")
			RegisterOverlay(ctx, OverlayRequest{
				Key: "child",
				Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
					order = append(order, "child")
					return layout.Dimensions{}
				},
			})
			return layout.Dimensions{}
		},
	})

	LayoutOverlays(ctx, gtx)
	if want := []string{"parent", "child"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("layout order = %v, want %v", order, want)
	}
}

func TestOverlayHostRestoresRegistrationKeyScope(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	pop := PushKey(ctx, "settings")
	var got string
	RegisterOverlay(ctx, OverlayRequest{
		Key: "dialog",
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			got = FullKey(ctx, "close")
			return layout.Dimensions{}
		},
	})
	pop()

	LayoutOverlays(ctx, gtx)
	if got != "settings/close" {
		t.Fatalf("overlay key scope = %q, want settings/close", got)
	}
	if root := FullKey(ctx, "close"); root != "close" {
		t.Fatalf("root key scope leaked after overlay layout: %q", root)
	}
}

func TestAfterOverlaysRunsAfterDynamicLayout(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	var order []string
	RegisterOverlay(ctx, OverlayRequest{
		Key: "menu",
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			order = append(order, "layout")
			return layout.Dimensions{}
		},
	})
	AfterOverlays(ctx, func() { order = append(order, "cleanup") })

	LayoutOverlays(ctx, gtx)
	if want := []string{"layout", "cleanup"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("callback order = %v, want %v", order, want)
	}
}

func TestOverlayTailCannotRegisterAnotherOverlay(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	RegisterOverlay(ctx, OverlayRequest{
		Key: "modal",
		Layout: func(layout.Context, image.Rectangle, bool) layout.Dimensions {
			return layout.Dimensions{}
		},
		Tail: func(layout.Context) {
			RegisterOverlay(ctx, OverlayRequest{
				Key: "too-late",
				Layout: func(layout.Context, image.Rectangle, bool) layout.Dimensions {
					return layout.Dimensions{}
				},
			})
		},
	})

	defer func() {
		if recover() == nil {
			t.Fatal("overlay tail registered new overlay work")
		}
	}()
	LayoutOverlays(ctx, gtx)
}

func TestOverlayHostPreservesDisabledContext(t *testing.T) {
	ctx, gtx := overlayTestContext(image.Pt(100, 100))
	enabled := true
	RegisterOverlay(ctx, OverlayRequest{
		Key:      "disabled",
		Disabled: true,
		Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			enabled = gtx.Enabled()
			return layout.Dimensions{}
		},
	})
	LayoutOverlays(ctx, gtx)
	if enabled {
		t.Fatal("disabled registration was laid out with an enabled context")
	}
}

func TestBeginFrameReleasesPreviousOverlayRequests(t *testing.T) {
	ctx, _ := overlayTestContext(image.Pt(100, 100))
	RegisterOverlay(ctx, OverlayRequest{
		Key: "menu",
		Layout: func(_ layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			return layout.Dimensions{}
		},
	})
	if len(ctx.overlays.requests) != 1 {
		t.Fatalf("overlay requests = %d, want 1", len(ctx.overlays.requests))
	}

	BeginFrameWithViewport(ctx, image.Pt(100, 100))
	if len(ctx.overlays.requests) != 0 {
		t.Fatalf("overlay requests after BeginFrame = %d, want 0", len(ctx.overlays.requests))
	}
	if slots := ctx.overlays.requests[:cap(ctx.overlays.requests)]; len(slots) > 0 && slots[0] != nil {
		t.Fatal("BeginFrame retained the previous overlay request")
	}
}

func registerOrderProbe(ctx *Context, key string, layer OverlayLayer, calls *[]string) {
	RegisterOverlay(ctx, OverlayRequest{
		Key:   key,
		Layer: layer,
		Layout: func(_ layout.Context, _ image.Rectangle, interactive bool) layout.Dimensions {
			*calls = append(*calls, key+":"+map[bool]string{false: "false", true: "true"}[interactive])
			return layout.Dimensions{}
		},
	})
}

func overlayTestContext(viewport image.Point) (*Context, layout.Context) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrameWithViewport(ctx, viewport)
	router := new(input.Router)
	return ctx, layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}
}
