package input

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/internal/frame"
)

type inputGroupState struct {
	inputState
	actionBounds []image.Rectangle
}

func inputGroupStateFor(ctx *frame.Context, key string) *inputGroupState {
	return frame.UseState[inputGroupState](ctx, key, stateSlotInputGroup)
}

func (s *inputGroupState) setActionBounds(bounds ...image.Rectangle) {
	s.actionBounds = s.actionBounds[:0]
	for _, bound := range bounds {
		if !bound.Empty() {
			s.actionBounds = append(s.actionBounds, bound)
		}
	}
}

func (s *inputGroupState) shouldFocusPress(event pointer.Event) bool {
	if event.Kind != pointer.Press {
		return true
	}
	for _, bound := range s.actionBounds {
		if pointInRect(bound, event.Position) {
			return false
		}
	}
	return true
}

func pointInRect(rect image.Rectangle, point f32.Point) bool {
	return point.X >= float32(rect.Min.X) && point.X < float32(rect.Max.X) &&
		point.Y >= float32(rect.Min.Y) && point.Y < float32(rect.Max.Y)
}
