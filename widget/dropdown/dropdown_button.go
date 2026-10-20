package dropdown

import (
	"github.com/qianniancn/flowui/style"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
)

// ButtonWidget combines a primary action button with a Dropdown trigger. It
// keeps the two controls visually joined while reusing the normal Button and
// Dropdown interaction paths.
type ButtonWidget struct {
	menuOverride   *theme.MenuOverride
	key            string
	action         button.Widget
	trigger        button.Widget
	dropdown       Widget
	disabled       bool
	hasDisabled    bool
	loading        bool
	defaultTrigger bool
}

// Button creates a split button with a provider-owned default menu trigger.
func Button(key string, action button.Widget, items []Item) ButtonWidget {
	trigger := button.New("").Key(key + ":trigger").Content(nil).IconOnly()
	return ButtonWidget{
		key:            key,
		action:         action,
		trigger:        trigger,
		dropdown:       New(key+":dropdown", trigger, items),
		defaultTrigger: true,
	}
}

// Trigger replaces the menu trigger button.
func (b ButtonWidget) Trigger(value button.Widget) ButtonWidget {
	b.trigger = value
	b.defaultTrigger = false
	return b
}

// OnClick handles the primary action half.
func (b ButtonWidget) OnClick(fn func()) ButtonWidget {
	b.action = b.action.OnClick(fn)
	return b
}

// Variant applies the same button variant to both halves.
func (b ButtonWidget) Variant(value button.Variant) ButtonWidget {
	b.action = b.action.Variant(value)
	b.trigger = b.trigger.Variant(value)
	return b
}

// Size applies the same button size to both halves.
func (b ButtonWidget) Size(value button.Size) ButtonWidget {
	b.action = b.action.Size(value)
	b.trigger = b.trigger.Size(value)
	return b
}

// Disabled disables both halves.
func (b ButtonWidget) Disabled(value bool) ButtonWidget {
	b.disabled = value
	b.hasDisabled = true
	return b
}

// Loading puts the primary action into its loading state and disables the
// menu half until loading finishes.
func (b ButtonWidget) Loading(value bool) ButtonWidget {
	b.action = b.action.Loading(value)
	b.loading = value
	return b
}

// Sections replaces the dropdown's item sections.
func (b ButtonWidget) Sections(value []Section) ButtonWidget {
	b.dropdown = b.dropdown.Sections(value)
	return b
}

// AutoSeparateSections controls automatic separators between menu sections.
func (b ButtonWidget) AutoSeparateSections(value bool) ButtonWidget {
	b.dropdown = b.dropdown.AutoSeparateSections(value)
	return b
}

// BeforeContent adds content above the popup menu items.
func (b ButtonWidget) BeforeContent(value core.Widget) ButtonWidget {
	b.dropdown = b.dropdown.BeforeContent(value)
	return b
}

// AfterContent adds content below the popup menu items.
func (b ButtonWidget) AfterContent(value core.Widget) ButtonWidget {
	b.dropdown = b.dropdown.AfterContent(value)
	return b
}

// EmptyText customizes the message shown by an empty popup menu.
func (b ButtonWidget) EmptyText(value string) ButtonWidget {
	b.dropdown = b.dropdown.EmptyText(value)
	return b
}

// DataVersion lets the popup reuse flattened menu entries and AutoWidth
// measurements until the supplied version changes.
func (b ButtonWidget) DataVersion(value uint64) ButtonWidget {
	b.dropdown = b.dropdown.DataVersion(value)
	return b
}

// SelectionMode configures selection behavior for the dropdown menu.
func (b ButtonWidget) SelectionMode(value SelectionMode) ButtonWidget {
	b.dropdown = b.dropdown.SelectionMode(value)
	return b
}

// SelectedKey marks one selected menu item.
func (b ButtonWidget) SelectedKey(value string) ButtonWidget {
	b.dropdown = b.dropdown.SelectedKey(value)
	return b
}

// SelectedKeys marks multiple selected menu items.
func (b ButtonWidget) SelectedKeys(value []string) ButtonWidget {
	b.dropdown = b.dropdown.SelectedKeys(value)
	return b
}

// DisabledKeys disables selected menu items.
func (b ButtonWidget) DisabledKeys(value []string) ButtonWidget {
	b.dropdown = b.dropdown.DisabledKeys(value)
	return b
}

// OnChange handles single-selection changes.
func (b ButtonWidget) OnChange(fn func(string)) ButtonWidget {
	b.dropdown = b.dropdown.OnChange(fn)
	return b
}

// OnSelectionChange handles multiple-selection changes.
func (b ButtonWidget) OnSelectionChange(fn func([]string)) ButtonWidget {
	b.dropdown = b.dropdown.OnSelectionChange(fn)
	return b
}

// OnCheckedChange reports checkbox item changes from the popup menu.
func (b ButtonWidget) OnCheckedChange(fn func(string, bool)) ButtonWidget {
	b.dropdown = b.dropdown.OnCheckedChange(fn)
	return b
}

// OnRadioChange reports radio item changes from the popup menu.
func (b ButtonWidget) OnRadioChange(fn func(string, string)) ButtonWidget {
	b.dropdown = b.dropdown.OnRadioChange(fn)
	return b
}

// CloseOnSelect controls whether selecting an item closes the menu.
func (b ButtonWidget) CloseOnSelect(value bool) ButtonWidget {
	b.dropdown = b.dropdown.CloseOnSelect(value)
	return b
}

