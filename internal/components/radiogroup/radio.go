package radiogroup

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type RadioItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
	Invalid     bool
}

type RadioGroupWidget struct {
	key                string
	selectedKey        string
	hasSelectedKey     bool
	defaultSelectedKey string
	hasDefaultSelected bool
	items              []RadioItem
	onChange           func(string)
	variant            RadioGroupVariant
	disabled           bool
	invalid            bool
	horizontal         bool
	customStyle        flowstyle.Style
}

type RadioGroupVariant int

const (
	RadioPrimary RadioGroupVariant = iota
	RadioSecondary
)

const (
	radioSelectDuration   = 200 * time.Millisecond
	radioFocusDuration    = 100 * time.Millisecond
	radioPressInDuration  = 80 * time.Millisecond
	radioPressOutDuration = 140 * time.Millisecond
)

func RadioGroup(key, selectedKey string, items []RadioItem) RadioGroupWidget {
	if selectedKey != "" {
		// non-empty selectedKey → controlled mode
		return RadioGroupWidget{
			key:            key,
			selectedKey:    selectedKey,
			hasSelectedKey: true,
			items:          items,
		}
	}
	// empty selectedKey → uncontrolled mode
	return RadioGroupWidget{
		key:   key,
		items: items,
	}
}

func (r RadioGroupWidget) OnChange(fn func(string)) RadioGroupWidget {
	r.onChange = fn
	return r
}

func (r RadioGroupWidget) SelectedKey(key string) RadioGroupWidget {
	r.selectedKey = key
	r.hasSelectedKey = true
	return r
}

func (r RadioGroupWidget) DefaultSelectedKey(key string) RadioGroupWidget {
	r.defaultSelectedKey = key
	r.hasDefaultSelected = true
	return r
}

func (r RadioGroupWidget) Disabled(disabled bool) RadioGroupWidget {
	r.disabled = disabled
	return r
}

func (r RadioGroupWidget) Invalid(invalid bool) RadioGroupWidget {
	r.invalid = invalid
	return r
}

func (r RadioGroupWidget) Variant(variant RadioGroupVariant) RadioGroupWidget {
	r.variant = variant
	return r
}

func (r RadioGroupWidget) Horizontal() RadioGroupWidget {
	r.horizontal = true
	return r
}

func (r RadioGroupWidget) Style(value flowstyle.Style) RadioGroupWidget {
	r.customStyle = value
	return r
}

func (r RadioGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := radioGroupStateFor(ctx, r.key)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(r.items)

	// Bind disclosure state
	state.bind(r)
	selectedKey := state.currentSelectedKey(r)

	disabled := r.disabled || !gtx.Enabled()
	if disabled {
		gtx = gtx.Disabled()
	} else if key, ok := state.updateKeys(gtx, r.items, selectedKey); ok {
		if key != selectedKey {
			selectedKey = state.requestSelectedKey(r, key)
		}
		frame.RequestFocus(ctx, &state.item(key).clickable)
	}

	children := make([]layout.Widget, 0, len(r.items))
	for _, item := range r.items {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return r.layoutItem(ctx, gtx, state, item, selectedKey)
		})
	}

	theme := frame.ActiveTheme(ctx).Components.RadioGroup
	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, item := range r.items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		itemFocused := gtx.Focused(&itemState.clickable)
		focused = focused || itemFocused
		focusVisible = focusVisible || frame.FocusVisible(ctx, &itemState.clickable, itemFocused)
	}
	return layoutui.LayoutStyled(ctx, gtx, frame.FullKey(ctx, r.key), flowstyle.StyleState{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		Selected:     r.selectedKey != "",
		Checked:      r.selectedKey != "",
		Invalid:      r.invalid,
	}, r.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.LayoutItems(ctx, gtx, r.horizontal, gtx.Dp(theme.HorizontalGap), gtx.Dp(theme.VerticalGap), children)
	}))
}
