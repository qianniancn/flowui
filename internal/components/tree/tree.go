package tree

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

// Item describes one node in a Tree.
type Item struct {
	Key              string
	Label            string
	Description      string
	Content          frame.Widget
	Leading          frame.Widget
	ExpandedLeading  frame.Widget
	Trailing         frame.Widget
	Switcher         frame.Widget
	ExpandedSwitcher frame.Widget
	Children         []Item
	ChildrenState    ChildrenState
	LoadError        string
	AcceptsChildren  bool
	Renamable        bool
	DisableCheckbox  bool
	DragDisabled     bool
	DropDisabled     bool
	Disabled         bool
}

// SimpleItem describes one flat Tree node. Items whose ParentKey is empty or
// unknown become roots when passed to ItemsFromSimple.
type SimpleItem struct {
	Key              string
	ParentKey        string
	Label            string
	Description      string
	Content          frame.Widget
	Leading          frame.Widget
	ExpandedLeading  frame.Widget
	Trailing         frame.Widget
	Switcher         frame.Widget
	ExpandedSwitcher frame.Widget
	ChildrenState    ChildrenState
	LoadError        string
	AcceptsChildren  bool
	Renamable        bool
	DisableCheckbox  bool
	DragDisabled     bool
	DropDisabled     bool
	Disabled         bool
}

// NewSimpleItem creates a flat Tree item for ItemsFromSimple.
func NewSimpleItem(key, parentKey, label string) SimpleItem {
	return SimpleItem{Key: key, ParentKey: parentKey, Label: label}
}

// ItemsFromSimple builds a Tree hierarchy from flat items while preserving
// input order. Items with unknown parents and invalid parent cycles are roots.
func ItemsFromSimple(simpleItems []SimpleItem) []Item {
	if len(simpleItems) == 0 {
		return nil
	}

	indexByKey := make(map[string]int, len(simpleItems))
	items := make([]Item, len(simpleItems))
	for index, simple := range simpleItems {
		if simple.Key == "" {
			panic("flowui: empty simple tree item key")
		}
		if _, exists := indexByKey[simple.Key]; exists {
			panic(fmt.Sprintf("flowui: duplicate simple tree item key %q", simple.Key))
		}
		indexByKey[simple.Key] = index
		items[index] = Item{
			Key:              simple.Key,
			Label:            simple.Label,
			Description:      simple.Description,
			Content:          simple.Content,
			Leading:          simple.Leading,
			ExpandedLeading:  simple.ExpandedLeading,
			Trailing:         simple.Trailing,
			Switcher:         simple.Switcher,
			ExpandedSwitcher: simple.ExpandedSwitcher,
			ChildrenState:    simple.ChildrenState,
			LoadError:        simple.LoadError,
			AcceptsChildren:  simple.AcceptsChildren,
			Renamable:        simple.Renamable,
			DisableCheckbox:  simple.DisableCheckbox,
			DragDisabled:     simple.DragDisabled,
			DropDisabled:     simple.DropDisabled,
			Disabled:         simple.Disabled,
		}
	}

	parents := make([]int, len(simpleItems))
	for index := range parents {
		parents[index] = -1
	}
	for index, simple := range simpleItems {
		if simple.ParentKey == "" {
			continue
		}
		if parentIndex, found := indexByKey[simple.ParentKey]; found {
			parents[index] = parentIndex
		}
	}
	validParent := simpleTreeValidParents(parents)
	children := make(map[int][]int, len(simpleItems))
	roots := make([]int, 0, len(simpleItems))
	for index, parentIndex := range parents {
		if parentIndex < 0 || !validParent[index] {
			roots = append(roots, index)
			continue
		}
		children[parentIndex] = append(children[parentIndex], index)
	}

	var build func(int) Item
	build = func(index int) Item {
		item := items[index]
		for _, child := range children[index] {
			item.Children = append(item.Children, build(child))
		}
		return item
	}
	result := make([]Item, 0, len(roots))
	for _, root := range roots {
		result = append(result, build(root))
	}
	return result
}

func simpleTreeValidParents(parents []int) []bool {
	const (
		simpleTreeVisiting uint8 = iota + 1
		simpleTreeResolved
	)
	status := make([]uint8, len(parents))
	valid := make([]bool, len(parents))
	var resolve func(int) bool
	resolve = func(index int) bool {
		switch status[index] {
		case simpleTreeResolved:
			return valid[index]
		case simpleTreeVisiting:
			return false
		}
		status[index] = simpleTreeVisiting
		parent := parents[index]
		valid[index] = parent < 0 || resolve(parent)
		status[index] = simpleTreeResolved
		return valid[index]
	}
	for index := range parents {
		resolve(index)
	}
	return valid
}

// ChildrenState describes controlled asynchronous child loading.
type ChildrenState uint8

const (
	ChildrenLoaded ChildrenState = iota
	ChildrenUnloaded
	ChildrenLoading
	ChildrenError
)

