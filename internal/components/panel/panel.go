// Package panel provides lifecycle-aware hosts for mutually exclusive views.
//
// A Host lays out one visible panel at a time. Hidden panels are deliberately
// kept out of the paint, input, focus, and semantic streams. Callers can opt
// into retaining their transient widget state or initializing them lazily in
// an isolated hidden pass. The host does not render a tab strip; Tabs, a dock
// header, or an application-specific navigation control can own selection.
package panel

import (
	"fmt"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

const stateSlotPanelHost = "panel-host"

// Item describes one view hosted by a Host.
type Item struct {
	Key      string
	Content  frame.Widget
	Disabled bool
}

// Host presents one selected panel and owns its hidden-panel lifecycle.
type HostWidget struct {
	key                string
	selectedKey        string
	hasSelectedKey     bool
	defaultSelectedKey string
	hasDefaultSelected bool
	items              []Item
	onChange           func(string)
	keepAlive          bool
	destroyOnHidden    bool
	forceRender        bool
	disabled           bool
	customStyle        flowstyle.Style
}

type hostState struct {
	disclosure      disclosure.Binding[string]
	retainedPanels  map[string]struct{}
	renderedPanels  map[string]struct{}
	itemKeys        map[string]struct{}
	forceRender     bool
	destroyOnHidden bool
}

// Host creates a lifecycle-aware panel host. A non-empty selectedKey uses
// controlled mode; an empty selectedKey uses uncontrolled mode and may be
// seeded with DefaultSelectedKey.
func Host(key, selectedKey string, items []Item) HostWidget {
	return HostWidget{
		key:            key,
		selectedKey:    selectedKey,
		hasSelectedKey: selectedKey != "",
		items:          append([]Item(nil), items...),
	}
}

// ViewStack is an alias-style constructor for callers that prefer the view
// stack terminology. It has the same lifecycle contract as Host.
func ViewStack(key, selectedKey string, items []Item) HostWidget {
	return Host(key, selectedKey, items)
}

func (h HostWidget) SelectedKey(key string) HostWidget {
	h.selectedKey = key
	h.hasSelectedKey = true
	return h
}

func (h HostWidget) DefaultSelectedKey(key string) HostWidget {
	h.defaultSelectedKey = key
	h.hasDefaultSelected = true
	return h
}

func (h HostWidget) OnChange(fn func(string)) HostWidget {
	h.onChange = fn
	return h
}

// BindChange composes an observer with the existing selection callback. It
// lets a Workbench controller link PanelHost selection without replacing the
// application's own model callback.
func (h HostWidget) BindChange(fn func(string)) HostWidget {
	if fn == nil {
		return h
	}
	previous := h.onChange
	h.onChange = func(key string) {
		if previous != nil {
			previous(key)
		}
		fn(key)
	}
	return h
}

// KeepAlive retains transient state created by hidden panels without laying
// them out again. It is ignored when DestroyOnHidden is enabled.
func (h HostWidget) KeepAlive(enabled bool) HostWidget {
	h.keepAlive = enabled
	return h
}

// ForceRender initializes each panel once in a hidden, disabled layout pass.
// The initialized state is retained until the item is removed or
// DestroyOnHidden is enabled.
func (h HostWidget) ForceRender(enabled bool) HostWidget {
	h.forceRender = enabled
	return h
}

// Lazy is the inverse spelling of ForceRender. Lazy(true) is the default.
func (h HostWidget) Lazy(enabled bool) HostWidget {
	h.forceRender = !enabled
	return h
}

// DestroyOnHidden releases a panel's retained state whenever it stops being
// selected. It takes precedence over KeepAlive and ForceRender.
func (h HostWidget) DestroyOnHidden(enabled bool) HostWidget {
	h.destroyOnHidden = enabled
	return h
}

func (h HostWidget) Disabled(disabled bool) HostWidget {
	h.disabled = disabled
	return h
}

func (h HostWidget) Style(value flowstyle.Style) HostWidget {
	h.customStyle = value
	return h
}

func (h HostWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	rootKey := frame.ClaimKey(ctx, state.KindPanelHost, h.key)
	value := frame.UseStateWith[hostState](ctx, rootKey, stateSlotPanelHost, func() *hostState {
		return &hostState{}
	})
	h.begin(value)
	h.checkItems(value)

	cfg := disclosure.Config[string]{
		Controlled: h.hasSelectedKey,
		Value:      h.selectedKey,
		HasDefault: h.hasDefaultSelected,
		Default:    h.defaultSelectedKey,
		OnChange:   h.onChange,
	}
	value.disclosure.Bind(cfg)
	selectedKey := value.disclosure.Current(cfg)
	selectedKey = h.normalizeSelection(value, cfg, selectedKey)
	if h.forceRender && !h.destroyOnHidden && selectedKey != "" {
		value.renderedPanels[selectedKey] = struct{}{}
	}
	h.syncLifecycle(ctx, value)

	if h.forceRender && !h.destroyOnHidden {
		h.layoutForcedPanels(ctx, gtx, value, selectedKey)
	}

	item, ok := h.item(selectedKey)
	if ok && item.Content != nil {
		h.preparePanelAssociations(ctx, item)
	}
	return layoutui.LayoutStyled(ctx, gtx, rootKey, flowstyle.StyleState{
		Disabled: h.disabled || !gtx.Enabled(),
		Selected: selectedKey != "",
	}, h.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if !ok || item.Content == nil {
			return layout.Dimensions{Size: gtx.Constraints.Constrain(gtx.Constraints.Min)}
		}
		if h.disabled {
			gtx = gtx.Disabled()
		}
		return h.layoutPanel(ctx, gtx, item)
	}))
}

