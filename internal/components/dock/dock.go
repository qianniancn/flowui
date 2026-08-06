// Package dock provides a declarative, recursively splittable workbench
// layout. It deliberately owns geometry and layout state only; panel headers,
// tabs, and collapse buttons remain ordinary FlowUI widgets supplied by the
// caller.
package dock

import (
	"fmt"
	"maps"
	"math"
	"slices"

	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

const stateSlotDockLayout = "dock-layout"

// Orientation controls the axis of a split node.
type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

// CollapseState stores the controlled collapsed state of one split node.
type CollapseState struct {
	First  bool
	Second bool
}

// Snapshot is the serializable geometry state of a DockLayout. Applications
// can keep it in their model and feed it back through Snapshot for persistent
// workbench layouts.
type Snapshot struct {
	// Version is the persistence schema version. Zero is accepted as a legacy
	// snapshot and upgraded during migration.
	Version uint16
	// RootKey identifies the declarative dock tree this snapshot belongs to.
	RootKey      string
	Ratios       map[string]float32
	Collapsed    map[string]CollapseState
	MaximizedKey string
}

const SnapshotVersion uint16 = 1

// Clone returns an independent snapshot suitable for storing in application
// state or passing to a callback.
func (s Snapshot) Clone() Snapshot {
	clone := Snapshot{
		Version:      s.Version,
		RootKey:      s.RootKey,
		Ratios:       make(map[string]float32, len(s.Ratios)),
		Collapsed:    make(map[string]CollapseState, len(s.Collapsed)),
		MaximizedKey: s.MaximizedKey,
	}
	maps.Copy(clone.Ratios, s.Ratios)
	maps.Copy(clone.Collapsed, s.Collapsed)
	return clone
}

// Migrate upgrades a snapshot and filters keys that no longer exist in the
// current tree. Aliases are applied transitively so node renames remain
// explicit and deterministic.
func (s Snapshot) Migrate(rootKey string, validKeys map[string]struct{}, aliases map[string]string) Snapshot {
	clone := s.Clone()
	clone.Version = SnapshotVersion
	if rootKey != "" {
		clone.RootKey = rootKey
	}
	resolve := func(key string) string {
		seen := make(map[string]struct{})
		for key != "" {
			if _, exists := seen[key]; exists {
				break
			}
			seen[key] = struct{}{}
			next, exists := aliases[key]
			if !exists || next == "" {
				break
			}
			key = next
		}
		return key
	}
	valid := func(key string) bool {
		if len(validKeys) == 0 {
			return key != ""
		}
		_, ok := validKeys[key]
		return key != "" && ok
	}
	ratios := make(map[string]float32, len(clone.Ratios))
	for _, source := range slices.Sorted(maps.Keys(clone.Ratios)) {
		key := resolve(source)
		if !valid(key) {
			continue
		}
		if _, exists := ratios[key]; !exists || source == key {
			ratios[key] = normalizeRatio(clone.Ratios[source])
		}
	}
	collapsed := make(map[string]CollapseState, len(clone.Collapsed))
	for _, source := range slices.Sorted(maps.Keys(clone.Collapsed)) {
		key := resolve(source)
		if !valid(key) {
			continue
		}
		if _, exists := collapsed[key]; !exists || source == key {
			collapsed[key] = clone.Collapsed[source]
		}
	}
	clone.Ratios, clone.Collapsed = ratios, collapsed
	clone.MaximizedKey = resolve(clone.MaximizedKey)
	if !valid(clone.MaximizedKey) {
		clone.MaximizedKey = ""
	}
	return clone
}

type nodeKind uint8

const (
	nodePanel nodeKind = iota
	nodeSplit
)

// Node is one panel or split in a DockLayout tree. Nodes are immutable value
// descriptors; fluent methods return an updated copy.
type Node struct {
	kind            nodeKind
	key             string
	content         frame.Widget
	first           *Node
	second          *Node
	orientation     Orientation
	ratio           float32
	hasRatio        bool
	minFirst        unit.Dp
	minSecond       unit.Dp
	collapsedFirst  bool
	collapsedSecond bool
}

// Panel creates a leaf node containing content.
func Panel(key string, content frame.Widget) Node {
	return Node{kind: nodePanel, key: key, content: content}
}

// Split creates a recursively composable split node. Nested Split calls can
// describe any workbench topology while each divider keeps independent state.
func Split(key string, orientation Orientation, first, second Node) Node {
	if orientation != Horizontal && orientation != Vertical {
		panic("flowui: invalid dock orientation")
	}
	return Node{
		kind:        nodeSplit,
		key:         key,
		first:       &first,
		second:      &second,
		orientation: orientation,
	}
}

// Ratio sets the initial ratio of a split node. User resizing is retained in
// the DockLayout state and takes precedence on subsequent frames.
func (n Node) Ratio(ratio float32) Node {
	n.ratio = normalizeRatio(ratio)
	n.hasRatio = true
	return n
}

func (n Node) MinFirst(dp int) Node {
	n.minFirst = unit.Dp(maxInt(dp, 0))
	return n
}

func (n Node) MinSecond(dp int) Node {
	n.minSecond = unit.Dp(maxInt(dp, 0))
	return n
}

// FirstCollapsed controls whether the first branch is removed from the
// visible layout. The last ratio remains available when it is expanded again.
func (n Node) FirstCollapsed(collapsed bool) Node {
	n.collapsedFirst = collapsed
	return n
}

// SecondCollapsed controls whether the second branch is removed from the
// visible layout. The last ratio remains available when it is expanded again.
func (n Node) SecondCollapsed(collapsed bool) Node {
	n.collapsedSecond = collapsed
	return n
}

// DockLayoutWidget lays out a Dock tree and retains divider geometry.
type DockLayoutWidget struct {
	key             string
	root            Node
	snapshot        Snapshot
	hasSnapshot     bool
	defaultSnapshot Snapshot
	hasDefault      bool
	onChange        func(Snapshot)
	onRatioChange   func(string, float32)
	keepAlive       bool
	disabled        bool
	customStyle     flowstyle.Style
	maximizedKey    string
	hasMaximizedKey bool
	defaultMaxKey   string
	hasDefaultMax   bool
}

type dockState struct {
	ratios         map[string]float32
	collapsed      map[string]CollapseState
	keys           map[string]struct{}
	retainedNodes  map[string]struct{}
	initialized    bool
	controlled     bool
	keepAlive      bool
	maximized      string
	maxInitialized bool
	rootKey        string
	retentionDepth int
}

// New creates a DockLayoutWidget from a declarative root node.
func New(key string, root Node) DockLayoutWidget {
	return DockLayoutWidget{key: key, root: root}
}

// Snapshot supplies controlled divider and collapse state. When set, the
// application owns the snapshot and should feed changes from OnChange back on
// the next frame.
func (d DockLayoutWidget) Snapshot(value Snapshot) DockLayoutWidget {
	d.snapshot = value.Clone()
	d.hasSnapshot = true
	return d
}

// DefaultSnapshot seeds uncontrolled state on the first frame.
func (d DockLayoutWidget) DefaultSnapshot(value Snapshot) DockLayoutWidget {
	d.defaultSnapshot = value.Clone()
	d.hasDefault = true
	return d
}

// MaximizedKey sets controlled maximization. Pass an empty key to restore the
// normal tree layout.
func (d DockLayoutWidget) MaximizedKey(key string) DockLayoutWidget {
	d.maximizedKey = key
	d.hasMaximizedKey = true
	return d
}

// DefaultMaximizedKey seeds uncontrolled maximization on the first frame.
func (d DockLayoutWidget) DefaultMaximizedKey(key string) DockLayoutWidget {
	d.defaultMaxKey = key
	d.hasDefaultMax = true
	return d
}

// OnChange reports a complete snapshot whenever a divider is resized.
func (d DockLayoutWidget) OnChange(fn func(Snapshot)) DockLayoutWidget {
	d.onChange = fn
	return d
}

// BindChange composes an observer with the existing snapshot callback. Shell
// controllers use it to link Dock geometry without replacing app callbacks.
func (d DockLayoutWidget) BindChange(fn func(Snapshot)) DockLayoutWidget {
	if fn == nil {
		return d
	}
	previous := d.onChange
	d.onChange = func(snapshot Snapshot) {
		if previous != nil {
			previous(snapshot)
		}
		fn(snapshot)
	}
	return d
}

// OnRatioChange reports the split key and effective ratio for every resize.
func (d DockLayoutWidget) OnRatioChange(fn func(string, float32)) DockLayoutWidget {
	d.onRatioChange = fn
	return d
}

// KeepAlive initializes collapsed branches in an isolated pass and retains
// their transient widget state while they are hidden.
func (d DockLayoutWidget) KeepAlive(enabled bool) DockLayoutWidget {
	d.keepAlive = enabled
	return d
}

func (d DockLayoutWidget) Disabled(disabled bool) DockLayoutWidget {
	d.disabled = disabled
	return d
}

func (d DockLayoutWidget) Style(value flowstyle.Style) DockLayoutWidget {
	d.customStyle = value
	return d
}

func (d DockLayoutWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	rootKey := frame.ClaimKey(ctx, state.KindDockLayout, d.key)
	value := frame.UseStateWith[dockState](ctx, rootKey, stateSlotDockLayout, func() *dockState {
		return &dockState{}
	})
	value.rootKey = rootKey
	value.retentionDepth = frame.StateRetentionDepth(ctx)
	d.initialize(ctx, value)
	clear(value.keys)
	validateNode(d.root, value.keys)
	d.syncSnapshot(ctx, value)
	maximizedKey := d.resolveMaximized(value)
	restoreRoot := frame.PushKey(ctx, d.key)
	defer restoreRoot()
	if maximizedKey != "" {
		if _, ok := findNode(d.root, maximizedKey); ok {
			d.prepareMaximizedAssociations(ctx, d.root, maximizedKey, value)
		} else {
			d.prepareNodeAssociations(ctx, d.root, value)
		}
	} else {
		d.prepareNodeAssociations(ctx, d.root, value)
	}
	return layoutui.LayoutStyled(ctx, gtx, rootKey, flowstyle.StyleState{
		Disabled: d.disabled || !gtx.Enabled(),
	}, d.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if d.disabled {
			gtx = gtx.Disabled()
		}
		if maximizedKey != "" {
			if _, ok := findNode(d.root, maximizedKey); ok {
				return d.layoutMaximized(ctx, gtx, d.root, maximizedKey, value)
			}
		}
		return d.layoutNode(ctx, gtx, d.root, value)
	}))
}

