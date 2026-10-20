package alert

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type incompleteAlertProvider struct {
	theme.ComponentVisualProvider
}

func (p incompleteAlertProvider) Alert(active *theme.Theme, request theme.AlertRequest) theme.AlertVisual {
	visual := p.ComponentVisualProvider.Alert(active, request)
	visual.Defaults = style.Style{}
	return visual
}

func TestAlertVisualValidationDoesNotCommitIncompleteProvider(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteAlertProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, &active, locale.LanguageAuto)
	cache := alertVisualCache{}
	value := New("incomplete", "provider")

	func() {
		defer func() {
			recovered := recover()
			if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
				t.Fatalf("incomplete alert provider panic = %v", recovered)
			}
		}()
		value.visualCache = &cache
		value.visualWithCache(ctx, nil)
	}()
	if stats := cache.stats(); stats.Misses != 0 || stats.Hits != 0 {
		t.Fatalf("incomplete alert visual polluted cache = %#v", stats)
	}

	active.VisualProvider = testtheme.BehaviorProvider()
	core.ReplaceTheme(ctx, active)
	value.visualWithCache(ctx, nil)
	if stats := cache.stats(); stats.Misses != 1 || stats.Hits != 0 {
		t.Fatalf("complete alert provider did not load after failed visual = %#v", stats)
	}
}
