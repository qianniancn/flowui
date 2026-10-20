package closebutton

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type incompleteCloseButtonProvider struct {
	theme.ComponentVisualProvider
}

func (incompleteCloseButtonProvider) CloseButton(*theme.Theme, theme.CloseButtonRequest) theme.CloseButtonVisual {
	return theme.CloseButtonVisual{}
}

func TestCloseButtonIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteCloseButtonProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("close")
	for range 2 {
		func() {
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete close button visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			value.visualWithCache(ctx, nil)
		}()
	}
	if stats := value.visualCache.Stats(); stats.Hits != 0 {
		t.Fatalf("incomplete close button visual produced cache hit: %#v", stats)
	}
}

func TestCloseButtonMissingProviderFailsExplicitly(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	defer func() {
		if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing close button provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	New("close").Layout(ctx, closeButtonTestContext())
}

func TestCloseButtonVisualCacheTracksTheme(t *testing.T) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("close")

	value.visualWithCache(firstContext, nil)
	value.visualWithCache(firstContext, nil)
	value.visualWithCache(secondContext, nil)

	stats := value.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.ThemeInvalidations != 1 {
		t.Fatalf("close button cache stats = %#v, want one hit, two misses and one theme invalidation", stats)
	}
}
