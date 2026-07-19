package dropdown

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/menu"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type SelectionMode = menu.SelectionMode
type ItemVariant = menu.ItemVariant
type IndicatorType = menu.IndicatorType
type Item = menu.Item
type Section = menu.Section

const (
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
)

type Widget struct {
	theme            func(*theme.Theme)
	key              string
	trigger          frame.Widget
	menu             menu.Widget
	onOpenChange     func(bool)
	open             bool
	hasOpen          bool
	defaultOpen      bool
	hasDefaultOpen   bool
	triggerMode      TriggerMode
	placement        overlay.PopoverPlacement
	offset           unit.Dp
	hasOffset        bool
	shouldFlip       bool
	hasShouldFlip    bool
	avoidOverflow    bool
	hasAvoidOverflow bool
	disabled         bool
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

func (d Widget) Sections(sections []Section) Widget {
	d.menu = d.menu.Sections(sections).AutoSeparateSections(false)
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

func (d Widget) OnAction(fn func(string)) Widget {
	d.menu = d.menu.OnAction(fn)
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

func (d Widget) OnOpenChange(fn func(bool)) Widget {
	d.onOpenChange = fn
	return d
}

func (d Widget) TriggerMode(mode TriggerMode) Widget {
	d.triggerMode = mode
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

func (d Widget) Width(dp int) Widget {
	d.menu = d.menu.Width(dp)
	return d
}

func (d Widget) Theme(fn func(*theme.Theme)) Widget {
	d.theme = fn
	return d
}

func (d Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, d.theme); restore != nil {
		defer restore()
	}
	return d.layoutRoot(ctx, gtx)
}

func (d Widget) flipEnabled() bool {
	return !d.hasShouldFlip || d.shouldFlip
}

func (d Widget) overflowAvoidanceEnabled() bool {
	return !d.hasAvoidOverflow || d.avoidOverflow
}
