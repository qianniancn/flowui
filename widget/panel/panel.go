// Package panel provides lifecycle-aware hosts for mutually exclusive views.
//
// A Widget lays out one visible panel at a time. Hidden panels are deliberately
// kept out of the paint, input, focus, and semantic streams. Callers can opt
// into retaining their transient widget state or initializing them lazily in
// an isolated hidden pass. The host does not render a tab strip; Tabs, a dock
// header, or an application-specific navigation control can own selection.
package panel

import (
	"fmt"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotPanelHost = "panel-host"

// Item describes one view hosted by a Widget.
type Item struct {
	Key      string
	Content  core.Widget
	Disabled bool
}

// Widget presents one selected panel and owns its hidden-panel lifecycle.
type Widget struct {
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
}

type hostState struct {
	disclosure      disclosure.Binding[string]
	retainedPanels  map[string]struct{}
	renderedPanels  map[string]struct{}
	itemKeys        map[string]struct{}
	forceRender     bool
	destroyOnHidden bool
}

// New creates a lifecycle-aware panel host. A non-empty selectedKey uses
// controlled mode; an empty selectedKey uses uncontrolled mode and may be
// seeded with DefaultSelectedKey.
func New(key, selectedKey string, items []Item) Widget {
	return Widget{
		key:            key,
		selectedKey:    selectedKey,
		hasSelectedKey: selectedKey != "",
		items:          append([]Item(nil), items...),
	}
}

func (h Widget) SelectedKey(key string) Widget {
	h.selectedKey = key
	h.hasSelectedKey = true
	return h
}

func (h Widget) DefaultSelectedKey(key string) Widget {
	h.defaultSelectedKey = key
	h.hasDefaultSelected = true
	return h
}

func (h Widget) OnChange(fn func(string)) Widget {
	h.onChange = fn
	return h
}

// BindChange composes an observer with the existing selection callback. It
// lets a Workbench controller link PanelHost selection without replacing the
// application's own model callback.
func (h Widget) BindChange(fn func(string)) Widget {
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
func (h Widget) KeepAlive(enabled bool) Widget {
	h.keepAlive = enabled
	return h
}

// ForceRender initializes each panel once in a hidden, disabled layout pass.
// The initialized state is retained until the item is removed or
// DestroyOnHidden is enabled.
func (h Widget) ForceRender(enabled bool) Widget {
	h.forceRender = enabled
	return h
}

// Lazy is the inverse spelling of ForceRender. Lazy(true) is the default.
func (h Widget) Lazy(enabled bool) Widget {
	h.forceRender = !enabled
	return h
}

// DestroyOnHidden releases a panel's retained state whenever it stops being
// selected. It takes precedence over KeepAlive and ForceRender.
func (h Widget) DestroyOnHidden(enabled bool) Widget {
	h.destroyOnHidden = enabled
	return h
}

func (h Widget) Disabled(disabled bool) Widget {
	h.disabled = disabled
	return h
}

func (h Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	rootKey := core.ClaimKey(ctx, core.KindPanelHost, h.key)
	value := core.UseStateWith[hostState](ctx, rootKey, stateSlotPanelHost, func() *hostState {
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
	if !ok || item.Content == nil {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(gtx.Constraints.Min)}
	}
	if h.disabled {
		gtx = gtx.Disabled()
	}
	return h.layoutPanel(ctx, gtx, item)
}

func (h Widget) begin(value *hostState) {
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

func (h Widget) checkItems(value *hostState) {
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

func (h Widget) syncLifecycle(ctx *core.Context, value *hostState) {
	for _, item := range h.items {
		scope := h.panelStateScope(ctx, item.Key)
		keep := h.keepAlive && !h.destroyOnHidden
		if !keep && h.forceRender && !h.destroyOnHidden {
			_, keep = value.renderedPanels[item.Key]
		}
		if keep {
			core.RetainState(ctx, scope)
			value.retainedPanels[item.Key] = struct{}{}
			continue
		}
		core.ReleaseStateRetention(ctx, scope)
		delete(value.retainedPanels, item.Key)
	}
	for itemKey := range value.retainedPanels {
		if _, exists := value.itemKeys[itemKey]; exists {
			continue
		}
		core.ReleaseStateRetention(ctx, h.panelStateScope(ctx, itemKey))
		delete(value.retainedPanels, itemKey)
	}
	for itemKey := range value.renderedPanels {
		if _, exists := value.itemKeys[itemKey]; exists {
			continue
		}
		delete(value.renderedPanels, itemKey)
	}
}

func (h Widget) normalizeSelection(value *hostState, cfg disclosure.Config[string], selected string) string {
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

func (h Widget) item(key string) (Item, bool) {
	for _, item := range h.items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (h Widget) panelStateScope(ctx *core.Context, itemKey string) string {
	return core.DerivedKey(ctx, core.FullKey(ctx, h.key), "panel-state:"+itemKey)
}

func (h Widget) layoutForcedPanels(ctx *core.Context, gtx layout.Context, value *hostState, selectedKey string) {
	for _, item := range h.items {
		if item.Key == selectedKey || item.Content == nil {
			continue
		}
		if _, rendered := value.renderedPanels[item.Key]; rendered {
			continue
		}
		scope := h.panelStateScope(ctx, item.Key)
		core.LayoutHidden(ctx, gtx, scope, core.WidgetFunc(func(ctx *core.Context, hiddenGtx layout.Context) layout.Dimensions {
			return h.layoutPanel(ctx, hiddenGtx, item)
		}))
		value.renderedPanels[item.Key] = struct{}{}
	}
}

func (h Widget) layoutPanel(ctx *core.Context, gtx layout.Context, item Item) layout.Dimensions {
	scope := h.panelStateScope(ctx, item.Key)
	restoreHost := core.PushKey(ctx, h.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, item.Key)
	defer restoreItem()
	if (h.keepAlive || h.forceRender) && !h.destroyOnHidden {
		restoreRetention := core.PushStateRetention(ctx, scope)
		defer restoreRetention()
	}
	return item.Content.Layout(ctx, gtx)
}

func (h Widget) preparePanelAssociations(ctx *core.Context, item Item) {
	restoreHost := core.PushKey(ctx, h.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, item.Key)
	defer restoreItem()
	layoutui.PrepareFieldAssociations(ctx, item.Content)
}
