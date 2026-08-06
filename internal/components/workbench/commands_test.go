package workbench

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestCommandScopeRoutesWorkbenchShortcuts(t *testing.T) {
	controller := NewController(NewState([]Group{{
		Key:  "editor",
		Tabs: []Tab{{Key: "one", Closable: true}, {Key: "two", Closable: true}},
	}}))
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	layoutCommandsFrame(ctx, router, controller, time.Unix(1, 0))
	router.Queue(key.Event{Name: key.NameTab, Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandsFrame(ctx, router, controller, time.Unix(1, int64(time.Millisecond)))
	if got := controller.State().ActiveTab("editor"); got != "two" {
		t.Fatalf("next tab = %q, want two", got)
	}
	router.Queue(key.Event{Name: key.Name("W"), Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandsFrame(ctx, router, controller, time.Unix(1, int64(2*time.Millisecond)))
	if group, _ := controller.State().Group("editor"); len(group.Tabs) != 1 || group.Tabs[0].Key != "one" {
		t.Fatalf("closed tab state = %#v", group)
	}
	router.Queue(key.Event{Name: key.Name("B"), Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandsFrame(ctx, router, controller, time.Unix(1, int64(3*time.Millisecond)))
	if controller.State().Chrome.SidebarVisible {
		t.Fatal("sidebar shortcut did not toggle visibility")
	}
}

func layoutCommandsFrame(ctx *frame.Context, router *input.Router, controller *Controller, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(320, 180)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, image.Pt(320, 180))
	controller.CommandScope(nil).Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
