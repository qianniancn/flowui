package listbox

import (
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

type ItemFocusState struct {
	TargetOpacity float32
}

func Item(ctx *frame.Context, key, itemKey string) (*widget.Clickable, ItemFocusState, bool) {
	return item(ctx, frame.FullKey(ctx, key), itemKey)
}

func DerivedItem(ctx *frame.Context, owner, role, itemKey string) (*widget.Clickable, ItemFocusState, bool) {
	return item(ctx, frame.DerivedKey(ctx, owner, role), itemKey)
}

func item(ctx *frame.Context, stateKey, itemKey string) (*widget.Clickable, ItemFocusState, bool) {
	state, ok := frame.PeekState[listBoxState](ctx, stateKey, stateSlotListBox)
	if !ok || state.items[itemKey] == nil {
		return nil, ItemFocusState{}, false
	}
	item := state.items[itemKey]
	return &item.Clickable, ItemFocusState{
		TargetOpacity: item.FocusTargetOpacity(),
	}, true
}

func EnsureItem(ctx *frame.Context, key, itemKey string) *widget.Clickable {
	return ensureItem(ctx, frame.FullKey(ctx, key), itemKey)
}

func EnsureDerivedItem(ctx *frame.Context, owner, role, itemKey string) *widget.Clickable {
	return ensureItem(ctx, frame.DerivedKey(ctx, owner, role), itemKey)
}

func ensureItem(ctx *frame.Context, stateKey, itemKey string) *widget.Clickable {
	state := frame.UseStateWith(ctx, stateKey, stateSlotListBox, func() *listBoxState {
		return new(listBoxState)
	})
	return &stateutil.EnsureFrameMap(&state.items, itemKey).Clickable
}

func HasState(ctx *frame.Context, key string) bool {
	return hasState(ctx, frame.FullKey(ctx, key))
}

func HasDerivedState(ctx *frame.Context, owner, role string) bool {
	return hasState(ctx, frame.DerivedKey(ctx, owner, role))
}

func hasState(ctx *frame.Context, stateKey string) bool {
	_, ok := frame.PeekState[listBoxState](ctx, stateKey, stateSlotListBox)
	return ok
}
