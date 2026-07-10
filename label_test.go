package flowui

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestLabelOptions(t *testing.T) {
	label := Label("Email").
		For("email").
		Required(true).
		Disabled(true).
		Invalid(true)

	if label.text != "Email" || label.forKey != "email" {
		t.Fatalf("label identity = (%q, %q), want (Email, email)", label.text, label.forKey)
	}
	if !label.required || !label.disabled || !label.invalid {
		t.Fatal("label state options were not retained")
	}
}

func TestLabelStyleMatchesHeroUIStates(t *testing.T) {
	theme := DefaultTheme()

	base := labelStyleFor(&theme, theme.Palette.Foreground, false, false)
	if base.text != theme.Palette.Foreground {
		t.Fatalf("base text = %v, want foreground %v", base.text, theme.Palette.Foreground)
	}
	if base.required != theme.Palette.Danger {
		t.Fatalf("required mark = %v, want danger %v", base.required, theme.Palette.Danger)
	}

	invalid := labelStyleFor(&theme, theme.Palette.Foreground, false, true)
	if invalid.text != theme.Palette.Danger {
		t.Fatalf("invalid text = %v, want danger %v", invalid.text, theme.Palette.Danger)
	}

	disabled := labelStyleFor(&theme, theme.Palette.Foreground, true, true)
	if disabled.text != theme.DisabledColor(theme.Palette.Danger) {
		t.Fatalf("disabled invalid text = %v, want disabled danger", disabled.text)
	}
	if disabled.required != theme.DisabledColor(theme.Palette.Danger) {
		t.Fatalf("disabled required mark = %v, want disabled danger", disabled.required)
	}
}

func TestLabelStyleUsesContextualForeground(t *testing.T) {
	theme := DefaultTheme()
	foreground := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}

	style := labelStyleFor(&theme, foreground, false, false)
	if style.text != foreground {
		t.Fatalf("label text = %v, want contextual foreground %v", style.text, foreground)
	}
	invalid := labelStyleFor(&theme, foreground, false, true)
	if invalid.text != theme.Palette.Danger {
		t.Fatalf("invalid label text = %v, want danger %v", invalid.text, theme.Palette.Danger)
	}
}

func TestLabelRequiredMarkAddsWidth(t *testing.T) {
	ctx := newContext(nil)
	if ctx.Theme.Components.Label.TextSize != 14 {
		t.Fatalf("label text size = %v, want 14", ctx.Theme.Components.Label.TextSize)
	}
	if ctx.Theme.Components.Label.RequiredMarkOffset != 2 {
		t.Fatalf("required mark offset = %v, want 2", ctx.Theme.Components.Label.RequiredMarkOffset)
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

func TestLabelFocusesAssociatedControls(t *testing.T) {
	tests := []struct {
		name   string
		field  Widget
		target func(*Context) event.Tag
	}{
		{
			name:  "input",
			field: Input("field", "").Hint("Type a value"),
			target: func(ctx *Context) event.Tag {
				return ctx.editors["field"]
			},
		},
		{
			name: "combobox",
			field: ComboBox("field", "", []ComboBoxItem{
				{Key: "go", Label: "Go"},
			}),
			target: func(ctx *Context) event.Tag {
				return &ctx.combos["field"].editor
			},
		},
		{
			name: "select",
			field: Select("field", "", []SelectItem{
				{Key: "go", Label: "Go"},
			}),
			target: func(ctx *Context) event.Tag {
				return &ctx.selects["field"].trigger
			},
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
			target := test.target(ctx)
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

	if !router.Source().Focused(&ctx.selects["language"].trigger) {
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
	target := ctx.editors["field"]
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
	target := ctx.editors["field"]
	clickLabel(router)
	layoutLabelFieldFrame(ctx, router, field, time.Unix(1, int64(time.Millisecond)))

	if router.Source().Focused(target) {
		t.Fatal("enabled control gained focus from a disabled label")
	}
}

func TestLabelAssociationsAreSwept(t *testing.T) {
	ctx := newContext(nil)

	ctx.beginFrame()
	Label("Field").For("field").Layout(ctx, testLayoutContext())
	Input("field", "").Layout(ctx, testLayoutContext())
	ctx.endFrame()
	if ctx.fieldLabel("field") != "Field" || ctx.fieldFocus["field"].tag == nil {
		t.Fatal("field association was not registered")
	}

	ctx.beginFrame()
	ctx.endFrame()
	if _, ok := ctx.fieldLabels["field"]; ok {
		t.Fatal("stale field label was not swept")
	}
	if _, ok := ctx.fieldFocus["field"]; ok {
		t.Fatal("stale field focus target was not swept")
	}
}

func TestFieldFocusIsSweptWhenKeyChangesToNonField(t *testing.T) {
	ctx := newContext(nil)

	ctx.beginFrame()
	Input("field", "").Layout(ctx, testLayoutContext())
	ctx.endFrame()
	if ctx.fieldFocus["field"].tag == nil {
		t.Fatal("field focus target was not registered")
	}

	ctx.beginFrame()
	Button("field", Text("Action")).Layout(ctx, testLayoutContext())
	ctx.endFrame()
	if _, ok := ctx.fieldFocus["field"]; ok {
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
	ctx.beginFrame()
	field.Layout(ctx, gtx)
	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
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
