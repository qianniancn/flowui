// Package workbench contains the interaction model shared by editor-like
// shells. It owns tab/group state and the dock snapshot, but deliberately does
// not know how an editor renders or stores document contents.
package workbench

import (
	"errors"
	"fmt"

	"github.com/qianniancn/flowui/internal/components/command"
	"github.com/qianniancn/flowui/internal/components/dock"
	"github.com/qianniancn/flowui/internal/components/panel"
	"github.com/qianniancn/flowui/internal/components/tabs"
	"github.com/qianniancn/flowui/internal/frame"
)

const SnapshotVersion uint16 = 1

var (
	ErrEmptyKey           = errors.New("flowui: workbench key must not be empty")
	ErrDuplicateKey       = errors.New("flowui: duplicate workbench key")
	ErrUnknownGroup       = errors.New("flowui: unknown workbench group")
	ErrUnknownTab         = errors.New("flowui: unknown workbench tab")
	ErrInvalidPosition    = errors.New("flowui: invalid workbench tab position")
	ErrUnsupportedVersion = errors.New("flowui: unsupported workbench snapshot version")
)

// Tab is the model-side identity of a tab. Content remains a tabs.TabItem in
// the view layer, so this type never embeds an editor implementation.
type Tab struct {
	Key      string
	Title    string
	GroupKey string
	Disabled bool
	Closable bool
}

// Group is an ordered tab group. ActiveKey is kept per group so moving to a
// different dock group does not destroy the previous group's selection.
type Group struct {
	Key       string
	Tabs      []Tab
	ActiveKey string
	Collapsed bool
}

// ChromeState stores visibility of the shell regions around the work area.
// Content widgets, including an editor, remain outside this state.
type ChromeState struct {
	SidebarVisible     bool
	BottomPanelVisible bool
	StatusBarVisible   bool
}

// State is the serializable, application-owned Workbench interaction model.
// Call Clone before retaining a value in an MVU message if the source may be
// mutated by the caller.
type State struct {
	Groups       []Group
	ActiveGroup  string
	FocusedGroup string
	Dock         dock.Snapshot
	Chrome       ChromeState
}

// NewState creates a normalized model from declarative groups.
func NewState(groups []Group) State {
	state := State{Groups: cloneGroups(groups)}
	state.normalize()
	state.Dock = state.Dock.Migrate(state.Dock.RootKey, nil, nil)
	state.Chrome = defaultChromeState()
	return state
}

func defaultChromeState() ChromeState {
	return ChromeState{SidebarVisible: true, BottomPanelVisible: true, StatusBarVisible: true}
}

// Clone returns an independent state value.
func (s State) Clone() State {
	clone := s
	clone.Groups = cloneGroups(s.Groups)
	clone.Dock = s.Dock.Clone()
	return clone
}

func cloneGroups(groups []Group) []Group {
	clone := make([]Group, len(groups))
	for index, group := range groups {
		clone[index] = group
		clone[index].Tabs = append([]Tab(nil), group.Tabs...)
	}
	return clone
}

func (s *State) normalize() {
	seenGroups := make(map[string]struct{}, len(s.Groups))
	groups := s.Groups[:0]
	for _, group := range s.Groups {
		if group.Key == "" {
			continue
		}
		if _, exists := seenGroups[group.Key]; exists {
			continue
		}
		seenGroups[group.Key] = struct{}{}
		seenTabs := make(map[string]struct{}, len(group.Tabs))
		for index := range group.Tabs {
			if group.Tabs[index].Key == "" {
				continue
			}
			if _, exists := seenTabs[group.Tabs[index].Key]; exists {
				continue
			}
			seenTabs[group.Tabs[index].Key] = struct{}{}
			group.Tabs[index].GroupKey = group.Key
		}
		group.Tabs = filterTabs(group.Tabs, seenTabs)
		group.ActiveKey = normalizeTabSelection(group)
		groups = append(groups, group)
	}
	s.Groups = groups
	if !s.hasGroup(s.ActiveGroup) {
		if len(s.Groups) > 0 {
			s.ActiveGroup = s.Groups[0].Key
		} else {
			s.ActiveGroup = ""
		}
	}
	if !s.hasGroup(s.FocusedGroup) {
		s.FocusedGroup = s.ActiveGroup
	}
}

