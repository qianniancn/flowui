package interact

import (
	"bytes"
	"io"
	"testing"

	"gioui.org/io/transfer"
)

func TestReadDropDataRespectsBoundsAndClosesReader(t *testing.T) {
	closed := false
	event := transfer.DataEvent{Open: func() io.ReadCloser {
		return closeProbe{Reader: bytes.NewReader([]byte("payload")), closed: &closed}
	}}
	data, ok := ReadDropData(event, 8)
	if !ok || string(data) != "payload" || !closed {
		t.Fatalf("drop data = %q ok %v closed %v", data, ok, closed)
	}
	if _, ok := ReadDropData(transfer.DataEvent{Open: func() io.ReadCloser {
		return io.NopCloser(bytes.NewReader(make([]byte, 9)))
	}}, 8); ok {
		t.Fatal("oversized payload was accepted")
	}
	if _, ok := ReadDropData(transfer.DataEvent{}, 8); ok {
		t.Fatal("missing transfer reader was accepted")
	}
}

type closeProbe struct {
	io.Reader
	closed *bool
}

func (p closeProbe) Close() error {
	*p.closed = true
	return nil
}
