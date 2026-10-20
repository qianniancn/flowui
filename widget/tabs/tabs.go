package tabs

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type TabItem struct {
	Key             string
	Label           string
	AccessibleLabel string
	Description     string
	Icon            []byte
	Content         core.Widget
	Leading         core.Widget
	Trailing        core.Widget
	Panel           core.Widget
	Disabled        bool
	Closable        bool
	Editable        bool
}

type Variant int

const (
	Primary Variant = iota
	Secondary
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// Placement controls which side of the panel owns the tab strip.
type Placement int

const (
	Top Placement = iota
	Bottom
	Start
	End
)

type Size int

const (
	Medium Size = iota
	Small
	Large
)

type Color int

const (
	ColorDefault Color = iota
	ColorAccent
)

// ActivationMode controls whether arrow-key navigation changes selection
// immediately or only moves focus until an activation key is pressed.
type ActivationMode int

const (
	ActivationAutomatic ActivationMode = iota
	ActivationManual
)

// PanelTransition controls how the selected panel enters the layout.
// PanelNone preserves the immediate switch behavior; PanelFade fades
// the newly selected panel in while it remains the only interactive panel.
type PanelTransition int

const (
	PanelNone PanelTransition = iota
	PanelFade
)

// IndicatorAlign controls the selected indicator's position along the
// tab's main axis when IndicatorWidth is narrower than the tab.
type IndicatorAlign int

const (
	IndicatorStart IndicatorAlign = iota
	IndicatorCenter
	IndicatorEnd
)

// OverflowMode controls how a tab strip behaves when its items do not fit
// in the available main-axis space. Scroll is the default; Menu replaces
// hidden items with a More menu; Auto uses the menu only when the strip
// actually overflows.
type OverflowMode int

const (
	OverflowScroll OverflowMode = iota
	OverflowMenu
	OverflowAuto
)

type Widget struct {
	override            *theme.TabsOverride
	key                 string
	selectedKey         string
	hasSelectedKey      bool
	defaultSelectedKey  string
	hasDefaultSelected  bool
	items               []TabItem
	onChange            func(string)
	variant             Variant
	orientation         Orientation
	placement           Placement
	size                Size
	color               Color
	disabled            bool
	separators          bool
	fit                 bool
	centered            bool
	leading             core.Widget
	trailing            core.Widget
	onClose             func(string)
	closeOnHover        bool
	onAdd               func()
	onEdit              func(string, string)
	editingKey          string
	hasEditingKey       bool
	overflowMode        OverflowMode
	moreLabel           string
	moreContent         core.Widget
	keepAlive           bool
	destroyOnHidden     bool
	forceRender         bool
	activation          ActivationMode
	panelTransition     PanelTransition
	panelDuration       time.Duration
	indicatorAlign      IndicatorAlign
	indicatorAlignSet   bool
	indicatorWidth      unit.Dp
	indicatorVisible    bool
	indicatorVisibleSet bool
	animated            bool
	animationSet        bool
	animationDuration   time.Duration
	indicatorDuration   time.Duration
	visualCache         *tabsVisualCache
}

// New creates a tabs widget. When selectedKey is non-empty, it starts in
// controlled mode with that tab selected. When selectedKey is empty, it starts
// in uncontrolled mode (use DefaultSelectedKey to set initial selection).
func New(key, selectedKey string, items []TabItem) Widget {
	if selectedKey != "" {
		// non-empty selectedKey → controlled mode
		return Widget{
			key:            key,
			selectedKey:    selectedKey,
			hasSelectedKey: true,
			items:          items,
			visualCache:    &tabsVisualCache{},
		}
	}
	// empty selectedKey → uncontrolled mode
	return Widget{
		key:         key,
		items:       items,
		visualCache: &tabsVisualCache{},
	}
}

// Uncontrolled uses the current selection only as the initial internal value.
func (t Widget) Uncontrolled() Widget {
	t.defaultSelectedKey = t.selectedKey
	t.hasDefaultSelected = true
	t.hasSelectedKey = false
	return t
}

func (t Widget) OnChange(fn func(string)) Widget {
	t.onChange = fn
	return t
}

// BindChange composes an observer with the existing selection callback. It is
// intended for shell controllers that need to observe Tabs without taking
// ownership away from the application view model.
func (t Widget) BindChange(fn func(string)) Widget {
	if fn == nil {
		return t
	}
	previous := t.onChange
	t.onChange = func(key string) {
		if previous != nil {
			previous(key)
		}
		fn(key)
	}
	return t
}

// SelectedKey sets the tabs to controlled mode with the given selected key.
func (t Widget) SelectedKey(key string) Widget {
	t.selectedKey = key
	t.hasSelectedKey = true
	return t
}

// DefaultSelectedKey sets the initial selected key for uncontrolled mode.
func (t Widget) DefaultSelectedKey(key string) Widget {
	t.defaultSelectedKey = key
	t.hasDefaultSelected = true
	return t
}

func (t Widget) Variant(variant Variant) Widget {
	t.variant = variant
	return t
}

func (t Widget) Orientation(orientation Orientation) Widget {
	t.orientation = orientation
	if orientation == Vertical {
		t.placement = Start
	} else {
		t.placement = Top
	}
	return t
}

func (t Widget) Vertical() Widget {
	t.orientation = Vertical
	t.placement = Start
	return t
}

// Placement controls whether the strip is above, below, before, or after the
// panel. It also updates the strip orientation to match the placement.
func (t Widget) Placement(placement Placement) Widget {
	t.placement = placement
	if placement == Start || placement == End {
		t.orientation = Vertical
	} else {
		t.orientation = Horizontal
	}
	return t
}

func (t Widget) Size(size Size) Widget {
	t.size = size
	return t
}

func (t Widget) Color(color Color) Widget {
	t.color = color
	return t
}

func (t Widget) Fit() Widget {
	t.fit = true
	return t
}

// Centered keeps the tab labels at their natural width and centers a
// horizontal strip in its available bar area when there is extra room.
func (t Widget) Centered(enabled bool) Widget {
	t.centered = enabled
	return t
}

func (t Widget) Disabled(disabled bool) Widget {
	t.disabled = disabled
	return t
}

func (t Widget) Separators(visible bool) Widget {
	t.separators = visible
	return t
}

// Leading places a widget before the tab strip. For vertical tabs it is laid
// out above the strip.
func (t Widget) Leading(widget core.Widget) Widget {
	t.leading = widget
	return t
}

// Trailing places a widget after the tab strip. For vertical tabs it is laid
// out below the strip.
func (t Widget) Trailing(widget core.Widget) Widget {
	t.trailing = widget
	return t
}

// OnClose handles close actions for items marked Closable. The caller owns the
// item slice and should remove the item from the next model update.
func (t Widget) OnClose(fn func(string)) Widget {
	t.onClose = fn
	return t
}

// CloseOnHover keeps closable tab controls hidden until the pointer enters
// the tab. The close slot remains reserved so hovering does not move the tab
// strip or its selection indicator.
func (t Widget) CloseOnHover(enabled bool) Widget {
	t.closeOnHover = enabled
	return t
}

// OnAdd adds an inline add-tab action after the tab strip. The callback owns
// the item model and should append the new item before the next core.
func (t Widget) OnAdd(fn func()) Widget {
	t.onAdd = fn
	return t
}

// OnEdit receives a committed inline label edit. The caller owns updating the
// corresponding TabItem.Label in the next model update.
func (t Widget) OnEdit(fn func(string, string)) Widget {
	t.onEdit = fn
	return t
}

// EditingKey puts one editable tab into inline editing mode. Pass an empty
// key to leave controlled editing mode.
func (t Widget) EditingKey(key string) Widget {
	t.editingKey = key
	t.hasEditingKey = true
	return t
}

// Overflow selects the fallback used when the tab strip cannot show all
// items. The default is OverflowScroll.
func (t Widget) Overflow(mode OverflowMode) Widget {
	t.overflowMode = mode
	return t
}

// MoreLabel customizes the accessible label and visible text of the overflow
// trigger. An empty value restores the default "More" label.
func (t Widget) MoreLabel(label string) Widget {
	t.moreLabel = label
	return t
}

// OverflowTrigger replaces the default More trigger content. The widget is
// measured before overflow is calculated and remains inside the built-in
// Dropdown trigger, so its accessible label and keyboard behavior are kept.
func (t Widget) OverflowTrigger(widget core.Widget) Widget {
	t.moreContent = widget
	return t
}

// ForceRender initializes every panel the first time the Tabs widget sees it,
// while keeping inactive panels out of the visible operation, input and
// semantic streams. The initialized panel state is retained until its item is
// removed or DestroyOnHidden is enabled. DestroyOnHidden takes precedence.
func (t Widget) ForceRender(enabled bool) Widget {
	t.forceRender = enabled
	return t
}

// Lazy explicitly controls whether inactive panels are initialized only on
// first selection. Lazy(true) is the default and is equivalent to
// ForceRender(false); Lazy(false) is equivalent to ForceRender(true).
func (t Widget) Lazy(enabled bool) Widget {
	t.forceRender = !enabled
	return t
}

// Activation selects automatic (default) or manual keyboard activation.
// Manual activation moves focus with arrow keys and commits the focused tab on
// Enter, Return, or Space.
func (t Widget) Activation(mode ActivationMode) Widget {
	t.activation = mode
	return t
}

// PanelTransition selects the selected-panel entry animation.
func (t Widget) PanelTransition(transition PanelTransition) Widget {
	t.panelTransition = transition
	return t
}

// PanelAnimationDuration overrides the selected-panel transition duration.
func (t Widget) PanelAnimationDuration(duration time.Duration) Widget {
	t.panelDuration = maxDuration(duration)
	return t
}

// IndicatorAlign positions a narrowed selected indicator along the tab's
// main axis. It has no effect while the indicator uses the full tab width.
func (t Widget) IndicatorAlign(align IndicatorAlign) Widget {
	t.indicatorAlign = align
	t.indicatorAlignSet = true
	return t
}

// IndicatorWidth limits the selected indicator length. Zero restores the
// theme's automatic length.
func (t Widget) IndicatorWidth(width unit.Dp) Widget {
	if width < 0 {
		width = 0
	}
	t.indicatorWidth = width
	return t
}

// IndicatorVisible controls whether the selected tab indicator is painted.
// It defaults to true for the line and primary variants.
func (t Widget) IndicatorVisible(visible bool) Widget {
	t.indicatorVisible = visible
	t.indicatorVisibleSet = true
	return t
}

// KeepAlive retains state created by hidden panels without laying them out or
// registering their input and semantic operations.
func (t Widget) KeepAlive(enabled bool) Widget {
	t.keepAlive = enabled
	return t
}

// DestroyOnHidden releases a panel's retained state when it is not selected.
// It takes precedence over KeepAlive when both are enabled.
func (t Widget) DestroyOnHidden(enabled bool) Widget {
	t.destroyOnHidden = enabled
	return t
}

// Animated controls label, indicator, and selected-panel transitions. The
// default is true.
func (t Widget) Animated(enabled bool) Widget {
	t.animated = enabled
	t.animationSet = true
	return t
}

// AnimationDuration overrides the tab label color transition duration.
func (t Widget) AnimationDuration(duration time.Duration) Widget {
	t.animationDuration = maxDuration(duration)
	return t
}

// IndicatorAnimationDuration overrides the selected indicator transition.
func (t Widget) IndicatorAnimationDuration(duration time.Duration) Widget {
	t.indicatorDuration = maxDuration(duration)
	return t
}

func (t Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return t.layoutWithOverride(ctx, gtx, t.override)
}

func (t Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.TabsOverride) layout.Dimensions {
	visual := t.visualWithCache(ctx, override, t.visualCache)
	return t.layoutWithVisual(ctx, gtx, visual)
}

func (t Widget) layoutWithVisual(ctx *core.Context, gtx layout.Context, visual theme.TabsVisual) layout.Dimensions {
	rootKey := core.FullKey(ctx, t.key)
	state := tabsStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	if state.retainedPanels == nil {
		state.retainedPanels = make(map[string]struct{}, len(t.items))
	}
	if state.renderedPanels == nil {
		state.renderedPanels = make(map[string]struct{}, len(t.items))
	}
	if state.forceRender != t.forceRender || state.destroyOnHidden != t.destroyOnHidden {
		clear(state.renderedPanels)
		state.forceRender = t.forceRender
		state.destroyOnHidden = t.destroyOnHidden
	}
	state.checkItems(t.items)
	if state.editingKey != "" {
		if _, ok := state.itemKeys[state.editingKey]; !ok {
			state.editingKey = ""
		}
	}
	if t.hasEditingKey {
		state.editingKey = t.editingKey
	}
	for _, item := range t.items {
		scope := t.panelStateScope(ctx, item.Key)
		keep := t.panelsKeepAlive()
		if !keep && t.panelsForceRender() {
			_, keep = state.renderedPanels[item.Key]
		}
		if keep {
			core.RetainState(ctx, scope)
			state.retainedPanels[item.Key] = struct{}{}
		} else {
			core.ReleaseStateRetention(ctx, scope)
			delete(state.retainedPanels, item.Key)
		}
	}
	for itemKey := range state.retainedPanels {
		if _, ok := state.itemKeys[itemKey]; ok {
			continue
		}
		core.ReleaseStateRetention(ctx, t.panelStateScope(ctx, itemKey))
		delete(state.retainedPanels, itemKey)
	}
	for itemKey := range state.renderedPanels {
		if _, ok := state.itemKeys[itemKey]; ok {
			continue
		}
		delete(state.renderedPanels, itemKey)
	}

	// Bind disclosure state
	state.bind(t)
	selectedKey := state.currentSelectedKey(t)
	selectedKey = state.normalizeSelection(t, selectedKey)
	if t.panelsForceRender() && selectedKey != "" {
		if item, ok := t.selectedItem(selectedKey); ok && item.Panel != nil {
			state.renderedPanels[selectedKey] = struct{}{}
		}
	}

	state.syncSelection(t.items, selectedKey)
	disabled := t.disabled || !gtx.Enabled()
	if !disabled {
		if selectionKey, focusKey, ok := state.updateKeys(gtx, t.items, selectedKey, t.orientation, t.activation); ok {
			if selectionKey != "" {
				selectedKey = state.requestSelectedKey(t, selectionKey)
			}
			if focusKey != "" {
				core.RequestFocus(ctx, &state.item(focusKey).clickable)
			}
		}
	}
	hovered, pressed, focused := false, false, false
	for _, item := range t.items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		focused = focused || gtx.Focused(&itemState.clickable)
	}
	return layoutui.Styled(ctx, gtx, rootKey, style.State{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  focused,
		Disabled: disabled,
		Selected: selectedKey != "",
	}, style.Join(visual.Root, visual.Override), core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return t.layout(ctx, gtx, state, selectedKey, disabled, visual)
	}))
}

