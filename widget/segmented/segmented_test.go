package segmented

import (
	"image"
	"testing"

	"gioui.org/f32"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestSelection(t *testing.T) {
	h := uitest.New(image.Pt(320, 48), base.Light())
	selected, changes := "one", 0
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		return New([]Item{
			{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}, {Key: "three", Label: "Three", Disabled: true},
		}).Key("view").SelectedKey(selected).FullWidth().OnChange(func(next string) { selected = next; changes++ }).Layout(ctx, gtx)
	})
	h.Frame(root)
	h.Frame(root)
	for _, label := range []string{"Two", "Two", "Three"} {
		node, ok := uitest.SemanticByLabel(h.Semantics(), label)
		if !ok {
			t.Fatalf("missing choice %q", label)
		}
		bounds := node.Desc.Bounds
		h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		h.Frame(root)
		h.Frame(root)
		if selected != "two" || changes != 1 {
			t.Fatalf("after %s: selected %q, changes %d; want two, 1", label, selected, changes)
		}
	}
}

func TestNewCopiesItems(t *testing.T) {
	items := []Item{{Key: "one", Label: "One"}}
	widget := New(items).Key("view").SelectedKey("one")
	items[0].Label = "changed"
	if widget.items[0].Label != "One" {
		t.Fatal("New should copy item definitions")
	}
}

func TestCompactAndFullWidth(t *testing.T) {
	items := []Item{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}}
	for _, fullWidth := range []bool{false, true} {
		h := uitest.New(image.Pt(400, 80), base.Light())
		widget := New(items).Key("view").SelectedKey("one")
		if fullWidth {
			widget = widget.FullWidth()
		}
		size := h.Frame(widget).Size
		if size.X <= 0 || size.X > 400 || size.Y <= 0 || size.Y >= 80 {
			t.Fatalf("fullWidth=%v: unexpected size %v", fullWidth, size)
		}
		if (size.X == 400) != fullWidth {
			t.Fatalf("fullWidth=%v: width = %d", fullWidth, size.X)
		}
	}
}
