package command

import (
	"image"
	"runtime"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/widget/button"
	inputui "github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/widget/tooltip"
)

func TestKeyShortcutNormalizesNamesAndModifiers(t *testing.T) {
	primary := "Ctrl"
	if runtime.GOOS == "darwin" {
		primary = "Command"
	}
	for _, test := range []struct {
		name      string
		shortcut  Shortcut
		wantName  key.Name
		wantMods  key.Modifiers
		wantLabel string
	}{
		{name: "primary letter", shortcut: KeyShortcut("s", ShortcutPrimary), wantName: key.Name("S"), wantMods: key.ModShortcut, wantLabel: primary + "+S"},
		{name: "named key", shortcut: KeyShortcut("escape", 0), wantName: key.NameEscape, wantLabel: "Esc"},
		{name: "function key", shortcut: KeyShortcut("f12", ShortcutAlt), wantName: key.Name("F12"), wantMods: key.ModAlt, wantLabel: "Alt+F12"},
		{name: "shifted plus", shortcut: KeyShortcut("+", ShortcutPrimary|ShortcutShift), wantName: key.Name("+"), wantMods: key.ModShortcut | key.ModShift, wantLabel: primary + "++"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.shortcut.name != test.wantName || test.shortcut.modifiers != test.wantMods || test.shortcut.String() != test.wantLabel {
				t.Fatalf("shortcut = name %q modifiers %v label %q, want %q %v %q", test.shortcut.name, test.shortcut.modifiers, test.shortcut, test.wantName, test.wantMods, test.wantLabel)
			}
		})
	}
}

func TestKeyShortcutRejectsUnsafeAndUnsupportedKeys(t *testing.T) {
	assertCommandPanic(t, func() { KeyShortcut("A", 0) })
	assertCommandPanic(t, func() { KeyShortcut("", ShortcutPrimary) })
	assertCommandPanic(t, func() { KeyShortcut("unknown", ShortcutPrimary) })
	assertCommandPanic(t, func() { KeyShortcut("A", ShortcutModifier(1<<7)) })
}

