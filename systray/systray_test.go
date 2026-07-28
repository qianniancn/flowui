package systray

import (
	"bytes"
	"errors"
	"testing"
)

func TestSetIconOwnsBytes(t *testing.T) {
	tray := New()
	icon := []byte{1, 2, 3}
	tray.SetIcon(icon)
	icon[0] = 9

	tray.mu.Lock()
	got := append([]byte(nil), tray.icon...)
	tray.mu.Unlock()
	if !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Fatalf("stored icon = %v, want owned copy", got)
	}
	tray.Destroy()
}

func TestMenuItemsReturnsSnapshot(t *testing.T) {
	menu := NewMenu()
	first := menu.Add("First")
	menu.Add("Second")

	items := menu.Items()
	items[0] = nil
	items = append(items, first)

	got := menu.Items()
	if len(got) != 2 || got[0] != first {
		t.Fatalf("menu items were mutated through snapshot: %#v", got)
	}
	menu.Destroy()
}

func TestDestroyedTrayCannotStartLater(t *testing.T) {
	tray := New()
	tray.Destroy()
	tray.Run()

	tray.mu.Lock()
	running := tray.running
	destroyed := tray.destroyed
	impl := tray.impl
	tray.mu.Unlock()
	if running || !destroyed || impl != nil {
		t.Fatalf("destroyed tray = running %v destroyed %v impl %T", running, destroyed, impl)
	}
	if GetTrayByID(tray.ID()) != nil {
		t.Fatal("destroyed tray remained in the global registry")
	}

	// Destroy remains safe for deferred cleanup after an earlier shutdown.
	tray.Destroy()
}

func TestTrayLifecycleReadyAndDone(t *testing.T) {
	tray := New()
	tray.mu.Lock()
	tray.running = true
	tray.mu.Unlock()
	if !tray.markReady() {
		t.Fatal("active tray was not marked ready")
	}
	select {
	case <-tray.Ready():
	default:
		t.Fatal("Ready was not closed")
	}

	tray.Destroy()
	select {
	case <-tray.Done():
	default:
		t.Fatal("Done was not closed")
	}
}

func TestTrayLifecycleFailureReportsError(t *testing.T) {
	tray := New()
	want := errors.New("startup failed")
	tray.fail(want)

	select {
	case got := <-tray.Errors():
		if !errors.Is(got, want) {
			t.Fatalf("Errors returned %v, want %v", got, want)
		}
	default:
		t.Fatal("startup error was not reported")
	}
	select {
	case <-tray.Done():
	default:
		t.Fatal("Done was not closed after startup failure")
	}
	select {
	case <-tray.Ready():
		t.Fatal("Ready closed after startup failure")
	default:
	}
	if GetTrayByID(tray.ID()) != nil {
		t.Fatal("failed tray remained in the global registry")
	}
}
