package label

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

func TestLabelProviderVisualStates(t *testing.T) {
	active := *testtheme.Behavior()
	provider := active.VisualProvider.(theme.LabelProvider)
	foreground := color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	visual := provider.Label(&active, theme.LabelRequest{Foreground: foreground})
	theme.ValidateLabelVisual(visual)

	textColor := func(state style.State) color.NRGBA {
		t.Helper()
		resolved := style.Cascade(state, visual.Defaults, visual.Override)
		if resolved.Text == nil {
			t.Fatalf("label declared no text properties for state %v", state)
		}
		value, ok := computed.Color(resolved.Text.Color)
		if !ok {
			t.Fatalf("label text color for state %v is not literal", state)
		}
		return value
	}

	if got := textColor(style.State{}); got != foreground {
		t.Fatalf("base label color = %#v, want %#v", got, foreground)
	}
	if got := textColor(style.State{Invalid: true}); got != active.Semantic().Tone(theme.ToneDanger).Fill {
		t.Fatalf("invalid label color = %#v", got)
	}
	if got := textColor(style.State{Disabled: true}); got != active.DisabledColor(foreground) {
		t.Fatalf("disabled label color = %#v", got)
	}
	if got := textColor(style.State{Disabled: true, Invalid: true}); got != active.DisabledColor(active.Semantic().Tone(theme.ToneDanger).Fill) {
		t.Fatalf("disabled invalid label color = %#v", got)
	}
}

func TestLabelStyledOverrideOnlyChangesRootVisual(t *testing.T) {
	active := *testtheme.Behavior()
	provider := active.VisualProvider.(theme.LabelProvider)
	foreground := color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	base := provider.Label(&active, theme.LabelRequest{Foreground: foreground})
	override := provider.Label(&active, theme.LabelRequest{
		Foreground: foreground,
		Override:   &theme.LabelOverride{Declaration: style.Style{}.TextColor(style.RGB(0x123456))},
	})
	baseRoot := style.Cascade(style.State{}, base.Defaults, base.Override)
	overrideRoot := style.Cascade(style.State{}, override.Defaults, override.Override)
	if baseRoot.Text == nil || overrideRoot.Text == nil || baseRoot.Text.FontSize == nil || overrideRoot.Text.FontSize == nil ||
		*baseRoot.Text.FontSize != *overrideRoot.Text.FontSize {
		t.Fatal("label override changed provider typography")
	}
	if override.Required.Hash64() != base.Required.Hash64() {
		t.Fatal("label override changed required-mark declaration")
	}
}

func TestLabelMissingProviderDoesNotCache(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	value := New("Name")
	defer func() {
		if recover() == nil {
			t.Fatal("label without a visual provider did not panic")
		}
		if stats := value.visualCache.stats(); stats.Misses != 0 || stats.Size != 0 {
			t.Fatalf("missing provider polluted cache: %#v", stats)
		}
	}()
	value.visualWithCache(ctx, nil)
}
