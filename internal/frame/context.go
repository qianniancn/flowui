package frame

import (
	"image"
	"image/color"
	"maps"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

// Context holds framework state for a running window.
//
// Application state belongs in the user's Model. Context only keeps Gio widget
// state that must survive from frame to frame.
type Context struct {
	window                       *app.Window
	requestWindowClose           func()
	theme                        *theme.Theme
	themeGeneration              uint64
	language                     locale.Language
	states                       state.Store
	retentionScopes              []string
	keys                         state.Keys
	exclusive                    state.Exclusive
	focus                        state.Focus
	focusGroup                   *FocusGroup
	focusCollector               *FocusCollector
	fieldFocus                   map[string]fieldFocusTarget
	fieldLabels                  map[string]string
	previousLabels               map[string]string
	preparedLabels               map[string]struct{}
	previousPreparedLabels       map[string]struct{}
	fieldDescriptions            map[string]string
	previousDescriptions         map[string]string
	preparedDescriptions         map[string]struct{}
	previousPreparedDescriptions map[string]struct{}
	viewport                     image.Point
	foreground                   color.NRGBA
	hasForeground                bool
	background                   color.NRGBA
	hasBackground                bool
	windowState                  WindowState
	overlays                     overlayHost
	styles                       []style.Style
	inheritedStyles              []style.Style
	hiddenLayoutDepth            int
	semantics                    []SemanticNode
	visualOverflowCollector      *VisualOverflowCollector
}

// SemanticRole identifies a framework-level role that Gio's semantic package
// does not model yet. Widgets still emit Gio semantic operations for existing
// clients; this additive registry carries relationships such as Tab -> Panel.
type SemanticRole uint8

const (
	SemanticUnknown SemanticRole = iota
	SemanticTabList
	SemanticTab
	SemanticTabPanel
)

// SemanticNode is the normalized semantic description registered during a
// frame. Key is a stable FlowUI identity. Controls links a Tab to its panel;
// Owner links a Tab to its list or a TabPanel to its owning list.
type SemanticNode struct {
	Key         string
	Role        SemanticRole
	Label       string
	Description string
	Controls    string
	Owner       string
	Selected    bool
	Disabled    bool
	Hidden      bool
	PosInSet    int
	SetSize     int
}

// New creates a per-window frame context. Application state remains in the
// user's MVU Model; Context owns only rendering and interaction services.
func New(w *app.Window, activeTheme *theme.Theme, language locale.Language) *Context {
	language = locale.Resolve(language)
	if activeTheme == nil {
		defaultTheme := theme.DefaultTheme()
		activeTheme = &defaultTheme
	}
	if theme.MaterialOf(activeTheme) == nil {
		theme.SyncMaterialTheme(activeTheme)
	}
	return &Context{
		window:   w,
		theme:    activeTheme,
		language: language,
	}
}

// Theme returns a snapshot of the active visual tokens. Mutating the returned
// value does not change the running application theme.
func (ctx *Context) Theme() theme.Theme {
	if ctx == nil || ctx.theme == nil {
		return theme.DefaultTheme()
	}
	snapshot := *ctx.theme
	theme.DetachMaterial(&snapshot)
	return snapshot
}

// Language returns the language resolved for this window.
func (ctx *Context) Language() locale.Language {
	if ctx == nil {
		return locale.LanguageEnglish
	}
	return ctx.language
}

// ReplaceTheme updates the visual tokens used by subsequent layout work.
func ReplaceTheme(ctx *Context, activeTheme theme.Theme) {
	if ctx == nil {
		return
	}
	theme.SyncMaterialTheme(&activeTheme)
	ctx.theme = &activeTheme
	ctx.themeGeneration++ // Bump generation to invalidate memoized results
}

// ReplaceLanguage updates the language used by localized components.
func ReplaceLanguage(ctx *Context, language locale.Language) {
	if ctx == nil {
		return
	}
	ctx.language = locale.Resolve(language)
}

type WindowMode uint8

const (
	Windowed WindowMode = iota
	Fullscreen
	Minimized
	Maximized
)

func (mode WindowMode) String() string {
	switch mode {
	case Windowed:
		return "windowed"
	case Fullscreen:
		return "fullscreen"
	case Minimized:
		return "minimized"
	case Maximized:
		return "maximized"
	default:
		return "unknown"
	}
}

// WindowState is the latest native state reported for this window. Size is
// the content size in physical pixels.
type WindowState struct {
	Size      image.Point
	Mode      WindowMode
	Focused   bool
	Decorated bool
	TopMost   bool
}

// WindowState returns the latest state for the current window.
func (ctx *Context) WindowState() WindowState {
	if ctx == nil {
		return WindowState{}
	}
	return ctx.windowState
}

func UpdateWindowConfig(ctx *Context, config app.Config) WindowState {
	ctx.windowState = WindowState{
		Size:      config.Size,
		Mode:      WindowMode(config.Mode),
		Focused:   config.Focused,
		Decorated: config.Decorated,
		TopMost:   config.TopMost,
	}
	return ctx.windowState
}

// ActiveTheme returns the mutable theme used internally while laying out a
// frame. It is only reachable by packages inside this module's internal tree.
// ActiveMaterial returns the Gio material bridge for the active theme.
// Internal text and editor helpers use this; applications theme via tokens.
func ActiveMaterial(ctx *Context) *material.Theme {
	return theme.MaterialOf(ActiveTheme(ctx))
}

func ActiveTheme(ctx *Context) *theme.Theme {
	return ctx.theme
}

// ThemeGeneration returns the current theme generation counter. The generation
// is incremented each time ReplaceTheme is called, allowing memoized
// computations to invalidate when the theme changes.
func ThemeGeneration(ctx *Context) uint64 {
	if ctx == nil {
		return 0
	}
	return ctx.themeGeneration
}

// ActiveLanguage returns the resolved language used by internal components.
func ActiveLanguage(ctx *Context) locale.Language {
	return ctx.language
}

// Invalidate asks Gio to draw another frame.
func (ctx *Context) Invalidate() {
	if ctx.hiddenLayoutDepth > 0 {
		return
	}
	if ctx.window != nil {
		ctx.window.Invalidate()
	}
}

// PushHiddenLayout marks a subtree as a state-initialization pass. Hidden
// layouts receive a private operation stream and disabled input, while their
// frame state can still be retained under the surrounding lifecycle scope.
// The marker also prevents focus, field-association, exclusive-action and
// overlay registrations from escaping into the visible frame.
func PushHiddenLayout(ctx *Context) func() {
	if ctx == nil {
		return func() {}
	}
	previous := ctx.hiddenLayoutDepth
	ctx.hiddenLayoutDepth++
	return func() {
		ctx.hiddenLayoutDepth = previous
	}
}

// LayoutHidden lays out a widget to initialize or retain its state without
// painting it or exposing it to input and semantics. A non-empty scope keeps
// identities observed by the hidden widget alive across later frames.
func LayoutHidden(ctx *Context, gtx layout.Context, scope string, child Widget) layout.Dimensions {
	if ctx == nil || child == nil {
		return layout.Dimensions{}
	}
	var hiddenOps op.Ops
	hiddenGtx := gtx
	hiddenGtx.Ops = &hiddenOps
	hiddenGtx.Source = input.Source{}
	hiddenGtx = hiddenGtx.Disabled()
	restoreHidden := PushHiddenLayout(ctx)
	defer restoreHidden()
	if scope != "" {
		restoreRetention := PushStateRetention(ctx, scope)
		defer restoreRetention()
	}
	return child.Layout(ctx, hiddenGtx)
}

// HiddenLayout reports whether the current widget is being laid out only to
// initialize hidden state. It is internal-facing but useful to composite
// components that need to suppress visible-only side effects.
func HiddenLayout(ctx *Context) bool {
	return ctx != nil && ctx.hiddenLayoutDepth > 0
}

// PushMeasurement isolates explicit key claims made by a fallback measurement
// layout. The measured widget may be laid out again in the same frame, so its
// temporary claims must not collide with the real paint pass.
func PushMeasurement(ctx *Context) func() {
	if ctx == nil {
		return func() {}
	}
	previous := ctx.keys.Frame()
	if previous == nil {
		ctx.keys.BeginFrame()
		previous = ctx.keys.Frame()
	}
	snapshot := make(map[string]state.Kind, len(previous))
	maps.Copy(snapshot, previous)
	clear(previous)
	return func() {
		clear(previous)
		maps.Copy(previous, snapshot)
	}
}

func PerformWindowActions(ctx *Context, actions system.Action) {
	if ctx != nil && ctx.hiddenLayoutDepth == 0 && ctx.window != nil && actions != 0 {
		ctx.window.Perform(actions)
	}
}

// SetWindowCloseRequest installs the application lifecycle callback used by
// client-side window decorations.
func SetWindowCloseRequest(ctx *Context, request func()) {
	if ctx != nil {
		ctx.requestWindowClose = request
	}
}

// RequestWindowClose routes a client-side close control through the
// application lifecycle. It reports whether a callback was installed.
func RequestWindowClose(ctx *Context) bool {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 || ctx.requestWindowClose == nil {
		return false
	}
	ctx.requestWindowClose()
	return true
}

func BeginFrame(ctx *Context) {
	ctx.overlays.beginFrame()
	ctx.states.BeginFrame()
	ctx.keys.BeginFrame()
	ctx.exclusive.BeginFrame()
	ctx.focus.BeginFrame()
	ctx.semantics = ctx.semantics[:0]
	ctx.fieldLabels, ctx.previousLabels = rotateStringMap(ctx.previousLabels, ctx.fieldLabels)
	ctx.fieldDescriptions, ctx.previousDescriptions = rotateStringMap(ctx.previousDescriptions, ctx.fieldDescriptions)
	ctx.preparedLabels, ctx.previousPreparedLabels = rotateSet(ctx.previousPreparedLabels, ctx.preparedLabels)
	ctx.preparedDescriptions, ctx.previousPreparedDescriptions = rotateSet(ctx.previousPreparedDescriptions, ctx.preparedDescriptions)
}

// RegisterSemantic records a framework-level semantic role for the current
// frame. Hidden layout passes are deliberately excluded so inactive panels do
// not remain focusable or visible to assistive technology.
func RegisterSemantic(ctx *Context, node SemanticNode) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 || node.Key == "" || node.Role == SemanticUnknown {
		return
	}
	for index := range ctx.semantics {
		if ctx.semantics[index].Key == node.Key {
			ctx.semantics[index] = node
			return
		}
	}
	ctx.semantics = append(ctx.semantics, node)
}

