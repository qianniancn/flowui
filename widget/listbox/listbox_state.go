package listbox

import (
	"fmt"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
	"github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/optionrow"
)

const stateSlotListBox = "listbox"

func listBoxStateFor(ctx *core.Context, key string) *listBoxState {
	key = core.ClaimKey(ctx, core.KindListBox, key)
	return core.UseState[listBoxState](ctx, key, stateSlotListBox)
}

func (l Widget) stateFor(ctx *core.Context) *listBoxState {
	if l.derivedOwner == "" {
		return listBoxStateFor(ctx, l.key)
	}
	key := core.ClaimDerivedResolvedKey(ctx, core.KindListBox, l.derivedOwner, l.derivedRole)
	return core.UseState[listBoxState](ctx, key, stateSlotListBox)
}

type listBoxState struct {
	scrollToKey      string
	scrollToIndex    int
	disclosureSingle disclosure.Binding[string]
	// For multiple selection, we manage state manually since []string is not comparable
	multipleValue     []string
	multipleReady     bool
	list              layout.List
	bar               widget.Scrollbar
	visualOutset      layoutui.VisualOutsetState
	items             map[string]*listBoxItemState
	frameItems        map[string]struct{}
	itemKeys          map[string]int
	keyFilters        []event.Filter
	dataCache         listBoxDataCache
	selectedKeys      core.StringSetCache
	disabledKeys      core.StringSetCache
	focusedKey        string
	pressedKey        key.Name
	pressedActionKey  string
	typeahead         navigation.Typeahead
	visualCache       listBoxVisualCache
	itemVisualCache   computed.ProviderCacheMap[theme.ListBoxItemRequest, theme.ListBoxItemVisual]
	resolvedPartCache listBoxResolvedPartCache
}

// listBoxItemVisualCacheCapacity bounds custom-provider state combinations.
const listBoxItemVisualCacheCapacity = 64

// listBoxVisualCache keeps the root Provider visual on the shared cache
// lifecycle. Instance overrides remain request-local because their declaration
// can be arbitrary and must not contaminate the base visual entry.
type listBoxVisualCache struct {
	cache computed.ProviderCache[theme.ListBoxRequest, theme.ListBoxVisual]
}

const listBoxResolvedPartCacheCapacity = 64

type listBoxResolvedPartKey struct {
	part    style.Part
	variant uint8
	state   style.StateMask
}

// listBoxResolvedPartCache stores immutable static Item Part snapshots. The
// item transition key stays outside this cache, so two rows can share a
// settled declaration without sharing animation state.
type listBoxResolvedPartCache struct {
	cache           computed.ProviderCacheMap[listBoxResolvedPartKey, style.ResolvedStyle]
	theme           *theme.Theme
	themeGeneration uint64
	scopeHash       uint64
	overrideHash    uint64
	enabled         bool
	initialized     bool
	hits            uint64
	misses          uint64
	themeInvalid    uint64
	scopeInvalid    uint64
	overrideInvalid uint64
}

type listBoxResolvedPartCacheStats struct {
	Hits                  uint64
	Misses                uint64
	ThemeInvalidations    uint64
	ScopeInvalidations    uint64
	OverrideInvalidations uint64
	Size                  int
	Evictions             uint64
}

func (c *listBoxResolvedPartCache) prepare(
	ctx *core.Context,
	active *theme.Theme,
	instanceStyle style.Style,
) {
	if c == nil {
		return
	}
	enabled := computed.StyleLayersCacheSafe(ctx, instanceStyle)
	scopeHash := uint64(0)
	overrideHash := uint64(0)
	if enabled {
		scopeHash = computed.StyleScopeHash(ctx)
		overrideHash = instanceStyle.Hash64()
	}
	generation := core.ThemeGeneration(ctx)
	changed := !c.initialized || c.theme != active || c.themeGeneration != generation ||
		c.scopeHash != scopeHash || c.overrideHash != overrideHash || c.enabled != enabled
	if !changed {
		return
	}
	if c.initialized {
		switch {
		case c.theme != active || c.themeGeneration != generation:
			c.themeInvalid++
		case c.scopeHash != scopeHash || c.enabled != enabled:
			c.scopeInvalid++
		case c.overrideHash != overrideHash:
			c.overrideInvalid++
		}
	}
	c.theme = active
	c.themeGeneration = generation
	c.scopeHash = scopeHash
	c.overrideHash = overrideHash
	c.enabled = enabled
	c.initialized = true
	c.cache = computed.ProviderCacheMap[listBoxResolvedPartKey, style.ResolvedStyle]{}
	c.cache.SetCapacity(listBoxResolvedPartCacheCapacity)
}