func (d DockLayoutWidget) initialize(ctx *frame.Context, value *dockState) {
	if value.ratios == nil {
		value.ratios = make(map[string]float32)
	}
	if value.collapsed == nil {
		value.collapsed = make(map[string]CollapseState)
	}
	if value.keys == nil {
		value.keys = make(map[string]struct{})
	}
	if value.retainedNodes == nil {
		value.retainedNodes = make(map[string]struct{})
	}
	if value.controlled != d.hasSnapshot {
		clear(value.ratios)
		clear(value.collapsed)
		value.controlled = d.hasSnapshot
		value.initialized = false
		value.maximized = ""
		value.maxInitialized = false
	}
	if value.keepAlive != d.keepAlive {
		d.releaseRetainedNodes(ctx, value)
		value.keepAlive = d.keepAlive
	}
	if !value.initialized && !d.hasSnapshot && d.hasDefault {
		copySnapshot(value, d.defaultSnapshot)
	}
	value.initialized = true
}

func (d DockLayoutWidget) resolveMaximized(value *dockState) string {
	if d.hasSnapshot {
		return value.maximized
	}
	if d.hasMaximizedKey {
		return d.maximizedKey
	}
	if !value.maxInitialized {
		if d.hasDefaultMax {
			value.maximized = d.defaultMaxKey
		}
		value.maxInitialized = true
	}
	return value.maximized
}

