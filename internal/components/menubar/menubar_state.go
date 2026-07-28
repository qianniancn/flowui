package menubar

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/disclosure"
	"github.com/qianniancn/FlowUI/internal/components/nav"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const (
	stateSlotMenubar = "menubar"
	menubarExclusive = "menubar"
	menubarEnter     = 150 * time.Millisecond
	menubarExit      = 100 * time.Millisecond
)

type menubarState struct {
	key           string
	triggers      map[string]*menubarTriggerState
	frameTriggers map[string]struct{}
	itemKeys      map[string]struct{}
	keyFilters    []event.Filter
	openKey       string // cached effective open key, updated by current/requestOpen
	wasOpen       bool
	panelKey      string
	focusPanelKey string
	focusLast     bool
	focusVisible  bool
	hoveredKey    string
	typeahead     nav.Typeahead
	dismiss       [16]overlay.ClickArea
	dialog        overlay.ClickArea
	transition    animation.FloatTransition
	disclosure    disclosure.Binding[string]
}

type menubarTriggerState struct {
	clickable  widget.Clickable
	focus      stateutil.FocusAnimation
	keyFilters stateutil.KeyFilterCache
}

func menubarStateFor(ctx *frame.Context, key string) *menubarState {
	key = frame.ClaimKey(ctx, stateutil.KindMenubar, key)
	value := frame.UseStateWith(ctx, key, stateSlotMenubar, func() *menubarState {
		return &menubarState{key: key}
	})
	frame.RegisterExclusive(ctx, menubarExclusive, key, value.closeForPeer)
	return value
}

// menubarDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func menubarDisclosureCfg(widget Widget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasOpenKey,
		Value:      widget.openKey,
		HasDefault: widget.hasDefaultOpenKey,
		Default:    widget.defaultOpenKey,
		OnChange:   widget.onOpenChange,
	}
}

func (s *menubarState) bind(widget Widget) {
	s.disclosure.Bind(menubarDisclosureCfg(widget))
}

func (s *menubarState) beginFrame(items []Item) {
	stateutil.BeginFrameMap(&s.frameTriggers)
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.key == "" {
			panic("flowui: empty menubar item key")
		}
		if item.label == "" {
			panic(fmt.Sprintf("flowui: empty menubar item label for key %q", item.key))
		}
		if _, exists := s.itemKeys[item.key]; exists {
			panic(fmt.Sprintf("flowui: duplicate menubar item key %q", item.key))
		}
		s.itemKeys[item.key] = struct{}{}
		s.trigger(item.key)
	}
}

func (s *menubarState) endFrame() {
	stateutil.SweepFrameMap(s.triggers, s.frameTriggers)
}

func (s *menubarState) trigger(key string) *menubarTriggerState {
	return stateutil.UseFrameMap(&s.triggers, &s.frameTriggers, key)
}

func (s *menubarState) current(widget Widget) string {
	key := s.disclosure.Current(menubarDisclosureCfg(widget))
	if widget.disabled || !widget.itemEnabled(key) {
		// If the current open key is now disabled, request close to clear it.
		if key != "" {
			s.openKey, _ = s.disclosure.Request(menubarDisclosureCfg(widget), "")
			return s.openKey
		}
		key = ""
	}
	s.openKey = key
	return key
}

func (s *menubarState) requestOpen(ctx *frame.Context, widget Widget, key string) string {
	if widget.disabled || !widget.itemEnabled(key) {
		key = ""
	}
	if key != "" {
		frame.ActivateExclusive(ctx, menubarExclusive, s.key)
	}
	s.openKey, _ = s.disclosure.Request(menubarDisclosureCfg(widget), key)
	if s.openKey == "" {
		frame.ReleaseExclusive(ctx, menubarExclusive, s.key)
	}
	return s.openKey
}

func (s *menubarState) closeForPeer() {
	if s.disclosure.PeerClose("") {
		s.openKey = ""
	}
}