func filterTabs(tabs []Tab, seen map[string]struct{}) []Tab {
	filtered := tabs[:0]
	for _, tab := range tabs {
		if _, ok := seen[tab.Key]; ok {
			filtered = append(filtered, tab)
			delete(seen, tab.Key)
		}
	}
	return filtered
}

func normalizeTabSelection(group Group) string {
	for _, tab := range group.Tabs {
		if tab.Key == group.ActiveKey && !tab.Disabled {
			return group.ActiveKey
		}
	}
	for _, tab := range group.Tabs {
		if !tab.Disabled {
			return tab.Key
		}
	}
	return ""
}

func (s State) hasGroup(key string) bool {
	return s.groupIndex(key) >= 0
}

func (s State) groupIndex(key string) int {
	if key == "" {
		return -1
	}
	for index, group := range s.Groups {
		if group.Key == key {
			return index
		}
	}
	return -1
}

func (s State) tabIndex(groupIndex int, key string) int {
	if groupIndex < 0 || groupIndex >= len(s.Groups) || key == "" {
		return -1
	}
	for index, tab := range s.Groups[groupIndex].Tabs {
		if tab.Key == key {
			return index
		}
	}
	return -1
}

// Group returns a copy of a group, or false when the key is unknown.
func (s State) Group(key string) (Group, bool) {
	index := s.groupIndex(key)
	if index < 0 {
		return Group{}, false
	}
	group := s.Groups[index]
	group.Tabs = append([]Tab(nil), group.Tabs...)
	return group, true
}

// ActiveTab returns the active tab key for a group.
func (s State) ActiveTab(groupKey string) string {
	if group, ok := s.Group(groupKey); ok {
		return group.ActiveKey
	}
	return ""
}

// SetActiveGroup changes the group receiving keyboard and tab-strip actions.
func (s *State) SetActiveGroup(key string) bool {
	if !s.hasGroup(key) {
		return false
	}
	s.ActiveGroup = key
	s.FocusedGroup = key
	return true
}

// SetFocusedGroup updates the group that owns roving focus without changing
// the visible group selection.
func (s *State) SetFocusedGroup(key string) bool {
	if !s.hasGroup(key) {
		return false
	}
	s.FocusedGroup = key
	return true
}

// ActivateTab selects an enabled tab and makes its group active.
func (s *State) ActivateTab(groupKey, tabKey string) bool {
	groupIndex := s.groupIndex(groupKey)
	if groupIndex < 0 {
		return false
	}
	tabIndex := s.tabIndex(groupIndex, tabKey)
	if tabIndex < 0 || s.Groups[groupIndex].Tabs[tabIndex].Disabled {
		return false
	}
	s.Groups[groupIndex].ActiveKey = tabKey
	s.ActiveGroup = groupKey
	s.FocusedGroup = groupKey
	return true
}

// AddGroup appends an ordered group.
func (s *State) AddGroup(group Group) error {
	if group.Key == "" {
		return ErrEmptyKey
	}
	if s.hasGroup(group.Key) {
		return fmt.Errorf("%w %q", ErrDuplicateKey, group.Key)
	}
	group.Tabs = append([]Tab(nil), group.Tabs...)
	seen := make(map[string]struct{}, len(group.Tabs))
	for index := range group.Tabs {
		if group.Tabs[index].Key == "" {
			return ErrEmptyKey
		}
		if _, exists := seen[group.Tabs[index].Key]; exists {
			return fmt.Errorf("%w %q", ErrDuplicateKey, group.Tabs[index].Key)
		}
		seen[group.Tabs[index].Key] = struct{}{}
		group.Tabs[index].GroupKey = group.Key
	}
	group.ActiveKey = normalizeTabSelection(group)
	s.Groups = append(s.Groups, group)
	if s.ActiveGroup == "" {
		s.ActiveGroup, s.FocusedGroup = group.Key, group.Key
	}
	return nil
}

// RemoveGroup removes a group and selects the nearest remaining group.
func (s *State) RemoveGroup(key string) bool {
	index := s.groupIndex(key)
	if index < 0 {
		return false
	}
	activeRemoved := s.ActiveGroup == key
	focusedRemoved := s.FocusedGroup == key
	s.Groups = append(s.Groups[:index], s.Groups[index+1:]...)
	if activeRemoved {
		if len(s.Groups) == 0 {
			s.ActiveGroup = ""
		} else {
			next := index
			if next >= len(s.Groups) {
				next = len(s.Groups) - 1
			}
			s.ActiveGroup = s.Groups[next].Key
		}
	}
	if focusedRemoved || activeRemoved {
		s.FocusedGroup = s.ActiveGroup
	}
	return true
}

