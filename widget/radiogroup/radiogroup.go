package radiogroup

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type RadioItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
	Invalid     bool
}

type Widget struct {
	override           *theme.RadioGroupOverride
	key                string
	selectedKey        string
	hasSelectedKey     bool
	defaultSelectedKey string
	hasDefaultSelected bool
	items              []RadioItem
	onChange           func(string)
	variant            Variant
	disabled           bool
	invalid            bool
	horizontal         bool
}

type Variant = theme.RadioGroupVariant

const (
	RadioPrimary   = theme.RadioPrimary
	RadioSecondary = theme.RadioSecondary
)

// New creates a radio group. It starts uncontrolled with nothing selected;
// SelectedKey sets the controlled value and DefaultSelectedKey the initial one.
func New(items []RadioItem) Widget {
	return Widget{
		key:   core.AutoKey(),
		items: items,
	}
}

// Key sets an explicit identity for the group state.
func (r Widget) Key(key string) Widget {
	r.key = key
	return r
}

func (r Widget) OnChange(fn func(string)) Widget {
	r.onChange = fn
	return r
}

func (r Widget) SelectedKey(key string) Widget {
	r.selectedKey = key
	r.hasSelectedKey = true
	return r
}

func (r Widget) DefaultSelectedKey(key string) Widget {
	r.defaultSelectedKey = key
	r.hasDefaultSelected = true
	return r
}

func (r Widget) Disabled(disabled bool) Widget {
	r.disabled = disabled
	return r
}

func (r Widget) Invalid(invalid bool) Widget {
	r.invalid = invalid
	return r
}

func (r Widget) Variant(variant Variant) Widget {
	r.variant = variant
	return r
}

func (r Widget) Horizontal() Widget {
	r.horizontal = true
	return r
}

func (r Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return r.layoutWithOverride(ctx, gtx, r.override)
}

func (r Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.RadioGroupOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
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
		core.RequestFocus(ctx, &state.item(key).clickable)
	}

	visual := state.visual(ctx, r.variant, override)
	children := make([]layout.Widget, 0, len(r.items))
	for _, item := range r.items {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return r.layoutItem(ctx, gtx, state, item, selectedKey, visual)
		})
	}

	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, item := range r.items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		itemFocused := gtx.Focused(&itemState.clickable)
		focused = focused || itemFocused
		focusVisible = focusVisible || core.FocusVisible(ctx, &itemState.clickable, itemFocused)
	}
	return layoutui.Styled(ctx, gtx, core.FullKey(ctx, r.key), style.State{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		Selected:     r.selectedKey != "",
		Checked:      r.selectedKey != "",
		Invalid:      r.invalid,
	}, visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Items(ctx, gtx, r.horizontal, gtx.Dp(visual.Metrics.HorizontalGap), gtx.Dp(visual.Metrics.VerticalGap), children)
	}))
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.RadioGroupOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
