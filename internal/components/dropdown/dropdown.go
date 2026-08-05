package dropdown

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type SelectionMode = menu.SelectionMode
type ItemKind = menu.ItemKind
type ItemVariant = menu.ItemVariant
type IndicatorType = menu.IndicatorType
type Item = menu.Item
type Section = menu.Section
type ActionEvent = menu.ActionEvent

const (
	ItemAction     = menu.ItemAction
	ItemCheckbox   = menu.ItemCheckbox
	ItemRadio      = menu.ItemRadio
	ItemSubmenu    = menu.ItemSubmenu
	ItemSeparator  = menu.ItemSeparator
	ItemGroupLabel = menu.ItemGroupLabel

	SelectionNone     = menu.SelectionNone
	SelectionSingle   = menu.SelectionSingle
	SelectionMultiple = menu.SelectionMultiple

	ItemDefault = menu.ItemDefault
	ItemDanger  = menu.ItemDanger

	IndicatorNone      = menu.IndicatorNone
	IndicatorCheckmark = menu.IndicatorCheckmark
	IndicatorDot       = menu.IndicatorDot
)

type TriggerMode uint8

const (
	TriggerPress TriggerMode = iota
	TriggerLongPress
	TriggerHover
	TriggerContextMenu
)

// OpenChangeSource identifies the interaction that requested a Dropdown state
// change.
type OpenChangeSource uint8

const (
	OpenChangeProgrammatic OpenChangeSource = iota
	OpenChangeTrigger
	OpenChangeMenu
	OpenChangeOutside
	OpenChangeKeyboard
	OpenChangePeer
	OpenChangeContextMenu
)

// OpenChangeEvent contains the state transition and its interaction source.
type OpenChangeEvent struct {
	Open   bool
	Source OpenChangeSource
}

type Widget struct {
	key                string
	trigger            frame.Widget
	menu               menu.Widget
	open               bool
	hasOpen            bool
	defaultOpen        bool
	hasDefaultOpen     bool
	triggerMode        TriggerMode
	placement          overlay.PopoverPlacement
	offset             unit.Dp
	hasOffset          bool
	shouldFlip         bool
	hasShouldFlip      bool
	avoidOverflow      bool
	hasAvoidOverflow   bool
	matchTriggerWidth  bool
	arrow              bool
	hoverOpenDelay     time.Duration
	hasHoverOpenDelay  bool
	hoverCloseDelay    time.Duration
	hasHoverCloseDelay bool
	longPressDelay     time.Duration
	hasLongPressDelay  bool
	disabled           bool
	customStyle        flowstyle.Style
	onOpenChangeEvent  func(OpenChangeEvent)
}

func New(key string, trigger frame.Widget, items []Item) Widget {
	return Widget{
		key:       key,
		trigger:   trigger,
		menu:      menu.Menu(key+":menu", items),
		placement: overlay.PopoverBottomStart,
	}
}

func NewSections(key string, trigger frame.Widget, sections []Section) Widget {
	return Widget{
		key:       key,
		trigger:   trigger,
		menu:      menu.MenuSections(key+":menu", sections).AutoSeparateSections(false),
		placement: overlay.PopoverBottomStart,
	}
}

func Separator() Item {
	return menu.MenuSeparator()
}

// GroupLabel returns a non-selectable heading for a dropdown section.
func GroupLabel(label string) Item {
	return menu.MenuGroupLabel(label)
}

func (d Widget) Sections(sections []Section) Widget {
	d.menu = d.menu.Sections(sections).AutoSeparateSections(false)
	return d
}

// AutoSeparateSections controls whether section groups receive automatic
// separators. Explicit separators remain available through Separator.
func (d Widget) AutoSeparateSections(enabled bool) Widget {
	d.menu = d.menu.AutoSeparateSections(enabled)
	return d
}

func (d Widget) BeforeContent(content frame.Widget) Widget {
	d.menu = d.menu.BeforeContent(content)
	return d
}

func (d Widget) AfterContent(content frame.Widget) Widget {
	d.menu = d.menu.AfterContent(content)
	return d
}

func (d Widget) EmptyText(text string) Widget {
	d.menu = d.menu.EmptyText(text)
	return d
}

// DataVersion lets Menu reuse flattened entries and AutoWidth measurements
// until the supplied version changes.
func (d Widget) DataVersion(version uint64) Widget {
	d.menu = d.menu.DataVersion(version)
	return d
}

func (d Widget) SelectionMode(mode SelectionMode) Widget {
	d.menu = d.menu.SelectionMode(mode)
	return d
}

func (d Widget) SelectedKey(key string) Widget {
	d.menu = d.menu.SelectedKey(key)
	return d
}

func (d Widget) SelectedKeys(keys []string) Widget {
	d.menu = d.menu.SelectedKeys(keys)
	return d
}

func (d Widget) DisabledKeys(keys []string) Widget {
	d.menu = d.menu.DisabledKeys(keys)
	return d
}

// OnActionEvent reports the complete activated item and its submenu path.
func (d Widget) OnActionEvent(fn func(ActionEvent)) Widget {
	d.menu = d.menu.OnActionEvent(fn)
	return d
}