// Semantics returns a snapshot of the framework-level semantic registry for
// the current frame. The returned slice can be modified by the caller.
func Semantics(ctx *Context) []SemanticNode {
	if ctx == nil {
		return nil
	}
	return append([]SemanticNode(nil), ctx.semantics...)
}

func rotateStringMap(current, previous map[string]string) (map[string]string, map[string]string) {
	if current == nil {
		current = make(map[string]string)
	} else {
		clear(current)
	}
	return current, previous
}

func rotateSet(current, previous map[string]struct{}) (map[string]struct{}, map[string]struct{}) {
	if current == nil {
		current = make(map[string]struct{})
	} else {
		clear(current)
	}
	return current, previous
}

func BeginFrameWithViewport(ctx *Context, viewport image.Point) {
	ctx.viewport = viewport
	ctx.windowState.Size = viewport
	BeginFrame(ctx)
}

func OverlayViewport(ctx *Context, fallback image.Point) image.Point {
	viewport := ctx.viewport
	if viewport.X <= 0 {
		viewport.X = fallback.X
	}
	if viewport.Y <= 0 {
		viewport.Y = fallback.Y
	}
	return viewport
}

func ApplyFrameCommands(ctx *Context, gtx layout.Context) {
	ctx.focus.CommitObservations()
	ctx.focus.ApplyFrameCommands(gtx)
}

