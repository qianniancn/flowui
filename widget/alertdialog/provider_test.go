package alertdialog

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type incompleteAlertDialogProvider struct {
	theme.ComponentVisualProvider
}

func (p incompleteAlertDialogProvider) AlertDialog(active *theme.Theme, request theme.AlertDialogRequest) theme.AlertDialogVisual {
	visual := p.ComponentVisualProvider.AlertDialog(active, request)
	visual.Title = style.Style{}
	return visual
}

func TestAlertDialogVisualValidationDoesNotCommitIncompleteProvider(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteAlertDialogProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, &active, locale.LanguageAuto)
	cache := alertDialogVisualCache{}
	value := New("Title", "Body").Key("incomplete")

	func() {
		defer func() {
			recovered := recover()
			if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
				t.Fatalf("incomplete alert dialog provider panic = %v", recovered)
			}
		}()
		value.visualCache = &cache
		value.visualWithCache(ctx, value.size, value.backdrop, value.status, nil)
	}()
	if stats := cache.stats(); stats.Misses != 0 || stats.Hits != 0 {
		t.Fatalf("incomplete alert dialog visual polluted cache = %#v", stats)
	}

	active.VisualProvider = testtheme.BehaviorProvider()
	core.ReplaceTheme(ctx, active)
	value.visualWithCache(ctx, value.size, value.backdrop, value.status, nil)
	if stats := cache.stats(); stats.Misses != 1 || stats.Hits != 0 {
		t.Fatalf("complete alert dialog provider did not load after failed visual = %#v", stats)
	}
}
