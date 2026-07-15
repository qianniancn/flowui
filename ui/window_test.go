package ui

import (
	"testing"

	"gioui.org/app"
)

func TestWindowSpecUsesIndependentIdentity(t *testing.T) {
	spec := NewWindow("details", 0, func(*int, int) {}, func(*Context, int, Send[int]) Widget { return Text("Details") }, Title("Details"))
	if spec.Key() != "details" || spec.run == nil || len(spec.options) != 1 {
		t.Fatalf("window spec = key %q run %v options %d", spec.Key(), spec.run != nil, len(spec.options))
	}
}

func TestWindowSpecRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name string
		make func()
	}{
		{"empty key", func() { NewWindow("", 0, func(*int, int) {}, func(*Context, int, Send[int]) Widget { return nil }) }},
		{"nil update", func() { NewWindow[int, int]("main", 0, nil, func(*Context, int, Send[int]) Widget { return nil }) }},
		{"nil view", func() { NewWindow[int, int]("main", 0, func(*int, int) {}, nil) }},
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
	if existing, added := windows.add("first", first); existing != nil || !added {
		t.Fatalf("first add = existing %p added %v", existing, added)
	}
	if existing, added := windows.add("second", second); existing != nil || !added {
		t.Fatalf("second add = existing %p added %v", existing, added)
	}
	if existing, added := windows.add("first", new(app.Window)); existing != first || added {
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
	_, _ = windows.add("main", window)
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
	_, _ = windows.add("details", first)
	windows.finishStarting()
	windows.deactivate("details", first)

	second := new(app.Window)
	if existing, added := windows.add("details", second); existing != nil || !added {
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
