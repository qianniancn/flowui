package input

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/theme"
)

type inputGroupState struct {
	// The group owns shell hover/focus state. Its editor state is kept by the
	// nested Input or TextArea, so embedding inputState here would duplicate
	// disclosure and editor-cache storage without sharing any behavior.
	field.State
	actionBounds []image.Rectangle
	styleCache   computed.ProviderCacheMap[inputVisualKey, theme.InputVisual]
}

func inputGroupStateFor(ctx *core.Context, key string) *inputGroupState {
	return core.UseState[inputGroupState](ctx, key, stateSlotInputGroup)
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
