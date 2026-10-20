package description

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestDescriptionProviderVisualUsesMutedForeground(t *testing.T) {
	active := *testtheme.Behavior()
	provider := active.VisualProvider.(theme.DescriptionProvider)
	visual := provider.Description(&active, theme.DescriptionRequest{})

	text := func(state style.State) color.NRGBA {
		t.Helper()
		resolved := style.Cascade(state, visual.Defaults, visual.Override)
		if resolved.Text == nil {
			t.Fatalf("description declared no text properties for state %v", state)
		}
		value, ok := computed.Color(resolved.Text.Color)
		if !ok {
			t.Fatalf("description text color for state %v is not a literal color", state)
		}
		return value
	}

	muted := active.Semantic().MutedForeground
	if got := text(style.State{}); got != muted {
		t.Fatalf("description did not use muted foreground: got %#v", got)
	}
	if got := text(style.State{Disabled: true}); got != active.DisabledColor(muted) {
		t.Fatalf("disabled description color was not applied: got %#v", got)
	}
}

func TestDescriptionMissingProviderDoesNotCache(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	value := New("Help")
	defer func() {
		if recover() == nil {
			t.Fatal("description without a visual provider did not panic")
		}
		if stats := value.visualCache.stats(); stats.Misses != 0 || stats.Size != 0 {
			t.Fatalf("missing provider polluted cache: %#v", stats)
		}
	}()
	value.visualWithCache(ctx, nil)
}
