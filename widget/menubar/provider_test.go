package menubar

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type incompleteMenubarProvider struct {
	theme.ComponentVisualProvider
}

func (incompleteMenubarProvider) Menubar(*theme.Theme, theme.MenubarRequest) theme.MenubarVisual {
	return theme.MenubarVisual{}
}

func TestMenubarVisualValidationRejectsIncompleteProviderWithoutCaching(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteMenubarProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := newMenubarContext(t, active)
	cache := &menubarVisualCache{}

	deferred := func() (recovered any) {
		defer func() { recovered = recover() }()
		(Widget{}).visualWithCache(ctx, nil, cache)
		return nil
	}()
	if recoveredErr, ok := deferred.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
		t.Fatalf("incomplete menubar visual panic = %v, want %v", deferred, theme.ErrIncompleteVisualProvider)
	}
	stats := cache.stats()
	if stats.Misses != 0 || stats.Hits != 0 {
		t.Fatalf("cache stats after rejected visual = %#v, want no hit or miss", stats)
	}

	active.VisualProvider = testtheme.BehaviorProvider()
	visual := (Widget{}).visualWithCache(ctx, nil, cache)
	if visual.Trigger == nil || visual.RenderTrigger == nil || visual.RenderLabel == nil {
		t.Fatal("complete provider did not load after rejected visual")
	}
	if stats = cache.stats(); stats.Misses != 1 || stats.Hits != 0 {
		t.Fatalf("cache stats after recovery = %#v, want one miss and no hit", stats)
	}
}

func newMenubarContext(t *testing.T, active *theme.Theme) *core.Context {
	t.Helper()
	return core.New(nil, active, locale.LanguageEnglish)
}
