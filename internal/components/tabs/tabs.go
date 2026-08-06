package tabs

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type TabItem struct {
	Key             string
	Label           string
	AccessibleLabel string
	Description     string
	Icon            []byte
	Content         frame.Widget
	Leading         frame.Widget
	Trailing        frame.Widget
	Panel           frame.Widget
	Disabled        bool
	Closable        bool
	Editable        bool
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

// TabsPlacement controls which side of the panel owns the tab strip.
type TabsPlacement int

const (
	TabsTop TabsPlacement = iota
	TabsBottom
	TabsStart
	TabsEnd
)

type TabsSize int

const (
	TabsMedium TabsSize = iota
	TabsSmall
	TabsLarge
)

type TabsColor int

const (
	TabsColorDefault TabsColor = iota
	TabsColorAccent
)

// TabsActivationMode controls whether arrow-key navigation changes selection
// immediately or only moves focus until an activation key is pressed.
type TabsActivationMode int

const (
	TabsActivationAutomatic TabsActivationMode = iota
	TabsActivationManual
)

// TabsPanelTransition controls how the selected panel enters the layout.
// TabsPanelNone preserves the immediate switch behavior; TabsPanelFade fades
// the newly selected panel in while it remains the only interactive panel.
type TabsPanelTransition int

const (
	TabsPanelNone TabsPanelTransition = iota
	TabsPanelFade
)

// TabsIndicatorAlign controls the selected indicator's position along the
// tab's main axis when IndicatorWidth is narrower than the tab.
type TabsIndicatorAlign int

const (
	TabsIndicatorStart TabsIndicatorAlign = iota
	TabsIndicatorCenter
	TabsIndicatorEnd
)

// TabsOverflowMode controls how a tab strip behaves when its items do not fit
// in the available main-axis space. Scroll is the default; Menu replaces
// hidden items with a More menu; Auto uses the menu only when the strip
// actually overflows.
type TabsOverflowMode int

const (
	TabsOverflowScroll TabsOverflowMode = iota
	TabsOverflowMenu
	TabsOverflowAuto
)

type TabsWidget struct {
	key                 string
	selectedKey         string
	hasSelectedKey      bool
	defaultSelectedKey  string
	hasDefaultSelected  bool
	items               []TabItem
	onChange            func(string)
	variant             TabsVariant
	orientation         TabsOrientation
	placement           TabsPlacement
	size                TabsSize
	color               TabsColor
	disabled            bool
	separators          bool
	fit                 bool
	centered            bool
	leading             frame.Widget
	trailing            frame.Widget
	onClose             func(string)
	closeOnHover        bool
	onAdd               func()
	onEdit              func(string, string)
	editingKey          string
	hasEditingKey       bool
	overflowMode        TabsOverflowMode
	moreLabel           string
	moreContent         frame.Widget
	keepAlive           bool
	destroyOnHidden     bool
	forceRender         bool
	activation          TabsActivationMode
	panelTransition     TabsPanelTransition
	panelDuration       time.Duration
	indicatorAlign      TabsIndicatorAlign
	indicatorAlignSet   bool
	indicatorWidth      unit.Dp
	indicatorVisible    bool
	indicatorVisibleSet bool
	animated            bool
	animationSet        bool
	animationDuration   time.Duration
	indicatorDuration   time.Duration
	customStyle         flowstyle.Style
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

// BindChange composes an observer with the existing selection callback. It is
// intended for shell controllers that need to observe Tabs without taking
// ownership away from the application view model.
func (t TabsWidget) BindChange(fn func(string)) TabsWidget {
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
	if orientation == TabsVertical {
		t.placement = TabsStart
	} else {
		t.placement = TabsTop
	}
	return t
}

func (t TabsWidget) Vertical() TabsWidget {
	t.orientation = TabsVertical
	t.placement = TabsStart
	return t
}

// Placement controls whether the strip is above, below, before, or after the
// panel. It also updates the strip orientation to match the placement.
func (t TabsWidget) Placement(placement TabsPlacement) TabsWidget {
	t.placement = placement
	if placement == TabsStart || placement == TabsEnd {
		t.orientation = TabsVertical
	} else {
		t.orientation = TabsHorizontal
	}
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

// Centered keeps the tab labels at their natural width and centers a
// horizontal strip in its available bar area when there is extra room.
func (t TabsWidget) Centered(enabled bool) TabsWidget {
	t.centered = enabled
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

// Leading places a widget before the tab strip. For vertical tabs it is laid
// out above the strip.
func (t TabsWidget) Leading(widget frame.Widget) TabsWidget {
	t.leading = widget
	return t
}

// Trailing places a widget after the tab strip. For vertical tabs it is laid
// out below the strip.
func (t TabsWidget) Trailing(widget frame.Widget) TabsWidget {
	t.trailing = widget
	return t
}

// OnClose handles close actions for items marked Closable. The caller owns the
// item slice and should remove the item from the next model update.
func (t TabsWidget) OnClose(fn func(string)) TabsWidget {
	t.onClose = fn
	return t
}

// CloseOnHover keeps closable tab controls hidden until the pointer enters
// the tab. The close slot remains reserved so hovering does not move the tab
// strip or its selection indicator.
func (t TabsWidget) CloseOnHover(enabled bool) TabsWidget {
	t.closeOnHover = enabled
	return t
}

// OnAdd adds an inline add-tab action after the tab strip. The callback owns
// the item model and should append the new item before the next frame.
func (t TabsWidget) OnAdd(fn func()) TabsWidget {
	t.onAdd = fn
	return t
}

// OnEdit receives a committed inline label edit. The caller owns updating the
// corresponding TabItem.Label in the next model update.
func (t TabsWidget) OnEdit(fn func(string, string)) TabsWidget {
	t.onEdit = fn
	return t
}

// EditingKey puts one editable tab into inline editing mode. Pass an empty
// key to leave controlled editing mode.
func (t TabsWidget) EditingKey(key string) TabsWidget {
	t.editingKey = key
	t.hasEditingKey = true
	return t
}

// Overflow selects the fallback used when the tab strip cannot show all
// items. The default is TabsOverflowScroll.
func (t TabsWidget) Overflow(mode TabsOverflowMode) TabsWidget {
	t.overflowMode = mode
	return t
}

// MoreLabel customizes the accessible label and visible text of the overflow
// trigger. An empty value restores the default "More" label.
func (t TabsWidget) MoreLabel(label string) TabsWidget {
	t.moreLabel = label
	return t
}

// OverflowTrigger replaces the default More trigger content. The widget is
// measured before overflow is calculated and remains inside the built-in
// Dropdown trigger, so its accessible label and keyboard behavior are kept.
func (t TabsWidget) OverflowTrigger(widget frame.Widget) TabsWidget {
	t.moreContent = widget
	return t
}

// ForceRender initializes every panel the first time the Tabs widget sees it,
// while keeping inactive panels out of the visible operation, input and
// semantic streams. The initialized panel state is retained until its item is
// removed or DestroyOnHidden is enabled. DestroyOnHidden takes precedence.
func (t TabsWidget) ForceRender(enabled bool) TabsWidget {
	t.forceRender = enabled
	return t
}

// Lazy explicitly controls whether inactive panels are initialized only on
// first selection. Lazy(true) is the default and is equivalent to
// ForceRender(false); Lazy(false) is equivalent to ForceRender(true).
func (t TabsWidget) Lazy(enabled bool) TabsWidget {
	t.forceRender = !enabled
	return t
}

// Activation selects automatic (default) or manual keyboard activation.
// Manual activation moves focus with arrow keys and commits the focused tab on
// Enter, Return, or Space.
func (t TabsWidget) Activation(mode TabsActivationMode) TabsWidget {
	t.activation = mode
	return t
}

// PanelTransition selects the selected-panel entry animation.
func (t TabsWidget) PanelTransition(transition TabsPanelTransition) TabsWidget {
	t.panelTransition = transition
	return t
}

// PanelAnimationDuration overrides the selected-panel transition duration.
func (t TabsWidget) PanelAnimationDuration(duration time.Duration) TabsWidget {
	t.panelDuration = maxDuration(duration)
	return t
}

// IndicatorAlign positions a narrowed selected indicator along the tab's
// main axis. It has no effect while the indicator uses the full tab width.
func (t TabsWidget) IndicatorAlign(align TabsIndicatorAlign) TabsWidget {
	t.indicatorAlign = align
	t.indicatorAlignSet = true
	return t
}

// IndicatorWidth limits the selected indicator length. Zero restores the
// theme's automatic length.
func (t TabsWidget) IndicatorWidth(width unit.Dp) TabsWidget {
	if width < 0 {
		width = 0
	}
	t.indicatorWidth = width
	return t
}

// IndicatorVisible controls whether the selected tab indicator is painted.
// It defaults to true for the line and primary variants.
func (t TabsWidget) IndicatorVisible(visible bool) TabsWidget {
	t.indicatorVisible = visible
	t.indicatorVisibleSet = true
	return t
}

// KeepAlive retains state created by hidden panels without laying them out or
// registering their input and semantic operations.
func (t TabsWidget) KeepAlive(enabled bool) TabsWidget {
	t.keepAlive = enabled
	return t
}

// DestroyOnHidden releases a panel's retained state when it is not selected.
// It takes precedence over KeepAlive when both are enabled.
func (t TabsWidget) DestroyOnHidden(enabled bool) TabsWidget {
	t.destroyOnHidden = enabled
	return t
}

// Animated controls label, indicator, and selected-panel transitions. The
// default is true.
func (t TabsWidget) Animated(enabled bool) TabsWidget {
	t.animated = enabled
	t.animationSet = true
	return t
}

// AnimationDuration overrides the tab label color transition duration.
func (t TabsWidget) AnimationDuration(duration time.Duration) TabsWidget {
	t.animationDuration = maxDuration(duration)
	return t
}

// IndicatorAnimationDuration overrides the selected indicator transition.
func (t TabsWidget) IndicatorAnimationDuration(duration time.Duration) TabsWidget {
	t.indicatorDuration = maxDuration(duration)
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
			frame.RetainState(ctx, scope)
			state.retainedPanels[item.Key] = struct{}{}
		} else {
			frame.ReleaseStateRetention(ctx, scope)
			delete(state.retainedPanels, item.Key)
		}
	}
	for itemKey := range state.retainedPanels {
		if _, ok := state.itemKeys[itemKey]; ok {
			continue
		}
		frame.ReleaseStateRetention(ctx, t.panelStateScope(ctx, itemKey))
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
				frame.RequestFocus(ctx, &state.item(focusKey).clickable)
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

func (t TabsWidget) panelsKeepAlive() bool {
	return t.keepAlive && !t.destroyOnHidden
}

func (t TabsWidget) panelsForceRender() bool {
	return t.forceRender && !t.destroyOnHidden
}

func (t TabsWidget) panelStateScope(ctx *frame.Context, itemKey string) string {
	return frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "panel-state:"+itemKey)
}

func (t TabsWidget) animationEnabled() bool {
	return !t.animationSet || t.animated
}

func (t TabsWidget) selectionDuration(ctx *frame.Context) time.Duration {
	if !t.animationEnabled() {
		return 0
	}
	if t.animationDuration > 0 {
		return t.animationDuration
	}
	return frame.ActiveTheme(ctx).Components.Tabs.ColorDuration
}

func (t TabsWidget) indicatorAnimationDuration(ctx *frame.Context) time.Duration {
	if !t.animationEnabled() {
		return 0
	}
	if t.indicatorDuration > 0 {
		return t.indicatorDuration
	}
	return frame.ActiveTheme(ctx).Components.Tabs.IndicatorDuration
}

func (t TabsWidget) panelAnimationDuration(ctx *frame.Context) time.Duration {
	if !t.animationEnabled() || t.panelTransition == TabsPanelNone {
		return 0
	}
	if t.panelDuration > 0 {
		return t.panelDuration
	}
	return frame.ActiveTheme(ctx).Components.Tabs.PanelDuration
}

func maxDuration(duration time.Duration) time.Duration {
	if duration < 0 {
		return 0
	}
	return duration
}

func (t TabsWidget) selectedItem(key string) (TabItem, bool) {
	for _, item := range t.items {
		if item.Key == key {
			return item, true
		}
	}
	return TabItem{}, false
}

func (t TabsWidget) usesIntrinsicTabWidths() bool {
	return t.fit || t.centered
}

func (t TabsWidget) closeButtonMetrics(ctx *frame.Context) (unit.Dp, unit.Dp) {
	theme := frame.ActiveTheme(ctx).Components.Tabs
	size, gap := theme.CloseButtonSize, theme.CloseButtonGap
	return size, gap
}
