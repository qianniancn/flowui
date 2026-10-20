package breadcrumb

import (
	"image"
	"testing"

	"gioui.org/f32"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/uitest"
)

func TestNavigation(t *testing.T) {
	h := uitest.New(image.Pt(460, 48), base.Light())
	selected, changes := "", 0
	root := New("path", []Item{
		{Key: "home", Label: "Home"},
		{Key: "locked", Label: "Locked", Disabled: true},
		{Key: "current", Label: "Current"},
	}).OnChange(func(key string) { selected = key; changes++ })
	h.Frame(root)
	h.Frame(root)
	for _, label := range []string{"Home", "Locked", "Current"} {
		node, ok := uitest.SemanticByLabel(h.Semantics(), label)
		if !ok {
			t.Fatalf("missing breadcrumb %q", label)
		}
		bounds := node.Desc.Bounds
		h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		h.Frame(root)
		h.Frame(root)
		if selected != "home" || changes != 1 {
			t.Fatalf("after %s: selected %q, changes %d; want home, 1", label, selected, changes)
		}
	}
}
