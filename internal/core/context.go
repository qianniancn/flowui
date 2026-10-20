package core

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
	giotext "gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Context holds framework state for a running window.
//
// Application state belongs in the user's Model. Context only keeps Gio widget
// state that must survive from frame to frame.
type Context struct {
	window                       *app.Window
	requestWindowClose           func()
	theme                        *theme.Theme
	strictVisuals                bool
	themeGeneration              uint64
	themeHashTheme               *theme.Theme
	themeHash                    uint64
	hasThemeHash                 bool
	language                     locale.Language
	translations                 i18n.Catalog
	states                       Store
	textMeasureCache             textMeasureCache
	retentionScopes              []string
	keys                         Keys
	exclusive                    Exclusive
	focus                        Focus
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
	instanceStyles               []style.Style
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
	return NewWithCatalog(w, activeTheme, language, i18n.Catalog{})
}

// NewWithCatalog creates a per-window frame context with application message
// overrides. The catalog is copied so its value is isolated from callers.
//
// This constructor keeps the low-level, theme-less context path used by unit
// tests and internal helpers. It does not install a visual provider: migrated
// widgets stay visually neutral unless the caller explicitly supplies one.
// Application windows must use NewWithCatalogStrict.
func NewWithCatalog(w *app.Window, activeTheme *theme.Theme, language locale.Language, translations i18n.Catalog) *Context {
	language = locale.Resolve(language)
	return newContext(w, activeTheme, language, translations, false)
}

// NewWithCatalogStrict creates an application context with an explicit,
// complete component visual provider. It is the only constructor used by the
// running UI loop. Keeping validation here makes the no-implicit-theme rule
// hold at the context boundary rather than relying on every widget to remember
// to validate its own provider.
func NewWithCatalogStrict(w *app.Window, activeTheme *theme.Theme, language locale.Language, translations i18n.Catalog) (*Context, error) {
	if activeTheme == nil {
		return nil, theme.ErrMissingVisualProvider
	}
	if err := theme.ValidateVisualProvider(*activeTheme); err != nil {
		return nil, err
	}
	language = locale.Resolve(language)
	return newContext(w, activeTheme, language, translations, true), nil
}

func newContext(w *app.Window, activeTheme *theme.Theme, language locale.Language, translations i18n.Catalog, strictVisuals bool) *Context {
	if theme.MaterialOf(activeTheme) == nil {
		theme.SyncMaterialTheme(activeTheme)
	}
	return &Context{
		window:        w,
		theme:         activeTheme,
		strictVisuals: strictVisuals,
		language:      language,
		translations:  translations.Clone(),
	}
}

// ActiveDesign is gone: a design was only ever "a theme plus the component
// callback overrides". Components now own their own visuals, so
// the theme is the complete visual input and ActiveTheme() is the whole answer.

