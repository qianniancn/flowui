package chip

import (
	"errors"
	"image"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type incompleteChipProvider struct {
	theme.ComponentVisualProvider
}

func (incompleteChipProvider) Chip(*theme.Theme, theme.ChipRequest) theme.ChipVisual {
	return theme.ChipVisual{}
}

func TestChipIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteChipProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("A")
	for range 2 {
		func() {
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete chip visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			value.Layout(ctx, chipLayoutContext(nil, image.Pt(80, 40)))
		}()
	}
	if stats := value.visualCache.Stats(); stats.Hits != 0 {
		t.Fatalf("incomplete chip visual produced cache hit: %#v", stats)
	}
}

func TestChipMissingProviderFailsExplicitly(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	defer func() {
		if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing chip provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	New("A").Layout(ctx, chipLayoutContext(nil, image.Pt(80, 40)))
}

func TestChipVisualCacheTracksKeyAndTheme(t *testing.T) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("A").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)

	value.visualWithCache(firstContext, nil)
	value.visualWithCache(firstContext, nil)
	value.Color(ColorDanger).visualWithCache(firstContext, nil)
	value.visualWithCache(secondContext, nil)

	stats := value.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.KeyInvalidations != 2 || stats.ThemeInvalidations != 1 {
		t.Fatalf("chip cache stats = %#v, want one hit, three misses, two key and one theme invalidation", stats)
	}
}
