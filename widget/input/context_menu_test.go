package input

import (
	"image"
	"io"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/widget/menu"
)

const contextMenuEnterDuration = 150 * time.Millisecond

const (
	contextMenuTestViewportX = 480
	contextMenuTestViewportY = 360
)

func layoutContextMenuInputFrame(ctx *core.Context, router *input.Router, i Widget, now time.Time, overlays bool) {
	viewport := image.Pt(contextMenuTestViewportX, contextMenuTestViewportY)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	i.Layout(ctx, gtx)
	if overlays {
		core.LayoutOverlays(ctx, gtx)
	}
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutContextMenuTextAreaFrame(ctx *core.Context, router *input.Router, t TextAreaWidget, now time.Time, overlays bool) {
	viewport := image.Pt(contextMenuTestViewportX, contextMenuTestViewportY)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	t.Layout(ctx, gtx)
	if overlays {
		core.LayoutOverlays(ctx, gtx)
	}
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func TestDefaultEditItemSpecsOrder(t *testing.T) {
	specs := defaultEditItemSpecs(EditSnapshot{Selection: 2, Length: 5}, editMenuLabels{"Cut", "Copy", "Paste", "Select All"})
	wantKeys := []string{"cut", "copy", "paste", "select-all"}
	if len(specs) != len(wantKeys) {
		t.Fatalf("specs = %d entries, want %d", len(specs), len(wantKeys))
	}
	for index, want := range wantKeys {
		if specs[index].key != want {
			t.Fatalf("spec[%d] = %q, want %q", index, specs[index].key, want)
		}
		if specs[index].label == "" || specs[index].shortcut == "" {
			t.Fatalf("spec %q missing label or shortcut", specs[index].key)
		}
	}
}

func TestDefaultEditItemSpecsMatrix(t *testing.T) {
	for _, test := range []struct {
		name     string
		snapshot EditSnapshot
		want     map[string]bool
	}{
		{
			name:     "editable with selection",
			snapshot: EditSnapshot{Selection: 3, Length: 10},
			want:     map[string]bool{"cut": false, "copy": false, "paste": false, "select-all": false},
		},
		{
			name:     "editable without selection",
			snapshot: EditSnapshot{Length: 10},
			want:     map[string]bool{"cut": true, "copy": true, "paste": false, "select-all": false},
		},
		{
			name:     "empty field",
			snapshot: EditSnapshot{},
			want:     map[string]bool{"cut": true, "copy": true, "paste": false, "select-all": true},
		},
		{
			name:     "read-only field",
			snapshot: EditSnapshot{Selection: 2, Length: 5, ReadOnly: true},
			want:     map[string]bool{"cut": true, "copy": false, "paste": true, "select-all": false},
		},
		{
			name:     "password field",
			snapshot: EditSnapshot{Selection: 2, Length: 5, Secret: true},
			want:     map[string]bool{"cut": true, "copy": true, "paste": false, "select-all": false},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, spec := range defaultEditItemSpecs(test.snapshot, editMenuLabels{"Cut", "Copy", "Paste", "Select All"}) {
				if spec.disabled != test.want[spec.key] {
					t.Fatalf("%s disabled = %v, want %v", spec.key, spec.disabled, test.want[spec.key])
				}
			}
		})
	}
}

func TestInputContextMenuQueuesCopy(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	i := New("Ada").Key("name")
	now := time.Unix(1, 0)

	layoutContextMenuInputFrame(ctx, router, i, now, false)
	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	editor.SetCaret(0, 2)
	state, _ := core.PeekState[inputState](ctx, "name", stateSlotInput)
	if state == nil {
		t.Fatal("missing input state")
	}
	state.queueEdit(EditCopy)

	layoutContextMenuInputFrame(ctx, router, i, now.Add(time.Millisecond), false)
	mime, content, ok := router.WriteClipboard()
	if !ok || mime != "application/text" || string(content) != "Ad" {
		t.Fatalf("clipboard = ok %v mime %q content %q", ok, mime, content)
	}
	if got := editor.Text(); got != "Ada" {
		t.Fatalf("copy changed editor text to %q", got)
	}
}

func TestInputContextMenuQueuesCut(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	changes := 0
	i := New("").Key("name").DefaultValue("Ada").OnChange(func(string) { changes++ })
	now := time.Unix(2, 0)

	layoutContextMenuInputFrame(ctx, router, i, now, false)
	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	editor.SetCaret(0, 2)
	state, _ := core.PeekState[inputState](ctx, "name", stateSlotInput)
	state.queueEdit(EditCut)

	layoutContextMenuInputFrame(ctx, router, i, now.Add(time.Millisecond), false)
	if got := editor.Text(); got != "a" {
		t.Fatalf("editor text after cut = %q, want a", got)
	}

	// The change surfaces through the editor on the following frame.
	layoutContextMenuInputFrame(ctx, router, i, now.Add(2*time.Millisecond), false)
	if changes != 1 {
		t.Fatalf("change callback count = %d, want 1", changes)
	}
	state, _ = core.PeekState[inputState](ctx, "name", stateSlotInput)
	if state.value != "a" {
		t.Fatalf("state value after cut = %q, want a", state.value)
	}
}

func TestInputContextMenuQueuesSelectAll(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	i := New("Ada").Key("name")
	now := time.Unix(3, 0)

	layoutContextMenuInputFrame(ctx, router, i, now, false)
	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	editor.SetCaret(1, 1)
	state, _ := core.PeekState[inputState](ctx, "name", stateSlotInput)
	state.queueEdit(EditSelectAll)

	layoutContextMenuInputFrame(ctx, router, i, now.Add(time.Millisecond), false)
	start, end := editor.Selection()
	if start != 0 || end != 3 {
		t.Fatalf("selection after select-all = (%d,%d), want (0,3)", start, end)
	}
}

func TestInputContextMenuQueuesPaste(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	i := New("Ada").Key("name")
	now := time.Unix(4, 0)

	layoutContextMenuInputFrame(ctx, router, i, now, false)
	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	editor.SetCaret(3, 3)
	state, _ := core.PeekState[inputState](ctx, "name", stateSlotInput)
	state.queueEdit(EditPaste)

	layoutContextMenuInputFrame(ctx, router, i, now.Add(time.Millisecond), false)
	router.Queue(transfer.DataEvent{
		Type: "application/text",
		Open: func() io.ReadCloser { return io.NopCloser(strings.NewReader("XYZ")) },
	})
	layoutContextMenuInputFrame(ctx, router, i, now.Add(2*time.Millisecond), false)
	if got := editor.Text(); got != "AdaXYZ" {
		t.Fatalf("editor text after paste = %q, want AdaXYZ", got)
	}
}

func TestInputContextMenuOpensOnRightClick(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	i := New("Ada").Key("name")
	start := time.Unix(5, 0)

	layoutContextMenuInputFrame(ctx, router, i, start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 20),
	})
	layoutContextMenuInputFrame(ctx, router, i, start.Add(time.Millisecond), true)
	layoutContextMenuInputFrame(ctx, router, i, start.Add(contextMenuEnterDuration+time.Millisecond), true)
	if !core.HasTopOverlay(ctx) {
		t.Fatal("right-click did not register the edit menu overlay")
	}
}

