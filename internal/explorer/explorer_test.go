package explorer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"testing"

	"gioui.org/app"
	"gioui.org/io/event"
	gioexplorer "gioui.org/x/explorer"
)

type fakeBackend struct {
	events           []event.Event
	chooseExtensions []string
	chooseFile       func() (io.ReadCloser, error)
	chooseFiles      func() ([]io.ReadCloser, error)
	createFile       func(string) (io.WriteCloser, error)
	chooseCalls      int
}

func (f *fakeBackend) ListenEvents(value event.Event) {
	f.events = append(f.events, value)
}

func (f *fakeBackend) ChooseFile(extensions ...string) (io.ReadCloser, error) {
	f.chooseCalls++
	f.chooseExtensions = append([]string(nil), extensions...)
	if len(extensions) > 0 {
		extensions[0] = "mutated"
	}
	return f.chooseFile()
}

func (f *fakeBackend) ChooseFiles(extensions ...string) ([]io.ReadCloser, error) {
	f.chooseExtensions = append([]string(nil), extensions...)
	return f.chooseFiles()
}

func (f *fakeBackend) CreateFile(name string) (io.WriteCloser, error) {
	return f.createFile(name)
}

type trackedReader struct {
	*bytes.Reader
	closed bool
}

func newTrackedReader(value string) *trackedReader {
	return &trackedReader{Reader: bytes.NewReader([]byte(value))}
}

func (r *trackedReader) Close() error {
	r.closed = true
	return nil
}

type bufferWriteCloser struct {
	bytes.Buffer
	closed bool
}

func (w *bufferWriteCloser) Close() error {
	w.closed = true
	return nil
}

func TestChooseFileOwnsExtensionsAndForwardsReader(t *testing.T) {
	reader := newTrackedReader("data")
	backend := &fakeBackend{chooseFile: func() (io.ReadCloser, error) { return reader, nil }}
	service := newService(backend)
	extensions := []string{".json", ".txt"}

	got, err := service.ChooseFile(context.Background(), extensions...)
	if err != nil {
		t.Fatal(err)
	}
	if got != reader {
		t.Fatalf("reader = %T, want original reader", got)
	}
	if !slices.Equal(backend.chooseExtensions, extensions) {
		t.Fatalf("extensions = %v, want %v", backend.chooseExtensions, extensions)
	}
	if extensions[0] != ".json" {
		t.Fatalf("caller extensions were mutated: %v", extensions)
	}
}

func TestChooseFileMapsDialogErrors(t *testing.T) {
	tests := []struct {
		name string
		in   error
		want error
	}{
		{name: "canceled", in: gioexplorer.ErrUserDecline, want: ErrCanceled},
		{name: "unavailable", in: gioexplorer.ErrNotAvailable, want: ErrUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := &fakeBackend{chooseFile: func() (io.ReadCloser, error) { return nil, test.in }}
			_, err := newService(backend).ChooseFile(context.Background())
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestChooseFileHonorsCanceledContext(t *testing.T) {
	backend := &fakeBackend{chooseFile: func() (io.ReadCloser, error) {
		t.Fatal("backend called for canceled context")
		return nil, nil
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := newService(backend).ChooseFile(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
	if backend.chooseCalls != 0 {
		t.Fatalf("backend calls = %d, want 0", backend.chooseCalls)
	}
}

func TestChooseFileClosesResultCanceledWhileDialogWasOpen(t *testing.T) {
	reader := newTrackedReader("data")
	ctx, cancel := context.WithCancel(context.Background())
	backend := &fakeBackend{chooseFile: func() (io.ReadCloser, error) {
		cancel()
		return reader, nil
	}}

	_, err := newService(backend).ChooseFile(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
	if !reader.closed {
		t.Fatal("reader was not closed after context cancellation")
	}
}

func TestChooseFilesClosesPartialResultsOnError(t *testing.T) {
	first := newTrackedReader("first")
	second := newTrackedReader("second")
	want := errors.New("dialog failed")
	backend := &fakeBackend{chooseFiles: func() ([]io.ReadCloser, error) {
		return []io.ReadCloser{first, second}, want
	}}

	_, err := newService(backend).ChooseFiles(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want wrapped backend error", err)
	}
	if !first.closed || !second.closed {
		t.Fatalf("partial readers closed = %v/%v, want true/true", first.closed, second.closed)
	}
}

func TestCreateFileAndEventForwarding(t *testing.T) {
	writer := new(bufferWriteCloser)
	var gotName string
	backend := &fakeBackend{
		createFile: func(name string) (io.WriteCloser, error) {
			gotName = name
			return writer, nil
		},
	}
	service := newService(backend)
	value := app.ConfigEvent{}
	service.ListenEvents(value)

	got, err := service.CreateFile(context.Background(), "report.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != writer || gotName != "report.txt" {
		t.Fatalf("create result = %T/%q, want writer/report.txt", got, gotName)
	}
	if len(backend.events) != 1 {
		t.Fatalf("forwarded events = %d, want 1", len(backend.events))
	}
}

func TestServiceContextRoundTrip(t *testing.T) {
	service := newService(&fakeBackend{})
	ctx := WithService(context.Background(), service)
	if got := FromContext(ctx); got != service {
		t.Fatalf("context service = %p, want %p", got, service)
	}
	if FromContext(nil) != nil {
		t.Fatal("nil context returned a service")
	}
}

func TestWithServiceDoesNotWrapContextForNilService(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	if got := WithService(ctx, nil); got != ctx {
		t.Fatal("nil service wrapped the context")
	}
}
