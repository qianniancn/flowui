package collapsible

import (
	"fmt"
	"image"
	"slices"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Item describes one entry in a Group.
type Item struct {
	Key      string
	Label    string
	Leading  core.Widget
	Trailing core.Widget
	Content  core.Widget
	Disabled bool
}

// Widget presents one controlled expandable section.
type Widget struct {
	override           *theme.CollapsibleOverride
	key                string
	expanded           bool
	hasExpanded        bool
	defaultExpanded    bool
	hasDefaultExpanded bool
	label              string
	content            core.Widget
	leading            core.Widget
	trailing           core.Widget
	disabled           bool
	onExpandedChange   func(bool)
}

// New creates one expandable section. When expanded is true, it starts
// in controlled mode with the section expanded. When expanded is false, it starts
// in uncontrolled mode with the section collapsed (use DefaultExpanded(true) for
// uncontrolled mode starting expanded).
// New creates a collapsible section. It starts uncontrolled and collapsed;
// Expanded sets the controlled value and DefaultExpanded the starting one.
func New(label string, content core.Widget) Widget {
	return Widget{
		key:     core.AutoKey(),
		label:   label,
		content: content,
	}
}

// Key sets an explicit identity for the section state.
func (w Widget) Key(key string) Widget {
	w.key = key
	return w
}

func (w Widget) Leading(leading core.Widget) Widget {
	w.leading = leading
	return w
}

func (w Widget) Trailing(trailing core.Widget) Widget {
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

// Expanded sets the collapsible to controlled mode with the given expanded state.
func (w Widget) Expanded(expanded bool) Widget {
	w.expanded = expanded
	w.hasExpanded = true
	return w
}

// DefaultExpanded sets the initial expanded state for uncontrolled mode.
func (w Widget) DefaultExpanded(expanded bool) Widget {
	w.defaultExpanded = expanded
	w.hasDefaultExpanded = true
	return w
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

// Style merges instance declarations into the section. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.CollapsibleOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}

// StyleTrigger styles the internal disclosure button.
func (w Widget) StyleTrigger(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.CollapsibleOverride{}
	}
	w.override.Trigger = style.Join(w.override.Trigger, declaration)
	return w
}

// StyleWithTrigger styles the section and its disclosure button together.
func (w Widget) StyleWithTrigger(root, trigger style.Style) Widget {
	if w.override == nil {
		w.override = &theme.CollapsibleOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, root)
	w.override.Trigger = style.Join(w.override.Trigger, trigger)
	return w
}

// Style merges instance declarations into the group.
func (g GroupWidget) Style(values ...style.Style) GroupWidget {
	declaration := style.Join(values...)
	if g.override == nil {
		g.override = &theme.CollapsibleOverride{}
	}
	g.override.Declaration = style.Join(g.override.Declaration, declaration)
	return g
}

// StyleWithTrigger styles every item root and its disclosure button.
func (g GroupWidget) StyleWithTrigger(root, trigger style.Style) GroupWidget {
	if g.override == nil {
		g.override = &theme.CollapsibleOverride{}
	}
	g.override.Declaration = style.Join(g.override.Declaration, root)
	g.override.Trigger = style.Join(g.override.Trigger, trigger)
	return g
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.CollapsibleOverride) layout.Dimensions {
	rootKey := core.FullKey(ctx, w.key)
	state := collapsibleStateFor(ctx, w.key)
	visual := state.visual(ctx, override)
	state.item.bind(w)
	expanded := state.item.isExpanded(w)

	disabled := w.disabled || !gtx.Enabled()
	presses := activePresses(&state.item)
	if !disabled {
		for state.item.clickable.Clicked(gtx) {
			state.item.requestExpanded(w, !expanded)
		}
		focusOnPress(ctx, &state.item, presses)
	}

	restore := core.PushKey(ctx, w.key)
	defer restore()
	focused := gtx.Focused(&state.item.clickable)
	return layoutui.Styled(ctx, gtx, rootKey, style.State{
		Hovered:      state.item.clickable.Hovered(),
		Pressed:      state.item.clickable.Pressed(),
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &state.item.clickable, focused),
		Disabled:     disabled,
		Expanded:     expanded,
	}, visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutItem(ctx, gtx, &state.item, Item{
			Key:      w.key,
			Label:    w.label,
			Leading:  w.leading,
			Trailing: w.trailing,
			Content:  w.content,
			Disabled: w.disabled,
		}, expanded, disabled, visual)
	}))
}

// GroupWidget coordinates a controlled collection of collapsible sections.
type GroupWidget struct {
	override              *theme.CollapsibleOverride
	key                   string
	expandedKeys          []string
	items                 []Item
	allowMultipleExpanded bool
	disabled              bool
	onExpandedChange      func([]string)
}

// Group creates a controlled group that allows one expanded item by default.
func Group(key string, expandedKeys []string, items []Item) GroupWidget {
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

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return g.layoutWithOverride(ctx, gtx, g.override)
}

func (g GroupWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.CollapsibleOverride) layout.Dimensions {
	rootKey := core.FullKey(ctx, g.key)
	state := collapsibleStateFor(ctx, g.key)
	visual := state.visual(ctx, override)
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
			core.RequestFocus(ctx, &state.itemFor(focusKey).clickable)
		}
	}

	restore := core.PushKey(ctx, g.key)
	defer restore()
	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, item := range g.items {
		itemState := state.itemFor(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		itemFocused := gtx.Focused(&itemState.clickable)
		focused = focused || itemFocused
		focusVisible = focusVisible || core.FocusVisible(ctx, &itemState.clickable, itemFocused)
	}
	return layoutui.Styled(ctx, gtx, rootKey, style.State{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     g.disabled || !gtx.Enabled(),
		Expanded:     len(g.expandedKeys) > 0,
	}, visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return g.layout(ctx, gtx, state, visual)
	}))
}

func (g GroupWidget) layout(ctx *core.Context, gtx layout.Context, state *collapsibleState, visual theme.CollapsibleVisual) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	children := make([]layout.FlexChild, len(g.items))
	placements := make([]core.OverlayPlacement, len(g.items))
	dimensions := make([]layout.Dimensions, len(g.items))
	for index, item := range g.items {
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			restore := core.PushKey(ctx, item.Key)
			defer restore()
			dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				disabled := g.disabled || item.Disabled || !gtx.Enabled()
				return layoutItem(ctx, gtx, state.itemFor(item.Key), item, slices.Contains(g.expandedKeys, item.Key), disabled, visual)
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
