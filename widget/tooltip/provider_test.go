package tooltip

import (
	"errors"
	"testing"
	"time"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type incompleteTooltipProvider struct {
	theme.ComponentVisualProvider
}

func (p incompleteTooltipProvider) Tooltip(active *theme.Theme, request theme.TooltipRequest) theme.TooltipVisual {
	visual := p.ComponentVisualProvider.Tooltip(active, request)
	visual.Arrow = nil
	return visual
}

type incompleteTooltipMotionProvider struct {
	theme.ComponentVisualProvider
}

func (p incompleteTooltipMotionProvider) Tooltip(active *theme.Theme, request theme.TooltipRequest) theme.TooltipVisual {
	visual := p.ComponentVisualProvider.Tooltip(active, request)
	visual.Metrics.EnterDuration = 0
	return visual
}

func TestTooltipVisualCacheHitAndInvalidation(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageAuto)
	cache := tooltipVisualCache{}

	tooltipVisualWithCache(ctx, nil, &cache)
	tooltipVisualWithCache(ctx, nil, &cache)
	stats := cache.stats()
	if stats.Misses != 1 || stats.Hits != 1 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 0 {
		t.Fatalf("initial tooltip visual cache stats = %#v", stats)
	}

	replacement := *testtheme.Behavior()
	core.ReplaceTheme(ctx, replacement)
	tooltipVisualWithCache(ctx, nil, &cache)
	stats = cache.stats()
	if stats.Misses != 2 || stats.ThemeInvalidations != 1 {
		t.Fatalf("theme replacement did not invalidate tooltip cache = %#v", stats)
	}

	override := &theme.TooltipOverride{Declaration: style.Style{}.Padding(4)}
	tooltipVisualWithCache(ctx, override, &cache)
	tooltipVisualWithCache(ctx, override, &cache)
	stats = cache.stats()
	if stats.Misses != 3 || stats.Hits != 2 || stats.KeyInvalidations != 1 {
		t.Fatalf("tooltip override cache key = %#v", stats)
	}
}

func TestTooltipRequiresVisualProvider(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageAuto)
	defer func() {
		if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing tooltip provider panic = %v", recovered)
		}
	}()
	New("missing-provider", nil, "").Content(nil).Layout(ctx, testContextAt(time.Time{}))
}

func TestTooltipVisualValidationDoesNotCommitIncompleteProvider(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteTooltipProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, &active, locale.LanguageAuto)
	cache := tooltipVisualCache{}

	func() {
		defer func() {
			recovered := recover()
			if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
				t.Fatalf("incomplete tooltip provider panic = %v", recovered)
			}
		}()
		tooltipVisualWithCache(ctx, nil, &cache)
	}()
	if stats := cache.stats(); stats.Misses != 0 || stats.Hits != 0 {
		t.Fatalf("incomplete tooltip visual polluted cache = %#v", stats)
	}
}

func TestTooltipVisualValidationRejectsIncompleteMotion(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteTooltipMotionProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, &active, locale.LanguageAuto)
	cache := tooltipVisualCache{}
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete tooltip motion panic = %v", recovered)
		}
		if stats := cache.stats(); stats.Misses != 0 || stats.Hits != 0 {
			t.Fatalf("incomplete tooltip motion polluted cache = %#v", stats)
		}
	}()
	tooltipVisualWithCache(ctx, nil, &cache)
}
