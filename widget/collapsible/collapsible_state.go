package collapsible

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotCollapsible = "collapsible"

type collapsibleItemState struct {
	clickable  widget.Clickable
	keyFilters core.KeyFilterCache
	disclosure disclosure.Binding[bool]
	expanded   bool // cached effective expanded, updated by isExpanded/requestExpanded
}

type collapsibleState struct {
	item        collapsibleItemState
	items       map[string]*collapsibleItemState
	frameItems  map[string]struct{}
	keyFilters  []event.Filter
	visualCache computed.ProviderCache[struct{}, theme.CollapsibleVisual]
}

// collapsibleDisclosureCfg builds a disclosure.Config from the widget's expanded-state fields.
func collapsibleDisclosureCfg(widget Widget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasExpanded,
		Value:      widget.expanded,
		HasDefault: widget.hasDefaultExpanded,
		Default:    widget.defaultExpanded,
		OnChange:   widget.onExpandedChange,
	}
}

func (s *collapsibleItemState) isExpanded(widget Widget) bool {
	s.expanded = s.disclosure.Current(collapsibleDisclosureCfg(widget))
	return s.expanded
}

func (s *collapsibleItemState) bind(widget Widget) {
	s.disclosure.Bind(collapsibleDisclosureCfg(widget))
}

func (s *collapsibleItemState) requestExpanded(widget Widget, expanded bool) bool {
	s.expanded, _ = s.disclosure.Request(collapsibleDisclosureCfg(widget), expanded)
	return s.expanded
}

func collapsibleStateFor(ctx *core.Context, key string) *collapsibleState {
	key = core.ClaimKey(ctx, core.KindCollapsible, key)
	return core.UseState[collapsibleState](ctx, key, stateSlotCollapsible)
}

func (s *collapsibleState) visual(ctx *core.Context, override *theme.CollapsibleOverride) theme.CollapsibleVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	load := func(request theme.CollapsibleRequest) theme.CollapsibleVisual {
		return theme.ResolveVisual(active, request,
			theme.ComponentVisualProvider.Collapsible, theme.ValidateCollapsibleVisual)
	}
	if override != nil {
		return load(theme.CollapsibleRequest{Override: override})
	}
	return s.visualCache.Resolve(active, struct{}{}, func() theme.CollapsibleVisual {
		return load(theme.CollapsibleRequest{})
	})
}

func (s *collapsibleState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *collapsibleState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *collapsibleState) itemFor(key string) *collapsibleItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *collapsibleState) checkItems(items []Item) {
	validateItems(items)
	for _, item := range items {
		s.itemFor(item.Key)
	}
}

func (s *collapsibleState) updateKeys(gtx layout.Context, items []Item) string {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		if item.Disabled {
			continue
		}
		itemState := s.itemFor(item.Key)
		tag := &itemState.clickable
		s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
			key.NameDownArrow,
			key.NameUpArrow,
			key.NameHome,
			key.NameEnd,
		)...)
	}
	if len(s.keyFilters) == 0 {
		return ""
	}

	current := s.focusedIndex(gtx, items)
	target := -1
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			target = moveEnabledIndex(items, current, 1)
		case key.NameUpArrow:
			target = moveEnabledIndex(items, current, -1)
		case key.NameHome:
			target = firstEnabledIndex(items)
		case key.NameEnd:
			target = lastEnabledIndex(items)
		}
		if target >= 0 {
			current = target
		}
	}
	if target < 0 {
		return ""
	}
	return items[target].Key
}

func (s *collapsibleState) focusedIndex(gtx layout.Context, items []Item) int {
	for index, item := range items {
		if !item.Disabled && gtx.Focused(&s.itemFor(item.Key).clickable) {
			return index
		}
	}
	return -1
}

func moveEnabledIndex(items []Item, current, delta int) int {
	if len(items) == 0 {
		return -1
	}
	if current < 0 || current >= len(items) {
		if delta < 0 {
			return lastEnabledIndex(items)
		}
		return firstEnabledIndex(items)
	}
	index := current
	for range len(items) {
		index = (index + delta + len(items)) % len(items)
		if !items[index].Disabled {
			return index
		}
	}
	return -1
}

func firstEnabledIndex(items []Item) int {
	for index, item := range items {
		if !item.Disabled {
			return index
		}
	}
	return -1
}

func lastEnabledIndex(items []Item) int {
	for index := len(items) - 1; index >= 0; index-- {
		if !items[index].Disabled {
			return index
		}
	}
	return -1
}

func activePresses(state *collapsibleItemState) int {
	return core.ActivePresses(state.clickable.History())
}

func focusOnPress(ctx *core.Context, state *collapsibleItemState, before int) {
	core.FocusOnPress(ctx, &state.clickable, state.clickable.History(), before)
}