func RequestFocus(ctx *Context, tag event.Tag) {
	RequestFocusVisible(ctx, tag, true)
}

// RequestFocus queues keyboard-visible focus for a custom widget.
func (ctx *Context) RequestFocus(tag event.Tag) {
	RequestFocus(ctx, tag)
}

func RequestFocusVisible(ctx *Context, tag event.Tag, visible bool) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	origin := state.FocusOriginKeyboard
	if !visible {
		origin = state.FocusOriginPointer
	}
	ctx.focus.Request(tag, origin)
}

// RequestFocusVisible queues focus and controls whether its focus ring is visible.
func (ctx *Context) RequestFocusVisible(tag event.Tag, visible bool) {
	RequestFocusVisible(ctx, tag, visible)
}

type FocusGroup struct {
	Items []FocusGroupItem
}

type FocusGroupItem struct {
	Tag event.Tag
}

// FocusCollector gathers focusable descendants laid out within a component.
type FocusCollector struct {
	Targets []event.Tag
}

func PushFocusCollector(ctx *Context, collector *FocusCollector) func() {
	previous := ctx.focusCollector
	ctx.focusCollector = collector
	return func() {
		ctx.focusCollector = previous
	}
}

func registerCollectedFocus(ctx *Context, tag event.Tag, enabled bool) {
	if ctx.hiddenLayoutDepth > 0 {
		return
	}
	if ctx.focusCollector != nil && tag != nil && enabled {
		ctx.focusCollector.Targets = append(ctx.focusCollector.Targets, tag)
	}
}