func (s *menubarState) observeOpen(ctx *frame.Context, key string) {
	if key != "" {
		s.panelKey = key
	}
	open := key != ""
	if open && !s.wasOpen {
		frame.ActivateExclusive(ctx, menubarExclusive, s.key)
	} else if !open && s.wasOpen {
		frame.ReleaseExclusive(ctx, menubarExclusive, s.key)
	}
	s.wasOpen = open
}

func (s *menubarState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := menubarExit
	if open {
		target = 1
		duration = menubarEnter
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *menubarState) updateInteractions(ctx *frame.Context, gtx layout.Context, widget Widget) {
	openKey := s.current(widget)
	for _, item := range widget.items {
		trigger := s.trigger(item.key)
		eventGtx := gtx
		disabled := widget.disabled || item.disabled || !gtx.Enabled()
		if disabled {
			eventGtx = eventGtx.Disabled()
		}
		presses := stateutil.ActivePresses(trigger.clickable.History())
		for {
			_, clicked := trigger.clickable.Update(eventGtx)
			if !clicked {
				break
			}
			if disabled {
				continue
			}
			visible := frame.FocusVisible(ctx, &trigger.clickable, gtx.Focused(&trigger.clickable))
			if openKey == item.key {
				openKey = s.requestOpen(ctx, widget, "")
				s.focusPanelKey = ""
			} else {
				openKey = s.requestOpen(ctx, widget, item.key)
				s.focusPanelKey = item.key
				s.focusLast = false
				s.focusVisible = visible
			}
		}
		if !disabled {
			frame.FocusOnPress(ctx, &trigger.clickable, trigger.clickable.History(), presses)
		}
	}

	hoveredKey := ""
	for _, item := range widget.items {
		if !widget.itemDisabled(item) && s.trigger(item.key).clickable.Hovered() {
			hoveredKey = item.key
			break
		}
	}
	if openKey != "" && hoveredKey != "" && hoveredKey != openKey && hoveredKey != s.hoveredKey {
		trigger := s.trigger(hoveredKey)
		openKey = s.requestOpen(ctx, widget, hoveredKey)
		s.focusPanelKey = ""
		s.focusLast = false
		s.focusVisible = false
		frame.RequestFocusVisible(ctx, &trigger.clickable, false)
	}
	s.hoveredKey = hoveredKey

	s.updateGlobalKeys(ctx, gtx, widget, openKey)
	s.updateTriggerKeys(ctx, gtx, widget, openKey)
}

// updateGlobalKeys handles desktop menubar accelerators that are not
// scoped to a focused trigger: F10 focuses the bar; Alt+letter opens AccessKey menus.
func (s *menubarState) updateGlobalKeys(ctx *frame.Context, gtx layout.Context, widget Widget, openKey string) {
	if widget.disabled || !gtx.Enabled() {
		return
	}
	filters := []event.Filter{
		key.Filter{Name: key.NameF10, Optional: key.ModAlt},
	}
	for _, item := range widget.items {
		if widget.itemDisabled(item) || item.accessKey == 0 {
			continue
		}
		name := key.Name(strings.ToUpper(string(item.accessKey)))
		filters = append(filters, key.Filter{
			Name:     name,
			Required: key.ModAlt,
		})
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		if event.Name == key.NameF10 {
			first := widget.firstEnabled()
			if first < 0 {
				continue
			}
			if openKey != "" {
				s.requestOpen(ctx, widget, "")
				s.focusPanelKey = ""
			}
			s.focusMovedTrigger(ctx, widget, first, "")
			continue
		}
		if event.Modifiers&key.ModAlt == 0 || event.Modifiers&^key.ModAlt != 0 {
			continue
		}
		text := nav.Printable(event.Name)
		if text == "" {
			continue
		}
		letter := unicode.ToLower(rune(text[0]))
		for index, item := range widget.items {
			if widget.itemDisabled(item) || item.accessKey == 0 {
				continue
			}
			if item.accessKey == letter {
				// Focus without reusing the previous openKey, then open this item.
				s.focusMovedTrigger(ctx, widget, index, "")
				s.openFromTrigger(ctx, widget, index, false)
				break
			}
		}
	}
}

func (s *menubarState) updateTriggerKeys(ctx *frame.Context, gtx layout.Context, widget Widget, openKey string) {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range widget.items {
		if widget.itemDisabled(item) {
			continue
		}
		trigger := s.trigger(item.key)
		tag := &trigger.clickable
		if widget.orientation == Vertical {
			s.keyFilters = append(s.keyFilters, trigger.keyFilters.Resolve(tag,
				key.NameHome,
				key.NameEnd,
				key.NameEscape,
				"",
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameRightArrow,
			)...)
		} else {
			s.keyFilters = append(s.keyFilters, trigger.keyFilters.Resolve(tag,
				key.NameHome,
				key.NameEnd,
				key.NameEscape,
				"",
				key.NameRightArrow,
				key.NameLeftArrow,
				key.NameDownArrow,
				key.NameUpArrow,
			)...)
		}
	}
	if len(s.keyFilters) == 0 {
		return
	}
	current := s.focusedIndex(gtx, widget)
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			return
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameHome:
			current = widget.firstEnabled()
			s.focusMovedTrigger(ctx, widget, current, openKey)
		case key.NameEnd:
			current = widget.lastEnabled()
			s.focusMovedTrigger(ctx, widget, current, openKey)
		case key.NameEscape:
			if openKey != "" {
				s.requestOpen(ctx, widget, "")
				s.focusPanelKey = ""
			}
		case key.NameRightArrow:
			if widget.orientation == Vertical {
				s.openFromTrigger(ctx, widget, current, false)
			} else {
				current = widget.moveIndex(current, 1)
				s.focusMovedTrigger(ctx, widget, current, openKey)
			}
		case key.NameLeftArrow:
			current = widget.moveIndex(current, -1)
			s.focusMovedTrigger(ctx, widget, current, openKey)
		case key.NameDownArrow:
			if widget.orientation == Vertical {
				current = widget.moveIndex(current, 1)
				s.focusMovedTrigger(ctx, widget, current, openKey)
			} else {
				s.openFromTrigger(ctx, widget, current, false)
			}
		case key.NameUpArrow:
			if widget.orientation == Vertical {
				current = widget.moveIndex(current, -1)
				s.focusMovedTrigger(ctx, widget, current, openKey)
			} else {
				s.openFromTrigger(ctx, widget, current, true)
			}
		default:
			if event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := nav.Printable(event.Name)
			if text == "" {
				continue
			}
			query := s.typeahead.Append(gtx.Now, text)
			next := widget.typeaheadIndex(current, query)
			if next < 0 && query != text {
				s.typeahead.Set(text)
				next = widget.typeaheadIndex(current, text)
			}
			if next >= 0 {
				current = next
				s.focusMovedTrigger(ctx, widget, current, openKey)
			}
		}
	}
}