func (d DockLayoutWidget) syncSnapshot(ctx *frame.Context, value *dockState) {
	clear(value.keys)
	collectNodeKeys(d.root, value.keys)
	if d.hasSnapshot {
		normalized := d.snapshot.Migrate(d.root.key, value.keys, nil)
		copySnapshot(value, normalized)
		value.maximized = normalized.MaximizedKey
		value.maxInitialized = true
	}
	for key := range value.ratios {
		if _, ok := value.keys[key]; !ok {
			delete(value.ratios, key)
		}
	}
	for key := range value.collapsed {
		if _, ok := value.keys[key]; !ok {
			delete(value.collapsed, key)
		}
	}
	for key := range value.retainedNodes {
		if _, exists := value.keys[key]; !exists {
			frame.ReleaseStateRetention(ctx, d.nodeStateScope(ctx, value, key))
			delete(value.retainedNodes, key)
			continue
		}
		if d.keepAlive {
			frame.RetainState(ctx, d.nodeStateScope(ctx, value, key))
		}
	}
}

func (d DockLayoutWidget) releaseRetainedNodes(ctx *frame.Context, value *dockState) {
	for key := range value.retainedNodes {
		frame.ReleaseStateRetention(ctx, d.nodeStateScope(ctx, value, key))
	}
	clear(value.retainedNodes)
}

func copySnapshot(value *dockState, snapshot Snapshot) {
	clear(value.ratios)
	for key, ratio := range snapshot.Ratios {
		value.ratios[key] = normalizeRatio(ratio)
	}
	clear(value.collapsed)
	maps.Copy(value.collapsed, snapshot.Collapsed)
}

