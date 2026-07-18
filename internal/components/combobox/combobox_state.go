package combobox

import (
	"fmt"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotComboBox = "combobox"

func comboBoxStateFor(ctx *frame.Context, key string) *comboBoxState {
	key = frame.ClaimKey(ctx, state.KindComboBox, key)
	return frame.UseState[comboBoxState](ctx, key, stateSlotComboBox)
}

type comboBoxState struct {
	editor             widget.Editor
	input              field.State
	trigger            widget.Clickable
	dialog             overlay.ClickArea
	list               layout.List
	bar                widget.Scrollbar
	open               bool
	wasFocused         bool
	highlight          int
	syncedSelectedKey  string
	syncedInputValue   string
	ignoreEditorChange bool
	popoverTransition  animation.FloatTransition
	iconTransition     animation.FloatTransition
	items              map[string]*comboBoxItemState
	frameItems         map[string]struct{}
	itemKeys           map[string]struct{}
	itemLabels         map[string]string
	dataVersion        uint64
	dataReady          bool
	visibleQuery       string
	visibleSelected    string
	cachedVisibleItems []int
	visibleReady       bool
}

func (s *comboBoxState) beginFrame() {
	state.BeginFrameMap(&s.frameItems)
}

func (s *comboBoxState) endFrame() {
	state.SweepFrameMap(s.items, s.frameItems)
}

func (s *comboBoxState) item(key string) *comboBoxItemState {
	return state.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *comboBoxState) checkItems(items []ComboBoxItem, versioned bool, version uint64) {
	if versioned && s.dataReady && s.dataVersion == version {
		return
	}
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	if s.itemLabels == nil {
		s.itemLabels = make(map[string]string, len(items))
	} else {
		clear(s.itemLabels)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty combobox item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate combobox item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
		s.itemLabels[item.Key] = item.Label
	}
	s.dataReady = versioned
	s.dataVersion = version
	s.visibleReady = false
}

func (s *comboBoxState) selectedLabel(widget ComboBoxWidget) (string, bool) {
	if widget.hasDataVersion {
		label, ok := s.itemLabels[widget.selectedKey]
		return label, ok
	}
	return widget.selectedLabel()
}

func (s *comboBoxState) visibleItems(widget ComboBoxWidget, query, selectedLabel string) []int {
	if widget.hasDataVersion && s.visibleReady && s.visibleQuery == query && s.visibleSelected == selectedLabel {
		return s.cachedVisibleItems
	}
	visible := comboBoxVisibleItems(widget.items, query, selectedLabel)
	if widget.hasDataVersion {
		s.visibleQuery = query
		s.visibleSelected = selectedLabel
		s.cachedVisibleItems = visible
		s.visibleReady = true
	}
	return visible
}

func (s *comboBoxState) syncEditor(editor *widget.Editor, c ComboBoxWidget) {
	if c.hasInputValue {
		if c.inputValue != s.syncedInputValue {
			s.setEditorText(editor, c.inputValue)
			s.syncedInputValue = c.inputValue
		}
		return
	}
	if label, ok := s.selectedLabel(c); ok {
		if c.selectedKey != s.syncedSelectedKey || label != s.syncedInputValue {
			s.setEditorText(editor, label)
			s.syncedInputValue = label
		}
	} else if !c.allowCustomValue {
		if c.selectedKey != s.syncedSelectedKey || s.syncedInputValue != "" {
			s.setEditorText(editor, "")
			s.syncedInputValue = ""
		}
	}
	s.syncedSelectedKey = c.selectedKey
}

func (s *comboBoxState) setEditorText(editor *widget.Editor, text string) {
	if editor.Text() == text {
		return
	}
	editor.SetText(text)
	s.ignoreEditorChange = true
}

func (s *comboBoxState) consumeEditorChange() bool {
	if !s.ignoreEditorChange {
		return false
	}
	s.ignoreEditorChange = false
	return true
}

func (s *comboBoxState) updateFocus(focused, disabled bool) {
	if disabled {
		s.open = false
		s.wasFocused = false
		return
	}
	if focused && !s.wasFocused {
		s.open = true
	}
	if !focused {
		s.open = false
	}
	s.wasFocused = focused
}

func (s *comboBoxState) updateKeys(gtx layout.Context, editor *widget.Editor, items []ComboBoxItem, visible []int) (int, bool) {
	if !s.wasFocused {
		return 0, false
	}
	filters := []event.Filter{
		key.Filter{Focus: editor, Name: key.NameDownArrow},
		key.Filter{Focus: editor, Name: key.NameUpArrow},
	}
	if s.open {
		filters = append(filters,
			key.Filter{Focus: editor, Name: key.NameReturn},
			key.Filter{Focus: editor, Name: key.NameEnter},
			key.Filter{Focus: editor, Name: key.NameEscape},
		)
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return 0, false
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			if !s.open {
				s.open = true
				s.highlight = comboBoxFirstEnabled(items, visible)
				continue
			}
			s.highlight = comboBoxMoveHighlight(items, visible, s.highlight, 1)
		case key.NameUpArrow:
			if !s.open {
				s.open = true
				s.highlight = comboBoxLastEnabled(items, visible)
				continue
			}
			s.highlight = comboBoxMoveHighlight(items, visible, s.highlight, -1)
		case key.NameReturn, key.NameEnter:
			if s.highlight >= 0 && s.highlight < len(visible) && !items[visible[s.highlight]].Disabled {
				return s.highlight, true
			}
		case key.NameEscape:
			s.open = false
		}
	}
}

func comboBoxFirstEnabled(items []ComboBoxItem, visible []int) int {
	for i, index := range visible {
		if !items[index].Disabled {
			return i
		}
	}
	return -1
}

func comboBoxLastEnabled(items []ComboBoxItem, visible []int) int {
	for i := len(visible) - 1; i >= 0; i-- {
		if !items[visible[i]].Disabled {
			return i
		}
	}
	return -1
}

func comboBoxMoveHighlight(items []ComboBoxItem, visible []int, current, delta int) int {
	count := len(visible)
	if count == 0 {
		return -1
	}
	if current < 0 || current >= count || items[visible[current]].Disabled {
		if delta < 0 {
			return comboBoxLastEnabled(items, visible)
		}
		return comboBoxFirstEnabled(items, visible)
	}
	next := current
	for range count {
		next = (next + delta + count) % count
		if !items[visible[next]].Disabled {
			return next
		}
	}
	return current
}

func (s *comboBoxState) clampHighlight(items []ComboBoxItem, visible []int) {
	if len(visible) == 0 {
		s.highlight = -1
		return
	}
	if s.highlight < 0 {
		return
	}
	if s.highlight >= 0 && s.highlight < len(visible) && !items[visible[s.highlight]].Disabled {
		return
	}
	s.highlight = comboBoxFirstEnabled(items, visible)
}

func (s *comboBoxState) popoverProgress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return s.popoverTransition.Value(gtx, target, comboBoxAnimationDuration, animation.EaseSmoothstep, motions...)
}

func (s *comboBoxState) iconProgress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return s.iconTransition.Value(gtx, target, comboBoxAnimationDuration, animation.EaseSmoothstep, motions...)
}

type comboBoxItemState = optionrow.State