func (s *menubarState) focusedIndex(gtx layout.Context, widget Widget) int {
	for index, item := range widget.items {
		if gtx.Focused(&s.trigger(item.key).clickable) {
			return index
		}
	}
	return -1
}

func (s *menubarState) focusMovedTrigger(ctx *frame.Context, widget Widget, index int, openKey string) {
	if index < 0 || index >= len(widget.items) {
		return
	}
	item := widget.items[index]
	trigger := s.trigger(item.key)
	frame.RequestFocusVisible(ctx, &trigger.clickable, true)
	if openKey != "" {
		s.requestOpen(ctx, widget, item.key)
		s.focusPanelKey = ""
		s.focusLast = false
	}
}

func (s *menubarState) openFromTrigger(ctx *frame.Context, widget Widget, index int, last bool) {
	if index < 0 || index >= len(widget.items) || widget.itemDisabled(widget.items[index]) {
		return
	}
	item := widget.items[index]
	s.requestOpen(ctx, widget, item.key)
	s.focusPanelKey = item.key
	s.focusLast = last
	s.focusVisible = true
}

func (s *menubarState) focusTrigger(ctx *frame.Context, key string, visible bool) {
	trigger := s.triggers[key]
	if trigger == nil {
		return
	}
	frame.RequestFocusVisible(ctx, &trigger.clickable, visible)
}