func PushFocusGroup(ctx *Context, group *FocusGroup) func() {
	previous := ctx.focusGroup
	ctx.focusGroup = group
	return func() {
		ctx.focusGroup = previous
	}
}

func RegisterFocusGroupItem(ctx *Context, tag event.Tag, enabled bool) {
	if ctx.hiddenLayoutDepth > 0 {
		return
	}
	registerCollectedFocus(ctx, tag, enabled)
	if ctx.focusGroup == nil || tag == nil || !enabled {
		return
	}
	ctx.focusGroup.Items = append(ctx.focusGroup.Items, FocusGroupItem{Tag: tag})
}

func FocusVisible(ctx *Context, tag event.Tag, focused bool) bool {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return false
	}
	return ctx.focus.Observe(tag, focused)
}

// FocusVisible reports whether a focused custom widget should draw its focus ring.
func (ctx *Context) FocusVisible(tag event.Tag, focused bool) bool {
	return FocusVisible(ctx, tag, focused)
}

type fieldFocusTarget struct {
	tag     event.Tag
	enabled bool
}

func RegisterFieldFocus(ctx *Context, key string, tag event.Tag, enabled bool) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	if ctx.fieldFocus == nil {
		ctx.fieldFocus = make(map[string]fieldFocusTarget)
	}
	ctx.fieldFocus[key] = fieldFocusTarget{tag: tag, enabled: enabled}
	registerCollectedFocus(ctx, tag, enabled)
}

func RequestFieldFocus(ctx *Context, key string) {
	RequestFieldFocusVisible(ctx, key, true)
}

func RequestFieldFocusVisible(ctx *Context, key string, visible bool) {
	target, ok := ctx.fieldFocus[key]
	if ok && target.enabled {
		RequestFocusVisible(ctx, target.tag, visible)
	}
}

func RegisterFieldLabel(ctx *Context, key, label string) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	if ctx.fieldLabels == nil {
		ctx.fieldLabels = make(map[string]string)
	}
	ctx.fieldLabels[key] = label
}

func PrepareFieldLabel(ctx *Context, key, label string) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	RegisterFieldLabel(ctx, key, label)
	if ctx.preparedLabels == nil {
		ctx.preparedLabels = make(map[string]struct{})
	}
	ctx.preparedLabels[key] = struct{}{}
}

func FieldLabel(ctx *Context, key string) string {
	if label, ok := ctx.fieldLabels[key]; ok {
		return label
	}
	if wasPrepared(ctx.previousPreparedLabels, ctx.preparedLabels, key) {
		return ""
	}
	return ctx.previousLabels[key]
}

func RegisterFieldDescription(ctx *Context, key, description string) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	if ctx.fieldDescriptions == nil {
		ctx.fieldDescriptions = make(map[string]string)
	}
	ctx.fieldDescriptions[key] = description
}

func PrepareFieldDescription(ctx *Context, key, description string) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	RegisterFieldDescription(ctx, key, description)
	if ctx.preparedDescriptions == nil {
		ctx.preparedDescriptions = make(map[string]struct{})
	}
	ctx.preparedDescriptions[key] = struct{}{}
}

func FieldDescription(ctx *Context, key string) string {
	if description, ok := ctx.fieldDescriptions[key]; ok {
		return description
	}
	if wasPrepared(ctx.previousPreparedDescriptions, ctx.preparedDescriptions, key) {
		return ""
	}
	return ctx.previousDescriptions[key]
}

func wasPrepared(previous, current map[string]struct{}, key string) bool {
	_, existed := previous[key]
	_, stillExists := current[key]
	return existed && !stillExists
}

func WithFieldSemantics(ctx *Context, key string, child layout.Widget) layout.Widget {
	label := FieldLabel(ctx, key)
	description := FieldDescription(ctx, key)
	if label == "" && description == "" {
		return child
	}
	return func(gtx layout.Context) layout.Dimensions {
		if label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		return child(gtx)
	}
}

func FocusOnPress(ctx *Context, tag event.Tag, history []widget.Press, before int) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	ctx.focus.OnPress(tag, history, before)
}

