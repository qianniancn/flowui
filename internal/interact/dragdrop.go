package interact

import (
	"bytes"
	"image"
	"io"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

// DefaultMaxDropData limits a drop payload to 64 KiB unless a component needs
// a different bound.
const DefaultMaxDropData = 64 * 1024

// DragState owns the Gio state needed for one drag source.
type DragState struct {
	draggable widget.Draggable
	press     f32.Point
	tag       byte
}

// Layout draws content and its drag preview through Gio's draggable widget.
func (s *DragState) Layout(gtx layout.Context, content, preview layout.Widget) layout.Dimensions {
	return s.draggable.Layout(gtx, content, preview)
}

// RegisterSource registers bounds where a primary press may begin dragging.
func (s *DragState) RegisterSource(gtx layout.Context, bounds image.Rectangle) {
	if bounds.Empty() {
		return
	}
	area := clip.Rect(bounds).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &s.tag)
	pass.Pop()
	area.Pop()
}

// Update offers payload while dragging and reports whether the drag exceeds
// slop. payload is copied only when Gio requests transfer data.
func (s *DragState) Update(gtx layout.Context, mime string, payload []byte, slop float32) bool {
	for {
		eventValue, ok := NextPointerEvent(gtx, &s.tag, pointer.Press)
		if !ok {
			break
		}
		if IsPrimaryPointerPress(eventValue) {
			s.press = eventValue.Position
		}
	}
	if mime == "" {
		return false
	}
	s.draggable.Type = mime
	if requestedMIME, requested := s.draggable.Update(gtx); requested {
		data := bytes.Clone(payload)
		s.draggable.Offer(gtx, requestedMIME, io.NopCloser(bytes.NewReader(data)))
	}
	position := s.draggable.Pos()
	return s.draggable.Dragging() && position.X*position.X+position.Y*position.Y > slop*slop
}

// Press reports the most recent primary press position in source coordinates.
func (s *DragState) Press() f32.Point {
	return s.press
}

// Position reports the current drag offset in source coordinates.
func (s *DragState) Position() f32.Point {
	return s.draggable.Pos()
}

// DropTarget owns one registered drop area.
type DropTarget struct {
	tag byte
}

// Register registers bounds as a transfer target.
func (t *DropTarget) Register(gtx layout.Context, bounds image.Rectangle) {
	if bounds.Empty() {
		return
	}
	area := clip.Rect(bounds).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &t.tag)
	pass.Pop()
	area.Pop()
}

// Next returns the next valid payload offered for mime. Invalid or oversized
// transfer events are discarded until no matching event remains.
func (t *DropTarget) Next(gtx layout.Context, mime string, maxBytes int) ([]byte, bool) {
	if mime == "" || maxBytes <= 0 {
		return nil, false
	}
	for {
		raw, ok := gtx.Event(transfer.TargetFilter{Target: &t.tag, Type: mime})
		if !ok {
			return nil, false
		}
		eventValue, ok := raw.(transfer.DataEvent)
		if !ok {
			continue
		}
		data, ok := ReadDropData(eventValue, maxBytes)
		if ok {
			return data, true
		}
	}
}

// ReadDropData reads one transfer payload with a strict size limit.
func ReadDropData(eventValue transfer.DataEvent, maxBytes int) ([]byte, bool) {
	if eventValue.Open == nil || maxBytes <= 0 {
		return nil, false
	}
	reader := eventValue.Open()
	if reader == nil {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(reader, int64(maxBytes)+1))
	closeErr := reader.Close()
	if err != nil || closeErr != nil || len(data) > maxBytes {
		return nil, false
	}
	return data, true
}