func (d Widget) OnChange(fn func(string)) Widget {
	d.menu = d.menu.OnChange(fn)
	return d
}

func (d Widget) OnSelectionChange(fn func([]string)) Widget {
	d.menu = d.menu.OnSelectionChange(fn)
	return d
}

// OnCheckedChange reports checkbox item changes from the popup menu.
func (d Widget) OnCheckedChange(fn func(string, bool)) Widget {
	d.menu = d.menu.OnCheckedChange(fn)
	return d
}

// OnRadioChange reports radio item changes from the popup menu.
func (d Widget) OnRadioChange(fn func(string, string)) Widget {
	d.menu = d.menu.OnRadioChange(fn)
	return d
}

func (d Widget) Open(open bool) Widget {
	d.open = open
	d.hasOpen = true
	return d
}

func (d Widget) DefaultOpen(open bool) Widget {
	d.defaultOpen = open
	d.hasDefaultOpen = true
	return d
}

// OnOpenChangeEvent reports state changes together with their source.
func (d Widget) OnOpenChangeEvent(fn func(OpenChangeEvent)) Widget {
	d.onOpenChangeEvent = fn
	return d
}

func (d Widget) TriggerMode(mode TriggerMode) Widget {
	d.triggerMode = mode
	return d
}

// HoverOpenDelay sets the delay before a hover-triggered popup opens.
func (d Widget) HoverOpenDelay(delay time.Duration) Widget {
	d.hoverOpenDelay = max(delay, 0)
	d.hasHoverOpenDelay = true
	return d
}

// HoverCloseDelay sets the delay before a hover-triggered popup closes.
func (d Widget) HoverCloseDelay(delay time.Duration) Widget {
	d.hoverCloseDelay = max(delay, 0)
	d.hasHoverCloseDelay = true
	return d
}

// LongPressDelay sets the delay before a long-press-triggered popup opens.
func (d Widget) LongPressDelay(delay time.Duration) Widget {
	d.longPressDelay = max(delay, 0)
	d.hasLongPressDelay = true
	return d
}

func (d Widget) Placement(placement overlay.PopoverPlacement) Widget {
	d.placement = placement
	return d
}

func (d Widget) Offset(dp int) Widget {
	d.offset = unit.Dp(max(dp, 0))
	d.hasOffset = true
	return d
}

func (d Widget) ShouldFlip(shouldFlip bool) Widget {
	d.shouldFlip = shouldFlip
	d.hasShouldFlip = true
	return d
}

func (d Widget) AvoidOverflow(avoid bool) Widget {
	d.avoidOverflow = avoid
	d.hasAvoidOverflow = true
	return d
}

func (d Widget) CloseOnSelect(close bool) Widget {
	d.menu = d.menu.CloseOnSelect(close)
	return d
}

func (d Widget) Disabled(disabled bool) Widget {
	d.disabled = disabled
	d.menu = d.menu.Disabled(disabled)
	return d
}

// Compact applies the menu's compact density to the popup.
func (d Widget) Compact(compact bool) Widget {
	d.menu = d.menu.Compact(compact)
	return d
}

func (d Widget) Width(dp int) Widget {
	d.menu = d.menu.Width(dp)
	return d
}

// AutoWidth sizes the popup to its content, subject to MinWidth and MaxWidth.
func (d Widget) AutoWidth() Widget {
	d.menu = d.menu.AutoWidth()
	return d
}

// MinWidth sets the minimum popup width when AutoWidth is enabled.
func (d Widget) MinWidth(dp int) Widget {
	d.menu = d.menu.MinWidth(dp)
	return d
}

// MaxWidth limits the popup width when AutoWidth is enabled.
func (d Widget) MaxWidth(dp int) Widget {
	d.menu = d.menu.MaxWidth(dp)
	return d
}

// MatchTriggerWidth makes the popup at least as wide as its trigger.
func (d Widget) MatchTriggerWidth(match bool) Widget {
	d.matchTriggerWidth = match
	return d
}

// Arrow controls whether the popup draws an anchor arrow.
func (d Widget) Arrow(show bool) Widget {
	d.arrow = show
	return d
}

func (d Widget) Style(value flowstyle.Style) Widget {
	d.customStyle = value
	return d
}

// MenuStyle customizes the popup menu without changing the dropdown trigger.
func (d Widget) MenuStyle(value flowstyle.Style) Widget {
	d.menu = d.menu.Style(value)
	return d
}

func (d Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return d.layoutRoot(ctx, gtx)
}

func (d Widget) flipEnabled() bool {
	return !d.hasShouldFlip || d.shouldFlip
}

func (d Widget) overflowAvoidanceEnabled() bool {
	return !d.hasAvoidOverflow || d.avoidOverflow
}

func (d Widget) hoverOpenDuration() time.Duration {
	if d.hasHoverOpenDelay {
		return d.hoverOpenDelay
	}
	return dropdownHoverOpen
}

func (d Widget) hoverCloseDuration() time.Duration {
	if d.hasHoverCloseDelay {
		return d.hoverCloseDelay
	}
	return dropdownHoverClose
}

func (d Widget) longPressDuration() time.Duration {
	if d.hasLongPressDelay {
		return d.longPressDelay
	}
	return dropdownLongPress
}