// SetGroupCollapsed controls a group's visibility state in a dock adapter.
func (s *State) SetGroupCollapsed(key string, collapsed bool) bool {
	index := s.groupIndex(key)
	if index < 0 {
		return false
	}
	s.Groups[index].Collapsed = collapsed
	return true
}

// SetSidebarVisible controls the primary navigation/sidebar region.
func (s *State) SetSidebarVisible(visible bool) {
	s.Chrome.SidebarVisible = visible
}

// ToggleSidebar switches the primary navigation/sidebar region.
func (s *State) ToggleSidebar() bool {
	s.Chrome.SidebarVisible = !s.Chrome.SidebarVisible
	return s.Chrome.SidebarVisible
}

// SetBottomPanelVisible controls the terminal/output-style bottom region.
func (s *State) SetBottomPanelVisible(visible bool) {
	s.Chrome.BottomPanelVisible = visible
}

// ToggleBottomPanel switches the terminal/output-style bottom region.
func (s *State) ToggleBottomPanel() bool {
	s.Chrome.BottomPanelVisible = !s.Chrome.BottomPanelVisible
	return s.Chrome.BottomPanelVisible
}

// SetStatusBarVisible controls the bottom status bar region.
func (s *State) SetStatusBarVisible(visible bool) {
	s.Chrome.StatusBarVisible = visible
}

// ToggleStatusBar switches the bottom status bar region.
func (s *State) ToggleStatusBar() bool {
	s.Chrome.StatusBarVisible = !s.Chrome.StatusBarVisible
	return s.Chrome.StatusBarVisible
}

// AddTab inserts a tab at index. A negative index appends the tab.
func (s *State) AddTab(groupKey string, tab Tab, index int) error {
	groupIndex := s.groupIndex(groupKey)
	if groupIndex < 0 {
		return fmt.Errorf("%w %q", ErrUnknownGroup, groupKey)
	}
	if tab.Key == "" {
		return ErrEmptyKey
	}
	if s.tabIndex(groupIndex, tab.Key) >= 0 {
		return fmt.Errorf("%w %q", ErrDuplicateKey, tab.Key)
	}
	if index < 0 {
		index = len(s.Groups[groupIndex].Tabs)
	}
	if index > len(s.Groups[groupIndex].Tabs) {
		return ErrInvalidPosition
	}
	tab.GroupKey = groupKey
	tabsValue := s.Groups[groupIndex].Tabs
	tabsValue = append(tabsValue, Tab{})
	copy(tabsValue[index+1:], tabsValue[index:])
	tabsValue[index] = tab
	s.Groups[groupIndex].Tabs = tabsValue
	if s.Groups[groupIndex].ActiveKey == "" && !tab.Disabled {
		s.Groups[groupIndex].ActiveKey = tab.Key
	}
	return nil
}

// ActivateNextTab moves selection by delta and wraps around enabled tabs.
func (s *State) ActivateNextTab(groupKey string, delta int) bool {
	groupIndex := s.groupIndex(groupKey)
	if groupIndex < 0 || delta == 0 {
		return false
	}
	items := s.Groups[groupIndex].Tabs
	if len(items) == 0 {
		return false
	}
	current := s.tabIndex(groupIndex, s.Groups[groupIndex].ActiveKey)
	for step := 1; step <= len(items); step++ {
		index := (current + delta*step) % len(items)
		if index < 0 {
			index += len(items)
		}
		if !items[index].Disabled {
			return s.ActivateTab(groupKey, items[index].Key)
		}
	}
	return false
}

// CloseTab removes a tab and selects the nearest enabled fallback.
func (s *State) CloseTab(groupKey, tabKey string) bool {
	groupIndex := s.groupIndex(groupKey)
	if groupIndex < 0 {
		return false
	}
	index := s.tabIndex(groupIndex, tabKey)
	if index < 0 || !s.Groups[groupIndex].Tabs[index].Closable || s.Groups[groupIndex].Tabs[index].Disabled {
		return false
	}
	items := s.Groups[groupIndex].Tabs
	wasActive := items[index].Key == s.Groups[groupIndex].ActiveKey
	remaining := append(items[:index], items[index+1:]...)
	s.Groups[groupIndex].Tabs = remaining
	if wasActive {
		s.Groups[groupIndex].ActiveKey = ""
		fallback := -1
		for candidate := index; candidate < len(remaining); candidate++ {
			if !remaining[candidate].Disabled {
				fallback = candidate
				break
			}
		}
		if fallback < 0 {
			for candidate := min(index-1, len(remaining)-1); candidate >= 0; candidate-- {
				if !remaining[candidate].Disabled {
					fallback = candidate
					break
				}
			}
		}
		if fallback >= 0 {
			s.Groups[groupIndex].ActiveKey = remaining[fallback].Key
		}
	}
	return true
}