// LoadEvent describes an asynchronous child-load request.
type LoadEvent struct {
	Key      string
	Retry    bool
	Expanded bool
}

// Variant selects the Tree container treatment.
type Variant uint8

const (
	VariantDefault Variant = iota
	VariantSurface
)

// SelectionMode controls whether rows can be selected.
type SelectionMode uint8

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
	SelectionNone
)

// Size selects the Tree density preset.
type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
)

// GuideStyle selects the indentation guide stroke pattern.
type GuideStyle uint8

const (
	GuideSolid GuideStyle = iota
	GuideDashed
)

// ExpandAction controls which row gesture toggles branch expansion.
type ExpandAction uint8

const (
	ExpandActionClick ExpandAction = iota
	ExpandActionDoubleClick
)

// DropPosition describes where a dragged item was dropped relative to its target.
type DropPosition uint8

const (
	DropBefore DropPosition = iota
	DropInside
	DropAfter
)

// DropEvent describes a requested move within one Tree.
type DropEvent struct {
	// SourceKey is the primary dragged item and is retained for single-item handlers.
	SourceKey string
	// SourceKeys contains all dragged items in tree order.
	SourceKeys []string
	TargetKey  string
	Position   DropPosition
}

// DragEvent describes a Tree drag operation or its current valid target.
// TargetKey is empty for drag start and end events without a target.
type DragEvent struct {
	SourceKey  string
	SourceKeys []string
	TargetKey  string
	Position   DropPosition
}

// Widget presents hierarchical, expandable data with controlled selection and expansion.
type Widget struct {
	key                 string
	selectedKey         string
	selectedKeys        []string
	selectedKeySet      stateutil.StringSet
	checkedKeys         []string
	halfCheckedKeys     []string
	checkedKeySet       stateutil.StringSet
	halfCheckedKeySet   stateutil.StringSet
	resolvedChecked     []string
	resolvedHalf        []string
	resolvedCheckedSet  stateutil.StringSet
	resolvedHalfSet     stateutil.StringSet
	items               []Item
	dataVersion         uint64
	hasDataVersion      bool
	expandedKeys        []string
	disabledKeys        []string
	disabledKeySet      stateutil.StringSet
	emptyText           string
	onChange            func(string)
	onSelectionChange   func([]string)
	onExpandedChange    func([]string)
	onAction            func(string)
	onCheckedChange     func([]string, []string)
	onDrop              func(DropEvent)
	canDrop             func(DropEvent) bool
	onDragStart         func(DragEvent)
	onDragEnter         func(DragEvent)
	onDragLeave         func(DragEvent)
	onDragOver          func(DragEvent)
	onDragEnd           func(DragEvent)
	onLoadChildren      func(string)
	onLoadChildrenEvent func(LoadEvent)
	onRename            func(string, string)
	renameRequestKey    string
	renameRequest       uint64
	contextMenu         menu.Widget
	hasContextMenu      bool
	onContextMenu       func(string)
	variant             Variant
	size                Size
	selectionMode       SelectionMode
	checkable           bool
	checkStrictly       bool
	disabled            bool
	allowEmpty          bool
	guides              bool
	guideConnectors     bool
	guideStyle          GuideStyle
	expandOnRowClick    bool
	expandAction        ExpandAction
	maxHeight           int
	filterQuery         string
	filterFunc          func(Item) bool
	filterVersion       uint64
	hasFilterVersion    bool
	customStyle         flowstyle.Style
}

// New creates a controlled Tree.
func New(key, selectedKey string, items []Item) Widget {
	return Widget{
		key:          key,
		selectedKey:  selectedKey,
		items:        items,
		emptyText:    "No items",
		expandAction: ExpandActionClick,
	}
}

// DataVersion enables validation and flattened-data reuse. Increase version
// whenever the item hierarchy or item content changes.
func (t Widget) DataVersion(version uint64) Widget {
	t.dataVersion = version
	t.hasDataVersion = true
	return t
}

// ExpandedKeys sets the controlled set of expanded branch keys.
func (t Widget) ExpandedKeys(keys []string) Widget {
	t.expandedKeys = keys
	return t
}

// OnChange registers a callback for selection changes.
func (t Widget) OnChange(fn func(string)) Widget {
	t.onChange = fn
	return t
}

// SelectedKeys sets the controlled selection for SelectionMultiple.
func (t Widget) SelectedKeys(keys []string) Widget {
	t.selectedKeys = keys
	return t
}

// OnSelectionChange registers a callback for multiple-selection changes.
func (t Widget) OnSelectionChange(fn func([]string)) Widget {
	t.onSelectionChange = fn
	return t
}

// OnExpandedChange registers a callback for expansion changes.
func (t Widget) OnExpandedChange(fn func([]string)) Widget {
	t.onExpandedChange = fn
	return t
}

