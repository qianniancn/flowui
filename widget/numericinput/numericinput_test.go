package numericinput

import (
	"image"
	"testing"
	"unicode/utf8"

	"gioui.org/f32"
	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/button"
)

func TestEditingAndCommit(t *testing.T) {
	h := uitest.New(image.Pt(360, 140), base.Light())
	value, changes := 2.0, 0
	disabled := false
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		return layout.Column(
			New("amount", value).Range(-10, 10).Step(0.25).Precision(2).Label("Amount").Disabled(disabled).
				OnChange(func(next float64) { value = next; changes++ }),
			button.New("Other"),
		).Gap(12).Layout(ctx, gtx)
	})
	frame := func() { h.Frame(root); h.Frame(root) }
	click := func(label string) {
		t.Helper()
		node, ok := uitest.SemanticByLabel(h.Semantics(), label)
		if !ok {
			t.Fatalf("missing control %q", label)
		}
		bounds := node.Desc.Bounds
		h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		frame()
	}
	edit := func(previous, next string) {
		h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: utf8.RuneCountInString(previous)}, Text: next})
		frame()
	}
	frame()
	click("Amount")
	edit("2.00", "-")
	if value != 2 || changes != 0 {
		t.Fatal("incomplete edit must not commit")
	}
	edit("-", "-3.126")
	if value != 2 {
		t.Fatal("valid draft must remain local until committed")
	}
	h.Key(key.NameReturn, 0)
	frame()
	if value != -3.13 || changes != 1 {
		t.Fatalf("Enter commit = %v (%d changes), want -3.13 (1 change)", value, changes)
	}
	edit("-3.13", "99")
	click("Other")
	if value != 10 || changes != 2 {
		t.Fatalf("blur commit = %v (%d changes), want 10 (2 changes)", value, changes)
	}
	click("Amount +")
	if value != 10 || changes != 2 {
		t.Fatal("increment at maximum must be disabled")
	}
	click("Amount −")
	if value != 9.75 || changes != 3 {
		t.Fatalf("step commit = %v (%d changes), want 9.75 (3 changes)", value, changes)
	}
	click("Amount")
	edit("9.75", "1e")
	h.Key(key.NameReturn, 0)
	frame()
	if value != 9.75 || changes != 3 {
		t.Fatal("invalid draft must restore controlled value without a callback")
	}
	h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: 4}, Text: "4.127"})
	h.Key(key.NameReturn, 0)
	frame()
	if value != 4.13 || changes != 4 {
		t.Fatalf("same-frame edit and submit = %v (%d changes), want 4.13 (4 changes)", value, changes)
	}
	disabled = true
	frame()
	click("Amount −")
	if value != 4.13 || changes != 4 {
		t.Fatal("disabled step must not commit")
	}
}

func TestRangeAndValueClamp(t *testing.T) {
	value := New("amount", 50).Range(10, 20).Value(99)
	if value.value != 20 {
		t.Fatalf("value = %v, want 20", value.value)
	}
}

func TestRangeRejectsInvertedBounds(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Range should reject inverted bounds")
		}
	}()
	New("amount", 1).Range(2, 1)
}

func TestStepRejectsNonPositive(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Step should reject non-positive values")
		}
	}()
	New("amount", 1).Step(0)
}
