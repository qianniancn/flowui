package collapsible

import (
	"fmt"
	"image"
	"slices"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Item describes one entry in a CollapsibleGroup.
type Item struct {
	Key      string
	Label    string
	Leading  frame.Widget
	Trailing frame.Widget
	Content  frame.Widget
	Disabled bool
}

// Widget presents one controlled expandable section.
type Widget struct {
	theme            func(*theme.Theme)
	key              string
	expanded         bool
	label            string
	content          frame.Widget
	leading          frame.Widget
	trailing         frame.Widget
	disabled         bool
	onExpandedChange func(bool)
}

// Collapsible creates one controlled expandable section.
func Collapsible(key string, expanded bool, label string, content frame.Widget) Widget {
	return Widget{key: key, expanded: expanded, label: label, content: content}
}

func (w Widget) Leading(leading frame.Widget) Widget {
	w.leading = leading
	return w
}

func (w Widget) Trailing(trailing frame.Widget) Widget {
	w.trailing = trailing
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) OnExpandedChange(fn func(bool)) Widget {
	w.onExpandedChange = fn
	return w
}

func (w Widget) Theme(fn func(*theme.Theme)) Widget {
	w.theme = fn
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, w.theme); restore != nil {
		defer restore()
	}
	state := collapsibleStateFor(ctx, w.key)
	disabled := w.disabled || !gtx.Enabled()
	presses := activePresses(&state.item)
	if !disabled {
		for state.item.clickable.Clicked(gtx) {
			if w.onExpandedChange != nil {
				w.onExpandedChange(!w.expanded)
			}
		}
		focusOnPress(ctx, &state.item, presses)
	}

	restore := frame.PushKey(ctx, w.key)
	defer restore()
	return layoutItem(ctx, gtx, &state.item, Item{
		Key:      w.key,
		Label:    w.label,
		Leading:  w.leading,
		Trailing: w.trailing,
		Content:  w.content,
		Disabled: w.disabled,
	}, w.expanded, disabled)
}

// GroupWidget coordinates a controlled collection of collapsible sections.
type GroupWidget struct {
	theme                 func(*theme.Theme)
	key                   string
	expandedKeys          []string
	items                 []Item
	allowMultipleExpanded bool
	disabled              bool
	onExpandedChange      func([]string)
}

// CollapsibleGroup creates a controlled group that allows one expanded item by default.
func CollapsibleGroup(key string, expandedKeys []string, items []Item) GroupWidget {
	return GroupWidget{
		key:          key,
		expandedKeys: append([]string(nil), expandedKeys...),
		items:        append([]Item(nil), items...),
	}
}

func (g GroupWidget) AllowMultipleExpanded(allow bool) GroupWidget {
	g.allowMultipleExpanded = allow
	return g
}

func (g GroupWidget) Disabled(disabled bool) GroupWidget {
	g.disabled = disabled
	return g
}

func (g GroupWidget) OnExpandedChange(fn func([]string)) GroupWidget {
	g.onExpandedChange = fn
	return g
}

func (g GroupWidget) Theme(fn func(*theme.Theme)) GroupWidget {
	g.theme = fn
	return g
}

func (g GroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, g.theme); restore != nil {
		defer restore()
	}
	state := collapsibleStateFor(ctx, g.key)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(g.items)

	for _, item := range g.items {
		itemState := state.itemFor(item.Key)
		disabled := g.disabled || item.Disabled || !gtx.Enabled()
		presses := activePresses(itemState)
		if disabled {
			continue
		}
		for itemState.clickable.Clicked(gtx) {
			if g.onExpandedChange != nil {
				g.onExpandedChange(toggleExpandedKeys(g.expandedKeys, item.Key, g.allowMultipleExpanded))
			}
		}
		focusOnPress(ctx, itemState, presses)
	}
	if !g.disabled && gtx.Enabled() {
		if focusKey := state.updateKeys(gtx, g.items); focusKey != "" {
			frame.RequestFocus(ctx, &state.itemFor(focusKey).clickable)
		}
	}

	restore := frame.PushKey(ctx, g.key)
	defer restore()
	return g.layout(ctx, gtx, state)
}

func (g GroupWidget) layout(ctx *frame.Context, gtx layout.Context, state *collapsibleState) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	children := make([]layout.FlexChild, len(g.items))
	placements := make([]frame.OverlayPlacement, len(g.items))
	dimensions := make([]layout.Dimensions, len(g.items))
	for index, item := range g.items {
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			restore := frame.PushKey(ctx, item.Key)
			defer restore()
			dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				disabled := g.disabled || item.Disabled || !gtx.Enabled()
				return layoutItem(ctx, gtx, state.itemFor(item.Key), item, slices.Contains(g.expandedKeys, item.Key), disabled)
			})
			dimensions[index], placements[index] = dims, placement
			return dims
		})
	}
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	y := 0
	for index := range placements {
		placements[index].PlaceOffset(image.Pt(0, y))
		y += dimensions[index].Size.Y
	}
	return dims
}

func toggleExpandedKeys(keys []string, key string, allowMultiple bool) []string {
	if slices.Contains(keys, key) {
		next := make([]string, 0, len(keys)-1)
		for _, current := range keys {
			if current != key {
				next = append(next, current)
			}
		}
		return next
	}
	if !allowMultiple {
		return []string{key}
	}
	next := append([]string(nil), keys...)
	return append(next, key)
}

func validateItems(items []Item) {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty collapsible item key")
		}
		if _, ok := seen[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate collapsible item key %q", item.Key))
		}
		seen[item.Key] = struct{}{}
	}
}