func collectNodeKeys(node Node, keys map[string]struct{}) {
	if node.key == "" {
		return
	}
	keys[node.key] = struct{}{}
	if node.kind != nodeSplit || node.first == nil || node.second == nil {
		return
	}
	collectNodeKeys(*node.first, keys)
	collectNodeKeys(*node.second, keys)
}

func (d DockLayoutWidget) collapseState(node Node, value *dockState) CollapseState {
	if collapsed, ok := value.collapsed[node.key]; ok {
		return collapsed
	}
	return CollapseState{First: node.collapsedFirst, Second: node.collapsedSecond}
}

func (d DockLayoutWidget) prepareNodeAssociations(ctx *frame.Context, node Node, value *dockState) {
	if node.kind == nodePanel {
		if node.content == nil {
			return
		}
		restore := frame.PushKey(ctx, node.key)
		defer restore()
		layoutui.PrepareFieldAssociations(ctx, node.content)
		return
	}
	if node.first == nil || node.second == nil {
		return
	}
	collapse := d.collapseState(node, value)
	if collapse.First && collapse.Second {
		return
	}
	if collapse.First {
		d.prepareNodeAssociations(ctx, *node.second, value)
		return
	}
	if collapse.Second {
		d.prepareNodeAssociations(ctx, *node.first, value)
		return
	}
	d.prepareNodeAssociations(ctx, *node.first, value)
	d.prepareNodeAssociations(ctx, *node.second, value)
}

func (d DockLayoutWidget) prepareMaximizedAssociations(ctx *frame.Context, node Node, target string, value *dockState) {
	if node.key == target {
		d.prepareNodeAssociations(ctx, node, value)
		return
	}
	if node.kind != nodeSplit || node.first == nil || node.second == nil {
		return
	}
	if containsNode(*node.first, target) {
		d.prepareMaximizedAssociations(ctx, *node.first, target, value)
		return
	}
	if containsNode(*node.second, target) {
		d.prepareMaximizedAssociations(ctx, *node.second, target, value)
	}
}

func findNode(node Node, key string) (Node, bool) {
	if node.key == key {
		return node, true
	}
	if node.kind != nodeSplit || node.first == nil || node.second == nil {
		return Node{}, false
	}
	if found, ok := findNode(*node.first, key); ok {
		return found, true
	}
	return findNode(*node.second, key)
}

func containsNode(node Node, key string) bool {
	_, ok := findNode(node, key)
	return ok
}

func (d DockLayoutWidget) layoutMaximized(ctx *frame.Context, gtx layout.Context, root Node, target string, value *dockState) layout.Dimensions {
	targetNode, ok := findNode(root, target)
	if !ok {
		return d.layoutNode(ctx, gtx, root, value)
	}
	d.layoutMaximizedHidden(ctx, gtx, root, target, value)
	return d.layoutNode(ctx, gtx, targetNode, value)
}

func (d DockLayoutWidget) layoutMaximizedHidden(ctx *frame.Context, gtx layout.Context, node Node, target string, value *dockState) {
	if node.key == target {
		return
	}
	if node.kind != nodeSplit || node.first == nil || node.second == nil {
		d.layoutHiddenBranch(ctx, gtx, value, node)
		return
	}
	if containsNode(*node.first, target) {
		d.layoutHiddenBranch(ctx, gtx, value, *node.second)
		d.layoutMaximizedHidden(ctx, gtx, *node.first, target, value)
		return
	}
	if containsNode(*node.second, target) {
		d.layoutHiddenBranch(ctx, gtx, value, *node.first)
		d.layoutMaximizedHidden(ctx, gtx, *node.second, target, value)
		return
	}
	d.layoutHiddenBranch(ctx, gtx, value, node)
}