// OnActionEvent reports the complete selected item and submenu path.
func (b ButtonWidget) OnActionEvent(fn func(ActionEvent)) ButtonWidget {
	b.dropdown = b.dropdown.OnActionEvent(fn)
	return b
}

// OnOpenChangeEvent reports open-state requests with their source.
func (b ButtonWidget) OnOpenChangeEvent(fn func(OpenChangeEvent)) ButtonWidget {
	b.dropdown = b.dropdown.OnOpenChangeEvent(fn)
	return b
}

func (b ButtonWidget) Open(open bool) ButtonWidget {
	b.dropdown = b.dropdown.Open(open)
	return b
}

func (b ButtonWidget) DefaultOpen(open bool) ButtonWidget {
	b.dropdown = b.dropdown.DefaultOpen(open)
	return b
}

func (b ButtonWidget) TriggerMode(mode TriggerMode) ButtonWidget {
	b.dropdown = b.dropdown.TriggerMode(mode)
	return b
}

func (b ButtonWidget) Offset(value int) ButtonWidget {
	b.dropdown = b.dropdown.Offset(value)
	return b
}

func (b ButtonWidget) ShouldFlip(value bool) ButtonWidget {
	b.dropdown = b.dropdown.ShouldFlip(value)
	return b
}

func (b ButtonWidget) AvoidOverflow(value bool) ButtonWidget {
	b.dropdown = b.dropdown.AvoidOverflow(value)
	return b
}

func (b ButtonWidget) HoverOpenDelay(value time.Duration) ButtonWidget {
	b.dropdown = b.dropdown.HoverOpenDelay(value)
	return b
}

func (b ButtonWidget) HoverCloseDelay(value time.Duration) ButtonWidget {
	b.dropdown = b.dropdown.HoverCloseDelay(value)
	return b
}

func (b ButtonWidget) LongPressDelay(value time.Duration) ButtonWidget {
	b.dropdown = b.dropdown.LongPressDelay(value)
	return b
}

func (b ButtonWidget) Placement(value core.PopoverPlacement) ButtonWidget {
	b.dropdown = b.dropdown.Placement(value)
	return b
}

func (b ButtonWidget) Width(dp int) ButtonWidget {
	b.dropdown = b.dropdown.Width(dp)
	return b
}

func (b ButtonWidget) AutoWidth() ButtonWidget {
	b.dropdown = b.dropdown.AutoWidth()
	return b
}

func (b ButtonWidget) MinWidth(dp int) ButtonWidget {
	b.dropdown = b.dropdown.MinWidth(dp)
	return b
}

func (b ButtonWidget) MaxWidth(dp int) ButtonWidget {
	b.dropdown = b.dropdown.MaxWidth(dp)
	return b
}

func (b ButtonWidget) MatchTriggerWidth(value bool) ButtonWidget {
	b.dropdown = b.dropdown.MatchTriggerWidth(value)
	return b
}

func (b ButtonWidget) Arrow(value bool) ButtonWidget {
	b.dropdown = b.dropdown.Arrow(value)
	return b
}

// Compact applies the menu's compact density to the popup.
func (b ButtonWidget) Compact(value bool) ButtonWidget {
	b.dropdown = b.dropdown.Compact(value)
	return b
}

func (b ButtonWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return b.layoutWithMenuOverride(ctx, gtx, b.menuOverride)
}

func (b ButtonWidget) layoutWithMenuOverride(ctx *core.Context, gtx layout.Context, menuOverride *theme.MenuOverride) layout.Dimensions {
	dropdown := b.dropdown
	action := button.GroupStart(b.action)
	trigger := b.trigger
	if b.defaultTrigger {
		state := dropdownStateFor(ctx, dropdown.key)
		visual := dropdown.visual(ctx, nil, &state.visualCache)
		dropdown.preparedState = state
		dropdown.preparedVisual = &visual
		trigger = trigger.Content(core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
			return visual.TriggerGlyph(gtx, theme.DropdownTriggerGlyphRequest{Color: ctx.ForegroundColor()})
		}))
	}
	if b.hasDisabled {
		action = action.Disabled(b.disabled)
		trigger = trigger.Disabled(b.disabled)
	}
	if action.DisabledState() {
		trigger = trigger.Disabled(true)
	}
	if b.loading {
		trigger = trigger.Disabled(true)
	}
	trigger = button.GroupEnd(trigger)
	effectiveDisabled := b.loading || action.DisabledState() || (b.hasDisabled && b.disabled) || dropdown.disabled
	if effectiveDisabled {
		dropdown = dropdown.Disabled(true)
	}
	dropdown.trigger = trigger
	var dropdownWidget core.Widget = dropdown
	if menuOverride != nil {
		dropdownWidget = dropdown.StyleMenu(menuOverride.Declaration)
	}
	children := layoutui.Row(action, dropdownWidget).Gap(0).AlignMiddle()
	return children.Layout(ctx, gtx)
}

// StyleMenu merges instance declarations into the popup menu.
func (b ButtonWidget) StyleMenu(values ...style.Style) ButtonWidget {
	declaration := style.Join(values...)
	if b.menuOverride == nil {
		b.menuOverride = &theme.MenuOverride{}
	}
	b.menuOverride.Declaration = style.Join(b.menuOverride.Declaration, declaration)
	return b
}
