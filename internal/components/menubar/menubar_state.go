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
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const (
	stateSlotMenubar   = "menubar"
	menubarExclusive   = "menubar"
	menubarTypeTimeout = 500 * time.Millisecond
	menubarEnter       = 150 * time.Millisecond
	menubarExit        = 100 * time.Millisecond
)

type menubarState struct {
	key            string
	triggers       map[string]*menubarTriggerState
	frameTriggers  map[string]struct{}
	itemKeys       map[string]struct{}
	keyFilters     []event.Filter
	openKey        string
	initialized    bool
	wasOpen        bool
	panelKey       string
	focusPanelKey  string
	focusLast      bool
	focusVisible   bool
	hoveredKey     string
	typeaheadText  string
	typeaheadAt    time.Time
	typeaheadReady bool
	dismiss        [16]overlay.ClickArea
	dialog         overlay.ClickArea
	transition     animation.FloatTransition
	binding        menubarBinding
}

type menubarTriggerState struct {
	clickable widget.Clickable
	focus     stateutil.FocusAnimation
}

type menubarBinding struct {
	controlled   bool
	openKey      string
	onOpenChange func(string)
}

func menubarStateFor(ctx *frame.Context, key string) *menubarState {
	key = frame.ClaimKey(ctx, stateutil.KindMenubar, key)
	value := frame.UseStateWith(ctx, key, stateSlotMenubar, func() *menubarState {
		return &menubarState{key: key}
	})
	frame.RegisterExclusive(ctx, menubarExclusive, key, value.closeForPeer)
	return value
}

func (s *menubarState) bind(widget Widget) {
	s.binding = menubarBinding{
		controlled:   widget.hasOpenKey,
		openKey:      widget.openKey,
		onOpenChange: widget.onOpenChange,
	}
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
	if !s.initialized {
		if widget.hasDefaultOpenKey {
			s.openKey = widget.defaultOpenKey
		}
		s.initialized = true
	}
	key := s.openKey
	if widget.hasOpenKey {
		key = widget.openKey
	}
	if widget.disabled || !widget.itemEnabled(key) {
		return ""
	}
	return key
}

func (s *menubarState) requestOpen(ctx *frame.Context, widget Widget, key string) string {
	if widget.disabled || !widget.itemEnabled(key) {
		key = ""
	}
	if key != "" {
		frame.ActivateExclusive(ctx, menubarExclusive, s.key)
	}
	current := s.current(widget)
	if widget.hasOpenKey {
		if widget.openKey != key && widget.onOpenChange != nil {
			widget.onOpenChange(key)
		}
		if current == "" {
			frame.ReleaseExclusive(ctx, menubarExclusive, s.key)
		}
		return current
	}
	if s.openKey != key {
		s.openKey = key
		if widget.onOpenChange != nil {
			widget.onOpenChange(key)
		}
	}
	if s.openKey == "" {
		frame.ReleaseExclusive(ctx, menubarExclusive, s.key)
	}
	return s.openKey
}

func (s *menubarState) closeForPeer() {
	if s.binding.controlled {
		if s.binding.openKey != "" && s.binding.onOpenChange != nil {
			s.binding.onOpenChange("")
		}
		return
	}
	if s.openKey != "" {
		s.openKey = ""
		if s.binding.onOpenChange != nil {
			s.binding.onOpenChange("")
		}
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

	s.updateTriggerKeys(ctx, gtx, widget, openKey)
}

func (s *menubarState) updateTriggerKeys(ctx *frame.Context, gtx layout.Context, widget Widget, openKey string) {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range widget.items {
		if widget.itemDisabled(item) {
			continue
		}
		tag := &s.trigger(item.key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
			key.Filter{Focus: tag, Name: key.NameEscape},
			key.Filter{Focus: tag},
		)
		if widget.orientation == Vertical {
			s.keyFilters = append(s.keyFilters,
				key.Filter{Focus: tag, Name: key.NameDownArrow},
				key.Filter{Focus: tag, Name: key.NameUpArrow},
				key.Filter{Focus: tag, Name: key.NameRightArrow},
			)
		} else {
			s.keyFilters = append(s.keyFilters,
				key.Filter{Focus: tag, Name: key.NameRightArrow},
				key.Filter{Focus: tag, Name: key.NameLeftArrow},
				key.Filter{Focus: tag, Name: key.NameDownArrow},
				key.Filter{Focus: tag, Name: key.NameUpArrow},
			)
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
			text := menubarKeyText(event.Name)
			if text == "" {
				continue
			}
			query := s.appendTypeahead(gtx.Now, text)
			next := widget.typeaheadIndex(current, query)
			if next < 0 && query != text {
				s.typeaheadText = text
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

func (s *menubarState) appendTypeahead(now time.Time, text string) string {
	if !s.typeaheadReady || now.Before(s.typeaheadAt) || now.Sub(s.typeaheadAt) > menubarTypeTimeout {
		s.typeaheadText = ""
	}
	s.typeaheadText += text
	s.typeaheadAt = now
	s.typeaheadReady = true
	return s.typeaheadText
}

func menubarKeyText(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
}