func (d DockLayoutWidget) layoutNode(ctx *frame.Context, gtx layout.Context, node Node, value *dockState) layout.Dimensions {
	if d.keepAlive {
		scope := d.nodeStateScope(ctx, value, node.key)
		restoreRetention := frame.PushStateRetention(ctx, scope)
		defer restoreRetention()
		value.retainedNodes[node.key] = struct{}{}
	}
	if node.kind == nodePanel {
		if node.content == nil {
			return layout.Dimensions{Size: gtx.Constraints.Constrain(gtx.Constraints.Min)}
		}
		restore := frame.PushKey(ctx, node.key)
		defer restore()
		return node.content.Layout(ctx, gtx)
	}
	if node.first == nil || node.second == nil {
		panic(fmt.Sprintf("flowui: dock split %q requires two children", node.key))
	}
	collapse := d.collapseState(node, value)
	if collapse.First && collapse.Second {
		d.layoutHiddenBranch(ctx, gtx, value, *node.first)
		d.layoutHiddenBranch(ctx, gtx, value, *node.second)
		return layout.Dimensions{Size: gtx.Constraints.Constrain(gtx.Constraints.Min)}
	}
	if collapse.First {
		d.layoutHiddenBranch(ctx, gtx, value, *node.first)
		return d.layoutChildNode(ctx, gtx, *node.second, value)
	}
	if collapse.Second {
		d.layoutHiddenBranch(ctx, gtx, value, *node.second)
		return d.layoutChildNode(ctx, gtx, *node.first, value)
	}

	ratio := node.ratio
	if !node.hasRatio {
		ratio = .5
	}
	if stored, ok := value.ratios[node.key]; ok {
		ratio = stored
	}
	first := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutChildNode(ctx, gtx, *node.first, value)
	})
	second := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutChildNode(ctx, gtx, *node.second, value)
	})
	split := layoutui.SplitPane(node.key, first, second).
		DefaultRatio(ratio).
		Ratio(ratio).
		MinFirst(int(node.minFirst)).
		MinSecond(int(node.minSecond)).
		Disabled(d.disabled || !gtx.Enabled())
	if node.orientation == Vertical {
		split = split.Vertical()
	} else {
		split = split.Horizontal()
	}
	split = split.OnRatioChange(func(next float32) {
		value.ratios[node.key] = normalizeRatio(next)
		if d.onRatioChange != nil {
			d.onRatioChange(node.key, value.ratios[node.key])
		}
		if d.onChange != nil {
			d.onChange(d.snapshotFor(value))
		}
	})
	return split.Layout(ctx, gtx)
}

func (d DockLayoutWidget) layoutHiddenBranch(ctx *frame.Context, gtx layout.Context, value *dockState, node Node) {
	if !d.keepAlive || d.subtreeRetained(value, node) {
		return
	}
	restoreBoundary := frame.PushStateRetentionBoundary(ctx, value.retentionDepth)
	defer restoreBoundary()
	frame.LayoutHidden(ctx, gtx, "", frame.WidgetFunc(func(ctx *frame.Context, hiddenGtx layout.Context) layout.Dimensions {
		return d.layoutNode(ctx, hiddenGtx, node, value)
	}))
}

func (d DockLayoutWidget) layoutChildNode(ctx *frame.Context, gtx layout.Context, node Node, value *dockState) layout.Dimensions {
	restoreBoundary := frame.PushStateRetentionBoundary(ctx, value.retentionDepth)
	defer restoreBoundary()
	return d.layoutNode(ctx, gtx, node, value)
}

func (d DockLayoutWidget) subtreeRetained(value *dockState, node Node) bool {
	if _, ok := value.retainedNodes[node.key]; !ok {
		return false
	}
	if node.kind != nodeSplit || node.first == nil || node.second == nil {
		return true
	}
	return d.subtreeRetained(value, *node.first) && d.subtreeRetained(value, *node.second)
}

func (d DockLayoutWidget) nodeStateScope(ctx *frame.Context, value *dockState, nodeKey string) string {
	return frame.DerivedKey(ctx, value.rootKey, "node-state:"+nodeKey)
}

func (d DockLayoutWidget) snapshotFor(value *dockState) Snapshot {
	snapshot := Snapshot{
		Version:      SnapshotVersion,
		RootKey:      d.root.key,
		Ratios:       make(map[string]float32, len(value.ratios)),
		Collapsed:    make(map[string]CollapseState, len(value.collapsed)),
		MaximizedKey: d.resolveMaximized(value),
	}
	maps.Copy(snapshot.Ratios, value.ratios)
	maps.Copy(snapshot.Collapsed, value.collapsed)
	return snapshot
}

func validateNode(node Node, keys map[string]struct{}) {
	if node.key == "" {
		panic("flowui: empty dock node key")
	}
	if _, exists := keys[node.key]; exists {
		panic(fmt.Sprintf("flowui: duplicate dock node key %q", node.key))
	}
	keys[node.key] = struct{}{}
	if node.kind == nodePanel {
		return
	}
	if node.first == nil || node.second == nil {
		panic(fmt.Sprintf("flowui: dock split %q requires two children", node.key))
	}
	validateNode(*node.first, keys)
	validateNode(*node.second, keys)
}

func normalizeRatio(value float32) float32 {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return .5
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func maxInt(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