// PreserveFocus prevents the frame's global pointer catcher from clearing the
// current focus after a pointer-only overlay surface consumed a press.
func PreserveFocus(ctx *Context) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	ctx.focus.Preserve()
}

// PreserveFocus prevents a pointer-only custom overlay from clearing focus this frame.
func (ctx *Context) PreserveFocus() {
	PreserveFocus(ctx)
}

func PushColors(ctx *Context, foreground, background color.NRGBA) func() {
	previousForeground, hadForeground := ctx.foreground, ctx.hasForeground
	previousBackground, hadBackground := ctx.background, ctx.hasBackground
	ctx.foreground, ctx.hasForeground = foreground, true
	ctx.background, ctx.hasBackground = background, true
	return func() {
		ctx.foreground, ctx.hasForeground = previousForeground, hadForeground
		ctx.background, ctx.hasBackground = previousBackground, hadBackground
	}
}

// PushStyle adds a cascading style for descendant component layout.
func PushStyle(ctx *Context, value style.Style) func() {
	if ctx == nil {
		return func() {}
	}
	previous := len(ctx.styles)
	ctx.styles = append(ctx.styles, value)
	return func() {
		ctx.styles = ctx.styles[:previous]
	}
}

// ActiveStyles returns the styles inherited by the current layout scope.
func ActiveStyles(ctx *Context) []style.Style {
	if ctx == nil {
		return nil
	}
	return append([]style.Style(nil), ctx.styles...)
}

// ActiveStylesReadOnly returns the styles inherited by the current layout scope
// without copying. The caller MUST NOT modify the returned slice or its elements.
// Use this only for append-only consumers that build a new slice.
func ActiveStylesReadOnly(ctx *Context) []style.Style {
	if ctx == nil {
		return nil
	}
	return ctx.styles
}

// PushInheritedStyle propagates a computed parent style with lower precedence
// than a child's variant, size, scope, and instance declarations.
func PushInheritedStyle(ctx *Context, value style.Style) func() {
	if ctx == nil {
		return func() {}
	}
	previous := len(ctx.inheritedStyles)
	ctx.inheritedStyles = append(ctx.inheritedStyles, value)
	return func() {
		ctx.inheritedStyles = ctx.inheritedStyles[:previous]
	}
}

func ActiveInheritedStyles(ctx *Context) []style.Style {
	if ctx == nil {
		return nil
	}
	return append([]style.Style(nil), ctx.inheritedStyles...)
}

// ActiveInheritedStylesReadOnly returns the inherited styles without copying.
// The caller MUST NOT modify the returned slice or its elements.
// Use this only for append-only consumers that build a new slice.
func ActiveInheritedStylesReadOnly(ctx *Context) []style.Style {
	if ctx == nil {
		return nil
	}
	return ctx.inheritedStyles
}

func (ctx *Context) ForegroundColor() color.NRGBA {
	if ctx.hasForeground {
		return ctx.foreground
	}
	return ctx.theme.Palette.Foreground
}

func (ctx *Context) BackgroundColor() color.NRGBA {
	if ctx.hasBackground {
		return ctx.background
	}
	return ctx.theme.Palette.Background
}

func EndFrame(ctx *Context) {
	ctx.overlays.runAfterLayout()
	frameKeys := ctx.keys.Frame()
	for key := range ctx.fieldFocus {
		switch frameKeys[key] {
		case state.KindInput, state.KindComboBox, state.KindSelect, state.KindSlider, state.KindDateField, state.KindDatePicker, state.KindDateRangePicker:
		default:
			delete(ctx.fieldFocus, key)
		}
	}
	if fieldAssociationsChanged(ctx) {
		ctx.Invalidate()
	}
	ctx.focus.CommitObservations()
	ctx.states.EndFrame()
	ctx.exclusive.EndFrame()
}

func fieldAssociationsChanged(ctx *Context) bool {
	return !sameStringMap(ctx.fieldLabels, ctx.previousLabels) ||
		!sameStringMap(ctx.fieldDescriptions, ctx.previousDescriptions) ||
		!sameSet(ctx.preparedLabels, ctx.previousPreparedLabels) ||
		!sameSet(ctx.preparedDescriptions, ctx.previousPreparedDescriptions)
}

func sameStringMap(first, second map[string]string) bool {
	if len(first) != len(second) {
		return false
	}
	for key, value := range first {
		other, ok := second[key]
		if !ok || other != value {
			return false
		}
	}
	return true
}