// OnAction registers a callback for row activation, independent of selection mode.
func (t Widget) OnAction(fn func(string)) Widget {
	t.onAction = fn
	return t
}

// Checkable enables checkbox controls on each item row.
func (t Widget) Checkable(enabled bool) Widget {
	t.checkable = enabled
	return t
}

// CheckedKeys sets the controlled checked item keys.
func (t Widget) CheckedKeys(keys []string) Widget {
	t.checkedKeys = keys
	return t
}

// HalfCheckedKeys sets the controlled half-checked item keys used by strict mode.
func (t Widget) HalfCheckedKeys(keys []string) Widget {
	t.halfCheckedKeys = keys
	return t
}

// CheckStrictly disables parent/child checkbox association.
func (t Widget) CheckStrictly(enabled bool) Widget {
	t.checkStrictly = enabled
	return t
}

// OnCheckedChange reports the next checked and half-checked keys.
func (t Widget) OnCheckedChange(fn func([]string, []string)) Widget {
	t.onCheckedChange = fn
	return t
}

// OnDrop enables item dragging and reports a requested move.
func (t Widget) OnDrop(fn func(DropEvent)) Widget {
	t.onDrop = fn
	return t
}

// CanDrop validates a proposed drop for both indicator and delivery. It may be
// called every frame while dragging and must not mutate application state.
func (t Widget) CanDrop(fn func(DropEvent) bool) Widget {
	t.canDrop = fn
	return t
}

// OnDragStart reports when item dragging begins.
func (t Widget) OnDragStart(fn func(DragEvent)) Widget {
	t.onDragStart = fn
	return t
}

// OnDragEnter reports when dragging enters a valid drop target.
func (t Widget) OnDragEnter(fn func(DragEvent)) Widget {
	t.onDragEnter = fn
	return t
}

// OnDragLeave reports when dragging leaves a previously valid drop target.
func (t Widget) OnDragLeave(fn func(DragEvent)) Widget {
	t.onDragLeave = fn
	return t
}

// OnDragOver reports the current valid drop target during dragging. It can be
// called every frame and must not mutate application state.
func (t Widget) OnDragOver(fn func(DragEvent)) Widget {
	t.onDragOver = fn
	return t
}

// OnDragEnd reports when item dragging ends. TargetKey identifies the last
// valid target when one was active.
func (t Widget) OnDragEnd(fn func(DragEvent)) Widget {
	t.onDragEnd = fn
	return t
}

// OnLoadChildren handles initial loading and retries for asynchronous branches.
func (t Widget) OnLoadChildren(fn func(string)) Widget {
	t.onLoadChildren = fn
	return t
}

// OnLoadChildrenEvent handles asynchronous child loading with retry and
// current expansion state. It can be used together with OnLoadChildren.
func (t Widget) OnLoadChildrenEvent(fn func(LoadEvent)) Widget {
	t.onLoadChildrenEvent = fn
	return t
}

// OnRename handles committed inline label edits for Renamable items.
func (t Widget) OnRename(fn func(string, string)) Widget {
	t.onRename = fn
	return t
}

// RequestRename starts inline editing once when revision changes.
func (t Widget) RequestRename(key string, revision uint64) Widget {
	t.renameRequestKey = key
	t.renameRequest = revision
	return t
}

// ContextMenu sets the menu opened from an enabled item row.
func (t Widget) ContextMenu(content menu.Widget) Widget {
	t.contextMenu = content
	t.hasContextMenu = true
	return t
}

// OnContextMenu reports the item whose context menu is opening.
func (t Widget) OnContextMenu(fn func(string)) Widget {
	t.onContextMenu = fn
	return t
}

// Variant sets the container treatment.
func (t Widget) Variant(variant Variant) Widget {
	t.variant = variant
	return t
}

// Size sets the Tree density preset.
func (t Widget) Size(size Size) Widget {
	t.size = size
	return t
}

// SelectionMode controls whether activation changes the selected key.
func (t Widget) SelectionMode(mode SelectionMode) Widget {
	t.selectionMode = mode
	return t
}

// DisabledKeys disables individual nodes by key.
func (t Widget) DisabledKeys(keys []string) Widget {
	t.disabledKeys = keys
	return t
}

// Disabled disables all Tree interaction.
func (t Widget) Disabled(disabled bool) Widget {
	t.disabled = disabled
	return t
}

// AllowEmptySelection allows activating the selected row to clear selection.
func (t Widget) AllowEmptySelection() Widget {
	t.allowEmpty = true
	return t
}

// Guides controls whether expanded branches draw indentation guides.
func (t Widget) Guides(enabled bool) Widget {
	t.guides = enabled
	return t
}

// GuideConnectors controls whether indentation guides connect horizontally to child content.
func (t Widget) GuideConnectors(enabled bool) Widget {
	t.guideConnectors = enabled
	return t
}

// GuideStyle sets the indentation guide stroke pattern.
func (t Widget) GuideStyle(style GuideStyle) Widget {
	t.guideStyle = style
	return t
}