// ReorderTab moves a tab within its group. The target is the final index.
func (s *State) ReorderTab(groupKey, tabKey string, target int) bool {
	groupIndex := s.groupIndex(groupKey)
	if groupIndex < 0 {
		return false
	}
	items := s.Groups[groupIndex].Tabs
	from := s.tabIndex(groupIndex, tabKey)
	if from < 0 || target < 0 || target >= len(items) {
		return false
	}
	if from == target {
		return true
	}
	tab := items[from]
	items = append(items[:from], items[from+1:]...)
	items = append(items, Tab{})
	copy(items[target+1:], items[target:])
	items[target] = tab
	s.Groups[groupIndex].Tabs = items
	return true
}

// MoveTab moves a tab between groups and activates it in the destination.
func (s *State) MoveTab(fromGroup, toGroup, tabKey string, target int) bool {
	fromIndex, toIndex := s.groupIndex(fromGroup), s.groupIndex(toGroup)
	if fromIndex < 0 || toIndex < 0 {
		return false
	}
	index := s.tabIndex(fromIndex, tabKey)
	if index < 0 || s.Groups[fromIndex].Tabs[index].Disabled {
		return false
	}
	if fromGroup == toGroup {
		if target < 0 {
			target = len(s.Groups[fromIndex].Tabs) - 1
		}
		if !s.ReorderTab(fromGroup, tabKey, target) {
			return false
		}
		return s.ActivateTab(toGroup, tabKey)
	}
	if s.tabIndex(toIndex, tabKey) >= 0 {
		return false
	}
	if target < 0 {
		target = len(s.Groups[toIndex].Tabs)
	}
	if target > len(s.Groups[toIndex].Tabs) {
		return false
	}
	tab := s.Groups[fromIndex].Tabs[index]
	s.Groups[fromIndex].Tabs = append(s.Groups[fromIndex].Tabs[:index], s.Groups[fromIndex].Tabs[index+1:]...)
	tab.GroupKey = toGroup
	items := append(s.Groups[toIndex].Tabs, Tab{})
	copy(items[target+1:], items[target:])
	items[target] = tab
	s.Groups[toIndex].Tabs = items
	if s.Groups[fromIndex].ActiveKey == tabKey {
		s.Groups[fromIndex].ActiveKey = normalizeTabSelection(s.Groups[fromIndex])
	}
	return s.ActivateTab(toGroup, tabKey)
}

// SetDockSnapshot applies dock geometry owned by the Workbench model.
func (s *State) SetDockSnapshot(snapshot dock.Snapshot) {
	s.Dock = snapshot.Migrate(snapshot.RootKey, nil, nil)
}

// Snapshot is the versioned persistence contract for a Workbench model. It
// stores stable keys and order, never widget pointers or editor contents.
type Snapshot struct {
	Version      uint16
	ActiveGroup  string
	FocusedGroup string
	Groups       []GroupSnapshot
	Dock         dock.Snapshot
	Chrome       ChromeState
}

type GroupSnapshot struct {
	Key       string
	TabOrder  []string
	ActiveKey string
	Collapsed bool
}

// Migration remaps keys when a persisted layout is renamed. Unknown entries
// are dropped during Restore, allowing removed panels to disappear cleanly.
type Migration struct {
	GroupAliases map[string]string
	TabAliases   map[string]string
	DockAliases  map[string]string
}

func (s State) Snapshot() Snapshot {
	snapshot := Snapshot{Version: SnapshotVersion, ActiveGroup: s.ActiveGroup, FocusedGroup: s.FocusedGroup, Dock: s.Dock.Migrate(s.Dock.RootKey, nil, nil), Chrome: s.Chrome}
	snapshot.Groups = make([]GroupSnapshot, len(s.Groups))
	for index, group := range s.Groups {
		order := make([]string, len(group.Tabs))
		for tabIndex, tab := range group.Tabs {
			order[tabIndex] = tab.Key
		}
		snapshot.Groups[index] = GroupSnapshot{Key: group.Key, TabOrder: order, ActiveKey: group.ActiveKey, Collapsed: group.Collapsed}
	}
	return snapshot
}

