package frame

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Context holds framework state for a running window.
//
// Application state belongs in the user's Model. Context only keeps Gio widget
// state that must survive from frame to frame.
type Context struct {
	window                       *app.Window
	theme                        *theme.Theme
	language                     locale.Language
	states                       state.Store
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
}

// New creates a per-window frame context. Application state remains in the
// user's MVU Model; Context owns only rendering and interaction services.
func New(w *app.Window, activeTheme *theme.Theme, language locale.Language) *Context {
	language = locale.Resolve(language)
	if activeTheme == nil {
		defaultTheme := theme.DefaultTheme()
		activeTheme = &defaultTheme
	}
	if activeTheme.Material == nil {
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
	if snapshot.Material != nil {
		materialTheme := *snapshot.Material
		snapshot.Material = &materialTheme
	}
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
func ActiveTheme(ctx *Context) *theme.Theme {
	return ctx.theme
}

// PushInstanceTheme customizes a copy of the active theme for the current layout scope.
func PushInstanceTheme(ctx *Context, customize func(*theme.Theme)) func() {
	if ctx == nil || customize == nil {
		return nil
	}
	activeTheme := ctx.Theme()
	customize(&activeTheme)
	theme.SyncMaterialTheme(&activeTheme)
	previous := ctx.theme
	ctx.theme = &activeTheme
	return func() {
		ctx.theme = previous
	}
}

// ActiveLanguage returns the resolved language used by internal components.
func ActiveLanguage(ctx *Context) locale.Language {
	return ctx.language
}

// Invalidate asks Gio to draw another frame.
func (ctx *Context) Invalidate() {
	if ctx.window != nil {
		ctx.window.Invalidate()
	}
}

func PerformWindowActions(ctx *Context, actions system.Action) {
	if ctx != nil && ctx.window != nil && actions != 0 {
		ctx.window.Perform(actions)
	}
}

func BeginFrame(ctx *Context) {
	ctx.overlays.beginFrame()
	ctx.states.BeginFrame()
	ctx.keys.BeginFrame()
	ctx.exclusive.BeginFrame()
	ctx.focus.BeginFrame()
	ctx.fieldLabels, ctx.previousLabels = rotateStringMap(ctx.previousLabels, ctx.fieldLabels)
	ctx.fieldDescriptions, ctx.previousDescriptions = rotateStringMap(ctx.previousDescriptions, ctx.fieldDescriptions)
	ctx.preparedLabels, ctx.previousPreparedLabels = rotateSet(ctx.previousPreparedLabels, ctx.preparedLabels)
	ctx.preparedDescriptions, ctx.previousPreparedDescriptions = rotateSet(ctx.previousPreparedDescriptions, ctx.preparedDescriptions)
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
	registerCollectedFocus(ctx, tag, enabled)
	if ctx.focusGroup == nil || tag == nil || !enabled {
		return
	}
	ctx.focusGroup.Items = append(ctx.focusGroup.Items, FocusGroupItem{Tag: tag})
}

func FocusVisible(ctx *Context, tag event.Tag, focused bool) bool {
	return ctx.focus.Visible(tag, focused)
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
	if ctx.fieldFocus == nil {
		ctx.fieldFocus = make(map[string]fieldFocusTarget)
	}
	ctx.fieldFocus[key] = fieldFocusTarget{tag: tag, enabled: enabled}
	registerCollectedFocus(ctx, tag, enabled)
}

func RequestFieldFocus(ctx *Context, key string) {
	target, ok := ctx.fieldFocus[key]
	if ok && target.enabled {
		RequestFocus(ctx, target.tag)
	}
}

func RegisterFieldLabel(ctx *Context, key, label string) {
	if ctx.fieldLabels == nil {
		ctx.fieldLabels = make(map[string]string)
	}
	ctx.fieldLabels[key] = label
}

func PrepareFieldLabel(ctx *Context, key, label string) {
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
	if ctx.fieldDescriptions == nil {
		ctx.fieldDescriptions = make(map[string]string)
	}
	ctx.fieldDescriptions[key] = description
}

func PrepareFieldDescription(ctx *Context, key, description string) {
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
	ctx.focus.OnPress(tag, history, before)
}

// PreserveFocus prevents the frame's global pointer catcher from clearing the
// current focus after a pointer-only overlay surface consumed a press.
func PreserveFocus(ctx *Context) {
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
	return state.Use[T](&ctx.states, state.Identity{Key: key, Slot: slot}, nil)
}

func UseStateWith[T any](ctx *Context, key, slot string, factory func() *T) *T {
	return state.Use(&ctx.states, state.Identity{Key: key, Slot: slot}, factory)
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

func RegisterExclusive(ctx *Context, group, key string, close func()) {
	ctx.exclusive.Register(group, key, close)
}

func ActivateExclusive(ctx *Context, group, key string) {
	ctx.exclusive.Activate(group, key)
}

func ReleaseExclusive(ctx *Context, group, key string) {
	ctx.exclusive.Release(group, key)
}

func ActiveExclusive(ctx *Context, group string) string {
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