func sameSet(first, second map[string]struct{}) bool {
	if len(first) != len(second) {
		return false
	}
	for key := range first {
		if _, ok := second[key]; !ok {
			return false
		}
	}
	return true
}

const (
	stateSlotClickable = "clickable"
	stateSlotDraggable = "draggable"
	stateSlotEditor    = "editor"
	stateSlotBool      = "bool"
	stateSlotList      = "list"
	stateSlotScroll    = "scroll"
)

func UseState[T any](ctx *Context, key, slot string) *T {
	id := state.Identity{Key: key, Slot: slot}
	value := state.Use[T](&ctx.states, id, nil)
	ctx.recordStateRetention(id)
	return value
}

func UseStateWith[T any](ctx *Context, key, slot string, factory func() *T) *T {
	id := state.Identity{Key: key, Slot: slot}
	value := state.Use(&ctx.states, id, factory)
	ctx.recordStateRetention(id)
	return value
}

// PushStateRetention makes states used by descendant widgets part of a named
// lifecycle scope. The scope remembers identities observed while mounted and
// can keep them alive while the subtree is hidden.
func PushStateRetention(ctx *Context, scope string) func() {
	if ctx == nil {
		return func() {}
	}
	ctx.states.Retain(scope)
	previous := len(ctx.retentionScopes)
	ctx.retentionScopes = append(ctx.retentionScopes, scope)
	return func() {
		ctx.retentionScopes = ctx.retentionScopes[:previous]
	}
}

// StateRetentionDepth reports the number of active lifecycle scopes. Composite
// widgets can capture the depth before opening their own scopes and preserve
// only outer owners while delegating lifecycle ownership to child widgets.
func StateRetentionDepth(ctx *Context) int {
	if ctx == nil {
		return 0
	}
	return len(ctx.retentionScopes)
}

// PushStateRetentionBoundary temporarily excludes scopes created at or after
// depth. It is useful when a composite child owns a separate lifecycle: outer
// scopes remain intact while the child's state is not retained by stale parent
// scopes. The caller must pass a depth captured from StateRetentionDepth.
func PushStateRetentionBoundary(ctx *Context, depth int) func() {
	if ctx == nil {
		return func() {}
	}
	if depth < 0 {
		depth = 0
	}
	if depth > len(ctx.retentionScopes) {
		depth = len(ctx.retentionScopes)
	}
	previous := ctx.retentionScopes
	ctx.retentionScopes = previous[:depth:depth]
	return func() {
		ctx.retentionScopes = previous
	}
}

// RetainState keeps the identities previously observed in scope alive for the
// current frame without laying out the hidden subtree.
func RetainState(ctx *Context, scope string) {
	if ctx == nil {
		return
	}
	ctx.states.Retain(scope)
}

// ReleaseStateRetention forgets the identities associated with scope.
func ReleaseStateRetention(ctx *Context, scope string) {
	if ctx == nil {
		return
	}
	ctx.states.ReleaseRetention(scope)
}

func (ctx *Context) recordStateRetention(id state.Identity) {
	for _, scope := range ctx.retentionScopes {
		ctx.states.RecordRetention(scope, id)
	}
}

// StateStore returns the underlying state store for advanced use cases such as
// memoization. Most code should use UseState or UseStateWith instead.
func StateStore(ctx *Context) *state.Store {
	if ctx == nil {
		return nil
	}
	return &ctx.states
}

func PeekState[T any](ctx *Context, key, slot string) (*T, bool) {
	return state.Peek[T](&ctx.states, state.Identity{Key: key, Slot: slot})
}

func DeleteState(ctx *Context, key, slot string) {
	ctx.states.Delete(state.Identity{Key: key, Slot: slot})
}

func StateLen(ctx *Context) int {
	return ctx.states.Len()
}

// RegisterExclusive registers a widget in an exclusive group. When another widget
// in the same group activates, this widget's close callback will be invoked.
//
// Exclusive groups are used for navigation widgets (dropdown, select, menubar,
// context-menu) and transient hints (tooltip) where only one should be open at a time.
// General-purpose overlays (popover, modal, alertdialog) do not participate in
// exclusive groups, allowing applications to control their lifecycle independently
// and support nested or stacked scenarios.
func RegisterExclusive(ctx *Context, group, key string, close func()) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	ctx.exclusive.Register(group, key, close)
}

