package selectfield

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotSelect = "select"

func selectStateFor(ctx *core.Context, key string) *selectState {
	key = core.ClaimKey(ctx, core.KindSelect, key)
	value := core.UseStateWith(ctx, key, stateSlotSelect, func() *selectState {
		return &selectState{key: key}
	})
	core.RegisterExclusive(ctx, "select", key, value.closeForPeer)
	return value
}

func activateSelect(ctx *core.Context, value *selectState) {
	if value != nil && value.key != "" {
		core.ActivateExclusive(ctx, "select", value.key)
	}
}

func releaseSelect(ctx *core.Context, value *selectState) {
	if value != nil {
		core.ReleaseExclusive(ctx, "select", value.key)
	}
}

type selectState struct {
	key                string
	trigger            widget.Clickable
	dismiss            [16]core.ClickArea
	dialog             core.ClickArea
	open               bool // cached effective open, updated by isOpen/requestOpen
	disclosure         disclosure.Binding[bool]
	selectedDisclosure disclosure.Binding[string]
	selectedKey        string
	selectedKeys       []string
	wasOpen            bool
	focusIntent        selectFocusIntent
	focusVisibleIntent bool
	triggerRect        image.Rectangle
	transition         animation.FloatTransition
	iconTransition     animation.FloatTransition
	skipRestore        bool
	peerClosePending   bool
	dataVersion        uint64
	dataReady          bool
	cachedItems        []SelectItem
	styleCache         computed.ProviderCache[SelectVariant, selectDeclarationBundle]
	styleKeyOwner      string
	stylePartKeys      [6]string
}

func (s *selectState) BeginFrame() {
	s.peerClosePending = false
}

func (s *selectState) itemsFor(widget Widget) []SelectItem {
	if !widget.hasDataVersion {
		return widget.allItems()
	}
	if s.dataReady && s.dataVersion == widget.dataVersion {
		return s.cachedItems
	}
	s.cachedItems = widget.allItems()
	s.dataVersion = widget.dataVersion
	s.dataReady = true
	return s.cachedItems
}

// selectDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func selectDisclosureCfg(widget Widget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

// selectSelectedDisclosureCfg builds a disclosure.Config for selectedKey (single selection mode).
func selectSelectedDisclosureCfg(widget Widget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultSelected,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *selectState) currentSelectedKey(widget Widget) string {
	if widget.selectionMode != SelectSelectionSingle {
		return widget.selectedKey
	}
	s.selectedKey = s.selectedDisclosure.Current(selectSelectedDisclosureCfg(widget))
	return s.selectedKey
}

func (s *selectState) bindSelected(widget Widget) {
	if widget.selectionMode == SelectSelectionSingle {
		s.selectedDisclosure.Bind(selectSelectedDisclosureCfg(widget))
	}
}

func (s *selectState) requestSelectedKey(widget Widget, key string) string {
	if widget.selectionMode != SelectSelectionSingle {
		return key
	}
	s.selectedKey, _ = s.selectedDisclosure.Request(selectSelectedDisclosureCfg(widget), key)
	return s.selectedKey
}

// Multi-selection uses manual state management since []string is not comparable
func (s *selectState) currentSelectedKeys(widget Widget) []string {
	if widget.selectionMode != SelectSelectionMultiple {
		return widget.selectedKeys
	}
	if widget.hasSelectedKeys {
		s.selectedKeys = widget.selectedKeys
		return s.selectedKeys
	}
	if len(s.selectedKeys) == 0 && widget.hasDefaultSelecteds {
		s.selectedKeys = widget.defaultSelectedKeys
	}
	return s.selectedKeys
}

func (s *selectState) requestSelectedKeys(widget Widget, keys []string) []string {
	if widget.selectionMode != SelectSelectionMultiple {
		return keys
	}
	if !widget.hasSelectedKeys {
		s.selectedKeys = keys
	}
	if widget.onSelectionChange != nil {
		widget.onSelectionChange(keys)
	}
	return s.selectedKeys
}

type selectFocusIntent uint8

const (
	selectFocusNone selectFocusIntent = iota
	selectFocusSelected
	selectFocusFirst
	selectFocusLast
)

func (s *selectState) isOpen(widget Widget) bool {
	s.open = s.disclosure.Current(selectDisclosureCfg(widget))
	return s.open
}

func (s *selectState) bind(widget Widget) {
	s.disclosure.Bind(selectDisclosureCfg(widget))
}

func (s *selectState) requestOpen(ctx *core.Context, widget Widget, open bool) bool {
	if widget.disabled {
		open = false
	}
	if open {
		s.skipRestore = false
		activateSelect(ctx, s)
	}
	s.open, _ = s.disclosure.Request(selectDisclosureCfg(widget), open)
	if !s.open {
		releaseSelect(ctx, s)
	}
	return s.open
}

