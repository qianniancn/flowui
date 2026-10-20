package uitest_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
)

func TestSemanticTreeQueriesFindReportedState(t *testing.T) {
	harness := uitest.New(image.Pt(240, 160), base.Light())
	clicks := 0
	root := layout.Column(
		layout.Box(text.New("Save")).Key("save").Label("Save file").OnClick(func() { clicks++ }),
		layout.Box(text.New("Delete")).Key("delete").Label("Delete file").OnClick(func() {}).Disabled(true),
	)
	harness.Frame(root)

	nodes := harness.Semantics()
	if len(nodes) == 0 {
		t.Fatal("frame did not produce a semantic tree")
	}

	node, ok := uitest.SemanticByLabel(nodes, "Save file")
	if !ok {
		t.Fatalf("semantic tree is missing the save label: %v", uitest.SemanticLabels(nodes))
	}
	if node.Desc.Class != semantic.Button {
		t.Fatalf("save class = %v, want %v", node.Desc.Class, semantic.Button)
	}
	if node.Desc.Disabled {
		t.Fatal("enabled control reported Disabled")
	}

	if class, ok := uitest.SemanticClassForLabel(nodes, "Save file"); !ok || class != semantic.Button {
		t.Fatalf("SemanticClassForLabel = %v, %v; want %v, true", class, ok, semantic.Button)
	}
	if !uitest.HasSemanticLabel(nodes, "Delete file") {
		t.Fatalf("disabled control label is missing: %v", uitest.SemanticLabels(nodes))
	}
	if disabled := uitest.DisabledSemanticLabels(nodes); len(disabled) != 1 || disabled[0] != "Delete file" {
		t.Fatalf("disabled labels = %v, want [Delete file]", disabled)
	}

	bounds, ok := uitest.SemanticBoundsForLabel(nodes, "Save file")
	if !ok || bounds.Empty() {
		t.Fatalf("save bounds = %v, ok = %v; want a non-empty rectangle", bounds, ok)
	}

	// The reported bounds must be a usable hit target, so drive the control
	// through the coordinates the accessibility tree advertises.
	center := f32.Pt(
		float32(bounds.Min.X+bounds.Max.X)/2,
		float32(bounds.Min.Y+bounds.Max.Y)/2,
	)
	harness.Click(center)
	harness.Frame(root)
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestWalkSemanticsVisitsDepthFirstAndStops(t *testing.T) {
	tree := []uitest.Semantic{
		{Desc: input.SemanticDesc{Label: "a"}, Children: []uitest.Semantic{
			{Desc: input.SemanticDesc{Label: "b"}},
		}},
		{Desc: input.SemanticDesc{Label: "c"}},
	}

	var all []string
	uitest.WalkSemantics(tree, func(node uitest.Semantic) bool {
		all = append(all, node.Desc.Label)
		return true
	})
	if len(all) != 3 || all[0] != "a" || all[1] != "b" || all[2] != "c" {
		t.Fatalf("walk order = %v, want [a b c]", all)
	}

	var stopped []string
	uitest.WalkSemantics(tree, func(node uitest.Semantic) bool {
		stopped = append(stopped, node.Desc.Label)
		return false
	})
	if len(stopped) != 1 || stopped[0] != "a" {
		t.Fatalf("stopped walk = %v, want [a]", stopped)
	}

	if _, ok := uitest.FindSemantic(tree, nil); ok {
		t.Fatal("nil matcher unexpectedly found a node")
	}
}

func TestTouchInputDrivesControls(t *testing.T) {
	harness := uitest.New(image.Pt(240, 160), base.Light())
	taps := 0
	root := layout.Box(text.New("Tap")).Key("tap").Label("Tap target").OnClick(func() { taps++ })
	harness.Frame(root)

	// A queued tap is delivered by the following frame.
	harness.TouchTap(f32.Pt(20, 16))
	harness.Frame(root)
	if taps != 1 {
		t.Fatalf("taps after TouchTap = %d, want 1", taps)
	}

	// A press and release split across frames must still complete one click.
	harness.TouchDown(f32.Pt(20, 16))
	harness.Frame(root)
	harness.TouchMove(f32.Pt(22, 16))
	harness.Frame(root)
	harness.TouchUp(f32.Pt(22, 16))
	harness.Frame(root)
	if taps != 2 {
		t.Fatalf("taps after TouchDown/TouchUp = %d, want 2", taps)
	}
}

func TestTouchAndMouseKeepDistinctPointerIdentity(t *testing.T) {
	harness := uitest.New(image.Pt(240, 160), base.Light())
	clicks := 0
	root := layout.Box(text.New("Mixed")).Key("mixed").Label("Mixed input").OnClick(func() { clicks++ })
	harness.Frame(root)

	harness.Click(f32.Pt(20, 16))
	harness.Frame(root)
	harness.TouchTap(f32.Pt(20, 16))
	harness.Frame(root)

	if clicks != 2 {
		t.Fatalf("clicks = %d, want 2; mouse and touch must not share a pointer identity", clicks)
	}
}

func TestRelationsExposeTabPanelLinks(t *testing.T) {
	harness := uitest.New(image.Pt(320, 200), base.Light())
	root := tabs.New("tabs", "a", []tabs.TabItem{
		{Key: "a", Label: "A", Panel: text.New("Panel A")},
		{Key: "b", Label: "B", Panel: text.New("Panel B")},
	})
	harness.Frame(root)

	relations := harness.Relations()
	if len(relations) == 0 {
		t.Fatal("tabs did not register semantic relations")
	}

	var list, tab, panel ui.SemanticNode
	for _, node := range relations {
		switch node.Role {
		case ui.SemanticTabList:
			list = node
		case ui.SemanticTab:
			if tab.Key == "" {
				tab = node
			}
		case ui.SemanticTabPanel:
			if panel.Key == "" {
				panel = node
			}
		}
	}
	if list.Key == "" || tab.Key == "" || panel.Key == "" {
		t.Fatalf("relations are missing a tab list, tab, or panel: %#v", relations)
	}
	if tab.Owner != list.Key {
		t.Fatalf("tab owner = %q, want tab list %q", tab.Owner, list.Key)
	}
	if tab.Controls != panel.Key {
		t.Fatalf("tab controls = %q, want panel %q", tab.Controls, panel.Key)
	}
	if panel.Owner != list.Key {
		t.Fatalf("panel owner = %q, want tab list %q", panel.Owner, list.Key)
	}
	if tab.SetSize != 2 || tab.PosInSet != 1 {
		t.Fatalf("tab set position = %d/%d, want 1/2", tab.PosInSet, tab.SetSize)
	}
}
