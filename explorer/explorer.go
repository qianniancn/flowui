// Package explorer provides native file dialogs for FlowUI commands.
//
// Dialog functions require the context passed to a ui.Cmd. FlowUI binds that
// context to the command's window, so dialogs remain isolated in multi-window
// applications. Calls made with another context return [ErrUnavailable].
package explorer

import (
	"context"
	"io"

	internalexplorer "github.com/qianniancn/flowui/internal/explorer"
)

var (
	// ErrCanceled reports that the user dismissed a native file dialog.
	ErrCanceled = internalexplorer.ErrCanceled
	// ErrUnavailable reports that no FlowUI window service or native dialog is
	// available on the current platform.
	ErrUnavailable = internalexplorer.ErrUnavailable
)

// ChooseFile opens a native dialog and returns the selected file for reading.
// Extensions may be written as ".json" or "json". The caller must close the
// returned reader.
func ChooseFile(ctx context.Context, extensions ...string) (io.ReadCloser, error) {
	service := internalexplorer.FromContext(ctx)
	if service == nil {
		return nil, ErrUnavailable
	}
	return service.ChooseFile(ctx, extensions...)
}

// ChooseFiles opens a native multi-file dialog. The caller must close every
// returned reader. Platforms without multi-selection return ErrUnavailable.
func ChooseFiles(ctx context.Context, extensions ...string) ([]io.ReadCloser, error) {
	service := internalexplorer.FromContext(ctx)
	if service == nil {
		return nil, ErrUnavailable
	}
	return service.ChooseFiles(ctx, extensions...)
}

// CreateFile opens a native save dialog and returns the selected destination.
// Some platforms commit the file only when the returned writer is closed.
func CreateFile(ctx context.Context, name string) (io.WriteCloser, error) {
	service := internalexplorer.FromContext(ctx)
	if service == nil {
		return nil, ErrUnavailable
	}
	return service.CreateFile(ctx, name)
}