func (c *listBoxResolvedPartCache) resolve(
	active *theme.Theme,
	key listBoxResolvedPartKey,
	load func() style.ResolvedStyle,
) style.ResolvedStyle {
	value, hit := c.cache.ResolveWithStatus(active, key, load)
	if hit {
		c.hits++
	} else {
		c.misses++
	}
	return value
}

func (c *listBoxResolvedPartCache) stats() listBoxResolvedPartCacheStats {
	if c == nil {
		return listBoxResolvedPartCacheStats{}
	}
	cache := c.cache.Stats()
	return listBoxResolvedPartCacheStats{
		Hits:                  c.hits,
		Misses:                c.misses,
		ThemeInvalidations:    c.themeInvalid,
		ScopeInvalidations:    c.scopeInvalid,
		OverrideInvalidations: c.overrideInvalid,
		Size:                  cache.Size,
		Evictions:             cache.Evictions,
	}
}

func (c *listBoxVisualCache) resolve(
	active *theme.Theme,
	request theme.ListBoxRequest,
	load func() theme.ListBoxVisual,
) theme.ListBoxVisual {
	if c == nil || request.Override != nil {
		return load()
	}
	return c.cache.Resolve(active, request, load)
}

type listBoxDataCache struct {
	ready   bool
	version uint64
	entries []listBoxEntry
	items   []Item
}

const listBoxTypeaheadTimeout = 500 * time.Millisecond

type listBoxKeyResult struct {
	focusKey  string
	actionKey string
}

// listBoxVisualFor keeps the test and cache inspection helpers on the neutral
// request path used by production layout.
func (s *listBoxState) listBoxVisualFor(ctx *core.Context) theme.ListBoxVisual {
	return s.listBoxVisualForRequest(ctx, theme.ListBoxRequest{})
}

// listBoxVisualForRequest resolves an instance-specific visual plan without
// polluting the shared declaration cache. An override can carry arbitrary
// style declarations, so its provider result is intentionally request-local.
func (s *listBoxState) listBoxVisualForRequest(ctx *core.Context, request theme.ListBoxRequest) theme.ListBoxVisual {
	active := core.ActiveTheme(ctx)
	return s.visualCache.resolve(active, request, func() theme.ListBoxVisual {
		return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.ListBox, theme.ValidateListBoxVisual)
	})
}

func (s *listBoxState) itemVisualFor(
	active *theme.Theme,
	visual theme.ListBoxVisual,
	request theme.ListBoxItemRequest,
) theme.ListBoxItemVisual {
	if visual.Item == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	s.itemVisualCache.SetCapacity(listBoxItemVisualCacheCapacity)
	return s.itemVisualCache.Resolve(active, request, func() theme.ListBoxItemVisual {
		itemVisual := visual.Item(request)
		theme.ValidateListBoxItemVisual(itemVisual, request.State)
		itemVisual.CacheUnsafe = itemVisual.Defaults.CacheUnsafe() || itemVisual.Variant.CacheUnsafe()
		return itemVisual
	})
}

func (s *listBoxState) resolvedItemPart(
	ctx *core.Context,
	gtx layout.Context,
	active *theme.Theme,
	transitionKey string,
	request theme.ListBoxItemRequest,
	itemVisual theme.ListBoxItemVisual,
	instanceStyle style.Style,
) style.ResolvedStyle {
	// CacheUnsafe is normalized when the provider result enters the item visual
	// cache, so the steady-state item path does not rescan immutable declarations.
	if !s.resolvedPartCache.enabled || itemVisual.CacheUnsafe {
		return computed.ResolvePartWithKey(
			ctx,
			gtx,
			transitionKey,
			style.PartItem,
			request.State,
			itemVisual.Defaults,
			itemVisual.Variant,
			style.Style{},
			instanceStyle,
		)
	}
	key := listBoxResolvedPartKey{
		part:    style.PartItem,
		variant: request.Variant,
		state:   request.State.Bits(),
	}
	static := s.resolvedPartCache.resolve(active, key, func() style.ResolvedStyle {
		return computed.ResolvePartStaticShared(
			ctx,
			style.PartItem,
			request.State,
			itemVisual.Defaults,
			itemVisual.Variant,
			style.Style{},
			instanceStyle,
		)
	})
	return computed.ApplyTransitionsWithSharedKey(ctx, gtx, transitionKey, static)
}

