package listbox

import (
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

type ItemFocusState struct {
	TargetOpacity float32
}

func ItemState(ctx *core.Context, key, itemKey string) (*widget.Clickable, ItemFocusState, bool) {
	return item(ctx, core.FullKey(ctx, key), itemKey)
}

func DerivedItem(ctx *core.Context, owner, role, itemKey string) (*widget.Clickable, ItemFocusState, bool) {
	return item(ctx, core.DerivedKey(ctx, owner, role), itemKey)
}

func item(ctx *core.Context, stateKey, itemKey string) (*widget.Clickable, ItemFocusState, bool) {
	state, ok := core.PeekState[listBoxState](ctx, stateKey, stateSlotListBox)
	if !ok || state.items[itemKey] == nil {
		return nil, ItemFocusState{}, false
	}
	item := state.items[itemKey]
	return &item.Clickable, ItemFocusState{
		TargetOpacity: item.FocusTargetOpacity(),
	}, true
}

func EnsureItem(ctx *core.Context, key, itemKey string) *widget.Clickable {
	return ensureItem(ctx, core.FullKey(ctx, key), itemKey)
}

func EnsureDerivedItem(ctx *core.Context, owner, role, itemKey string) *widget.Clickable {
	return ensureItem(ctx, core.DerivedKey(ctx, owner, role), itemKey)
}

func ensureItem(ctx *core.Context, stateKey, itemKey string) *widget.Clickable {
	state := core.UseStateWith(ctx, stateKey, stateSlotListBox, func() *listBoxState {
		return new(listBoxState)
	})
	return &core.EnsureFrameMap(&state.items, itemKey).Clickable
}

func HasState(ctx *core.Context, key string) bool {
	return hasState(ctx, core.FullKey(ctx, key))
}

func HasDerivedState(ctx *core.Context, owner, role string) bool {
	return hasState(ctx, core.DerivedKey(ctx, owner, role))
}

func hasState(ctx *core.Context, stateKey string) bool {
	_, ok := core.PeekState[listBoxState](ctx, stateKey, stateSlotListBox)
	return ok
}