func TestInputContextMenuOptOut(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	i := New("Ada").Key("name").DefaultContextMenu(false)
	start := time.Unix(6, 0)

	layoutContextMenuInputFrame(ctx, router, i, start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 20),
	})
	layoutContextMenuInputFrame(ctx, router, i, start.Add(time.Millisecond), true)
	layoutContextMenuInputFrame(ctx, router, i, start.Add(contextMenuEnterDuration+time.Millisecond), true)
	if core.HasTopOverlay(ctx) {
		t.Fatal("opted-out input still registered the edit menu overlay")
	}
}

func TestInputContextMenuCustomReplacesDefault(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	custom := menu.New("custom", []menu.Item{{Key: "clear", Label: "Clear"}})
	i := New("Ada").Key("name").ContextMenu(custom)
	start := time.Unix(7, 0)

	layoutContextMenuInputFrame(ctx, router, i, start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 20),
	})
	layoutContextMenuInputFrame(ctx, router, i, start.Add(time.Millisecond), true)
	layoutContextMenuInputFrame(ctx, router, i, start.Add(contextMenuEnterDuration+time.Millisecond), true)
	if !core.HasTopOverlay(ctx) {
		t.Fatal("custom context menu did not register an overlay")
	}
}

func TestTextAreaContextMenuOpensOnRightClick(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	ta := NewTextArea("Ada").Key("body")
	start := time.Unix(8, 0)

	layoutContextMenuTextAreaFrame(ctx, router, ta, start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 20),
	})
	layoutContextMenuTextAreaFrame(ctx, router, ta, start.Add(time.Millisecond), true)
	layoutContextMenuTextAreaFrame(ctx, router, ta, start.Add(contextMenuEnterDuration+time.Millisecond), true)
	if !core.HasTopOverlay(ctx) {
		t.Fatal("right-click did not register the text area edit menu overlay")
	}
}

func layoutContextMenuGroupFrame(ctx *core.Context, router *input.Router, g GroupWidget, now time.Time, overlays bool) {
	viewport := image.Pt(contextMenuTestViewportX, contextMenuTestViewportY)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	g.Layout(ctx, gtx)
	if overlays {
		core.LayoutOverlays(ctx, gtx)
	}
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func TestGroupContextMenuOpensOnRightClick(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	g := Group(New("Ada").Key("name"))
	start := time.Unix(9, 0)

	layoutContextMenuGroupFrame(ctx, router, g, start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 20),
	})
	layoutContextMenuGroupFrame(ctx, router, g, start.Add(time.Millisecond), true)
	layoutContextMenuGroupFrame(ctx, router, g, start.Add(contextMenuEnterDuration+time.Millisecond), true)
	if !core.HasTopOverlay(ctx) {
		t.Fatal("right-click did not register the group edit menu overlay")
	}
}

func TestGroupTextAreaContextMenuOpensOnRightClick(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	g := GroupTextArea(NewTextArea("Ada").Key("body"))
	start := time.Unix(10, 0)

	layoutContextMenuGroupFrame(ctx, router, g, start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 20),
	})
	layoutContextMenuGroupFrame(ctx, router, g, start.Add(time.Millisecond), true)
	layoutContextMenuGroupFrame(ctx, router, g, start.Add(contextMenuEnterDuration+time.Millisecond), true)
	if !core.HasTopOverlay(ctx) {
		t.Fatal("right-click did not register the group text area edit menu overlay")
	}
}

func TestEditMenuLabelsLocalized(t *testing.T) {
	zh := editMenuLabelsFor(core.New(nil, testtheme.Behavior(), locale.LanguageChinese))
	if zh.cut != "剪切" || zh.copy != "复制" || zh.paste != "粘贴" || zh.selectAll != "全选" {
		t.Fatalf("zh labels = %+v", zh)
	}
	en := editMenuLabelsFor(core.New(nil, testtheme.Behavior(), locale.LanguageEnglish))
	if en.cut != "Cut" || en.copy != "Copy" || en.paste != "Paste" || en.selectAll != "Select All" {
		t.Fatalf("en labels = %+v", en)
	}
}