func (s *listBoxState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *listBoxState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *listBoxState) item(key string) *listBoxItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *listBoxState) resolveEntries(widget Widget) ([]listBoxEntry, []Item) {
	if !widget.hasDataVersion {
		entries, items := widget.entriesAndItems()
		s.checkItems(items)
		s.dataCache.ready = false
		return entries, items
	}
	if s.dataCache.ready && s.dataCache.version == widget.dataVersion {
		return s.dataCache.entries, s.dataCache.items
	}
	entries, items := widget.entriesAndItems()
	s.checkItems(items)
	s.dataCache = listBoxDataCache{
		ready:   true,
		version: widget.dataVersion,
		entries: entries,
		items:   items,
	}
	return entries, items
}

func (s *listBoxState) checkItems(items []Item) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]int, len(items))
	} else {
		clear(s.itemKeys)
	}
	for index, item := range items {
		if item.Key == "" {
			panic("flowui: empty listbox item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate listbox item key %q", item.Key))
		}
		s.itemKeys[item.Key] = index
	}
}

func (s *listBoxState) updateKeys(gtx layout.Context, items []Item, widget Widget, selectedKey string) listBoxKeyResult {
	s.keyFilters = s.keyFilters[:0]
	current := s.focusedIndex(gtx)
	if current < 0 {
		if index, ok := s.itemKeys[selectedKey]; ok {
			current = index
			s.focusedKey = selectedKey
		}
	}
	for itemKey, itemState := range s.items {
		index, ok := s.itemKeys[itemKey]
		if !ok || index < 0 || index >= len(items) {
			continue
		}
		tag := &itemState.Clickable
		if widget.itemDisabled(items[index]) {
			s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameHome,
				key.NameEnd,
			)...)
		} else {
			s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameHome,
				key.NameEnd,
				key.NameEnter,
				key.NameReturn,
				key.NameSpace,
				"",
			)...)
		}
	}
	if len(s.keyFilters) == 0 {
		return listBoxKeyResult{}
	}

	// itemDisabled resolves in O(1) via the widget's disabled-key set, so the
	// per-keystroke navigation below does not linearly scan disabledKeys.
	list := navigation.List{
		Count:    len(items),
		Disabled: func(i int) bool { return widget.itemDisabled(items[i]) },
		Label:    func(i int) string { return items[i].Label },
	}

	result := listBoxKeyResult{}
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			if event.State != key.Press {
				continue
			}
			if next, ok := navigation.Move(list, current, 1, false); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameUpArrow:
			if event.State != key.Press {
				continue
			}
			if next, ok := navigation.Move(list, current, -1, false); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameHome:
			if event.State != key.Press {
				continue
			}
			if next, ok := navigation.First(list); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameEnd:
			if event.State != key.Press {
				continue
			}
			if next, ok := navigation.Last(list); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			switch event.State {
			case key.Press:
				s.pressedKey = event.Name
				s.pressedActionKey = ""
				if current >= 0 && current < len(items) && !widget.itemDisabled(items[current]) {
					s.pressedActionKey = items[current].Key
				}
			case key.Release:
				if s.pressedKey != event.Name || s.pressedActionKey == "" {
					s.pressedKey = ""
					s.pressedActionKey = ""
					continue
				}
				actionKey := s.pressedActionKey
				s.pressedKey = ""
				s.pressedActionKey = ""
				if item, ok := listBoxItemByKey(items, actionKey); ok && !widget.itemDisabled(item) {
					result.actionKey = actionKey
				}
			}
		default:
			if event.State != key.Press || event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := navigation.Printable(event.Name)
			if text == "" {
				continue
			}
			query := s.typeahead.Append(gtx.Now, text)
			next, ok := navigation.Match(list, current, query)
			if !ok && query != text {
				s.typeahead.Set(text)
				next, ok = navigation.Match(list, current, text)
			}
			if ok {
				current = next
				result.focusKey = items[next].Key
			}
		}
	}
	return result
}

// appendTypeahead delegates to the shared navigation.Typeahead accumulator. It remains
// a method so keyboard tests can drive the state directly.
func (s *listBoxState) appendTypeahead(now time.Time, text string) string {
	return s.typeahead.Append(now, text)
}

func (s *listBoxState) focusedIndex(gtx layout.Context) int {
	if s.focusedKey != "" {
		if index, ok := s.itemKeys[s.focusedKey]; ok {
			if itemState := s.items[s.focusedKey]; itemState != nil && gtx.Focused(&itemState.Clickable) {
				return index
			}
		}
	}
	for key, itemState := range s.items {
		if index, ok := s.itemKeys[key]; ok && gtx.Focused(&itemState.Clickable) {
			s.focusedKey = key
			return index
		}
	}
	return -1
}

