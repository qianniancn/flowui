package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
)

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

	frame.BeginFrame(ctx)
	Description("Supporting text").For("field").Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)
	if got := frame.FieldDescription(ctx, "field"); got != "Supporting text" {
		t.Fatalf("field description = %q, want Supporting text", got)
	}

	frame.BeginFrame(ctx)
	frame.EndFrame(ctx)
	if frame.HasFieldDescription(ctx, "field") {
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
			field := Column(
				Label("Field").For("field"),
				test.field,
				Description("Supporting text").For("field"),
			).Gap(6)
			layoutDescriptionFrame(ctx, router, field, time.Unix(1, 0))
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

func TestDescriptionRemovalUpdatesControlSemanticsInSameFrame(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	withDescription := Column(
		Label("Email").For("email"),
		Input("email", ""),
		Description("Used for notifications").For("email"),
	).Gap(6)
	withoutDescription := Column(
		Label("Email").For("email"),
		Input("email", ""),
	).Gap(6)

	layoutDescriptionFrame(ctx, router, withDescription, time.Unix(1, 0))
	if tree := router.AppendSemantics(nil); !semanticTreeHasDescription(tree, "Used for notifications") {
		t.Fatal("first-frame semantic tree does not contain the description")
	}

	layoutDescriptionFrame(ctx, router, withoutDescription, time.Unix(1, int64(time.Millisecond)))
	if tree := router.AppendSemantics(nil); semanticTreeHasDescription(tree, "Used for notifications") {
		t.Fatal("semantic tree retained the removed description")
	}
}

func TestDescriptionAssociationRespectsKeyScope(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Key("profile", Column(
		Label("Email").For("email"),
		Input("email", ""),
		Description("Scoped supporting text").For("email"),
	).Gap(6))

	layoutDescriptionFrame(ctx, router, field, time.Unix(1, 0))
	if tree := router.AppendSemantics(nil); !semanticTreeHasDescription(tree, "Scoped supporting text") {
		t.Fatal("semantic tree does not contain the scoped description")
	}
}

func TestDescriptionAssociationFallsBackForCustomContainers(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	field := Column(manualDescriptionField{description: "Custom supporting text"})

	layoutDescriptionFrame(ctx, router, field, time.Unix(1, 0))
	layoutDescriptionFrame(ctx, router, field, time.Unix(1, int64(time.Millisecond)))
	if tree := router.AppendSemantics(nil); !semanticTreeHasDescription(tree, "Custom supporting text") {
		t.Fatal("custom container did not receive the previous-frame description association")
	}
}

func TestDescriptionAssociationMigratesFromBuiltInToCustomContainer(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	builtIn := Column(
		Label("Email").For("custom-email"),
		Input("custom-email", ""),
		Description("Custom supporting text").For("custom-email"),
	)
	custom := Column(manualDescriptionField{description: "Custom supporting text"})

	layoutDescriptionFrame(ctx, router, builtIn, time.Unix(1, 0))
	layoutDescriptionFrame(ctx, router, custom, time.Unix(1, int64(time.Millisecond)))
	layoutDescriptionFrame(ctx, router, custom, time.Unix(1, int64(2*time.Millisecond)))
	if tree := router.AppendSemantics(nil); !semanticTreeHasDescription(tree, "Custom supporting text") {
		t.Fatal("association did not recover after moving to a custom container")
	}
}

func TestSelectAndSwitchDescriptionsReuseAssociation(t *testing.T) {
	ctx := newContext(nil)

	frame.BeginFrame(ctx)
	Select("language", "", nil).Description("Choose a language").Layout(ctx, testLayoutContext())
	Switch("updates", false, "Updates").Description("Receive product updates").Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)

	if got := frame.FieldDescription(ctx, "language"); got != "Choose a language" {
		t.Fatalf("select field description = %q, want Choose a language", got)
	}
	if got := frame.FieldDescription(ctx, "updates"); got != "Receive product updates" {
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

func TestInternalFieldAssociationsAreRemovedWithoutStaleSemantics(t *testing.T) {
	tests := []struct {
		name   string
		before Widget
		after  Widget
		class  semantic.ClassOp
		label  string
	}{
		{
			name:   "select description",
			before: Select("language", "go", []SelectItem{{Key: "go", Label: "Go"}}).Description("Choose a language"),
			after:  Select("language", "go", []SelectItem{{Key: "go", Label: "Go"}}),
			class:  semantic.Button,
			label:  "Go",
		},
		{
			name:   "select label",
			before: Select("language", "go", []SelectItem{{Key: "go", Label: "Go"}}).Label("Language"),
			after:  Select("language", "go", []SelectItem{{Key: "go", Label: "Go"}}),
			class:  semantic.Button,
			label:  "Go",
		},
		{
			name:   "switch description",
			before: Switch("updates", false, "Updates").Description("Receive product updates"),
			after:  Switch("updates", false, "Updates"),
			class:  semantic.Switch,
			label:  "Updates",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			layoutDescriptionFrame(ctx, router, test.before, time.Unix(1, 0))
			layoutDescriptionFrame(ctx, router, test.after, time.Unix(1, int64(time.Millisecond)))

			node, ok := semanticNodeForClass(router.AppendSemantics(nil), test.class)
			if !ok {
				t.Fatalf("semantic tree does not contain class %v", test.class)
			}
			if node.Desc.Label != test.label || node.Desc.Description != "" {
				t.Fatalf("semantics = label %q description %q, want label %q without description", node.Desc.Label, node.Desc.Description, test.label)
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
	frame.BeginFrame(ctx)
	child.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
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

type manualDescriptionField struct {
	description string
}

func (field manualDescriptionField) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Label("Email").For("custom-email").Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Input("custom-email", "").Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Description(field.description).For("custom-email").Layout(ctx, gtx)
		}),
	)
}

const stateSlotEditor = "editor"

func newContext(_ any) *Context {
	return frame.New(nil, nil, LanguageAuto)
}

func testComponentState[T any](ctx *Context, key, slot string) *T {
	state, _ := frame.PeekState[T](ctx, key, slot)
	return state
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}
