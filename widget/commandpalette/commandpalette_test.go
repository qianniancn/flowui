package commandpalette

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/modal"
	"github.com/qianniancn/flowui/widget/text"
)

func TestSearchKeyboardAndExecution(t *testing.T) {
	h := uitest.New(image.Pt(720, 600), base.Light())
	open, executed := true, ""
	commands := []command.Command{
		command.New("disabled", "Disabled").Disabled(true).OnExecute(func() { t.Fatal("disabled command executed") }),
	}
	for index := range 20 {
		id := fmt.Sprintf("item-%02d", index)
		commands = append(commands, command.New(id, "Command "+id).Description("项目命令").
			OnExecute(func() { executed = id }))
	}
	commands = append(commands, command.New("stay", "Keep open").KeepOpen(true).
		OnExecute(func() { executed = "stay" }))
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		return New("palette", open, commands).OnOpenChange(func(next bool) { open = next }).Layout(ctx, gtx)
	})
	frame := func() { h.Frame(root); h.Frame(root) }
	frame()
	h.Advance(time.Second)
	frame()
	for range 15 {
		h.Key(key.NameDownArrow, 0)
		frame()
	}
	selected, ok := uitest.SemanticByLabel(h.Semantics(), "Command item-15")
	if !ok || selected.Desc.Bounds.Min.Y < 0 || selected.Desc.Bounds.Max.Y > 600 {
		t.Fatalf("keyboard navigation must reveal offscreen results: found=%v, bounds=%v", ok, selected.Desc.Bounds)
	}
	h.Key(key.NameReturn, 0)
	frame()
	if open || executed != "item-15" {
		t.Fatalf("execution = %q, open = %v; want item-15, false", executed, open)
	}
	h.Advance(time.Second)
	frame()
	open = true
	frame()
	h.Advance(time.Second)
	frame()
	h.Router().Queue(key.EditEvent{Text: "项目 ITEM-19"})
	frame()
	h.Key(key.NameReturn, 0)
	frame()
	if open || executed != "item-19" {
		t.Fatalf("filtered execution = %q, open = %v; want item-19, false", executed, open)
	}
	h.Advance(time.Second)
	frame()
	open = true
	frame()
	h.Advance(time.Second)
	frame()
	h.Router().Queue(key.EditEvent{Text: "Keep open"})
	frame()
	h.Key(key.NameReturn, 0)
	frame()
	if !open || executed != "stay" {
		t.Fatal("KeepOpen command must execute without closing")
	}
	h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: 9}, Text: "no-match"})
	frame()
	h.Key(key.NameReturn, 0)
	frame()
	if !open || executed != "stay" {
		t.Fatal("empty search must not execute a previous selection")
	}
	h.Key(key.NameEscape, 0)
	frame()
	if open {
		t.Fatal("Escape must close the palette")
	}
}

func TestSearchWaitsForCoveringModal(t *testing.T) {
	h := uitest.New(image.Pt(720, 600), base.Light())
	covered, open, executed := true, true, false
	commands := []command.Command{command.New("save", "Save").OnExecute(func() { executed = true })}
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		return layout.Column(
			New("palette", open, commands).OnOpenChange(func(next bool) { open = next }),
			modal.New("Other dialog", text.New("Close this first")).Key("cover").Open(covered).
				OnOpenChange(func(next bool) { covered = next }),
		).Layout(ctx, gtx)
	})
	frame := func() { h.Frame(root); h.Frame(root) }
	frame()
	h.Advance(time.Second)
	frame()
	h.Key(key.NameReturn, 0)
	frame()
	if executed || !open {
		t.Fatal("covered palette stole keyboard focus")
	}
	h.Key(key.NameEscape, 0)
	frame()
	if covered || !open {
		t.Fatal("Escape must dismiss only the top modal")
	}
	h.Advance(time.Second)
	frame()
	h.Router().Queue(key.EditEvent{Text: "save"})
	h.Key(key.NameReturn, 0)
	frame()
	if !executed || open {
		t.Fatal("uncovered palette did not focus its search field")
	}
}