// Theme returns a snapshot of the active visual tokens. Mutating the returned
// value does not change the running application theme.
func (ctx *Context) Theme() theme.Theme {
	if ctx == nil || ctx.theme == nil {
		return theme.Theme{}
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

// Text resolves a FlowUI-owned message for this window's language.
func (ctx *Context) Text(key i18n.Key) string {
	if ctx == nil {
		return (i18n.Catalog{}).Text(locale.LanguageEnglish, key)
	}
	return ctx.translations.Text(ctx.language, key)
}

// Textf resolves and formats a FlowUI-owned dynamic message for this window.
func (ctx *Context) Textf(key i18n.Key, args ...any) string {
	if ctx == nil {
		return (i18n.Catalog{}).Textf(locale.LanguageEnglish, key, args...)
	}
	return ctx.translations.Textf(ctx.language, key, args...)
}

// ReplaceTheme updates the visual tokens used by subsequent layout work.
func ReplaceTheme(ctx *Context, activeTheme theme.Theme) {
	if ctx == nil {
		return
	}
	if ctx.strictVisuals {
		if err := theme.ValidateVisualProvider(activeTheme); err != nil {
			panic(err)
		}
	}
	theme.SyncMaterialTheme(&activeTheme)
	ctx.theme = &activeTheme
	ctx.themeGeneration++ // Bump generation to invalidate memoized results
	ctx.textMeasureCache.reset()
}

// ReplaceLanguage updates the language used by localized components.
//
// Localized components derive both their strings and their geometry from the
// language, so this follows the same invalidation contract as ReplaceTheme and
// ReplaceDesign. Without the appearance-generation bump, component caches keyed
// on ThemeGeneration - menu item widths, tree and list row metrics, chart axis
// labels - would keep the previous language's dimensions whenever the
// application changes only the language and not its data version.
func ReplaceLanguage(ctx *Context, language locale.Language) {
	if ctx == nil {
		return
	}
	resolved := locale.Resolve(language)
	if resolved == ctx.language {
		return
	}
	ctx.language = resolved
	ctx.themeGeneration++
	ctx.textMeasureCache.reset()
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
	if ctx == nil {
		return nil
	}
	return ctx.theme
}

// RequireVisualProvider revalidates the active provider at a migrated widget
// boundary. Strict application contexts must never silently render an empty
// declaration after a caller mutates or replaces the theme provider; the
// low-level non-strict context remains available for neutral behavior tests.
func RequireVisualProvider(ctx *Context) {
	if ctx == nil || !ctx.strictVisuals {
		return
	}
	active := ActiveTheme(ctx)
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	if err := theme.ValidateVisualProvider(*active); err != nil {
		panic(err)
	}
}

// ThemeGeneration returns the current appearance generation counter. The
// generation is incremented by ReplaceTheme, ReplaceDesign, and ReplaceLanguage
// so that memoized computations invalidate when the visual configuration or the
// localized text that feeds component geometry changes.
func ThemeGeneration(ctx *Context) uint64 {
	if ctx == nil {
		return 0
	}
	return ctx.themeGeneration
}

// CachedThemeHash memoizes a theme-derived value for the active theme pointer.
// Computed style resolution calls this once per context/theme pair instead of
// re-hashing the same immutable theme for every visible child.
func CachedThemeHash(ctx *Context, active *theme.Theme, compute func() uint64) uint64 {
	if ctx == nil || compute == nil {
		if compute == nil {
			return 0
		}
		return compute()
	}
	if ctx.hasThemeHash && ctx.themeHashTheme == active {
		return ctx.themeHash
	}
	value := compute()
	ctx.themeHashTheme = active
	ctx.themeHash = value
	ctx.hasThemeHash = true
	return value
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
	snapshot := make(map[string]Kind, len(previous))
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
	beginAutoKeyFrame()
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
	origin := FocusOriginKeyboard
	if !visible {
		origin = FocusOriginPointer
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

// PushInstanceStyle makes a wrapper-owned declaration available to the
// component currently being laid out. Unlike PushStyle, it is not inherited
// by descendant widgets; component resolvers consume it explicitly.
func PushInstanceStyle(ctx *Context, value style.Style) func() {
	if ctx == nil {
		return func() {}
	}
	previous := len(ctx.instanceStyles)
	ctx.instanceStyles = append(ctx.instanceStyles, value)
	return func() {
		ctx.instanceStyles = ctx.instanceStyles[:previous]
	}
}

// InstanceStyle returns the innermost wrapper-owned declaration.
func InstanceStyle(ctx *Context) style.Style {
	if ctx == nil || len(ctx.instanceStyles) == 0 {
		return style.Style{}
	}
	return ctx.instanceStyles[len(ctx.instanceStyles)-1]
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
	if ctx == nil {
		return color.NRGBA{}
	}
	if ctx.hasForeground {
		return ctx.foreground
	}
	if ctx.theme == nil {
		return color.NRGBA{}
	}
	return ctx.theme.Semantic().Foreground
}

func (ctx *Context) BackgroundColor() color.NRGBA {
	if ctx == nil {
		return color.NRGBA{}
	}
	if ctx.hasBackground {
		return ctx.background
	}
	if ctx.theme == nil {
		return color.NRGBA{}
	}
	return ctx.theme.Semantic().Background
}

// TextShaper returns the shaper that lays out text for the active theme.
//
// Widgets that arrange mixed-style text themselves (inline markup, link spans,
// code spans) cannot use text.New, which resolves a single style for a whole
// paragraph. They need this shaper to hand spans to Gio's styled text layout.
// The value belongs to the theme's material bridge: do not replace it, and do
// not retain it past the frame that returned it.
func (ctx *Context) TextShaper() *giotext.Shaper {
	if ctx == nil {
		return nil
	}
	material := ActiveMaterial(ctx)
	if material == nil {
		return nil
	}
	return material.Shaper
}

func EndFrame(ctx *Context) {
	ctx.overlays.runAfterLayout()
	frameKeys := ctx.keys.Frame()
	for key := range ctx.fieldFocus {
		switch frameKeys[key] {
		case KindInput, KindComboBox, KindSelect, KindSlider, KindDateField, KindDatePicker, KindDateRangePicker:
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
	id := Identity{Key: key, Slot: slot}
	value := Use[T](&ctx.states, id, nil)
	ctx.recordStateRetention(id)
	return value
}

func UseStateWith[T any](ctx *Context, key, slot string, factory func() *T) *T {
	id := Identity{Key: key, Slot: slot}
	value := Use(&ctx.states, id, factory)
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

func (ctx *Context) recordStateRetention(id Identity) {
	for _, scope := range ctx.retentionScopes {
		ctx.states.RecordRetention(scope, id)
	}
}

// StateStore returns the underlying state store for advanced use cases such as
// memoization. Most code should use UseState or UseStateWith instead.
func StateStore(ctx *Context) *Store {
	if ctx == nil {
		return nil
	}
	return &ctx.states
}

func PeekState[T any](ctx *Context, key, slot string) (*T, bool) {
	return Peek[T](&ctx.states, Identity{Key: key, Slot: slot})
}

func DeleteState(ctx *Context, key, slot string) {
	ctx.states.Delete(Identity{Key: key, Slot: slot})
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

func (ctx *Context) Clickable(key string) *widget.Clickable {
	_, clickable := ClickableWithKey(ctx, key)
	return clickable
}

// Draggable returns Gio drag state retained under key for a custom widget.
func (ctx *Context) Draggable(key string) *widget.Draggable {
	key = ClaimKey(ctx, KindDraggable, key)
	return UseState[widget.Draggable](ctx, key, stateSlotDraggable)
}

func ClickableWithKey(ctx *Context, key string) (string, *widget.Clickable) {
	key = ClaimKey(ctx, KindClickable, key)
	return key, UseState[widget.Clickable](ctx, key, stateSlotClickable)
}

func DerivedClickableWithKey(ctx *Context, owner, role string) (string, *widget.Clickable) {
	key := ClaimDerivedKey(ctx, KindClickable, owner, role)
	return key, UseState[widget.Clickable](ctx, key, stateSlotClickable)
}

func (ctx *Context) Editor(key string) *widget.Editor {
	_, e := EditorWithKey(ctx, key)
	return e
}

func EditorWithKey(ctx *Context, key string) (string, *widget.Editor) {
	return editorWithKind(ctx, KindEditor, key)
}

func InputEditor(ctx *Context, key string) (string, *widget.Editor) {
	return editorWithKind(ctx, KindInput, key)
}

func editorWithKind(ctx *Context, kind Kind, key string) (string, *widget.Editor) {
	key = ClaimKey(ctx, kind, key)
	return key, UseState[widget.Editor](ctx, key, stateSlotEditor)
}

func (ctx *Context) BoolState(key string) *widget.Bool {
	_, state := BoolStateWithKey(ctx, key)
	return state
}

func BoolStateWithKey(ctx *Context, key string) (string, *widget.Bool) {
	key = ClaimKey(ctx, KindCheckbox, key)
	return key, UseState[widget.Bool](ctx, key, stateSlotBool)
}

// ClaimListState resolves key and returns the list state together with the
// identity it was claimed under, so derived state can hang off the same key.
func (ctx *Context) ClaimListState(key string) (string, *layout.List) {
	key = ClaimKey(ctx, KindList, key)
	return key, UseStateWith(ctx, key, stateSlotList, func() *layout.List {
		return &layout.List{Axis: layout.Vertical}
	})
}

func (ctx *Context) ListState(key string) *layout.List {
	_, state := ctx.ClaimListState(key)
	return state
}

// ClaimScrollState resolves key and returns the scroll state together with the
// identity it was claimed under.
func (ctx *Context) ClaimScrollState(key string) (string, *layout.List) {
	key = ClaimKey(ctx, KindScroll, key)
	return key, UseStateWith(ctx, key, stateSlotScroll, func() *layout.List {
		return &layout.List{Axis: layout.Vertical}
	})
}

func (ctx *Context) ScrollState(key string) *layout.List {
	_, state := ctx.ClaimScrollState(key)
	return state
}

func PushKey(ctx *Context, key string) func() {
	return ctx.keys.Push(key)
}

func ClaimKey(ctx *Context, kind Kind, key string) string {
	return ctx.keys.Claim(kind, key)
}

func ClaimDerivedKey(ctx *Context, kind Kind, owner, role string) string {
	return ctx.keys.ClaimDerived(kind, owner, role)
}

func ClaimDerivedResolvedKey(ctx *Context, kind Kind, owner, role string) string {
	return ctx.keys.ClaimDerivedResolved(kind, owner, role)
}

func DerivedKey(ctx *Context, owner, role string) string {
	return ctx.keys.Derived(owner, role)
}

func FullKey(ctx *Context, key string) string {
	return ctx.keys.FullKey(key)
}