func (s Snapshot) Clone() Snapshot {
	clone := s
	clone.Groups = make([]GroupSnapshot, len(s.Groups))
	for index, group := range s.Groups {
		clone.Groups[index] = group
		clone.Groups[index].TabOrder = append([]string(nil), group.TabOrder...)
	}
	clone.Dock = s.Dock.Clone()
	return clone
}

// Migrate applies aliases and upgrades an older version without validating
// against the current model. Restore performs the model-aware filtering.
func (s Snapshot) Migrate(migration Migration) Snapshot {
	clone := s.Clone()
	clone.Version = SnapshotVersion
	for index := range clone.Groups {
		clone.Groups[index].Key = alias(clone.Groups[index].Key, migration.GroupAliases)
		clone.Groups[index].ActiveKey = alias(clone.Groups[index].ActiveKey, migration.TabAliases)
		for tabIndex, key := range clone.Groups[index].TabOrder {
			clone.Groups[index].TabOrder[tabIndex] = alias(key, migration.TabAliases)
		}
	}
	clone.ActiveGroup = alias(clone.ActiveGroup, migration.GroupAliases)
	clone.FocusedGroup = alias(clone.FocusedGroup, migration.GroupAliases)
	clone.Dock = clone.Dock.Migrate(clone.Dock.RootKey, nil, migration.DockAliases)
	return clone
}

func alias(key string, aliases map[string]string) string {
	seen := map[string]struct{}{}
	for key != "" {
		if _, ok := seen[key]; ok {
			break
		}
		seen[key] = struct{}{}
		next, ok := aliases[key]
		if !ok || next == "" {
			break
		}
		key = next
	}
	return key
}

// Restore applies a persisted snapshot to the current declarative model.
// New groups/tabs keep their current order, while known persisted entries are
// reordered and unknown/removed entries are ignored.
func (s *State) Restore(snapshot Snapshot, migration Migration) error {
	if s == nil {
		return errors.New("flowui: nil workbench state")
	}
	if snapshot.Version > SnapshotVersion {
		return fmt.Errorf("%w %d", ErrUnsupportedVersion, snapshot.Version)
	}
	if snapshot.Dock.Version > dock.SnapshotVersion {
		return fmt.Errorf("%w dock %d", ErrUnsupportedVersion, snapshot.Dock.Version)
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}
	legacy := snapshot.Version == 0
	snapshot = snapshot.Migrate(migration)
	if err := snapshot.Validate(); err != nil {
		return err
	}
	byGroup := make(map[string]GroupSnapshot, len(snapshot.Groups))
	for _, saved := range snapshot.Groups {
		if saved.Key != "" {
			byGroup[saved.Key] = saved
		}
	}
	for groupIndex := range s.Groups {
		group := &s.Groups[groupIndex]
		saved, ok := byGroup[group.Key]
		if !ok {
			group.ActiveKey = normalizeTabSelection(*group)
			continue
		}
		positions := make(map[string]int, len(group.Tabs))
		for index, tab := range group.Tabs {
			positions[tab.Key] = index
		}
		ordered := make([]Tab, 0, len(group.Tabs))
		seen := make(map[string]struct{}, len(group.Tabs))
		for _, key := range saved.TabOrder {
			if index, exists := positions[key]; exists {
				ordered = append(ordered, group.Tabs[index])
				seen[key] = struct{}{}
			}
		}
		for _, tab := range group.Tabs {
			if _, exists := seen[tab.Key]; !exists {
				ordered = append(ordered, tab)
			}
		}
		group.Tabs = ordered
		group.ActiveKey = saved.ActiveKey
		group.Collapsed = saved.Collapsed
		group.ActiveKey = normalizeTabSelection(*group)
	}
	if s.hasGroup(snapshot.ActiveGroup) {
		s.ActiveGroup = snapshot.ActiveGroup
	}
	if s.hasGroup(snapshot.FocusedGroup) {
		s.FocusedGroup = snapshot.FocusedGroup
	}
	s.Dock = snapshot.Dock.Clone()
	if !legacy {
		s.Chrome = snapshot.Chrome
	}
	s.normalize()
	return nil
}

