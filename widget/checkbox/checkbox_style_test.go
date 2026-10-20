package checkbox

import (
	"errors"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type countingCheckboxProvider struct {
	theme.ComponentVisualProvider
	calls *int
}

type incompleteCheckboxProvider struct {
	theme.ComponentVisualProvider
	mutate func(*theme.CheckboxVisual)
}

func (p incompleteCheckboxProvider) Checkbox(active *theme.Theme, request theme.CheckboxRequest) theme.CheckboxVisual {
	visual := p.ComponentVisualProvider.Checkbox(active, request)
	p.mutate(&visual)
	return visual
}

func (p countingCheckboxProvider) Checkbox(active *theme.Theme, request theme.CheckboxRequest) theme.CheckboxVisual {
	(*p.calls)++
	return p.ComponentVisualProvider.Checkbox(active, request)
}

func TestVisualCacheReusesVariantsAndInvalidatesTheme(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingCheckboxProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	var cache VisualCache

	cache.Resolve(active, Primary)
	cache.Resolve(active, Primary)
	cache.Resolve(active, Secondary)
	cache.Resolve(active, Secondary)
	if calls != 2 {
		t.Fatalf("provider calls = %d, want one per variant", calls)
	}
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 2 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 1 {
		t.Fatalf("cache stats = %+v", stats)
	}

	replacement := *active
	cache.Resolve(&replacement, Primary)
	if calls != 3 {
		t.Fatalf("theme replacement provider calls = %d, want 3", calls)
	}
	if got := cache.Stats().ThemeInvalidations; got != 1 {
		t.Fatalf("theme invalidations = %d, want 1", got)
	}
}

func TestCheckboxLayoutDoesNotResolveProviderTwice(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingCheckboxProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	var ops op.Ops
	New("").Key("once").Checked(false).Layout(ctx, layout.Context{Ops: &ops})
	if calls != 1 {
		t.Fatalf("provider calls during one layout = %d, want 1", calls)
	}
}

func TestIncompleteCheckboxVisualIsNotCached(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteCheckboxProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.CheckboxVisual) {
			visual.Metrics.IndicatorSize = 0
		},
	}
	var cache VisualCache

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete checkbox visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Size != 0 {
			t.Fatalf("incomplete checkbox visual polluted cache = %+v", stats)
		}
	}()
	cache.Resolve(active, Primary)
}
