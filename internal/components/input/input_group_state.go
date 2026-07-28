package input

import "github.com/qianniancn/flowui/internal/frame"

type inputGroupState struct {
	inputState
}

func inputGroupStateFor(ctx *frame.Context, key string) *inputGroupState {
	return frame.UseState[inputGroupState](ctx, key, stateSlotInputGroup)
}