// ExpandOnRowClick controls whether clicking a branch row also toggles it.
func (t Widget) ExpandOnRowClick(enabled bool) Widget {
	t.expandOnRowClick = enabled
	return t
}

// ExpandAction controls whether branch rows expand on single or double click.
func (t Widget) ExpandAction(action ExpandAction) Widget {
	t.expandAction = action
	t.expandOnRowClick = true
	return t
}

// EmptyText sets the message displayed when the Tree has no nodes.
func (t Widget) EmptyText(value string) Widget {
	t.emptyText = value
	return t
}

// Filter keeps nodes whose label or description matches query
// case-insensitively, plus ancestors of matches. Empty query shows the full
// tree. Matching branches are expanded for the duration of the filter.
// FilterFunc replaces this default matcher when configured.
func (t Widget) Filter(query string) Widget {
	t.filterQuery = strings.TrimSpace(query)
	return t
}

// FilterFunc keeps nodes matched by fn, plus their ancestors. It replaces the
// default Filter text matcher. Pair dynamic predicates with FilterVersion so a
// versioned Tree can invalidate its flattened-data cache when they change.
func (t Widget) FilterFunc(fn func(Item) bool) Widget {
	t.filterFunc = fn
	return t
}

// FilterVersion invalidates cached filtered entries when a FilterFunc changes
// its captured condition. It is only needed with FilterFunc and DataVersion.
func (t Widget) FilterVersion(version uint64) Widget {
	t.filterVersion = version
	t.hasFilterVersion = true
	return t
}

// MaxHeight overrides the theme maximum height in dp.
func (t Widget) MaxHeight(dp int) Widget {
	t.maxHeight = max(dp, 0)
	return t
}

