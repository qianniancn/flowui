// Package notify provides native desktop notifications through a stable
// FlowUI API. It intentionally exposes only capabilities shared well enough
// across supported platforms; notification actions and click callbacks are not
// currently portable.
package notify

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	gionotify "gioui.org/x/notify"
)

var (
	// ErrUnavailable reports that native notifications are unavailable on the
	// current platform or desktop session.
	ErrUnavailable = errors.New("flowui/notify: unavailable")
	// ErrUnsupported reports that a specific notification operation is not
	// implemented by the current platform backend.
	ErrUnsupported = errors.New("flowui/notify: operation unsupported")
)

type backend interface {
	CreateNotification(title, text string) (gionotify.Notification, error)
}

type iconBackend interface {
	UseIcon(path string)
}

// Notifier creates native notifications. A Notifier may be reused and is safe
// for concurrent use.
type Notifier struct {
	mu      sync.Mutex
	backend backend
}

// New initializes the notification backend for the current platform.
func New() (*Notifier, error) {
	if !platformSupported() {
		return nil, ErrUnavailable
	}
	value, err := gionotify.NewNotifier()
	if err != nil {
		return nil, unavailableError("initialize", err)
	}
	if value == nil {
		return nil, ErrUnavailable
	}
	return newNotifier(value), nil
}

func newNotifier(value backend) *Notifier {
	return &Notifier{backend: value}
}

// Push sends a native notification.
func (n *Notifier) Push(title, text string) (*Notification, error) {
	if n == nil || n.backend == nil {
		return nil, ErrUnavailable
	}
	n.mu.Lock()
	handle, err := n.backend.CreateNotification(title, text)
	n.mu.Unlock()
	if err != nil {
		return nil, unavailableError("push", err)
	}
	if handle == nil {
		return nil, errors.New("flowui/notify: backend returned a nil notification")
	}
	return newNotification(handle, platformCanCancel()), nil
}

// SupportsIcon reports whether SetIcon is implemented by the current backend.
func (n *Notifier) SupportsIcon() bool {
	if n == nil {
		return false
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	_, ok := n.backend.(iconBackend)
	return ok
}

// SetIcon configures the image used by future notifications. The value is a
// filesystem path. It currently applies to the Windows backend.
func (n *Notifier) SetIcon(path string) error {
	if n == nil || n.backend == nil {
		return ErrUnavailable
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	value, ok := n.backend.(iconBackend)
	if !ok {
		return ErrUnsupported
	}
	value.UseIcon(path)
	return nil
}

// Notification is a handle to a delivered native notification.
type Notification struct {
	mu         sync.Mutex
	handle     gionotify.Notification
	cancelable bool
	canceled   bool
}

func newNotification(handle gionotify.Notification, cancelable bool) *Notification {
	return &Notification{handle: handle, cancelable: cancelable}
}

// Cancelable reports whether the current platform can remove this
// notification after delivery.
func (n *Notification) Cancelable() bool {
	if n == nil {
		return false
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.cancelable
}

// Cancel removes the notification when supported. It is idempotent after a
// successful cancellation.
func (n *Notification) Cancel() error {
	if n == nil {
		return nil
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.canceled {
		return nil
	}
	if !n.cancelable {
		return ErrUnsupported
	}
	if n.handle == nil {
		return ErrUnavailable
	}
	if err := n.handle.Cancel(); err != nil {
		return fmt.Errorf("flowui/notify: cancel: %w", err)
	}
	n.canceled = true
	return nil
}

var defaultNotifier struct {
	sync.Mutex
	value *Notifier
}

// Push sends a notification through a lazily initialized package notifier.
// Initialization errors are not cached, so a temporarily unavailable desktop
// notification service can recover on a later call.
func Push(title, text string) (*Notification, error) {
	defaultNotifier.Lock()
	value := defaultNotifier.value
	if value == nil {
		var err error
		value, err = New()
		if err != nil {
			defaultNotifier.Unlock()
			return nil, err
		}
		defaultNotifier.value = value
	}
	defaultNotifier.Unlock()
	return value.Push(title, text)
}

func platformSupported() bool {
	switch runtime.GOOS {
	case "android", "darwin", "freebsd", "linux", "netbsd", "openbsd", "windows":
		return true
	default:
		return false
	}
}

func platformCanCancel() bool {
	return runtime.GOOS != "windows"
}

func unavailableError(operation string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrUnavailable, operation, err)
}