func (s *listBoxState) keyboardActiveKey(widget Widget) string {
	if widget.selectionMode != SelectionMultiple {
		return widget.selectedKey
	}
	for _, key := range widget.selectedKeys {
		if _, ok := s.itemKeys[key]; ok {
			return key
		}
	}
	return ""
}

func listBoxIndexByKey(items []Item, key string) int {
	for i, item := range items {
		if item.Key == key {
			return i
		}
	}
	return -1
}

func listBoxItemByKey(items []Item, key string) (Item, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func listBoxItemDisabled(item Item, disabledKeys []string) bool {
	return item.Disabled || listBoxContainsKey(disabledKeys, item.Key)
}

// sliceList adapts an item slice + disabled-key slice into a navigation.List for the
// exported one-shot helpers used by Select/ComboBox. These lookups are not on
// the per-keystroke path, so the slice-based disabled check is acceptable.
func sliceList(items []Item, disabledKeys []string) navigation.List {
	return navigation.List{
		Count:    len(items),
		Disabled: func(i int) bool { return listBoxItemDisabled(items[i], disabledKeys) },
		Label:    func(i int) string { return items[i].Label },
	}
}

func FirstEnabled(items []Item, disabledKeys []string) (int, bool) {
	return navigation.First(sliceList(items, disabledKeys))
}

func LastEnabled(items []Item, disabledKeys []string) (int, bool) {
	return navigation.Last(sliceList(items, disabledKeys))
}

func IndexByKey(items []Item, key string) int {
	return listBoxIndexByKey(items, key)
}

func ItemDisabled(item Item, disabledKeys []string) bool {
	return listBoxItemDisabled(item, disabledKeys)
}

func FocusItem(ctx *core.Context, key, itemKey string, visible bool) bool {
	return focusItem(ctx, core.FullKey(ctx, key), itemKey, visible)
}

func FocusDerivedItem(ctx *core.Context, owner, role, itemKey string, visible bool) bool {
	return focusItem(ctx, core.DerivedKey(ctx, owner, role), itemKey, visible)
}

func focusItem(ctx *core.Context, stateKey, itemKey string, visible bool) bool {
	clickable, _, ok := item(ctx, stateKey, itemKey)
	if !ok {
		return false
	}
	core.RequestFocusVisible(ctx, clickable, visible)
	return true
}

type listBoxItemState struct {
	optionrow.FocusableState
	keyFilters    core.KeyFilterCache
	styleKey      string
	transitionKey string
}

// Disclosure helpers for single selection mode
func listBoxSingleDisclosureCfg(widget Widget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultKey,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *listBoxState) currentSingleValue(widget Widget) string {
	return s.disclosureSingle.Current(listBoxSingleDisclosureCfg(widget))
}

func (s *listBoxState) bindSingle(widget Widget) {
	s.disclosureSingle.Bind(listBoxSingleDisclosureCfg(widget))
}

func (s *listBoxState) requestSingleValue(widget Widget, value string) string {
	newValue, _ := s.disclosureSingle.Request(listBoxSingleDisclosureCfg(widget), value)
	return newValue
}

// Manual disclosure for multiple selection mode ([]string is not comparable)
func (s *listBoxState) currentMultipleValue(widget Widget) []string {
	if widget.hasSelectedKeys {
		// Controlled mode
		return widget.selectedKeys
	}
	if !s.multipleReady {
		// First frame - initialize from default or empty
		if widget.hasDefaultKeys {
			s.multipleValue = append([]string(nil), widget.defaultSelectedKeys...)
		} else {
			s.multipleValue = nil
		}
		s.multipleReady = true
	}
	return s.multipleValue
}

func (s *listBoxState) bindMultiple(widget Widget) {
	// For controlled mode, just validate
	if widget.hasSelectedKeys {
		return
	}
	// For uncontrolled mode, ensure initialized
	if !s.multipleReady {
		if widget.hasDefaultKeys {
			s.multipleValue = append([]string(nil), widget.defaultSelectedKeys...)
		} else {
			s.multipleValue = nil
		}
		s.multipleReady = true
	}
}

func (s *listBoxState) requestMultipleValue(widget Widget, value []string) []string {
	if widget.hasSelectedKeys {
		// Controlled mode - just call onChange
		if widget.onSelectionChange != nil {
			widget.onSelectionChange(value)
		}
		return widget.selectedKeys
	}
	// Uncontrolled mode - update internal state
	s.multipleValue = append([]string(nil), value...)
	if widget.onSelectionChange != nil {
		widget.onSelectionChange(value)
	}
	return s.multipleValue
}