// EventKind identifies a state transition emitted by Controller.
type EventKind string

const (
	EventGroupActivated EventKind = "group-activated"
	EventTabActivated   EventKind = "tab-activated"
	EventTabAdded       EventKind = "tab-added"
	EventTabClosed      EventKind = "tab-closed"
	EventTabReordered   EventKind = "tab-reordered"
	EventTabMoved       EventKind = "tab-moved"
	EventGroupAdded     EventKind = "group-added"
	EventGroupRemoved   EventKind = "group-removed"
	EventGroupCollapsed EventKind = "group-collapsed"
	EventDockChanged    EventKind = "dock-changed"
	EventChromeChanged  EventKind = "chrome-changed"
	EventLayoutRestored EventKind = "layout-restored"
)

// CommandID identifies a standard Workbench action.
type CommandID string

const (
	CommandNextTab       CommandID = "workbench.next-tab"
	CommandPreviousTab   CommandID = "workbench.previous-tab"
	CommandCloseTab      CommandID = "workbench.close-tab"
	CommandToggleSidebar CommandID = "workbench.toggle-sidebar"
	CommandTogglePanel   CommandID = "workbench.toggle-bottom-panel"
	CommandToggleStatus  CommandID = "workbench.toggle-status-bar"
)

type Event struct {
	Kind          EventKind
	GroupKey      string
	TabKey        string
	FromGroup     string
	ToGroup       string
	Index         int
	PreviousIndex int
}

// Controller is the event-emitting adapter used to connect model state to
// Tabs and Dock widgets. It is safe to keep in an application model pointer.
type Controller struct {
	state     State
	listeners []func(Event)
}

func NewController(state State) *Controller {
	state = state.Clone()
	state.normalize()
	state.Dock = state.Dock.Migrate(state.Dock.RootKey, nil, nil)
	return &Controller{state: state}
}

func (c *Controller) State() State {
	if c == nil {
		return State{}
	}
	return c.state.Clone()
}

func (c *Controller) OnEvent(listener func(Event)) *Controller {
	if c != nil && listener != nil {
		c.listeners = append(c.listeners, listener)
	}
	return c
}

func (c *Controller) emit(event Event) {
	for _, listener := range append([]func(Event){}, c.listeners...) {
		listener(event)
	}
}

func (c *Controller) SetActiveGroup(key string) bool {
	if c == nil || !c.state.SetActiveGroup(key) {
		return false
	}
	c.emit(Event{Kind: EventGroupActivated, GroupKey: key})
	return true
}

func (c *Controller) SetFocusedGroup(key string) bool {
	if c == nil || !c.state.SetFocusedGroup(key) {
		return false
	}
	return true
}

func (c *Controller) AddGroup(group Group) error {
	if c == nil {
		return ErrUnknownGroup
	}
	if err := c.state.AddGroup(group); err != nil {
		return err
	}
	c.emit(Event{Kind: EventGroupAdded, GroupKey: group.Key})
	return nil
}

func (c *Controller) RemoveGroup(key string) bool {
	if c == nil || !c.state.RemoveGroup(key) {
		return false
	}
	c.emit(Event{Kind: EventGroupRemoved, GroupKey: key})
	return true
}

func (c *Controller) SetGroupCollapsed(key string, collapsed bool) bool {
	if c == nil || !c.state.SetGroupCollapsed(key, collapsed) {
		return false
	}
	c.emit(Event{Kind: EventGroupCollapsed, GroupKey: key})
	return true
}

func (c *Controller) SetSidebarVisible(visible bool) {
	if c == nil {
		return
	}
	if c.state.Chrome.SidebarVisible == visible {
		return
	}
	c.state.SetSidebarVisible(visible)
	c.emit(Event{Kind: EventChromeChanged})
}

func (c *Controller) ToggleSidebar() bool {
	if c == nil {
		return false
	}
	visible := c.state.ToggleSidebar()
	c.emit(Event{Kind: EventChromeChanged})
	return visible
}

func (c *Controller) SetBottomPanelVisible(visible bool) {
	if c == nil {
		return
	}
	if c.state.Chrome.BottomPanelVisible == visible {
		return
	}
	c.state.SetBottomPanelVisible(visible)
	c.emit(Event{Kind: EventChromeChanged})
}

