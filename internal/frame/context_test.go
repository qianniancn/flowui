package frame

import (
	"image/color"
	"testing"

	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/locale"
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

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
