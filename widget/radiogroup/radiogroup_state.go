package radiogroup

import (
	"fmt"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
	"github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/optionrow"
)

const stateSlotRadioGroup = "radio-group"

const radioSelectDuration = 200 * time.Millisecond

func radioGroupStateFor(ctx *core.Context, key string) *radioGroupState {
	key = core.ClaimKey(ctx, core.KindRadioGroup, key)
	return core.UseState[radioGroupState](ctx, key, stateSlotRadioGroup)
}

type radioGroupState struct {
	disclosure  disclosure.Binding[string]
	selectedKey string
	items       map[string]*radioItemState
	frameItems  map[string]struct{}
	itemKeys    map[string]struct{}
	keyFilters  []event.Filter
	visualCache VisualCache
}

// radioGroupDisclosureCfg builds a disclosure.Config from the widget's selected-key fields.
func radioGroupDisclosureCfg(widget Widget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultSelected,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *radioGroupState) currentSelectedKey(widget Widget) string {
	s.selectedKey = s.disclosure.Current(radioGroupDisclosureCfg(widget))
	return s.selectedKey
}

func (s *radioGroupState) bind(widget Widget) {
	s.disclosure.Bind(radioGroupDisclosureCfg(widget))
}

func (s *radioGroupState) requestSelectedKey(widget Widget, key string) string {
	s.selectedKey, _ = s.disclosure.Request(radioGroupDisclosureCfg(widget), key)
	return s.selectedKey
}

func (s *radioGroupState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *radioGroupState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *radioGroupState) item(key string) *radioItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *radioGroupState) visual(ctx *core.Context, variant Variant, override *theme.RadioGroupOverride) theme.RadioGroupVisual {
	active := core.ActiveTheme(ctx)
	return radioVisualCached(active, variant, override, &s.visualCache)
}

func (s *radioGroupState) checkItems(items []RadioItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty radio item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate radio item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
	}
}

func (s *radioGroupState) updateKeys(gtx layout.Context, items []RadioItem, selectedKey string) (string, bool) {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		if item.Disabled {
			continue
		}
		itemState := s.item(item.Key)
		tag := &itemState.clickable
		s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
			key.NameRightArrow,
			key.NameDownArrow,
			key.NameLeftArrow,
			key.NameUpArrow,
			key.NameHome,
			key.NameEnd,
		)...)
	}
	if len(s.keyFilters) == 0 {
		return "", false
	}

	target := selectedKey
	current := s.focusedIndex(gtx, items)
	if current < 0 {
		current = radioIndexByKey(items, target)
	}
	// Radio groups wrap arrow navigation (ARIA convention); no typeahead.
	list := radioNavList(items)
	changed := false
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
		case key.NameRightArrow, key.NameDownArrow:
			if next, ok := navigation.Move(list, current, 1, true); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		case key.NameLeftArrow, key.NameUpArrow:
			if next, ok := navigation.Move(list, current, -1, true); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		case key.NameHome:
			if next, ok := navigation.First(list); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		case key.NameEnd:
			if next, ok := navigation.Last(list); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		}
	}
	return target, changed
}

func radioNavList(items []RadioItem) navigation.List {
	return navigation.List{
		Count:    len(items),
		Disabled: func(i int) bool { return items[i].Disabled },
	}
}

func (s *radioGroupState) focusedIndex(gtx layout.Context, items []RadioItem) int {
	for i, item := range items {
		if item.Disabled {
			continue
		}
		if gtx.Focused(&s.item(item.Key).clickable) {
			return i
		}
	}
	return -1
}

func radioIndexByKey(items []RadioItem, key string) int {
	for i, item := range items {
		if item.Key == key {
			return i
		}
	}
	return -1
}

type radioItemState struct {
	clickable  widget.Clickable
	selected   animation.FloatTransition
	focus      animation.FocusAnimation
	keyFilters core.KeyFilterCache
}

func (s *radioItemState) selection(gtx layout.Context, selected bool, motions ...theme.MotionTheme) float32 {
	return s.selectionFor(gtx, selected, radioSelectDuration, motions...)
}

func (s *radioItemState) selectionFor(gtx layout.Context, selected bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selected.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *radioItemState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}

func radioPressScale(gtx layout.Context, history []widget.Press, visual theme.RadioGroupVisual, disabled bool) float32 {
	target := visual.Metrics.PressedScale
	if target <= 0 || target > 1 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return optionrow.PressScale(gtx, history, disabled, target, visual.PressInDuration, visual.PressOutDuration, visual.Motion)
}