func TestCommandScopeExecutesPressedShortcut(t *testing.T) {
	ctx := commandTestContext()
	router := new(input.Router)
	called := 0
	scope := Scope([]Command{
		New("save", "Save").
			Shortcut(KeyShortcut("S", ShortcutPrimary)).
			OnExecute(func() { called++ }),
	}, nil)
	start := time.Unix(1, 0)
	layoutCommandFrame(ctx, router, scope, start)
	router.Queue(key.Event{Name: key.Name("S"), Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandFrame(ctx, router, scope, start.Add(time.Millisecond))
	if called != 1 {
		t.Fatalf("command executions = %d, want 1", called)
	}
}

func TestCommandScopeIgnoresReleaseExtraModifiersAndDisabledCommands(t *testing.T) {
	ctx := commandTestContext()
	router := new(input.Router)
	called := 0
	shortcut := KeyShortcut("S", ShortcutPrimary)
	scope := Scope([]Command{
		New("save", "Save").Shortcut(shortcut).OnExecute(func() { called++ }),
		New("disabled", "Disabled").Shortcut(KeyShortcut("D", ShortcutPrimary)).Disabled(true),
	}, nil)
	start := time.Unix(2, 0)
	layoutCommandFrame(ctx, router, scope, start)
	router.Queue(
		key.Event{Name: key.Name("S"), Modifiers: key.ModShortcut, State: key.Release},
		key.Event{Name: key.Name("S"), Modifiers: key.ModShortcut | key.ModShift, State: key.Press},
		key.Event{Name: key.Name("D"), Modifiers: key.ModShortcut, State: key.Press},
	)
	layoutCommandFrame(ctx, router, scope, start.Add(time.Millisecond))
	if called != 0 {
		t.Fatalf("ignored shortcut executions = %d, want 0", called)
	}
}

func TestCommandScopeLetsChildConsumeShortcutFirst(t *testing.T) {
	ctx := commandTestContext()
	router := new(input.Router)
	shortcut := KeyShortcut("S", ShortcutPrimary)
	childCalls := 0
	commandCalls := 0
	scope := Scope(
		[]Command{New("save", "Save").Shortcut(shortcut).OnExecute(func() { commandCalls++ })},
		shortcutConsumer{shortcut: shortcut, calls: &childCalls},
	)
	start := time.Unix(3, 0)
	layoutCommandFrame(ctx, router, scope, start)
	router.Queue(key.Event{Name: key.Name("S"), Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandFrame(ctx, router, scope, start.Add(time.Millisecond))
	if childCalls != 1 || commandCalls != 0 {
		t.Fatalf("shortcut calls = child %d command %d, want 1/0", childCalls, commandCalls)
	}
}

func TestCommandScopeCanPauseWhileFieldIsFocused(t *testing.T) {
	ctx := commandTestContext()
	router := new(input.Router)
	called := 0
	scope := Scope(
		[]Command{New("duplicate", "Duplicate").Shortcut(KeyShortcut("D", ShortcutPrimary)).OnExecute(func() { called++ })},
		inputui.New("Ada").Key("name"),
	).DisableWhenFieldFocused()
	start := time.Unix(4, 0)
	layoutCommandFrame(ctx, router, scope, start)
	editor := core.FieldFocusTag(ctx, "name")
	router.Source().Execute(key.FocusCmd{Tag: editor})
	layoutCommandFrame(ctx, router, scope, start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.Name("D"), Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandFrame(ctx, router, scope, start.Add(2*time.Millisecond))
	if called != 0 {
		t.Fatalf("command executions with focused field = %d, want 0", called)
	}

	router.Source().Execute(key.FocusCmd{})
	layoutCommandFrame(ctx, router, scope, start.Add(3*time.Millisecond))
	router.Queue(key.Event{Name: key.Name("D"), Modifiers: key.ModShortcut, State: key.Press})
	layoutCommandFrame(ctx, router, scope, start.Add(4*time.Millisecond))
	if called != 1 {
		t.Fatalf("command executions without focused field = %d, want 1", called)
	}
}

func TestCommandScopeRejectsInvalidCommandSets(t *testing.T) {
	shortcut := KeyShortcut("S", ShortcutPrimary)
	tests := []Widget{
		Scope([]Command{
			New("save", "Save").OnExecute(func() {}),
			New("save", "Save again").OnExecute(func() {}),
		}, nil),
		Scope([]Command{
			New("save", "Save").Shortcut(shortcut).OnExecute(func() {}),
			New("store", "Store").Shortcut(shortcut).OnExecute(func() {}),
		}, nil),
		Scope([]Command{New("save", "Save")}, nil),
	}
	for index, scope := range tests {
		t.Run(string(rune('A'+index)), func(t *testing.T) {
			assertCommandPanic(t, func() {
				scope.Layout(commandTestContext(), commandLayoutContext(new(input.Router), time.Unix(4, 0)))
			})
		})
	}
}

func TestCommandAdaptersPreserveBehavior(t *testing.T) {
	called := 0
	base := New("save", "Save").
		Description("Save changes").
		Shortcut(KeyShortcut("S", ShortcutPrimary)).
		OnExecute(func() { called++ })
	item := MenuItem(base.Danger(true).KeepOpen(true))
	if item.Key != "save" || item.Label != "Save" || item.Description != "Save changes" || item.Shortcut == "" || item.Variant != menu.ItemDanger || !item.KeepOpen || item.OnAction == nil {
		t.Fatalf("menu item = %#v", item)
	}
	item.OnAction()
	if called != 1 {
		t.Fatalf("menu command executions = %d, want 1", called)
	}

	if _, ok := Button("save-text", base).(button.Widget); !ok {
		t.Fatal("text command did not produce a plain Button")
	}
	if _, ok := Button("bold-text", base.Toggle(false)).(togglebutton.Widget); !ok {
		t.Fatal("text toggle command did not produce a plain ToggleButton")
	}
	iconCommand := base.Icon(fixedCommandWidget{size: image.Pt(16, 16)})
	if Button("save-icon", iconCommand) == nil {
		t.Fatal("icon command button did not include a Tooltip")
	}
}

func TestMenuItemForLocalizesShortcutNames(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageChinese)
	command := New("move-left", "Move left").
		Shortcut(KeyShortcut("left", 0)).
		OnExecute(func() {})
	item := MenuItemFor(ctx, command)
	if item.Shortcut != "左" {
		t.Fatalf("localized shortcut = %q, want 左", item.Shortcut)
	}
	if got := MenuItem(command).Shortcut; got != "Left" {
		t.Fatalf("compatibility shortcut = %q, want Left", got)
	}
}

func TestCommandButtonTooltipKeyDoesNotCollideWithUserKey(t *testing.T) {
	ctx := commandTestContext()
	router := new(input.Router)
	gtx := commandLayoutContext(router, time.Unix(5, 0))
	command := New("save", "Save").
		Icon(fixedCommandWidget{size: image.Pt(16, 16)}).
		OnExecute(func() {})
	userTooltip := tooltip.New("save-button:tooltip", fixedCommandWidget{size: image.Pt(16, 16)}, "").Content(fixedCommandWidget{size: image.Pt(16, 16)})

	core.BeginFrame(ctx)
	Button("save-button", command).Layout(ctx, gtx)
	userTooltip.Layout(ctx, gtx)
	core.EndFrame(ctx)
}

type shortcutConsumer struct {
	shortcut Shortcut
	calls    *int
}

func (c shortcutConsumer) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	for {
		value, ok := gtx.Event(c.shortcut.filter())
		if !ok {
			break
		}
		if event, ok := value.(key.Event); ok && event.State == key.Press {
			(*c.calls)++
		}
	}
	return layout.Dimensions{Size: image.Pt(40, 20)}
}

type fixedCommandWidget struct {
	size image.Point
}

func (w fixedCommandWidget) Layout(*core.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: w.size}
}

func commandTestContext() *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func commandLayoutContext(router *input.Router, now time.Time) layout.Context {
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(400, 300)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
}

func layoutCommandFrame(ctx *core.Context, router *input.Router, scope Widget, now time.Time) {
	gtx := commandLayoutContext(router, now)
	core.BeginFrame(ctx)
	scope.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(gtx.Ops)
}

func assertCommandPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("operation did not panic")
		}
	}()
	fn()
}