func ActivateExclusive(ctx *Context, group, key string) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	ctx.exclusive.Activate(group, key)
}

func ReleaseExclusive(ctx *Context, group, key string) {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return
	}
	ctx.exclusive.Release(group, key)
}

func ActiveExclusive(ctx *Context, group string) string {
	if ctx == nil || ctx.hiddenLayoutDepth > 0 {
		return ""
	}
	return ctx.exclusive.Active(group)
}

func HasFieldFocus(ctx *Context, key string) bool {
	return ctx.fieldFocus[key].tag != nil
}

func FieldFocusTag(ctx *Context, key string) event.Tag {
	return ctx.fieldFocus[key].tag
}

func AnyFieldFocused(ctx *Context, gtx layout.Context) bool {
	for _, target := range ctx.fieldFocus {
		if target.enabled && gtx.Focused(target.tag) {
			return true
		}
	}
	return false
}

func HasFieldLabel(ctx *Context, key string) bool {
	_, ok := ctx.fieldLabels[key]
	return ok
}

func HasFieldDescription(ctx *Context, key string) bool {
	_, ok := ctx.fieldDescriptions[key]
	return ok
}

func (ctx *Context) Clickable(key string) *widget.Clickable {
	_, clickable := ClickableWithKey(ctx, key)
	return clickable
}

// Draggable returns Gio drag state retained under key for a custom widget.
func (ctx *Context) Draggable(key string) *widget.Draggable {
	key = ClaimKey(ctx, state.KindDraggable, key)
	return UseState[widget.Draggable](ctx, key, stateSlotDraggable)
}

func ClickableWithKey(ctx *Context, key string) (string, *widget.Clickable) {
	key = ClaimKey(ctx, state.KindClickable, key)
	return key, UseState[widget.Clickable](ctx, key, stateSlotClickable)
}

func DerivedClickableWithKey(ctx *Context, owner, role string) (string, *widget.Clickable) {
	key := ClaimDerivedKey(ctx, state.KindClickable, owner, role)
	return key, UseState[widget.Clickable](ctx, key, stateSlotClickable)
}

func (ctx *Context) Editor(key string) *widget.Editor {
	_, e := EditorWithKey(ctx, key)
	return e
}

func EditorWithKey(ctx *Context, key string) (string, *widget.Editor) {
	return editorWithKind(ctx, state.KindEditor, key)
}

func InputEditor(ctx *Context, key string) (string, *widget.Editor) {
	return editorWithKind(ctx, state.KindInput, key)
}

func editorWithKind(ctx *Context, kind state.Kind, key string) (string, *widget.Editor) {
	key = ClaimKey(ctx, kind, key)
	return key, UseState[widget.Editor](ctx, key, stateSlotEditor)
}

func (ctx *Context) BoolState(key string) *widget.Bool {
	_, state := BoolStateWithKey(ctx, key)
	return state
}

func BoolStateWithKey(ctx *Context, key string) (string, *widget.Bool) {
	key = ClaimKey(ctx, state.KindCheckbox, key)
	return key, UseState[widget.Bool](ctx, key, stateSlotBool)
}

func (ctx *Context) ListState(key string) *layout.List {
	key = ClaimKey(ctx, state.KindList, key)
	return UseStateWith(ctx, key, stateSlotList, func() *layout.List {
		return &layout.List{Axis: layout.Vertical}
	})
}

func (ctx *Context) ScrollState(key string) *layout.List {
	key = ClaimKey(ctx, state.KindScroll, key)
	return UseStateWith(ctx, key, stateSlotScroll, func() *layout.List {
		return &layout.List{Axis: layout.Vertical}
	})
}

func PushKey(ctx *Context, key string) func() {
	return ctx.keys.Push(key)
}

func ClaimKey(ctx *Context, kind state.Kind, key string) string {
	return ctx.keys.Claim(kind, key)
}

func ClaimDerivedKey(ctx *Context, kind state.Kind, owner, role string) string {
	return ctx.keys.ClaimDerived(kind, owner, role)
}

func ClaimDerivedResolvedKey(ctx *Context, kind state.Kind, owner, role string) string {
	return ctx.keys.ClaimDerivedResolved(kind, owner, role)
}

func DerivedKey(ctx *Context, owner, role string) string {
	return ctx.keys.Derived(owner, role)
}

func FullKey(ctx *Context, key string) string {
	return ctx.keys.FullKey(key)
}
