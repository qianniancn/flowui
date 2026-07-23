package input

import (
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type inputState struct {
	field.State
}

func inputStateFor(ctx *frame.Context, key string) *inputState {
	return frame.UseState[inputState](ctx, key, stateSlotInput)
}