func (s *selectState) closeForPeer() {
	s.focusIntent = selectFocusNone
	s.focusVisibleIntent = false
	s.skipRestore = true
	s.peerClosePending = true
	if s.disclosure.PeerClose(false) {
		s.open = false
	}
}

func (s *selectState) handleTrigger(ctx *core.Context, gtx layout.Context, widget Widget, open bool) bool {
	presses := core.SnapshotPresses(s.trigger.History())
	if widget.disabled {
		s.open = false
		s.focusIntent = selectFocusNone
		s.focusVisibleIntent = false
		return false
	}
	for s.trigger.Clicked(gtx) {
		if open {
			s.focusIntent = selectFocusNone
			s.focusVisibleIntent = false
			open = s.requestOpen(ctx, widget, false)
		} else {
			s.focusIntent = selectFocusSelected
			s.focusVisibleIntent = false
			open = s.requestOpen(ctx, widget, true)
		}
		core.RequestFocusVisible(ctx, &s.trigger, presses.ClickFocusVisible(s.trigger.History()))
	}
	core.FocusOnPress(ctx, &s.trigger, s.trigger.History(), presses.Active())
	return s.handleTriggerKeys(ctx, gtx, widget, open)
}

func (s *selectState) handleTriggerKeys(ctx *core.Context, gtx layout.Context, widget Widget, open bool) bool {
	filters := []event.Filter{
		key.Filter{Focus: &s.trigger, Name: key.NameDownArrow},
		key.Filter{Focus: &s.trigger, Name: key.NameUpArrow},
		key.Filter{Focus: &s.trigger, Name: key.NameReturn},
		key.Filter{Focus: &s.trigger, Name: key.NameEnter},
		key.Filter{Focus: &s.trigger, Name: key.NameSpace},
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return open
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameUpArrow:
			if !open {
				s.focusIntent = selectFocusLast
				s.focusVisibleIntent = true
				open = s.requestOpen(ctx, widget, true)
			}
		case key.NameDownArrow:
			if !open {
				s.focusIntent = selectFocusFirst
				s.focusVisibleIntent = true
				open = s.requestOpen(ctx, widget, true)
			}
		default:
			if !open {
				s.focusIntent = selectFocusSelected
				s.focusVisibleIntent = true
				open = s.requestOpen(ctx, widget, true)
			}
		}
	}
}

func (s *selectState) handleOverlayEvents(ctx *core.Context, gtx layout.Context, widget Widget, open bool) bool {
	peerClosePending := s.peerClosePending
	s.peerClosePending = false
	for s.dialog.Clicked(gtx) {
	}
	if s.dialog.TakePressed() {
		core.PreserveFocus(ctx)
	}
	dismissed := false
	for i := range s.dismiss {
		for s.dismiss[i].Clicked(gtx) {
			dismissed = true
		}
		dismissed = s.dismiss[i].TakePressed() || dismissed
	}
	if !open {
		return open
	}
	closed := dismissed
	escape := false
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if ok && event.State == key.Press {
			closed = true
			escape = true
		}
	}
	if peerClosePending {
		return open
	}
	if !closed {
		return open
	}
	s.focusIntent = selectFocusNone
	s.focusVisibleIntent = false
	if escape {
		core.RequestFocus(ctx, &s.trigger)
	} else {
		s.skipRestore = true
	}
	open = s.requestOpen(ctx, widget, false)
	gtx.Execute(op.InvalidateCmd{})
	return open
}

func (s *selectState) observeOpen(ctx *core.Context, open, restoreFocus bool) {
	if open && !s.wasOpen && s.focusIntent == selectFocusNone {
		s.focusIntent = selectFocusSelected
		s.focusVisibleIntent = false
	}
	if !open && s.wasOpen {
		s.focusIntent = selectFocusNone
		s.focusVisibleIntent = false
		if restoreFocus && !s.skipRestore {
			core.RequestFocus(ctx, &s.trigger)
		}
		s.skipRestore = false
	}
	s.wasOpen = open
}

func (s *selectState) progress(gtx layout.Context, open bool, metrics theme.SelectMetrics, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	duration := metrics.OpenDuration
	if !open {
		duration = metrics.CloseDuration
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *selectState) iconProgress(gtx layout.Context, open bool, metrics theme.SelectMetrics, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	s.iconTransition.Initialize(0, gtx.Now)
	return s.iconTransition.Value(gtx, target, metrics.IndicatorDuration, animation.EaseSmoothstep, motions...)
}
