package text

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

func TestSelectableTextMenuHookQueuesCopy(t *testing.T) {
	ctx := textTestContext()
	router := new(input.Router)
	selectable := Selectable("hooked", "Copy this text")
	start := time.Unix(3, 0)
	layoutTextFrame(ctx, router, selectable, start, image.Pt(240, 80))

	stateValue, ok := core.PeekState[selectableTextState](ctx, "hooked", stateSlotSelectableText)
	if !ok {
		t.Fatal("selectable state is missing")
	}
	stateValue.selectable.SetCaret(0, 4)
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.selectable})

	orig := core.SelectableTextMenu
	defer func() { core.SelectableTextMenu = orig }()
	var hookKey string
	core.SelectableTextMenu = func(ctx *core.Context, gtx layout.Context, key string, field core.Widget, selectable *widget.Selectable, queue func(core.SelectableEditAction)) layout.Dimensions {
		hookKey = key
		queue(core.SelectableEditCopy)
		return field.Layout(ctx, gtx)
	}
	layoutTextFrame(ctx, router, selectable, start.Add(time.Millisecond), image.Pt(240, 80))
	if hookKey != "hooked" {
		t.Fatalf("menu hook key = %q", hookKey)
	}
	layoutTextFrame(ctx, router, selectable, start.Add(2*time.Millisecond), image.Pt(240, 80))
	mime, content, ok := router.WriteClipboard()
	if !ok || mime != "application/text" || string(content) != "Copy" {
		t.Fatalf("clipboard = ok %v mime %q content %q", ok, mime, content)
	}
}

func TestSelectableTextMenuHookSelectAll(t *testing.T) {
	ctx := textTestContext()
	router := new(input.Router)
	selectable := Selectable("selall", "Select this text")
	start := time.Unix(4, 0)
	layoutTextFrame(ctx, router, selectable, start, image.Pt(240, 80))

	stateValue, ok := core.PeekState[selectableTextState](ctx, "selall", stateSlotSelectableText)
	if !ok {
		t.Fatal("selectable state is missing")
	}
	orig := core.SelectableTextMenu
	defer func() { core.SelectableTextMenu = orig }()
	core.SelectableTextMenu = func(ctx *core.Context, gtx layout.Context, key string, field core.Widget, selectable *widget.Selectable, queue func(core.SelectableEditAction)) layout.Dimensions {
		queue(core.SelectableEditSelectAll)
		return field.Layout(ctx, gtx)
	}
	layoutTextFrame(ctx, router, selectable, start.Add(time.Millisecond), image.Pt(240, 80))
	layoutTextFrame(ctx, router, selectable, start.Add(2*time.Millisecond), image.Pt(240, 80))
	if stateValue.selectable.SelectionLen() != len([]rune("Select this text")) {
		t.Fatalf("select-all selection length = %d", stateValue.selectable.SelectionLen())
	}
}
