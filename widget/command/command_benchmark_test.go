package command

import (
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkCommandScopeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	scope := Scope([]Command{
		New("save", "Save").Shortcut(KeyShortcut("S", ShortcutPrimary)).OnExecute(func() {}),
		New("close", "Close").Shortcut(KeyShortcut("W", ShortcutPrimary)).OnExecute(func() {}),
	}, nil)
	layoutCommandFrame(ctx, router, scope, time.Unix(0, 0))
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		layoutCommandFrame(ctx, router, scope, time.Unix(0, int64(index+1)*int64(time.Millisecond)))
	}
}

func BenchmarkCommandShortcutDispatch(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	calls := 0
	scope := Scope([]Command{
		New("save", "Save").Shortcut(KeyShortcut("S", ShortcutPrimary)).OnExecute(func() { calls++ }),
	}, nil)
	start := time.Unix(1, 0)
	layoutCommandFrame(ctx, router, scope, start)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		router.Queue(key.Event{Name: key.Name("S"), Modifiers: key.ModShortcut, State: key.Press})
		layoutCommandFrame(ctx, router, scope, start.Add(time.Duration(index+1)*time.Millisecond))
	}
	b.StopTimer()
	if calls != b.N {
		b.Fatalf("shortcut calls = %d, want %d", calls, b.N)
	}
}