func (t Widget) Style(value flowstyle.Style) Widget {
	t.customStyle = value
	return t
}

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := treeStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	t.selectedKeySet = state.selectedKeys.Resolve(t.selectedKeys)
	t.checkedKeySet = state.checkedKeys.Resolve(t.checkedKeys)
	t.halfCheckedKeySet = state.halfCheckedKeys.Resolve(t.halfCheckedKeys)
	t.disabledKeySet = state.disabledKeys.Resolve(t.disabledKeys)
	t.resolveCheckState(state)
	visible := state.resolveVisible(t)
	if state.renameKey != "" && treeVisibleIndex(visible, state.renameKey) < 0 {
		state.finishRename(t, false)
	}
	state.applyRenameRequest(ctx, t, visible)
	state.updateRename(gtx, t)
	state.dragSource = ""
	state.dragSources = state.dragSources[:0]
	state.dropTarget = treeDropTarget{}
	var dragSelection []string
	if t.onDrop != nil && t.selectionMode == SelectionMultiple {
		dragSelection = t.selectedDragKeys(visible)
	}

	if !t.disabled {
		dragIndex := -1
		for itemKey, itemState := range state.items {
			index := state.visibleIndex(visible, itemKey, itemState)
			if index < 0 {
				continue
			}
			entry := visible[index]
			disabled := t.itemDisabled(entry.item)
			presses := stateutil.ActivePresses(itemState.clickable.History())
			for itemState.toggle.Clicked(gtx) {
				if !disabled && treeItemExpandable(entry.item) {
					t.requestItemToggle(entry.item)
				}
			}
			for itemState.check.Clicked(gtx) {
				if !t.itemCheckboxDisabled(entry.item) {
					t.requestCheck(entry.item.Key, !t.itemChecked(entry.item.Key))
				}
			}
			for {
				click, ok := itemState.clickable.Update(gtx)
				if !ok {
					break
				}
				if disabled {
					continue
				}
				t.activateWithModifiers(state, visible, entry.item.Key, click.Modifiers)
				selectionModifier := click.Modifiers.Contain(key.ModShift) || click.Modifiers.Contain(key.ModCtrl) || click.Modifiers.Contain(key.ModCommand)
				if t.rowClickExpands(click.NumClicks) && !selectionModifier && treeItemExpandable(entry.item) {
					t.requestItemToggle(entry.item)
				}
			}
			if !disabled {
				frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
			}
			if t.onDrop != nil && !disabled {
				if !entry.item.DropDisabled {
					t.updateDropEvents(gtx, state, itemState, visible, entry)
				}
				sources := []string{entry.item.Key}
				if t.selectionMode == SelectionMultiple && treeContainsKey(t.selectedKeys, entry.item.Key) && len(dragSelection) > 0 {
					sources = dragSelection
				}
				if !entry.item.DragDisabled && itemState.updateDrag(gtx, state.dragMIME, sources) && dragIndex < 0 {
					state.dragSource = entry.item.Key
					state.dragSources = append(state.dragSources[:0], sources...)
					dragIndex = index
				}
			}
		}
		if dragIndex >= 0 {
			tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
			geometry := state.dragGeometryFor(
				t,
				visible,
				gtx.Dp(tokens.RowHeight),
				gtx.Dp(tokens.DescriptionRowHeight),
				gtx.Dp(tokens.Gap),
			)
			sourceState := state.item(state.dragSource)
			pointerY := sourceState.drag.Press().Y + sourceState.drag.Position().Y
			target := geometry.dropTargetAt(visible, dragIndex, pointerY)
			viewportHeight := gtx.Constraints.Max.Y - 2*gtx.Dp(tokens.Padding)
			maxHeight := tokens.MaxHeight
			if t.maxHeight > 0 {
				maxHeight = unit.Dp(t.maxHeight)
			}
			if maxHeight > 0 {
				viewportHeight = min(viewportHeight, gtx.Dp(maxHeight)-2*gtx.Dp(tokens.Padding))
			}
			pointerViewportY := geometry.viewportY(state.list.Position, dragIndex, pointerY)
			state.updateDragScroll(gtx, pointerViewportY, max(viewportHeight, 0), gtx.Dp(unit.Dp(24)), len(visible))
			validTarget := treeDropTarget{}
			if t.dragDropAllowed(visible, geometry, state.dragSources, target) {
				validTarget = geometry.dropIndicatorTarget(visible, target)
				state.dropTarget = validTarget
			}
			t.updateDragLifecycle(state, state.dragSources, validTarget)
			targetIndex := treeVisibleIndex(visible, state.dropTarget.key)
			hoverExpandable := targetIndex >= 0 && state.dropTarget.position == DropInside && treeItemExpandable(visible[targetIndex].item) && !treeContainsKey(t.expandedKeys, state.dropTarget.key)
			if state.updateDragHover(gtx, state.dropTarget.key, hoverExpandable) {
				t.requestItemToggle(visible[targetIndex].item)
			}
		} else {
			t.endDragLifecycle(state)
			state.resetDragAssist()
		}
		result := state.updateKeys(gtx, t, visible)
		if result.focusKey != "" {
			if result.rangeKey == "" && t.selectionMode == SelectionMultiple {
				state.selectionAnchor = result.focusKey
			}
			frame.RequestFocus(ctx, &state.item(result.focusKey).clickable)
			state.ensureVisible(treeVisibleIndex(visible, result.focusKey))
		}
		if result.loadEvent.Key != "" {
			t.requestLoad(result.loadEvent)
		}
		if result.toggleKey != "" {
			t.requestToggle(result.toggleKey)
		}
		if result.renameKey != "" {
			index := treeVisibleIndex(visible, result.renameKey)
			if index >= 0 {
				state.beginRename(visible[index].item)
				frame.RequestFocus(ctx, &state.renameEditor)
			}
		}
		if result.rangeKey != "" {
			t.activateWithModifiers(state, visible, result.rangeKey, key.ModShift)
		}
		if result.checkKey != "" {
			t.requestCheck(result.checkKey, !t.itemChecked(result.checkKey))
		}
		if result.actionKey != "" {
			t.activateWithModifiers(state, visible, result.actionKey, result.actionModifiers)
		}
	} else {
		t.endDragLifecycle(state)
		state.resetDragAssist()
	}
	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, item := range state.items {
		hovered = hovered || item.clickable.Hovered() || item.toggle.Hovered() || item.check.Hovered()
		pressed = pressed || item.clickable.Pressed() || item.check.Pressed()
		itemFocused := gtx.Focused(&item.clickable)
		focused = focused || itemFocused
		focusVisible = focusVisible || frame.FocusVisible(ctx, &item.clickable, itemFocused)
	}
	dims := layoutui.LayoutStyled(ctx, gtx, frame.FullKey(ctx, t.key), flowstyle.StyleState{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     t.disabled || !gtx.Enabled(),
		Selected:     t.selectedKey != "" || len(t.selectedKeys) > 0,
		Expanded:     len(t.expandedKeys) > 0,
		Dragging:     state.dragSource != "",
		DropTarget:   state.dropTarget.key != "",
	}, t.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return t.layout(ctx, gtx, state, visible)
	}))
	if state.dragSource != "" {
		dragIndex := treeVisibleIndex(visible, state.dragSource)
		if dragIndex >= 0 && !treeListPositionContains(state.list.Position, dragIndex) {
			tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
			geometry := state.dragGeometryFor(
				t,
				visible,
				gtx.Dp(tokens.RowHeight),
				gtx.Dp(tokens.DescriptionRowHeight),
				gtx.Dp(tokens.Gap),
			)
			sourceState := state.item(state.dragSource)
			sourceTop := geometry.viewportY(state.list.Position, dragIndex, 0)
			press := sourceState.drag.Press().Round()
			sourceOffset := op.Offset(image.Pt(gtx.Dp(tokens.Padding), gtx.Dp(tokens.Padding)+int(sourceTop))).Push(gtx.Ops)
			label := visible[dragIndex].item.Label
			if count := len(state.dragSources); count > 1 {
				label = fmt.Sprintf("%s +%d", label, count-1)
			}
			offscreenGtx := gtx
			offscreenGtx.Constraints.Min = image.Point{}
			sourceState.drag.Layout(
				offscreenGtx,
				func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(1, 1)} },
				t.dragPreview(ctx, label, press.Add(image.Pt(gtx.Dp(unit.Dp(12)), gtx.Dp(unit.Dp(12))))),
			)
			sourceOffset.Pop()
		}
	}
	return dims
}

