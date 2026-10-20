package ui

import (
	"github.com/qianniancn/flowui/internal/core"
)

const customStateSlot = "custom"

// UseState retains transient custom-widget state while key is rendered.
// Application and business state still belongs in the MVU Model.
// Values with a BeginFrame method are notified at the start of retained frames.
func UseState[T any](ctx *Context, key string) *T {
	return UseStateWith[T](ctx, key, nil)
}

// UseStateWith retains transient custom-widget state and initializes it with factory.
func UseStateWith[T any](ctx *Context, key string, factory func() *T) *T {
	key = core.ClaimKey(ctx, core.KindCustom, key)
	return core.UseStateWith(ctx, key, customStateSlot, factory)
}