func (t Widget) panelsKeepAlive() bool {
	return t.keepAlive && !t.destroyOnHidden
}

func (t Widget) panelsForceRender() bool {
	return t.forceRender && !t.destroyOnHidden
}

func (t Widget) panelStateScope(ctx *core.Context, itemKey string) string {
	return core.DerivedKey(ctx, core.FullKey(ctx, t.key), "panel-state:"+itemKey)
}

func (t Widget) animationEnabled() bool {
	return !t.animationSet || t.animated
}

func (t Widget) selectionDuration(ctx *core.Context, visual theme.TabsVisual) time.Duration {
	if !t.animationEnabled() {
		return 0
	}
	if t.animationDuration > 0 {
		return t.animationDuration
	}
	return visual.Metrics.ColorDuration
}

func (t Widget) indicatorAnimationDuration(ctx *core.Context, visual theme.TabsVisual) time.Duration {
	if !t.animationEnabled() {
		return 0
	}
	if t.indicatorDuration > 0 {
		return t.indicatorDuration
	}
	return visual.Metrics.IndicatorDuration
}

func (t Widget) panelAnimationDuration(ctx *core.Context, visual theme.TabsVisual) time.Duration {
	if !t.animationEnabled() || t.panelTransition == PanelNone {
		return 0
	}
	if t.panelDuration > 0 {
		return t.panelDuration
	}
	return visual.Metrics.PanelDuration
}

func maxDuration(duration time.Duration) time.Duration {
	if duration < 0 {
		return 0
	}
	return duration
}

func (t Widget) selectedItem(key string) (TabItem, bool) {
	for _, item := range t.items {
		if item.Key == key {
			return item, true
		}
	}
	return TabItem{}, false
}

func (t Widget) usesIntrinsicTabWidths() bool {
	return t.fit || t.centered
}

func (t Widget) closeButtonMetrics(visual theme.TabsVisual) (unit.Dp, unit.Dp) {
	return visual.Metrics.CloseButtonSize, visual.Metrics.CloseButtonGap
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.TabsOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
