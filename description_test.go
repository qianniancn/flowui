package flowui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestDescriptionOptions(t *testing.T) {
	description := Description("Supporting text").For("field").Disabled(true)

	if description.text != "Supporting text" || description.forKey != "field" {
		t.Fatalf("description identity = (%q, %q), want (Supporting text, field)", description.text, description.forKey)
	}
	if !description.disabled {
		t.Fatal("description disabled option was not retained")
	}
}

func TestDescriptionStyleMatchesHeroUI(t *testing.T) {
	theme := DefaultTheme()
	if theme.Components.Description.TextSize != 12 {
		t.Fatalf("description text size = %v, want 12", theme.Components.Description.TextSize)
	}

	base := descriptionStyleFor(&theme, false)
	if base.text != theme.Palette.MutedForeground {
		t.Fatalf("description text = %v, want muted foreground %v", base.text, theme.Palette.MutedForeground)
	}
	disabled := descriptionStyleFor(&theme, true)
	if disabled.text != theme.DisabledColor(theme.Palette.MutedForeground) {
		t.Fatalf("disabled description = %v, want disabled muted foreground", disabled.text)
	}
}

func TestDescriptionWrapsLongWords(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(80, 200)},
		Ops:         &ops,
	}
	dims := Description("averylongdescriptionwithoutspaces").Layout(ctx, gtx)
	singleLine := Description("short").Layout(ctx, gtx)

	if dims.Size.X > gtx.Constraints.Max.X {
		t.Fatalf("description width = %d, want at most %d", dims.Size.X, gtx.Constraints.Max.X)
	}
	if dims.Size.Y <= singleLine.Size.Y {
		t.Fatalf("wrapped height = %d, want greater than single-line height %d", dims.Size.Y, singleLine.Size.Y)
	}
}

func TestDescriptionAssociationIsSwept(t *testing.T) {
	ctx := newContext(nil)

	ctx.beginFrame()
	Description("Supporting text").For("field").Layout(ctx, testLayoutContext())
	ctx.endFrame()
	if got := ctx.fieldDescription("field"); got != "Supporting text" {
		t.Fatalf("field description = %q, want Supporting text", got)
	}

	ctx.beginFrame()
	ctx.endFrame()
	if _, ok := ctx.fieldDescriptions["field"]; ok {
		t.Fatal("stale field description was not swept")
	}
}

func TestDescriptionAssociatesWithControls(t *testing.T) {
	tests := []struct {
		name              string
		field             Widget
		class             semantic.ClassOp
		directDescription bool
	}{
		{name: "input", field: Input("field", ""), class: semantic.Editor},
		{name: "combobox", field: ComboBox("field", "", []ComboBoxItem{{Key: "go", Label: "Go"}}), class: semantic.Editor},
		{name: "select", field: Select("field", "go", []SelectItem{{Key: "go", Label: "Go"}}), class: semantic.Button, directDescription: true},
		{name: "switch", field: Switch("field", false, "Updates"), class: semantic.Switch, directDescription: true},
		{name: "checkbox", field: Checkbox("field", false, "Updates"), class: semantic.CheckBox, directDescription: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			layoutDescriptionFrame(ctx, router, Description("Supporting text").For("field"), time.Unix(1, 0))
			layoutDescriptionFrame(ctx, router, test.field, time.Unix(1, int64(time.Millisecond)))
			tree := router.AppendSemantics(nil)

			if !semanticTreeHasDescription(tree, "Supporting text") {
				t.Fatal("control semantic tree does not contain the associated description")
			}
			description, ok := semanticDescriptionForClass(tree, test.class)
			if !ok {
				t.Fatalf("control semantic tree does not contain class %v", test.class)
			}
			if test.directDescription && description != "Supporting text" {
				t.Fatalf("control description = %q, want Supporting text", description)
			}
		})
	}
}

func TestSelectAndSwitchDescriptionsReuseAssociation(t *testing.T) {
	ctx := newContext(nil)

	ctx.beginFrame()
	Select("language", "", nil).Description("Choose a language").Layout(ctx, testLayoutContext())
	Switch("updates", false, "Updates").Description("Receive product updates").Layout(ctx, testLayoutContext())
	ctx.endFrame()

	if got := ctx.fieldDescription("language"); got != "Choose a language" {
		t.Fatalf("select field description = %q, want Choose a language", got)
	}
	if got := ctx.fieldDescription("updates"); got != "Receive product updates" {
		t.Fatalf("switch field description = %q, want Receive product updates", got)
	}
}

func TestInternalDescriptionsPreserveControlSemantics(t *testing.T) {
	tests := []struct {
		name        string
		field       Widget
		class       semantic.ClassOp
		label       string
		description string
	}{
		{
			name:        "select",
			field:       Select("language", "go", []SelectItem{{Key: "go", Label: "Go"}}).Label("Language").Description("Choose a language"),
			class:       semantic.Button,
			label:       "Language, Go",
			description: "Choose a language",
		},
		{
			name:        "switch",
			field:       Switch("updates", false, "Updates").Description("Receive product updates"),
			class:       semantic.Switch,
			label:       "Updates",
			description: "Receive product updates",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			layoutDescriptionFrame(ctx, router, test.field, time.Unix(1, 0))
			node, ok := semanticNodeForClass(router.AppendSemantics(nil), test.class)
			if !ok {
				t.Fatalf("semantic tree does not contain class %v", test.class)
			}
			if node.Desc.Label != test.label || node.Desc.Description != test.description {
				t.Fatalf("semantics = label %q description %q, want %q and %q", node.Desc.Label, node.Desc.Description, test.label, test.description)
			}
		})
	}
}

func TestInvalidSelectHidesDescriptionSemantics(t *testing.T) {
	tests := []struct {
		name        string
		field       SelectWidget
		description string
	}{
		{
			name:  "without error",
			field: Select("language", "", nil).Description("Choose a language").Invalid(true),
		},
		{
			name:        "with error",
			field:       Select("language", "", nil).Description("Choose a language").Invalid(true).ErrorMessage("Language is required"),
			description: "Language is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			layoutDescriptionFrame(ctx, router, test.field, time.Unix(1, 0))
			node, ok := semanticNodeForClass(router.AppendSemantics(nil), semantic.Button)
			if !ok {
				t.Fatal("semantic tree does not contain select button")
			}
			if node.Desc.Description != test.description {
				t.Fatalf("select description = %q, want %q", node.Desc.Description, test.description)
			}
		})
	}
}

func layoutDescriptionFrame(ctx *Context, router *input.Router, child Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 240)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	ctx.beginFrame()
	child.Layout(ctx, gtx)
	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
	router.Frame(&ops)
}

func semanticDescriptionForClass(nodes []input.SemanticNode, class semantic.ClassOp) (string, bool) {
	node, ok := semanticNodeForClass(nodes, class)
	return node.Desc.Description, ok
}

func semanticNodeForClass(nodes []input.SemanticNode, class semantic.ClassOp) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == class {
			return node, true
		}
		if found, ok := semanticNodeForClass(node.Children, class); ok {
			return found, true
		}
	}
	return input.SemanticNode{}, false
}

func semanticTreeHasDescription(nodes []input.SemanticNode, description string) bool {
	for _, node := range nodes {
		if node.Desc.Description == description || semanticTreeHasDescription(node.Children, description) {
			return true
		}
	}
	return false
}