func (h HostWidget) begin(value *hostState) {
	if value.retainedPanels == nil {
		value.retainedPanels = make(map[string]struct{}, len(h.items))
	}
	if value.renderedPanels == nil {
		value.renderedPanels = make(map[string]struct{}, len(h.items))
	}
	if value.itemKeys == nil {
		value.itemKeys = make(map[string]struct{}, len(h.items))
	}
	if value.forceRender != h.forceRender || value.destroyOnHidden != h.destroyOnHidden {
		clear(value.renderedPanels)
		value.forceRender = h.forceRender
		value.destroyOnHidden = h.destroyOnHidden
	}
}

func (h HostWidget) checkItems(value *hostState) {
	clear(value.itemKeys)
	for _, item := range h.items {
		if item.Key == "" {
			panic("flowui: empty panel key")
		}
		if _, exists := value.itemKeys[item.Key]; exists {
			panic(fmt.Sprintf("flowui: duplicate panel key %q", item.Key))
		}
		value.itemKeys[item.Key] = struct{}{}
	}
}

func (h HostWidget) syncLifecycle(ctx *frame.Context, value *hostState) {
	for _, item := range h.items {
		scope := h.panelStateScope(ctx, item.Key)
		keep := h.keepAlive && !h.destroyOnHidden
		if !keep && h.forceRender && !h.destroyOnHidden {
			_, keep = value.renderedPanels[item.Key]
		}
		if keep {
			frame.RetainState(ctx, scope)
			value.retainedPanels[item.Key] = struct{}{}
			continue
		}
		frame.ReleaseStateRetention(ctx, scope)
		delete(value.retainedPanels, item.Key)
	}
	for itemKey := range value.retainedPanels {
		if _, exists := value.itemKeys[itemKey]; exists {
			continue
		}
		frame.ReleaseStateRetention(ctx, h.panelStateScope(ctx, itemKey))
		delete(value.retainedPanels, itemKey)
	}
	for itemKey := range value.renderedPanels {
		if _, exists := value.itemKeys[itemKey]; exists {
			continue
		}
		delete(value.renderedPanels, itemKey)
	}
}

func (h HostWidget) normalizeSelection(value *hostState, cfg disclosure.Config[string], selected string) string {
	if item, ok := h.item(selected); ok && !item.Disabled {
		return selected
	}
	for _, item := range h.items {
		if item.Disabled {
			continue
		}
		if selected != item.Key {
			value.disclosure.Request(cfg, item.Key)
		}
		if h.hasSelectedKey {
			return item.Key
		}
		return value.disclosure.Current(cfg)
	}
	return ""
}

func (h HostWidget) item(key string) (Item, bool) {
	for _, item := range h.items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (h HostWidget) panelStateScope(ctx *frame.Context, itemKey string) string {
	return frame.DerivedKey(ctx, frame.FullKey(ctx, h.key), "panel-state:"+itemKey)
}

func (h HostWidget) layoutForcedPanels(ctx *frame.Context, gtx layout.Context, value *hostState, selectedKey string) {
	for _, item := range h.items {
		if item.Key == selectedKey || item.Content == nil {
			continue
		}
		if _, rendered := value.renderedPanels[item.Key]; rendered {
			continue
		}
		scope := h.panelStateScope(ctx, item.Key)
		frame.LayoutHidden(ctx, gtx, scope, frame.WidgetFunc(func(ctx *frame.Context, hiddenGtx layout.Context) layout.Dimensions {
			return h.layoutPanel(ctx, hiddenGtx, item)
		}))
		value.renderedPanels[item.Key] = struct{}{}
	}
}

func (h HostWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, item Item) layout.Dimensions {
	scope := h.panelStateScope(ctx, item.Key)
	restoreHost := frame.PushKey(ctx, h.key)
	defer restoreHost()
	restoreItem := frame.PushKey(ctx, item.Key)
	defer restoreItem()
	if (h.keepAlive || h.forceRender) && !h.destroyOnHidden {
		restoreRetention := frame.PushStateRetention(ctx, scope)
		defer restoreRetention()
	}
	return item.Content.Layout(ctx, gtx)
}

func (h HostWidget) preparePanelAssociations(ctx *frame.Context, item Item) {
	restoreHost := frame.PushKey(ctx, h.key)
	defer restoreHost()
	restoreItem := frame.PushKey(ctx, item.Key)
	defer restoreItem()
	layoutui.PrepareFieldAssociations(ctx, item.Content)
}
