package explorer

import (
	"context"
	"errors"
	"fmt"
	"io"

	"gioui.org/app"
	"gioui.org/io/event"
	gioexplorer "gioui.org/x/explorer"
)

var (
	ErrCanceled    = errors.New("flowui/explorer: dialog canceled")
	ErrUnavailable = errors.New("flowui/explorer: unavailable")
)

type backend interface {
	ListenEvents(event.Event)
	ChooseFile(...string) (io.ReadCloser, error)
	ChooseFiles(...string) ([]io.ReadCloser, error)
	CreateFile(string) (io.WriteCloser, error)
}

// Service owns the native file-dialog integration for one Gio window.
type Service struct {
	backend backend
}

func New(window *app.Window) *Service {
	return &Service{backend: gioexplorer.NewExplorer(window)}
}

func newService(value backend) *Service {
	return &Service{backend: value}
}

// ListenEvents forwards native view events required by some platforms.
func (s *Service) ListenEvents(value event.Event) {
	if s == nil || s.backend == nil {
		return
	}
	s.backend.ListenEvents(value)
}

func (s *Service) ChooseFile(ctx context.Context, extensions ...string) (io.ReadCloser, error) {
	if s == nil || s.backend == nil {
		return nil, ErrUnavailable
	}
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	reader, err := s.backend.ChooseFile(cloneStrings(extensions)...)
	if err != nil {
		return nil, normalizeError("choose file", err)
	}
	if reader == nil {
		return nil, errors.New("flowui/explorer: choose file returned a nil reader")
	}
	if err := contextError(ctx); err != nil {
		_ = reader.Close()
		return nil, err
	}
	return reader, nil
}

func (s *Service) ChooseFiles(ctx context.Context, extensions ...string) ([]io.ReadCloser, error) {
	if s == nil || s.backend == nil {
		return nil, ErrUnavailable
	}
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	readers, err := s.backend.ChooseFiles(cloneStrings(extensions)...)
	if err != nil {
		closeReaders(readers)
		return nil, normalizeError("choose files", err)
	}
	for _, reader := range readers {
		if reader == nil {
			closeReaders(readers)
			return nil, errors.New("flowui/explorer: choose files returned a nil reader")
		}
	}
	if err := contextError(ctx); err != nil {
		closeReaders(readers)
		return nil, err
	}
	return readers, nil
}

func (s *Service) CreateFile(ctx context.Context, name string) (io.WriteCloser, error) {
	if s == nil || s.backend == nil {
		return nil, ErrUnavailable
	}
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	writer, err := s.backend.CreateFile(name)
	if err != nil {
		return nil, normalizeError("create file", err)
	}
	if writer == nil {
		return nil, errors.New("flowui/explorer: create file returned a nil writer")
	}
	if err := contextError(ctx); err != nil {
		_ = writer.Close()
		return nil, err
	}
	return writer, nil
}

type serviceContextKey struct{}

func WithService(ctx context.Context, service *Service) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == nil {
		return ctx
	}
	return context.WithValue(ctx, serviceContextKey{}, service)
}

func FromContext(ctx context.Context) *Service {
	if ctx == nil {
		return nil
	}
	service, _ := ctx.Value(serviceContextKey{}).(*Service)
	return service
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func normalizeError(operation string, err error) error {
	switch {
	case errors.Is(err, gioexplorer.ErrUserDecline):
		return ErrCanceled
	case errors.Is(err, gioexplorer.ErrNotAvailable):
		return ErrUnavailable
	default:
		return fmt.Errorf("flowui/explorer: %s: %w", operation, err)
	}
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}

func closeReaders(readers []io.ReadCloser) {
	for _, reader := range readers {
		if reader != nil {
			_ = reader.Close()
		}
	}
}
