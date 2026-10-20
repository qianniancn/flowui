package radiogroup

import (
	"errors"
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type countingRadioProvider struct {
	theme.ComponentVisualProvider
	calls *int
}

type incompleteRadioProvider struct {
	theme.ComponentVisualProvider
	mutate func(*theme.RadioGroupVisual)
}

func (p countingRadioProvider) RadioGroup(active *theme.Theme, request theme.RadioGroupRequest) theme.RadioGroupVisual {
	*p.calls++
	return p.ComponentVisualProvider.RadioGroup(active, request)
}

func (p incompleteRadioProvider) RadioGroup(active *theme.Theme, request theme.RadioGroupRequest) theme.RadioGroupVisual {
	visual := p.ComponentVisualProvider.RadioGroup(active, request)
	p.mutate(&visual)
	return visual
}

func TestVisualCacheReusesVariantsAndInvalidatesTheme(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingRadioProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	var cache VisualCache

	cache.Resolve(active, RadioPrimary)
	cache.Resolve(active, RadioPrimary)
	cache.Resolve(active, RadioSecondary)
	cache.Resolve(active, RadioSecondary)
	if calls != 2 {
		t.Fatalf("provider calls = %d, want one per variant", calls)
	}
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 2 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 1 {
		t.Fatalf("cache stats = %+v", stats)
	}

	replacement := *active
	cache.Resolve(&replacement, RadioPrimary)
	if calls != 3 {
		t.Fatalf("theme replacement provider calls = %d, want 3", calls)
	}
	if got := cache.Stats().ThemeInvalidations; got != 1 {
		t.Fatalf("theme invalidations = %d, want 1", got)
	}
}

func TestRadioVisualCacheOverrideBypassesCache(t *testing.T) {
	active := testtheme.Behavior()
	var cache VisualCache
	cache.Resolve(active, RadioPrimary)
	before := cache.Stats()

	_ = radioVisualCached(active, RadioPrimary, &theme.RadioGroupOverride{}, &cache)
	if got := cache.Stats(); got != before {
		t.Fatalf("override changed visual cache stats: before=%#v after=%#v", before, got)
	}
}

func TestRadioGroupLayoutDoesNotResolveProviderTwice(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingRadioProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	var ops op.Ops
	value := New([]RadioItem{
		{Key: "basic", Label: "Basic"},
		{Key: "pro", Label: "Pro"},
	}).Key("once").SelectedKey("basic")
	value.Layout(ctx, layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(320, 120)}})
	if calls != 1 {
		t.Fatalf("provider calls during one layout = %d, want 1", calls)
	}
}

func TestIncompleteRadioVisualIsNotCached(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*theme.RadioGroupVisual)
	}{
		{name: "missing indicator resolver", mutate: func(visual *theme.RadioGroupVisual) {
			visual.IndicatorColor = nil
		}},
		{name: "missing indicator renderer", mutate: func(visual *theme.RadioGroupVisual) {
			visual.IndicatorRenderer = nil
		}},
		{name: "invalid pressed scale", mutate: func(visual *theme.RadioGroupVisual) {
			visual.Metrics.PressedScale = 0
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			active := testtheme.Behavior()
			active.VisualProvider = incompleteRadioProvider{
				ComponentVisualProvider: testtheme.BehaviorProvider(),
				mutate:                  test.mutate,
			}
			var cache VisualCache

			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete radio visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
				stats := cache.Stats()
				if stats.Size != 0 {
					t.Fatalf("incomplete radio visual polluted cache = %+v", stats)
				}
				if stats.Misses == 0 {
					t.Fatalf("incomplete radio visual did not record a failed miss = %+v", stats)
				}
			}()
			cache.Resolve(active, RadioPrimary)
		})
	}
}
