package main

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/FlowUI/uitest"
)

func TestCustomTriggerClickAndSemantics(t *testing.T) {
	clicked := false
	button := customTrigger{
		key:     "test-trigger",
		label:   "Test action",
		onClick: func() { clicked = true },
	}
	harness := uitest.New(image.Pt(320, 120))

	harness.Frame(button)
	if node, ok := semanticNode(harness.Router().AppendSemantics(nil), semantic.Button, "Test action"); !ok || node.Desc.Disabled {
		t.Fatal("custom trigger did not expose enabled button semantics")
	}
	harness.Click(f32.Pt(20, 18))
	harness.Frame(button)
	if !clicked {
		t.Fatal("custom trigger did not handle the pointer click")
	}
}

func TestTransformedButtonAnimatesItsHitArea(t *testing.T) {
	clicked := false
	button := transformedButton{button: customTrigger{
		key:     "test-transform",
		label:   "Transform",
		onClick: func() { clicked = true },
	}}
	harness := uitest.New(image.Pt(320, 120))

	harness.Frame(button)
	harness.Click(f32.Pt(270, 60))
	harness.Frame(button)
	if clicked {
		t.Fatal("untransformed button handled a click outside its bounds")
	}

	button.transformed = true
	harness.Frame(button)
	harness.Advance(240 * time.Millisecond)
	harness.Frame(button)
	harness.Click(f32.Pt(270, 60))
	harness.Frame(button)
	if !clicked {
		t.Fatal("transformed button hit area did not follow its Gio transform")
	}
}

func TestTransformedButtonsKeepIndependentTweenState(t *testing.T) {
	first := transformedButton{
		button:      customTrigger{key: "first-transform", label: "First"},
		transformed: true,
	}
	second := transformedButton{
		button:      customTrigger{key: "second-transform", label: "Second"},
		transformed: true,
	}
	siblingTween := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		ui.Tween("first-transform:motion", 1).Value(ctx, gtx)
		return layout.Dimensions{}
	})
	harness := uitest.New(image.Pt(640, 240))

	harness.Frame(ui.Row(first, second, siblingTween))
	if _, ok := semanticNode(harness.Router().AppendSemantics(nil), semantic.Button, "First"); !ok {
		t.Fatal("first transformed button is missing")
	}
	if _, ok := semanticNode(harness.Router().AppendSemantics(nil), semantic.Button, "Second"); !ok {
		t.Fatal("second transformed button is missing")
	}
}

func TestPortalPanelDisabledAndCloseInteraction(t *testing.T) {
	closed := false
	harness := uitest.New(image.Pt(360, 260))
	anchor := image.Rect(20, 10, 210, 46)
	enabled := portalPanel(anchor, true, func() { closed = true })
	harness.Frame(enabled)
	node, ok := semanticButtonContainingLabel(harness.Router().AppendSemantics(nil), "Close")
	if !ok {
		t.Fatal("interactive portal panel did not expose its close button")
	}
	center := node.Desc.Bounds.Min.Add(node.Desc.Bounds.Max).Div(2)

	disabled := portalPanel(anchor, false, func() { closed = true })
	harness.Frame(disabled)
	harness.Click(f32.Pt(float32(center.X), float32(center.Y)))
	harness.Frame(disabled)
	if closed {
		t.Fatal("non-interactive portal panel handled a close click")
	}

	harness.Frame(enabled)
	harness.Click(f32.Pt(float32(center.X), float32(center.Y)))
	harness.Frame(enabled)
	if !closed {
		t.Fatal("portal panel close button did not handle the click")
	}
}

func semanticNode(nodes []input.SemanticNode, class semantic.ClassOp, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == class && node.Desc.Label == label {
			return node, true
		}
		if child, ok := semanticNode(node.Children, class, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func semanticButtonContainingLabel(nodes []input.SemanticNode, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button && semanticTreeHasLabel(node.Children, label) {
			return node, true
		}
		if child, ok := semanticButtonContainingLabel(node.Children, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func semanticTreeHasLabel(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || semanticTreeHasLabel(node.Children, label) {
			return true
		}
	}
	return false
}