func (c *Controller) ToggleBottomPanel() bool {
	if c == nil {
		return false
	}
	visible := c.state.ToggleBottomPanel()
	c.emit(Event{Kind: EventChromeChanged})
	return visible
}

func (c *Controller) SetStatusBarVisible(visible bool) {
	if c == nil {
		return
	}
	if c.state.Chrome.StatusBarVisible == visible {
		return
	}
	c.state.SetStatusBarVisible(visible)
	c.emit(Event{Kind: EventChromeChanged})
}

func (c *Controller) ToggleStatusBar() bool {
	if c == nil {
		return false
	}
	visible := c.state.ToggleStatusBar()
	c.emit(Event{Kind: EventChromeChanged})
	return visible
}

func (c *Controller) ActivateTab(groupKey, tabKey string) bool {
	if c == nil || !c.state.ActivateTab(groupKey, tabKey) {
		return false
	}
	c.emit(Event{Kind: EventTabActivated, GroupKey: groupKey, TabKey: tabKey})
	return true
}

func (c *Controller) ActivateNextTab(groupKey string, delta int) bool {
	if c == nil || !c.state.ActivateNextTab(groupKey, delta) {
		return false
	}
	c.emit(Event{Kind: EventTabActivated, GroupKey: groupKey, TabKey: c.state.ActiveTab(groupKey)})
	return true
}

func (c *Controller) AddTab(groupKey string, tab Tab, index int) error {
	if c == nil {
		return ErrUnknownGroup
	}
	if err := c.state.AddTab(groupKey, tab, index); err != nil {
		return err
	}
	if index < 0 {
		index = len(c.state.Groups[c.state.groupIndex(groupKey)].Tabs) - 1
	}
	c.emit(Event{Kind: EventTabAdded, GroupKey: groupKey, TabKey: tab.Key, Index: index})
	return nil
}

func (c *Controller) CloseTab(groupKey, tabKey string) bool {
	if c == nil || !c.state.CloseTab(groupKey, tabKey) {
		return false
	}
	c.emit(Event{Kind: EventTabClosed, GroupKey: groupKey, TabKey: tabKey})
	return true
}

func (c *Controller) ReorderTab(groupKey, tabKey string, target int) bool {
	if c == nil {
		return false
	}
	groupIndex := c.state.groupIndex(groupKey)
	previous := c.state.tabIndex(groupIndex, tabKey)
	if !c.state.ReorderTab(groupKey, tabKey, target) {
		return false
	}
	c.emit(Event{Kind: EventTabReordered, GroupKey: groupKey, TabKey: tabKey, Index: target, PreviousIndex: previous})
	return true
}

func (c *Controller) MoveTab(fromGroup, toGroup, tabKey string, target int) bool {
	if c == nil {
		return false
	}
	fromIndex := c.state.groupIndex(fromGroup)
	toIndex := c.state.groupIndex(toGroup)
	previous := c.state.tabIndex(fromIndex, tabKey)
	finalTarget := target
	if fromIndex >= 0 && toIndex >= 0 && target < 0 {
		if fromGroup == toGroup {
			finalTarget = len(c.state.Groups[fromIndex].Tabs) - 1
		} else {
			finalTarget = len(c.state.Groups[toIndex].Tabs)
		}
	}
	if !c.state.MoveTab(fromGroup, toGroup, tabKey, target) {
		return false
	}
	c.emit(Event{Kind: EventTabMoved, FromGroup: fromGroup, ToGroup: toGroup, TabKey: tabKey, Index: finalTarget, PreviousIndex: previous})
	return true
}

func (c *Controller) SetDockSnapshot(snapshot dock.Snapshot) {
	if c == nil {
		return
	}
	c.state.SetDockSnapshot(snapshot)
	c.emit(Event{Kind: EventDockChanged})
}

func (c *Controller) SetDockMaximized(key string) {
	if c == nil {
		return
	}
	c.state.Dock.MaximizedKey = key
	c.emit(Event{Kind: EventDockChanged})
}

func (c *Controller) SetDockCollapsed(key string, collapsed dock.CollapseState) {
	if c == nil {
		return
	}
	if c.state.Dock.Collapsed == nil {
		c.state.Dock.Collapsed = make(map[string]dock.CollapseState)
	}
	c.state.Dock.Collapsed[key] = collapsed
	c.emit(Event{Kind: EventDockChanged})
}

