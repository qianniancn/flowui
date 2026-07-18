package ui

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/app"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestWindowSpecUsesIndependentIdentity(t *testing.T) {
	spec := NewWindow("details", func() int { return 0 }, func(*int, int) {}, func(*Context, int, Send[int]) Widget { return Text("Details") }, Title("Details"))
	if spec.Key() != "details" || spec.run == nil || len(spec.options) != 1 {
		t.Fatalf("window spec = key %q run %v options %d", spec.Key(), spec.run != nil, len(spec.options))
	}
}

func TestProgramWindowSpec(t *testing.T) {
	program := Program[int, int]{
		Init:   func() (int, Cmd[int]) { return 1, nil },
		Update: func(*int, int) Cmd[int] { return nil },
		View:   func(*Context, int, Send[int]) Widget { return Text("Program") },
		WindowStateMessage: func(state WindowState) int {
			return state.Size.X
		},
	}
	spec := NewProgramWindow("program", program, Title("Program"))
	if spec.Key() != "program" || spec.run == nil || len(spec.options) != 1 {
		t.Fatalf("program window = key %q run %v options %d", spec.Key(), spec.run != nil, len(spec.options))
	}
}

func TestWindowSpecRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name string
		make func()
	}{
		{"empty key", func() {
			NewWindow("", func() int { return 0 }, func(*int, int) {}, func(*Context, int, Send[int]) Widget { return nil })
		}},
		{"nil initializer", func() {
			NewWindow[int, int]("main", nil, func(*int, int) {}, func(*Context, int, Send[int]) Widget { return nil })
		}},
		{"nil subscription initializer", func() {
			NewWindowWithSubscriptions[int, int]("main", nil, func(*int, int) Cmd[int] { return nil }, nil, func(*Context, int, Send[int]) Widget { return nil })
		}},
		{"nil update", func() {
			NewWindow[int, int]("main", func() int { return 0 }, nil, func(*Context, int, Send[int]) Widget { return nil })
		}},
		{"nil view", func() { NewWindow[int, int]("main", func() int { return 0 }, func(*int, int) {}, nil) }},
		{"nil program init", func() {
			NewProgramWindow("main", Program[int, int]{Update: func(*int, int) Cmd[int] { return nil }, View: func(*Context, int, Send[int]) Widget { return nil }})
		}},
		{"nil program update", func() {
			NewProgramWindow("main", Program[int, int]{Init: func() (int, Cmd[int]) { return 0, nil }, View: func(*Context, int, Send[int]) Widget { return nil }})
		}},
		{"nil program view", func() {
			NewProgramWindow("main", Program[int, int]{Init: func() (int, Cmd[int]) { return 0, nil }, Update: func(*int, int) Cmd[int] { return nil }})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid window definition did not panic")
				}
			}()
			test.make()
		})
	}
}