func (t Widget) activateWithModifiers(state *treeState, visible []flatItem, itemKey string, modifiers key.Modifiers) {
	rangeSelection := t.selectionMode == SelectionMultiple && modifiers.Contain(key.ModShift) && state != nil
	if t.onAction != nil && !rangeSelection {
		t.onAction(itemKey)
	}
	switch t.selectionMode {
	case SelectionSingle:
		next := itemKey
		if t.allowEmpty && itemKey == t.selectedKey {
			next = ""
		}
		if next != t.selectedKey && t.onChange != nil {
			t.onChange(next)
		}
	case SelectionMultiple:
		var next []string
		switch {
		case rangeSelection:
			next = t.rangeKeys(visible, state.selectionAnchor, itemKey)
		case modifiers.Contain(key.ModCtrl) || modifiers.Contain(key.ModCommand):
			state.selectionAnchor = itemKey
			next = toggleTreeKey(t.selectedKeys, itemKey)
		default:
			state.selectionAnchor = itemKey
			if t.allowEmpty && len(t.selectedKeys) == 1 && t.selectedKeys[0] == itemKey {
				next = nil
			} else {
				next = []string{itemKey}
			}
		}
		if !treeKeysEqual(next, t.selectedKeys) && t.onSelectionChange != nil {
			t.onSelectionChange(next)
		}
	}
}

func (t Widget) requestToggle(key string) {
	if t.onExpandedChange != nil {
		t.onExpandedChange(toggleTreeKey(t.expandedKeys, key))
	}
}

func (t Widget) requestItemToggle(item Item) {
	expanded := treeContainsKey(t.expandedKeys, item.Key)
	if item.ChildrenState == ChildrenError {
		t.requestLoad(treeLoadEvent(item, expanded))
		if !expanded {
			t.requestToggle(item.Key)
		}
		return
	}
	if item.ChildrenState == ChildrenUnloaded && !expanded {
		t.requestLoad(treeLoadEvent(item, expanded))
	}
	t.requestToggle(item.Key)
}

func (t Widget) requestLoad(event LoadEvent) {
	if t.onLoadChildren != nil {
		t.onLoadChildren(event.Key)
	}
	if t.onLoadChildrenEvent != nil {
		t.onLoadChildrenEvent(event)
	}
}

func treeLoadEvent(item Item, expanded bool) LoadEvent {
	return LoadEvent{Key: item.Key, Retry: item.ChildrenState == ChildrenError, Expanded: expanded}
}

func (t *Widget) resolveCheckState(state *treeState) {
	if !t.checkable {
		t.resolvedChecked, t.resolvedHalf = nil, nil
		t.resolvedCheckedSet, t.resolvedHalfSet = nil, nil
		return
	}
	if t.checkStrictly {
		t.resolvedChecked = append([]string(nil), t.checkedKeys...)
		t.resolvedHalf = append([]string(nil), t.halfCheckedKeys...)
		t.resolvedCheckedSet = state.checkedKeys.Resolve(t.resolvedChecked)
		t.resolvedHalfSet = state.halfCheckedKeys.Resolve(t.resolvedHalf)
		return
	}
	checked, half := deriveCascadeCheckedKeysForKeys(t.items, t.checkedKeys, t.checkedKeySet, t.itemCheckboxBoundary)
	t.resolvedChecked = checked
	t.resolvedHalf = half
	t.resolvedCheckedSet = state.checkedKeys.Resolve(checked)
	t.resolvedHalfSet = state.halfCheckedKeys.Resolve(half)
}

func (t Widget) requestCheck(key string, checked bool) {
	if !t.checkable || t.onCheckedChange == nil {
		return
	}
	if t.checkStrictly {
		next := append([]string(nil), t.checkedKeys...)
		if checked {
			next = appendTreeKey(next, key)
		} else {
			next = removeTreeKey(next, key)
		}
		half := removeTreeKey(append([]string(nil), t.halfCheckedKeys...), key)
		t.onCheckedChange(next, half)
		return
	}
	base := treeKeySet(t.resolvedChecked)
	if checked {
		treeSetCheckSubtree(base, t.items, key, true, t.itemCheckboxBoundary)
	} else {
		treeSetCheckSubtree(base, t.items, key, false, t.itemCheckboxBoundary)
		for _, ancestor := range treeCheckAncestorKeys(t.items, key, t.itemCheckboxBoundary) {
			delete(base, ancestor)
		}
	}
	nextChecked, nextHalf := deriveCascadeCheckedKeysWith(t.items, base, t.itemCheckboxBoundary)
	t.onCheckedChange(nextChecked, nextHalf)
}

