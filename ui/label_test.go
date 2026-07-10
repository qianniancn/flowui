package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestLabelRequiredMarkAddsWidth(t *testing.T) {
	ctx := newContext(nil)
	activeTheme := ctx.Theme()
	if activeTheme.Components.Label.TextSize != 14 {
		t.Fatalf("label text size = %v, want 14", activeTheme.Components.Label.TextSize)
	}
	if activeTheme.Components.Label.RequiredMarkOffset != 2 {
		t.Fatalf("required mark offset = %v, want 2", activeTheme.Components.Label.RequiredMarkOffset)
	}
	gtx := testLayoutContext()
	base := Label("Email").Layout(ctx, gtx)
	required := Label("Email").Required(true).Layout(ctx, gtx)

	if required.Size.X <= base.Size.X {
		t.Fatalf("required width = %d, want greater than base width %d", required.Size.X, base.Size.X)
	}
	if required.Size.Y != base.Size.Y {
		t.Fatalf("required height = %d, want base height %d", required.Size.Y, base.Size.Y)
	}
}

func TestLabelInternalClickableDoesNotCollideWithUserKey(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Column(
		Label("Email").For("email"),
		Button("email:label", Text("Action")),
		Input("email", ""),
	).Gap(6)

	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
	if testComponentState[widget.Clickable](ctx, "email:label", "clickable") == nil {
		t.Fatal("user clickable state was not registered")
	}
	derivedKey := frame.DerivedKey(ctx, "email", "label")
	if testComponentState[widget.Clickable](ctx, derivedKey, "clickable") == nil {
		t.Fatal("label clickable state was not registered under its derived key")
	}
}

func TestLabelFocusesAssociatedControls(t *testing.T) {
	tests := []struct {
		name  string
		field Widget
	}{
		{
			name:  "input",
			field: Input("field", "").Hint("Type a value"),
		},
		{
			name: "combobox",
			field: ComboBox("field", "", []ComboBoxItem{
				{Key: "go", Label: "Go"},
			}),
		},
		{
			name: "select",
			field: Select("field", "", []SelectItem{
				{Key: "go", Label: "Go"},
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			field := Column(
				Label("Field").For("field"),
				test.field,
			).Gap(6)

			layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
			target := frame.FieldFocusTag(ctx, "field")
			clickLabel(router)
			layoutLabelFieldFrame(ctx, router, field, time.Unix(1, int64(time.Millisecond)))

			if !router.Source().Focused(target) {
				t.Fatal("associated control did not gain focus")
			}
		})
	}
}

func TestSelectInternalLabelFocusesTrigger(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Select("language", "", []SelectItem{{Key: "go", Label: "Go"}}).
		Label("Language")

	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
	clickLabel(router)
	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, int64(time.Millisecond)))

	if !router.Source().Focused(frame.FieldFocusTag(ctx, "language")) {
		t.Fatal("select trigger did not gain focus from its internal label")
	}
}

func TestLabelDoesNotFocusDisabledTarget(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Column(
		Label("Field").For("field"),
		Input("field", "").Disabled(true),
	).Gap(6)

	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
	target := testComponentState[widget.Editor](ctx, "field", stateSlotEditor)
	clickLabel(router)
	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, int64(time.Millisecond)))

	if router.Source().Focused(target) {
		t.Fatal("disabled associated control gained focus")
	}
}

func TestDisabledLabelDoesNotFocusEnabledTarget(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Column(
		Label("Field").For("field").Disabled(true),
		Input("field", ""),
	).Gap(6)

	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
	target := testComponentState[widget.Editor](ctx, "field", stateSlotEditor)
	clickLabel(router)
	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, int64(time.Millisecond)))

	if router.Source().Focused(target) {
		t.Fatal("enabled control gained focus from a disabled label")
	}
}

func TestLabelAssociationsAreSwept(t *testing.T) {
	ctx := newContext(nil)

	frame.BeginFrame(ctx)
	Label("Field").For("field").Layout(ctx, testLayoutContext())
	Input("field", "").Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)
	if frame.FieldLabel(ctx, "field") != "Field" || !frame.HasFieldFocus(ctx, "field") {
		t.Fatal("field association was not registered")
	}

	frame.BeginFrame(ctx)
	frame.EndFrame(ctx)
	if frame.HasFieldLabel(ctx, "field") {
		t.Fatal("stale field label was not swept")
	}
	if frame.HasFieldFocus(ctx, "field") {
		t.Fatal("stale field focus target was not swept")
	}
}

func TestFieldFocusIsSweptWhenKeyChangesToNonField(t *testing.T) {
	ctx := newContext(nil)

	frame.BeginFrame(ctx)
	Input("field", "").Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)
	if !frame.HasFieldFocus(ctx, "field") {
		t.Fatal("field focus target was not registered")
	}

	frame.BeginFrame(ctx)
	Button("field", Text("Action")).Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)
	if frame.HasFieldFocus(ctx, "field") {
		t.Fatal("field focus target survived a key kind change")
	}
}

func TestLabelProvidesEditorSemanticLabel(t *testing.T) {
	tests := []struct {
		name  string
		field Widget
	}{
		{name: "input", field: Input("field", "").Hint("name@example.com")},
		{name: "combobox", field: ComboBox("field", "", []ComboBoxItem{{Key: "go", Label: "Go"}})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			field := Column(
				Label("Field label").For("field"),
				test.field,
			).Gap(6)

			layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
			tree := router.AppendSemantics(nil)
			if !semanticTreeHasLabel(tree, "Field label") {
				t.Fatal("field semantic tree does not contain the associated label")
			}
			if _, ok := semanticLabelForClass(tree, semantic.Editor); !ok {
				t.Fatal("field semantic tree does not contain the editor")
			}
		})
	}
}

func TestLabelProvidesSelectSemanticLabel(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Column(
		Label("Language").For("language"),
		Select("language", "go", []SelectItem{{Key: "go", Label: "Go"}}),
	).Gap(6)

	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, 0))
	if label, ok := semanticLabelForClass(router.AppendSemantics(nil), semantic.Button); !ok || label != "Language, Go" {
		t.Fatalf("select semantic label = %q, %v; want Language, Go, true", label, ok)
	}
}

func clickLabel(router *input.Router) {
	router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  f32.Pt(8, 8),
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  f32.Pt(8, 8),
		},
	)
}

func layoutLabelFieldFrame(ctx *Context, router *input.Router, field Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 240)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	field.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func semanticLabelForClass(nodes []input.SemanticNode, class semantic.ClassOp) (string, bool) {
	for _, node := range nodes {
		if node.Desc.Class == class {
			return node.Desc.Label, true
		}
		if label, ok := semanticLabelForClass(node.Children, class); ok {
			return label, true
		}
	}
	return "", false
}

func semanticTreeHasLabel(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || semanticTreeHasLabel(node.Children, label) {
			return true
		}
	}
	return false
}
