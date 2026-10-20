package combobox

import (
	"fmt"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/optionrow"
)

const stateSlotComboBox = "combobox"

func comboBoxStateFor(ctx *core.Context, key string) *comboBoxState {
	key = core.ClaimKey(ctx, core.KindComboBox, key)
	return core.UseState[comboBoxState](ctx, key, stateSlotComboBox)
}

type comboBoxState struct {
	editor             widget.Editor
	input              field.State
	trigger            widget.Clickable
	dialog             core.ClickArea
	list               layout.List
	bar                widget.Scrollbar
	visualOutset       layoutui.VisualOutsetState
	disclosure         disclosure.Binding[string]
	selectedKey        string
	open               bool
	wasFocused         bool
	highlight          int
	syncedSelectedKey  string
	syncedInputValue   string
	ignoreEditorChange bool
	pendingEdit        input.EditAction
	hasPendingEdit     bool
	popoverTransition  animation.FloatTransition
	iconTransition     animation.FloatTransition
	items              map[string]*comboBoxItemState
	frameItems         map[string]struct{}
	itemKeys           map[string]struct{}
	itemLabels         map[string]string
	itemDerivedKeys    map[string]string
	itemDerivedBase    string
	dataVersion        uint64
	dataReady          bool
	visibleQuery       string
	visibleSelected    string
	visibleDataVersion uint64
	cachedVisibleItems []int
	visibleReady       bool
	styleCache         computed.ProviderCache[field.Variant, comboDeclarationBundle]
	styleKeyOwner      string
	stylePartKeys      [5]string
}

// comboBoxDisclosureCfg builds a disclosure.Config from the widget's selected-key fields.
func comboBoxDisclosureCfg(widget Widget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultSelected,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *comboBoxState) currentSelectedKey(widget Widget) string {
	s.selectedKey = s.disclosure.Current(comboBoxDisclosureCfg(widget))
	return s.selectedKey
}

func (s *comboBoxState) bind(widget Widget) {
	s.disclosure.Bind(comboBoxDisclosureCfg(widget))
}

func (s *comboBoxState) requestSelectedKey(widget Widget, key string) string {
	s.selectedKey, _ = s.disclosure.Request(comboBoxDisclosureCfg(widget), key)
	return s.selectedKey
}

func (s *comboBoxState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *comboBoxState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *comboBoxState) item(key string) *comboBoxItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *comboBoxState) derivedItemKey(ctx *core.Context, base, itemKey string) string {
	if s.itemDerivedBase != base {
		s.itemDerivedBase = base
		clear(s.itemDerivedKeys)
	}
	if s.itemDerivedKeys == nil {
		s.itemDerivedKeys = make(map[string]string)
	}
	if key, ok := s.itemDerivedKeys[itemKey]; ok {
		return key
	}
	key := core.DerivedKey(ctx, base, "item:"+itemKey)
	s.itemDerivedKeys[itemKey] = key
	return key
}

func (s *comboBoxState) checkItems(items []Item, versioned bool, version uint64) {
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

func (s *comboBoxState) selectedLabel(widget Widget) (string, bool) {
	if widget.hasDataVersion {
		label, ok := s.itemLabels[widget.selectedKey]
		return label, ok
	}
	return widget.selectedLabel()
}

func (s *comboBoxState) visibleItems(widget Widget, query, selectedLabel string) []int {
	// Always check cache based on query and selectedLabel, regardless of DataVersion.
	if s.visibleReady && s.visibleQuery == query && s.visibleSelected == selectedLabel {
		// If DataVersion is set and changed, invalidate cache.
		if widget.hasDataVersion && s.visibleDataVersion != widget.dataVersion {
			// Fall through to recompute
		} else {
			return s.cachedVisibleItems
		}
	}
	visible := comboBoxVisibleItems(widget.items, query, selectedLabel)
	s.visibleQuery = query
	s.visibleSelected = selectedLabel
	s.cachedVisibleItems = visible
	s.visibleReady = true
	if widget.hasDataVersion {
		s.visibleDataVersion = widget.dataVersion
	}
	return visible
}

func (s *comboBoxState) syncEditor(editor *widget.Editor, c Widget) {
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

func (s *comboBoxState) queueEdit(action input.EditAction) {
	s.pendingEdit = action
	s.hasPendingEdit = true
}

// takePendingEdit runs a queued edit menu action. Call after syncEditor and
// before the editor's Update loop so the ChangeEvent surfaces like a keystroke.
func (s *comboBoxState) takePendingEdit(gtx layout.Context, editor *widget.Editor) {
	if !s.hasPendingEdit {
		return
	}
	s.hasPendingEdit = false
	input.RunEditAction(gtx, editor, s.pendingEdit, false)
}

func (s *comboBoxState) updateFocus(focused, disabled bool) {
	if disabled {
		s.open = false
		s.wasFocused = false
		return
	}
	// A trigger press focuses the editor through the field hit area before its
	// release reaches Clickable. Do not open here as well, or the same click
	// would open on focus and close again on release. Non-trigger focus (for
	// example keyboard navigation or a press on the editor content) still opens
	// the popup on the focus transition.
	if focused && !s.wasFocused && core.ActivePresses(s.trigger.History()) == 0 {
		s.open = true
	}
	if !focused {
		s.open = false
	}
	s.wasFocused = focused
}

func (s *comboBoxState) updateKeys(gtx layout.Context, editor *widget.Editor, items []Item, visible []int, includeEscape bool) (int, bool) {
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
		)
		if includeEscape {
			filters = append(filters, key.Filter{Focus: editor, Name: key.NameEscape})
		}
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

// comboBoxNavList adapts the filtered visible indices into a navigation.List; the
// highlight is a position within visible, and Disabled indexes back through it.
func comboBoxNavList(items []Item, visible []int) navigation.List {
	return navigation.List{
		Count:    len(visible),
		Disabled: func(i int) bool { return items[visible[i]].Disabled },
	}
}

func comboBoxFirstEnabled(items []Item, visible []int) int {
	index, _ := navigation.First(comboBoxNavList(items, visible))
	return index
}

func comboBoxLastEnabled(items []Item, visible []int) int {
	index, _ := navigation.Last(comboBoxNavList(items, visible))
	return index
}

func comboBoxMoveHighlight(items []Item, visible []int, current, delta int) int {
	list := comboBoxNavList(items, visible)
	// ComboBox jumps to the first/last enabled when the highlight is unset or
	// stale; from a valid highlight it wraps.
	if current < 0 || current >= list.Count || list.Disabled(current) {
		if delta < 0 {
			return comboBoxLastEnabled(items, visible)
		}
		return comboBoxFirstEnabled(items, visible)
	}
	if next, ok := navigation.Move(list, current, delta, true); ok {
		return next
	}
	return current
}

func (s *comboBoxState) clampHighlight(items []Item, visible []int) {
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

func (s *comboBoxState) popoverProgress(gtx layout.Context, open bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return s.popoverTransition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *comboBoxState) iconProgress(gtx layout.Context, open bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return s.iconTransition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

type comboBoxItemState = optionrow.State