func TestWindowSetWaitsForEveryWindowAndReportsFailure(t *testing.T) {
	var windows windowSet
	done := windows.begin()
	first := new(app.Window)
	second := new(app.Window)
	if existing, added := windows.add("first", first, new(windowAppearance)); existing != nil || !added {
		t.Fatalf("first add = existing %p added %v", existing, added)
	}
	if existing, added := windows.add("second", second, new(windowAppearance)); existing != nil || !added {
		t.Fatalf("second add = existing %p added %v", existing, added)
	}
	if existing, added := windows.add("first", new(app.Window), new(windowAppearance)); existing != first || added {
		t.Fatalf("duplicate add = existing %p added %v", existing, added)
	}
	windows.finishStarting()
	windows.deactivate("first", first)
	windows.complete(false)
	select {
	case code := <-done:
		t.Fatalf("application exited with second window open: %d", code)
	default:
	}
	windows.deactivate("second", second)
	windows.complete(true)
	if code := <-done; code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestWindowSetDoesNotExitDuringInitialStartup(t *testing.T) {
	var windows windowSet
	done := windows.begin()
	window := new(app.Window)
	_, _ = windows.add("main", window, new(windowAppearance))
	windows.deactivate("main", window)
	windows.complete(false)
	select {
	case code := <-done:
		t.Fatalf("application exited during startup: %d", code)
	default:
	}
	windows.finishStarting()
	if code := <-done; code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestWindowSetAllowsReopenWhileClosedWindowFinishes(t *testing.T) {
	var windows windowSet
	done := windows.begin()
	first := new(app.Window)
	_, _ = windows.add("details", first, new(windowAppearance))
	windows.finishStarting()
	windows.deactivate("details", first)

	second := new(app.Window)
	if existing, added := windows.add("details", second, new(windowAppearance)); existing != nil || !added {
		t.Fatalf("reopen = existing %p added %v", existing, added)
	}
	windows.complete(false)
	select {
	case code := <-done:
		t.Fatalf("application exited with reopened window active: %d", code)
	default:
	}

	windows.deactivate("details", second)
	windows.complete(false)
	if code := <-done; code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestApplicationConfiguresAndControlsActiveWindow(t *testing.T) {
	application := NewApplication()
	done := application.windows.begin()
	window := new(app.Window)
	_, _ = application.windows.add("main", window, new(windowAppearance))
	application.windows.finishStarting()

	if application.Configure("missing", Title("Missing")) {
		t.Fatal("configured a missing window")
	}
	if !application.Configure("main", Title("Workspace"), Size(800, 600), TopMost(true), Decorated(false)) {
		t.Fatal("active window was not configured")
	}
	if !application.Perform("main", WindowActionCenter) || !application.Perform("main", WindowActionMaximize) {
		t.Fatal("active window action was not performed")
	}
	if application.Perform("main", 0) || application.Perform("missing", WindowActionRaise) {
		t.Fatal("invalid window action succeeded")
	}

	application.windows.deactivate("main", window)
	application.windows.complete(false)
	if code := <-done; code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestApplicationChangesRuntimeAppearance(t *testing.T) {
	application := NewApplication()
	done := application.windows.begin()
	window := new(app.Window)
	appearance := new(windowAppearance)
	_, _ = application.windows.add("main", window, appearance)
	application.windows.finishStarting()

	activeTheme := DarkTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	wantAccent := activeTheme.Palette.Accent
	if !application.SetTheme("main", activeTheme) || !application.SetLanguage("main", LanguageChinese) {
		t.Fatal("active window appearance was not changed")
	}
	activeTheme.Palette.Accent = color.NRGBA{}
	if application.SetTheme("missing", DefaultTheme()) || application.SetLanguage("missing", LanguageEnglish) {
		t.Fatal("missing window appearance was changed")
	}

	ctx := frame.New(nil, nil, LanguageEnglish)
	appearance.apply(ctx)
	if got := ctx.Theme(); got.Palette.Accent != wantAccent || got.Material.Palette.ContrastBg != wantAccent {
		t.Fatalf("runtime theme = accent %#v material %#v", got.Palette.Accent, got.Material.Palette.ContrastBg)
	}
	if got := ctx.Language(); got != LanguageChinese {
		t.Fatalf("runtime language = %q, want %q", got, LanguageChinese)
	}

	application.windows.deactivate("main", window)
	if application.SetTheme("main", DefaultTheme()) || application.SetLanguage("main", LanguageEnglish) {
		t.Fatal("closed window appearance was changed")
	}
	application.windows.complete(false)
	if code := <-done; code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestApplicationReportsLatestWindowState(t *testing.T) {
	application := NewApplication()
	done := application.windows.begin()
	window := new(app.Window)
	_, _ = application.windows.add("main", window, new(windowAppearance))
	application.windows.finishStarting()
	if _, ok := application.WindowState("main"); ok {
		t.Fatal("window state was available before the first config event")
	}
	want := WindowState{Size: image.Pt(800, 600), Mode: WindowModeMaximized, Focused: true, Decorated: true, TopMost: true}
	application.windows.update("main", window, want)

	if got, ok := application.WindowState("main"); !ok || got != want {
		t.Fatalf("window state = %#v, %v; want %#v, true", got, ok, want)
	}
	application.windows.update("main", new(app.Window), WindowState{})
	if got, _ := application.WindowState("main"); got != want {
		t.Fatalf("stale window replaced state: %#v", got)
	}
	application.windows.deactivate("main", window)
	if _, ok := application.WindowState("main"); ok {
		t.Fatal("closed window retained state")
	}
	application.windows.complete(false)
	if code := <-done; code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}
