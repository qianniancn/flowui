package tabs

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
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
	key         string
	selectedKey string
	items       []TabItem
	onChange    func(string)
	variant     TabsVariant
	orientation TabsOrientation
	size        TabsSize
	color       TabsColor
	disabled    bool
	separators  bool
	fit         bool
}

func Tabs(key, selectedKey string, items []TabItem) TabsWidget {
	return TabsWidget{key: key, selectedKey: selectedKey, items: items}
}

func (t TabsWidget) OnChange(fn func(string)) TabsWidget {
	t.onChange = fn
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

func (t TabsWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := tabsStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(t.items)

	selectedKey := t.effectiveSelectedKey()
	state.syncSelection(t.items, selectedKey)
	disabled := t.disabled || !gtx.Enabled()
	if !disabled {
		if key, ok := state.updateKeys(gtx, t.items, selectedKey, t.orientation); ok {
			if t.onChange != nil {
				t.onChange(key)
			}
			frame.RequestFocus(ctx, &state.item(key).clickable)
		}
	}
	return t.layout(ctx, gtx, state, selectedKey, disabled)
}

func (t TabsWidget) effectiveSelectedKey() string {
	for _, item := range t.items {
		if item.Key == t.selectedKey {
			return item.Key
		}
	}
	for _, item := range t.items {
		if !item.Disabled {
			return item.Key
		}
	}
	return ""
}

func (t TabsWidget) selectedItem(key string) (TabItem, bool) {
	for _, item := range t.items {
		if item.Key == key {
			return item, true
		}
	}
	return TabItem{}, false
}