// Snapshot returns the current versioned layout snapshot.
func (c *Controller) Snapshot() Snapshot {
	if c == nil {
		return Snapshot{Version: SnapshotVersion}
	}
	return c.state.Snapshot()
}

// Commands returns the standard shell commands for the current active group.
// The callbacks are dynamic and resolve the active tab at execution time, so
// a command list can be built once per frame without stale tab captures.
func (c *Controller) Commands() []command.Command {
	if c == nil {
		return nil
	}
	groupKey := c.state.ActiveGroup
	group, _ := c.state.Group(groupKey)
	enabledTabs := 0
	activeClosable := false
	for _, tab := range group.Tabs {
		if !tab.Disabled {
			enabledTabs++
		}
		if tab.Key == group.ActiveKey && tab.Closable && !tab.Disabled {
			activeClosable = true
		}
	}
	return []command.Command{
		command.New(string(CommandNextTab), "Next Tab").
			Disabled(enabledTabs < 2).
			Shortcut(command.KeyShortcut("tab", command.ShortcutPrimary)).
			OnExecute(func() { c.ActivateNextTab(c.state.ActiveGroup, 1) }),
		command.New(string(CommandPreviousTab), "Previous Tab").
			Disabled(enabledTabs < 2).
			Shortcut(command.KeyShortcut("tab", command.ShortcutPrimary|command.ShortcutShift)).
			OnExecute(func() { c.ActivateNextTab(c.state.ActiveGroup, -1) }),
		command.New(string(CommandCloseTab), "Close Tab").
			Disabled(!activeClosable).
			Shortcut(command.KeyShortcut("w", command.ShortcutPrimary)).
			OnExecute(func() {
				group := c.state.ActiveGroup
				c.CloseTab(group, c.state.ActiveTab(group))
			}),
		command.New(string(CommandToggleSidebar), "Toggle Sidebar").
			Shortcut(command.KeyShortcut("b", command.ShortcutPrimary)).
			OnExecute(func() { c.ToggleSidebar() }),
		command.New(string(CommandTogglePanel), "Toggle Bottom Panel").
			Shortcut(command.KeyShortcut("j", command.ShortcutPrimary)).
			OnExecute(func() { c.ToggleBottomPanel() }),
		command.New(string(CommandToggleStatus), "Toggle Status Bar").
			Shortcut(command.KeyShortcut("s", command.ShortcutPrimary|command.ShortcutShift)).
			OnExecute(func() { c.ToggleStatusBar() }),
	}
}

// CommandScope installs the standard Workbench commands around a child. It
// uses FlowUI's normal command routing, so child widgets get first chance to
// consume a shortcut. Call DisableWhenFieldFocused on the returned scope when
// an embedded editor should own every shortcut while focused.
func (c *Controller) CommandScope(child frame.Widget) command.ScopeWidget {
	return command.Scope(c.Commands(), child)
}

func (c *Controller) Restore(snapshot Snapshot, migration Migration) error {
	if c == nil {
		return errors.New("flowui: nil workbench controller")
	}
	if err := c.state.Restore(snapshot, migration); err != nil {
		return err
	}
	c.emit(Event{Kind: EventLayoutRestored})
	return nil
}

// BindTabs connects a Tabs widget's selection callback to this controller.
// Existing callbacks are preserved and run before the controller transition.
func (c *Controller) BindTabs(groupKey string, widget tabs.TabsWidget) tabs.TabsWidget {
	if c == nil {
		return widget
	}
	return widget.SelectedKey(c.state.ActiveTab(groupKey)).BindChange(func(key string) {
		c.ActivateTab(groupKey, key)
	})
}

// BindDock connects a Dock widget's snapshot callback to this controller.
func (c *Controller) BindDock(widget dock.DockLayoutWidget) dock.DockLayoutWidget {
	if c == nil {
		return widget
	}
	return widget.Snapshot(c.state.Dock).BindChange(func(snapshot dock.Snapshot) {
		c.SetDockSnapshot(snapshot)
	})
}

// BindPanel connects a PanelHost's selection callback to a Workbench group.
// PanelHost still owns KeepAlive/DestroyOnHidden behavior; the controller only
// supplies the selected key and observes changes.
func (c *Controller) BindPanel(groupKey string, widget panel.HostWidget) panel.HostWidget {
	if c == nil {
		return widget
	}
	return widget.SelectedKey(c.state.ActiveTab(groupKey)).BindChange(func(key string) {
		c.ActivateTab(groupKey, key)
	})
}