func (t Widget) itemDisabled(item Item) bool {
	return t.disabled || item.Disabled || stateutil.StringSetContains(t.disabledKeys, t.disabledKeySet, item.Key)
}

func (t Widget) itemCheckboxDisabled(item Item) bool {
	return t.disabled || t.itemCheckboxBoundary(item)
}

func (t Widget) itemCheckboxBoundary(item Item) bool {
	return item.Disabled || stateutil.StringSetContains(t.disabledKeys, t.disabledKeySet, item.Key) || item.DisableCheckbox
}

func (t Widget) itemSelected(key string) bool {
	if t.selectionMode == SelectionNone {
		return false
	}
	if t.selectionMode == SelectionMultiple {
		return stateutil.StringSetContains(t.selectedKeys, t.selectedKeySet, key)
	}
	return key == t.selectedKey
}

func (t Widget) itemChecked(key string) bool {
	return t.checkable && stateutil.StringSetContains(t.resolvedChecked, t.resolvedCheckedSet, key)
}

func (t Widget) itemHalfChecked(key string) bool {
	return t.checkable && stateutil.StringSetContains(t.resolvedHalf, t.resolvedHalfSet, key)
}

func (t Widget) keyboardActiveIndex(visible []flatItem) int {
	if t.selectionMode == SelectionMultiple {
		for index := len(t.selectedKeys) - 1; index >= 0; index-- {
			visibleIndex := treeVisibleIndex(visible, t.selectedKeys[index])
			if visibleIndex >= 0 && !t.itemDisabled(visible[visibleIndex].item) {
				return visibleIndex
			}
		}
		return -1
	}
	index := treeVisibleIndex(visible, t.selectedKey)
	if index >= 0 && !t.itemDisabled(visible[index].item) {
		return index
	}
	return -1
}

func (t Widget) rangeKeys(visible []flatItem, anchor, target string) []string {
	anchorIndex := treeVisibleIndex(visible, anchor)
	targetIndex := treeVisibleIndex(visible, target)
	if anchorIndex < 0 {
		for _, selected := range t.selectedKeys {
			if index := treeVisibleIndex(visible, selected); index >= 0 {
				anchorIndex = index
				break
			}
		}
	}
	if targetIndex < 0 {
		return append([]string(nil), t.selectedKeys...)
	}
	if anchorIndex < 0 {
		anchorIndex = targetIndex
	}
	start, end := min(anchorIndex, targetIndex), max(anchorIndex, targetIndex)
	next := make([]string, 0, end-start+1)
	for _, selected := range t.selectedKeys {
		if index := treeVisibleIndex(visible, selected); index >= 0 && t.itemDisabled(visible[index].item) {
			next = append(next, selected)
		}
	}
	for index := start; index <= end; index++ {
		if !t.itemDisabled(visible[index].item) && !treeContainsKey(next, visible[index].item.Key) {
			next = append(next, visible[index].item.Key)
		}
	}
	return next
}

func (t Widget) selectedDragKeys(visible []flatItem) []string {
	if t.selectionMode != SelectionMultiple || len(t.selectedKeys) == 0 {
		return nil
	}
	selected := treeKeySet(t.selectedKeys)
	result := make([]string, 0, len(selected))
	for _, entry := range visible {
		if _, ok := selected[entry.item.Key]; !ok || t.itemDisabled(entry.item) || entry.item.DragDisabled {
			continue
		}
		nested := false
		for parent := entry.parentKey; parent != ""; {
			index := treeVisibleIndex(visible, parent)
			if index < 0 {
				break
			}
			parentItem := visible[index].item
			if _, ok := selected[parent]; ok && !t.itemDisabled(parentItem) && !parentItem.DragDisabled {
				nested = true
				break
			}
			parent = visible[index].parentKey
		}
		if !nested {
			result = append(result, entry.item.Key)
		}
	}
	return result
}

func (t Widget) rowClickExpands(numClicks int) bool {
	if !t.expandOnRowClick {
		return false
	}
	if t.expandAction == ExpandActionDoubleClick {
		return numClicks >= 2
	}
	return numClicks >= 1
}

func deriveCascadeCheckedKeys(items []Item, explicit stateutil.StringSet) ([]string, []string) {
	return deriveCascadeCheckedKeysWith(items, explicit, nil)
}

func deriveCascadeCheckedKeysWith(items []Item, explicit stateutil.StringSet, checkboxDisabled func(Item) bool) ([]string, []string) {
	return deriveCascadeCheckedKeysForKeys(items, nil, explicit, checkboxDisabled)
}

