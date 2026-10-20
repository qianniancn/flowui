package togglebutton

import "github.com/qianniancn/flowui/internal/core"

const stateSlotToggleButtonVisual = "togglebutton-visual"

type toggleButtonState struct {
	visualCache VisualCache
}

func toggleButtonStateFor(ctx *core.Context, key string) *toggleButtonState {
	return core.UseState[toggleButtonState](ctx, key, stateSlotToggleButtonVisual)
}
