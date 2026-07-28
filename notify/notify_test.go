package notify

import (
	"errors"
	"testing"

	gionotify "gioui.org/x/notify"
)

type fakeBackend struct {
	handle gionotify.Notification
	err    error
	title  string
	text   string
}

func (f *fakeBackend) CreateNotification(title, text string) (gionotify.Notification, error) {
	f.title = title
	f.text = text
	return f.handle, f.err
}

type fakeIconBackend struct {
	fakeBackend
	icon string
}

func (f *fakeIconBackend) UseIcon(path string) {
	f.icon = path
}

type fakeNotification struct {
	cancelCalls int
	err         error
}

func (f *fakeNotification) Cancel() error {
	f.cancelCalls++
	return f.err
}

func TestNotifierPushForwardsContent(t *testing.T) {
	handle := new(fakeNotification)
	backend := &fakeBackend{handle: handle}
	notifier := newNotifier(backend)

	got, err := notifier.Push("Build complete", "The application is ready.")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || backend.title != "Build complete" || backend.text != "The application is ready." {
		t.Fatalf("notification = %#v, content = %q/%q", got, backend.title, backend.text)
	}
}

func TestNotifierPushWrapsBackendError(t *testing.T) {
	want := errors.New("desktop service failed")
	notifier := newNotifier(&fakeBackend{err: want})
	_, err := notifier.Push("Title", "Body")
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want wrapped backend error", err)
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
}

func TestNotifierIconCapability(t *testing.T) {
	plain := newNotifier(&fakeBackend{})
	if plain.SupportsIcon() {
		t.Fatal("plain backend unexpectedly supports icons")
	}
	if err := plain.SetIcon("app.png"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("SetIcon error = %v, want ErrUnsupported", err)
	}

	backend := new(fakeIconBackend)
	withIcon := newNotifier(backend)
	if !withIcon.SupportsIcon() {
		t.Fatal("icon backend did not report icon support")
	}
	if err := withIcon.SetIcon("app.png"); err != nil {
		t.Fatal(err)
	}
	if backend.icon != "app.png" {
		t.Fatalf("icon = %q, want app.png", backend.icon)
	}
}

func TestNotificationCancelIsIdempotent(t *testing.T) {
	handle := new(fakeNotification)
	notification := newNotification(handle, true)
	if !notification.Cancelable() {
		t.Fatal("notification should be cancelable")
	}
	if err := notification.Cancel(); err != nil {
		t.Fatal(err)
	}
	if err := notification.Cancel(); err != nil {
		t.Fatal(err)
	}
	if handle.cancelCalls != 1 {
		t.Fatalf("cancel calls = %d, want 1", handle.cancelCalls)
	}
}

func TestNotificationCancelReportsUnsupported(t *testing.T) {
	handle := new(fakeNotification)
	notification := newNotification(handle, false)
	if notification.Cancelable() {
		t.Fatal("notification unexpectedly reports cancellation support")
	}
	if err := notification.Cancel(); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Cancel error = %v, want ErrUnsupported", err)
	}
	if handle.cancelCalls != 0 {
		t.Fatalf("cancel calls = %d, want 0", handle.cancelCalls)
	}
}

func TestNotificationRetriesFailedCancellation(t *testing.T) {
	want := errors.New("cancel failed")
	handle := &fakeNotification{err: want}
	notification := newNotification(handle, true)
	if err := notification.Cancel(); !errors.Is(err, want) {
		t.Fatalf("Cancel error = %v, want wrapped backend error", err)
	}
	handle.err = nil
	if err := notification.Cancel(); err != nil {
		t.Fatal(err)
	}
	if handle.cancelCalls != 2 {
		t.Fatalf("cancel calls = %d, want 2", handle.cancelCalls)
	}
}
