package avatar

import (
	"errors"
	"image"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type incompleteAvatarProvider struct {
	theme.ComponentVisualProvider
}

func (incompleteAvatarProvider) Avatar(*theme.Theme, theme.AvatarRequest) theme.AvatarVisual {
	return theme.AvatarVisual{}
}

func TestAvatarIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteAvatarProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("A")
	for range 2 {
		func() {
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete avatar visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			value.Layout(ctx, avatarLayoutContext(image.Pt(40, 40)))
		}()
	}
	if stats := value.visualCache.Stats(); stats.Hits != 0 {
		t.Fatalf("incomplete avatar visual produced cache hit: %#v", stats)
	}
}

func TestAvatarMissingProviderFailsExplicitly(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	defer func() {
		if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing avatar provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	New("A").Layout(ctx, avatarLayoutContext(image.Pt(40, 40)))
}

func TestAvatarVisualCacheTracksKeyAndTheme(t *testing.T) {
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
		t.Fatalf("avatar cache stats = %#v, want one hit, three misses, two key and one theme invalidation", stats)
	}
}