func deriveCascadeCheckedKeysForKeys(items []Item, explicitKeys []string, explicitSet stateutil.StringSet, checkboxDisabled func(Item) bool) ([]string, []string) {
	full := treeExpandedCheckedSetForKeys(items, explicitKeys, explicitSet, checkboxDisabled)
	checked := make([]string, 0)
	half := make([]string, 0)
	var walkItem func(Item) uint8
	walkItem = func(item Item) uint8 {
		itemFull := stateutil.StringSetContains(nil, full, item.Key)
		if treeCheckboxDisabled(item, checkboxDisabled) {
			for _, child := range item.Children {
				walkItem(child)
			}
			if itemFull {
				checked = append(checked, item.Key)
				return 2
			}
			return 0
		}
		if len(item.Children) == 0 {
			if itemFull {
				checked = append(checked, item.Key)
				return 2
			}
			return 0
		}
		allFull := true
		anyChecked := false
		checkableChildren := 0
		for _, child := range item.Children {
			state := walkItem(child)
			if treeCheckboxDisabled(child, checkboxDisabled) {
				continue
			}
			checkableChildren++
			allFull = allFull && state == 2
			anyChecked = anyChecked || state != 0
		}
		if itemFull || checkableChildren > 0 && allFull {
			checked = append(checked, item.Key)
			return 2
		}
		if anyChecked {
			half = append(half, item.Key)
			return 1
		}
		return 0
	}
	for _, item := range items {
		walkItem(item)
	}
	slices.Sort(checked)
	slices.Sort(half)
	return checked, half
}

func treeExpandedCheckedSet(items []Item, explicit stateutil.StringSet) stateutil.StringSet {
	return treeExpandedCheckedSetWith(items, explicit, nil)
}

func treeExpandedCheckedSetWith(items []Item, explicit stateutil.StringSet, checkboxDisabled func(Item) bool) stateutil.StringSet {
	return treeExpandedCheckedSetForKeys(items, nil, explicit, checkboxDisabled)
}

func treeExpandedCheckedSetForKeys(items []Item, explicitKeys []string, explicitSet stateutil.StringSet, checkboxDisabled func(Item) bool) stateutil.StringSet {
	full := make(stateutil.StringSet)
	var walk func([]Item, bool)
	walk = func(children []Item, inherited bool) {
		for _, item := range children {
			explicitlyChecked := stateutil.StringSetContains(explicitKeys, explicitSet, item.Key)
			checked := inherited || explicitlyChecked
			if treeCheckboxDisabled(item, checkboxDisabled) {
				checked = explicitlyChecked
			}
			if checked {
				full[item.Key] = struct{}{}
			}
			if treeCheckboxDisabled(item, checkboxDisabled) {
				walk(item.Children, false)
				continue
			}
			walk(item.Children, checked)
		}
	}
	walk(items, false)
	return full
}

func treeSetCheckSubtree(keys stateutil.StringSet, items []Item, target string, checked bool, checkboxDisabled func(Item) bool) bool {
	for _, item := range items {
		if item.Key == target {
			treeApplyCheckSubtree(keys, item, checked, checkboxDisabled)
			return true
		}
		if treeSetCheckSubtree(keys, item.Children, target, checked, checkboxDisabled) {
			return true
		}
	}
	return false
}

func treeApplyCheckSubtree(keys stateutil.StringSet, item Item, checked bool, checkboxDisabled func(Item) bool) {
	if treeCheckboxDisabled(item, checkboxDisabled) {
		return
	}
	if checked {
		keys[item.Key] = struct{}{}
	} else {
		delete(keys, item.Key)
	}
	for _, child := range item.Children {
		treeApplyCheckSubtree(keys, child, checked, checkboxDisabled)
	}
}

func treeCheckAncestorKeys(items []Item, target string, checkboxDisabled func(Item) bool) []string {
	var path []Item
	var walk func([]Item, []Item) bool
	walk = func(children []Item, parents []Item) bool {
		for _, item := range children {
			if item.Key == target {
				path = append([]Item(nil), parents...)
				return true
			}
			if walk(item.Children, append(parents, item)) {
				return true
			}
		}
		return false
	}
	walk(items, nil)
	ancestors := make([]string, 0, len(path))
	for index := len(path) - 1; index >= 0; index-- {
		if treeCheckboxDisabled(path[index], checkboxDisabled) {
			break
		}
		ancestors = append(ancestors, path[index].Key)
	}
	return ancestors
}

func treeCheckboxDisabled(item Item, checkboxDisabled func(Item) bool) bool {
	return checkboxDisabled != nil && checkboxDisabled(item)
}

func appendTreeKey(keys []string, key string) []string {
	if treeContainsKey(keys, key) {
		return keys
	}
	return append(keys, key)
}

func removeTreeKey(keys []string, key string) []string {
	index := slices.Index(keys, key)
	if index < 0 {
		return keys
	}
	return append(append([]string(nil), keys[:index]...), keys[index+1:]...)
}
