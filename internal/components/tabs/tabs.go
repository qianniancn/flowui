package tabs

import (
	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type TabItem struct {
	Key      string
	Label    string
	Panel    frame.Widget
	Disabled bool
}

type TabsVariant int

const (
	TabsPrimary TabsVariant = iota
	TabsSecondary
)

type TabsOrientation int

const (
	TabsHorizontal TabsOrientation = iota
	TabsVertical
)

type TabsSize int

const (
	TabsMedium TabsSize = iota
	TabsSmall
)

type TabsColor int

const (
	TabsColorDefault TabsColor = iota
	TabsColorAccent
)

type TabsWidget struct {
	key                string
	selectedKey        string
	hasSelectedKey     bool
	defaultSelectedKey string
	hasDefaultSelected bool
	items              []TabItem
	onChange           func(string)
	variant            TabsVariant
	orientation        TabsOrientation
	size               TabsSize
	color              TabsColor
	disabled           bool
	separators         bool
	fit                bool
	customStyle        flowstyle.Style
}

// Tabs creates a tabs widget. When selectedKey is non-empty, it starts in
// controlled mode with that tab selected. When selectedKey is empty, it starts
// in uncontrolled mode (use DefaultSelectedKey to set initial selection).
func Tabs(key, selectedKey string, items []TabItem) TabsWidget {
	if selectedKey != "" {
		// non-empty selectedKey → controlled mode
		return TabsWidget{
			key:            key,
			selectedKey:    selectedKey,
			hasSelectedKey: true,
			items:          items,
		}
	}
	// empty selectedKey → uncontrolled mode
	return TabsWidget{
		key:   key,
		items: items,
	}
}

func (t TabsWidget) OnChange(fn func(string)) TabsWidget {
	t.onChange = fn
	return t
}

// SelectedKey sets the tabs to controlled mode with the given selected key.
func (t TabsWidget) SelectedKey(key string) TabsWidget {
	t.selectedKey = key
	t.hasSelectedKey = true
	return t
}

// DefaultSelectedKey sets the initial selected key for uncontrolled mode.
func (t TabsWidget) DefaultSelectedKey(key string) TabsWidget {
	t.defaultSelectedKey = key
	t.hasDefaultSelected = true
	return t
}

func (t TabsWidget) Variant(variant TabsVariant) TabsWidget {
	t.variant = variant
	return t
}

func (t TabsWidget) Orientation(orientation TabsOrientation) TabsWidget {
	t.orientation = orientation
	return t
}

func (t TabsWidget) Vertical() TabsWidget {
	t.orientation = TabsVertical
	return t
}

func (t TabsWidget) Size(size TabsSize) TabsWidget {
	t.size = size
	return t
}

func (t TabsWidget) Color(color TabsColor) TabsWidget {
	t.color = color
	return t
}

func (t TabsWidget) Fit() TabsWidget {
	t.fit = true
	return t
}

func (t TabsWidget) Disabled(disabled bool) TabsWidget {
	t.disabled = disabled
	return t
}

func (t TabsWidget) Separators(visible bool) TabsWidget {
	t.separators = visible
	return t
}

func (t TabsWidget) Style(value flowstyle.Style) TabsWidget {
	t.customStyle = value
	return t
}

func (t TabsWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	rootKey := frame.FullKey(ctx, t.key)
	state := tabsStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(t.items)

	// Bind disclosure state
	state.bind(t)
	selectedKey := state.currentSelectedKey(t)

	state.syncSelection(t.items, selectedKey)
	disabled := t.disabled || !gtx.Enabled()
	if !disabled {
		if key, ok := state.updateKeys(gtx, t.items, selectedKey, t.orientation); ok {
			selectedKey = state.requestSelectedKey(t, key)
			frame.RequestFocus(ctx, &state.item(key).clickable)
		}
	}
	hovered, pressed, focused := false, false, false
	for _, item := range t.items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		focused = focused || gtx.Focused(&itemState.clickable)
	}
	return layoutui.LayoutStyled(ctx, gtx, rootKey, flowstyle.StyleState{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  focused,
		Disabled: disabled,
		Selected: selectedKey != "",
	}, t.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return t.layout(ctx, gtx, state, selectedKey, disabled)
	}))
}

func (t TabsWidget) selectedItem(key string) (TabItem, bool) {
	for _, item := range t.items {
		if item.Key == key {
			return item, true
		}
	}
	return TabItem{}, false
}
