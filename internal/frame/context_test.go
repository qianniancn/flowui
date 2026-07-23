package frame

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/app"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestNewUsesProvidedTheme(t *testing.T) {
	theme := theme.DefaultTheme()
	theme.Palette.Accent = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	ctx := New(nil, &theme, locale.LanguageAuto)
	if ActiveTheme(ctx) != &theme {
		t.Fatal("context did not use provided theme")
	}
	snapshot := ctx.Theme()
	snapshot.Palette.Accent = color.NRGBA{}
	if ActiveTheme(ctx).Palette.Accent != theme.Palette.Accent {
		t.Fatal("mutating the public theme snapshot changed the active theme")
	}
}

func TestPushStyleRestoresCascadeScope(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	first := style.Style{}.Background(style.RGB(0x112233))
	second := style.Style{}.PaddingX(12)
	restoreFirst := PushStyle(ctx, first)
	restoreSecond := PushStyle(ctx, second)

	active := ActiveStyles(ctx)
	if len(active) != 2 || active[0].Resolve(style.StyleState{}).Paint.Background != style.RGB(0x112233) {
		t.Fatalf("active styles = %#v", active)
	}
	restoreSecond()
	if len(ActiveStyles(ctx)) != 1 {
		t.Fatalf("nested style was not restored")
	}
	restoreFirst()
	if len(ActiveStyles(ctx)) != 0 {
		t.Fatalf("style scope leaked")
	}
}

func TestContextRejectsDuplicateExplicitKeys(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	ctx.Clickable("save")
	mustPanic(t, func() {
		ctx.Editor("save")
	})
}

func TestContextScopesExplicitKeys(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	pop := PushKey(ctx, "todo:1")
	ctx.Clickable("delete")
	pop()
	ctx.Clickable("delete")

	if state, _ := PeekState[widget.Clickable](ctx, "todo:1/delete", stateSlotClickable); state == nil {
		t.Fatal("missing scoped clickable")
	}
	if state, _ := PeekState[widget.Clickable](ctx, "delete", stateSlotClickable); state == nil {
		t.Fatal("missing root clickable")
	}
}

func TestContextRetainsDraggableStateByKey(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	first := ctx.Draggable("component")
	EndFrame(ctx)

	BeginFrame(ctx)
	second := ctx.Draggable("component")
	if first != second {
		t.Fatal("draggable state was not retained by key")
	}
}

func TestContextPublicFocusHelpersPreserveModality(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	BeginFrame(ctx)
	pointerTarget := new(int)
	ctx.RequestFocusVisible(pointerTarget, false)
	if ctx.FocusVisible(pointerTarget, true) {
		t.Fatal("pointer-originated custom focus was visible")
	}

	keyboardTarget := new(int)
	ctx.RequestFocus(keyboardTarget)
	if !ctx.FocusVisible(keyboardTarget, true) {
		t.Fatal("keyboard-originated custom focus was hidden")
	}
	ctx.RequestFocusVisible(keyboardTarget, false)
	ctx.PreserveFocus()
	if ctx.FocusVisible(keyboardTarget, true) {
		t.Fatal("pointer-originated focus remained visible after preservation")
	}
}

func TestContextScopesKeysWithSeparators(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)

	pop := PushKey(ctx, "a/b")
	parentSlashKey, parentSlash := ClickableWithKey(ctx, "c")
	pop()

	popA := PushKey(ctx, "a")
	popB := PushKey(ctx, "b")
	nestedKey, nested := ClickableWithKey(ctx, "c")
	popB()
	leafSlashKey, leafSlash := ClickableWithKey(ctx, "b/c")
	popA()

	rootKey, root := ClickableWithKey(ctx, "a/b/c")
	keys := []string{parentSlashKey, nestedKey, leafSlashKey, rootKey}
	states := []*widget.Clickable{parentSlash, nested, leafSlash, root}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] == keys[j] {
				t.Fatalf("distinct component paths produced key %q", keys[i])
			}
			if states[i] == states[j] {
				t.Fatalf("distinct component paths reused state for %q and %q", keys[i], keys[j])
			}
		}
	}
}

func TestContextSweepsUnmountedInteractionState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	ctx.Clickable("save")
	UseState[int](ctx, "custom", "test")
	EndFrame(ctx)

	BeginFrame(ctx)
	EndFrame(ctx)
	if StateLen(ctx) != 0 {
		t.Fatalf("interaction states = %d, want 0", StateLen(ctx))
	}
}

func TestPreparedAssociationTransitionRequestsAnotherFrame(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	ctx.fieldDescriptions = map[string]string{"email": "Supporting text"}
	ctx.previousDescriptions = map[string]string{"email": "Supporting text"}
	ctx.preparedDescriptions = map[string]struct{}{}
	ctx.previousPreparedDescriptions = map[string]struct{}{"email": {}}

	if !fieldAssociationsChanged(ctx) {
		t.Fatal("prepared-to-unprepared transition did not change field associations")
	}
}

func TestContextTracksWindowState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	state := UpdateWindowConfig(ctx, app.Config{
		Size:      image.Pt(640, 480),
		Mode:      app.Maximized,
		Focused:   true,
		Decorated: true,
		TopMost:   true,
	})
	if state != (WindowState{Size: image.Pt(640, 480), Mode: Maximized, Focused: true, Decorated: true, TopMost: true}) {
		t.Fatalf("window state = %#v", state)
	}
	BeginFrameWithViewport(ctx, image.Pt(620, 440))
	if got := ctx.WindowState(); got.Size != image.Pt(620, 440) || got.Mode != Maximized || !got.Focused {
		t.Fatalf("frame window state = %#v", got)
	}
}

func TestWindowModeString(t *testing.T) {
	for mode, want := range map[WindowMode]string{
		Windowed: "windowed", Fullscreen: "fullscreen", Minimized: "minimized", Maximized: "maximized", WindowMode(255): "unknown",
	} {
		if got := mode.String(); got != want {
			t.Fatalf("WindowMode(%d).String() = %q, want %q", mode, got, want)
		}
	}
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
