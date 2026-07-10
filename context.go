package flowui

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/widget"
	flowstate "github.com/qianniancn/FlowUI/state"
)

// Context holds framework state for a running window.
//
// Application state belongs in the user's Model. Context only keeps Gio widget
// state that must survive from frame to frame.
type Context struct {
	// Theme contains the active visual tokens used by widgets.
	Theme *Theme
	// Language contains the resolved language used by localized widgets.
	Language Language
	// DatePickerLocale contains the date formatting strings used by DatePicker.
	DatePickerLocale DatePickerLocale

	window          *app.Window
	clickables      map[string]*widget.Clickable
	buttons         map[string]*buttonState
	editors         map[string]*widget.Editor
	inputs          map[string]*inputState
	combos          map[string]*comboBoxState
	selects         map[string]*selectState
	datePickers     map[string]*datePickerState
	bools           map[string]*widget.Bool
	checkboxes      map[string]*checkboxState
	switches        map[string]*switchState
	radioGroups     map[string]*radioGroupState
	progressBars    map[string]*progressBarState
	listBoxes       map[string]*listBoxState
	popovers        map[string]*popoverState
	modals          map[string]*modalState
	lists           map[string]*layout.List
	scrolls         map[string]*layout.List
	keys            flowstate.Keys
	focus           flowstate.Focus
	activeSelectKey string
	viewport        image.Point
	foreground      color.NRGBA
	hasForeground   bool
}

func newContext(w *app.Window) *Context {
	return newContextWithTheme(w, nil)
}

func newContextWithTheme(w *app.Window, theme *Theme) *Context {
	return newContextWithThemeAndLanguage(w, theme, LanguageAuto)
}

func newContextWithThemeAndLanguage(w *app.Window, theme *Theme, language Language) *Context {
	language = resolvedLanguage(language)
	return newContextWithThemeLocaleAndLanguage(w, theme, datePickerLocaleForLanguage(language), language)
}

func newContextWithThemeAndLocale(w *app.Window, theme *Theme, datePickerLocale DatePickerLocale) *Context {
	return newContextWithThemeLocaleAndLanguage(w, theme, datePickerLocale, languageForDatePickerLocale(datePickerLocale))
}

func newContextWithThemeLocaleAndLanguage(w *app.Window, theme *Theme, datePickerLocale DatePickerLocale, language Language) *Context {
	if theme == nil {
		defaultTheme := DefaultTheme()
		theme = &defaultTheme
	}
	if theme.Material == nil {
		syncMaterialTheme(theme)
	}
	return &Context{
		Theme:            theme,
		Language:         language,
		DatePickerLocale: normalizeDatePickerLocale(datePickerLocale),
		window:           w,
	}
}

// Invalidate asks Gio to draw another frame.
func (ctx *Context) Invalidate() {
	if ctx.window != nil {
		ctx.window.Invalidate()
	}
}

func (ctx *Context) beginFrame() {
	for _, state := range ctx.selects {
		state.peerClosePending = false
	}
	ctx.keys.BeginFrame()
	ctx.focus.BeginFrame()
}

func (ctx *Context) beginFrameWithViewport(viewport image.Point) {
	ctx.viewport = viewport
	ctx.beginFrame()
}

func (ctx *Context) overlayViewport(fallback image.Point) image.Point {
	viewport := ctx.viewport
	if viewport.X <= 0 {
		viewport.X = fallback.X
	}
	if viewport.Y <= 0 {
		viewport.Y = fallback.Y
	}
	return viewport
}

func (ctx *Context) applyFrameCommands(gtx layout.Context) {
	ctx.focus.ApplyFrameCommands(gtx)
}

func (ctx *Context) requestFocus(tag event.Tag) {
	ctx.focus.Request(tag)
}

func (ctx *Context) focusOnPress(tag event.Tag, history []widget.Press, before int) {
	ctx.focus.OnPress(tag, history, before)
}

func (ctx *Context) pushForeground(foreground color.NRGBA) func() {
	previous, hadPrevious := ctx.foreground, ctx.hasForeground
	ctx.foreground, ctx.hasForeground = foreground, true
	return func() {
		ctx.foreground, ctx.hasForeground = previous, hadPrevious
	}
}

func (ctx *Context) foregroundColor() color.NRGBA {
	if ctx.hasForeground {
		return ctx.foreground
	}
	return ctx.Theme.Palette.Foreground
}

func activePresses(history []widget.Press) int {
	return flowstate.ActivePresses(history)
}

func (ctx *Context) endFrame() {
	frameKeys := ctx.keys.Frame()
	flowstate.Sweep(ctx.clickables, frameKeys, flowstate.KindClickable)
	flowstate.Sweep(ctx.buttons, frameKeys, flowstate.KindClickable)
	sweepEditorState(ctx.editors, frameKeys)
	flowstate.Sweep(ctx.inputs, frameKeys, flowstate.KindInput)
	flowstate.Sweep(ctx.combos, frameKeys, flowstate.KindComboBox)
	flowstate.Sweep(ctx.selects, frameKeys, flowstate.KindSelect)
	if ctx.activeSelectKey != "" && ctx.selects[ctx.activeSelectKey] == nil {
		ctx.activeSelectKey = ""
	}
	flowstate.Sweep(ctx.datePickers, frameKeys, flowstate.KindDatePicker)
	flowstate.Sweep(ctx.bools, frameKeys, flowstate.KindCheckbox)
	flowstate.Sweep(ctx.checkboxes, frameKeys, flowstate.KindCheckbox)
	flowstate.Sweep(ctx.switches, frameKeys, flowstate.KindSwitch)
	flowstate.Sweep(ctx.radioGroups, frameKeys, flowstate.KindRadioGroup)
	flowstate.Sweep(ctx.progressBars, frameKeys, flowstate.KindProgressBar)
	flowstate.Sweep(ctx.listBoxes, frameKeys, flowstate.KindListBox)
	flowstate.Sweep(ctx.popovers, frameKeys, flowstate.KindPopover)
	flowstate.Sweep(ctx.modals, frameKeys, flowstate.KindModal)
	flowstate.Sweep(ctx.lists, frameKeys, flowstate.KindList)
	flowstate.Sweep(ctx.scrolls, frameKeys, flowstate.KindScroll)
}

func sweepEditorState(states map[string]*widget.Editor, frameKeys map[string]flowstate.Kind) {
	for key := range states {
		kind := frameKeys[key]
		if kind != flowstate.KindEditor && kind != flowstate.KindInput {
			delete(states, key)
		}
	}
}

func (ctx *Context) clickable(key string) *widget.Clickable {
	_, clickable := ctx.clickableWithKey(key)
	return clickable
}

func (ctx *Context) clickableWithKey(key string) (string, *widget.Clickable) {
	key = ctx.claimKey(flowstate.KindClickable, key)
	if ctx.clickables == nil {
		ctx.clickables = make(map[string]*widget.Clickable)
	}
	if c := ctx.clickables[key]; c != nil {
		return key, c
	}
	c := new(widget.Clickable)
	ctx.clickables[key] = c
	return key, c
}

func (ctx *Context) buttonState(key string) *buttonState {
	if ctx.buttons == nil {
		ctx.buttons = make(map[string]*buttonState)
	}
	if state := ctx.buttons[key]; state != nil {
		return state
	}
	state := new(buttonState)
	ctx.buttons[key] = state
	return state
}

func (ctx *Context) editor(key string) *widget.Editor {
	_, e := ctx.editorWithKey(key)
	return e
}

func (ctx *Context) editorWithKey(key string) (string, *widget.Editor) {
	return ctx.editorWithKind(flowstate.KindEditor, key)
}

func (ctx *Context) inputEditor(key string) (string, *widget.Editor) {
	return ctx.editorWithKind(flowstate.KindInput, key)
}

func (ctx *Context) editorWithKind(kind flowstate.Kind, key string) (string, *widget.Editor) {
	key = ctx.claimKey(kind, key)
	if ctx.editors == nil {
		ctx.editors = make(map[string]*widget.Editor)
	}
	if e := ctx.editors[key]; e != nil {
		return key, e
	}
	e := new(widget.Editor)
	ctx.editors[key] = e
	return key, e
}

func (ctx *Context) inputState(key string) *inputState {
	if ctx.inputs == nil {
		ctx.inputs = make(map[string]*inputState)
	}
	if state := ctx.inputs[key]; state != nil {
		return state
	}
	state := new(inputState)
	ctx.inputs[key] = state
	return state
}

func (ctx *Context) comboBoxState(key string) *comboBoxState {
	key = ctx.claimKey(flowstate.KindComboBox, key)
	if ctx.combos == nil {
		ctx.combos = make(map[string]*comboBoxState)
	}
	if state := ctx.combos[key]; state != nil {
		return state
	}
	state := new(comboBoxState)
	ctx.combos[key] = state
	return state
}

func (ctx *Context) selectState(key string) *selectState {
	key = ctx.claimKey(flowstate.KindSelect, key)
	if ctx.selects == nil {
		ctx.selects = make(map[string]*selectState)
	}
	if state := ctx.selects[key]; state != nil {
		return state
	}
	state := &selectState{key: key}
	ctx.selects[key] = state
	return state
}

func (ctx *Context) activateSelect(state *selectState) {
	if state == nil || state.key == "" || ctx.activeSelectKey == state.key {
		return
	}
	previousKey := ctx.activeSelectKey
	ctx.activeSelectKey = ""
	if previous := ctx.selects[previousKey]; previous != nil {
		previous.closeForPeer()
	}
	ctx.activeSelectKey = state.key
}

func (ctx *Context) releaseSelect(state *selectState) {
	if state != nil && ctx.activeSelectKey == state.key {
		ctx.activeSelectKey = ""
	}
}

func (ctx *Context) datePickerState(key string) *datePickerState {
	key = ctx.claimKey(flowstate.KindDatePicker, key)
	if ctx.datePickers == nil {
		ctx.datePickers = make(map[string]*datePickerState)
	}
	if state := ctx.datePickers[key]; state != nil {
		return state
	}
	state := new(datePickerState)
	ctx.datePickers[key] = state
	return state
}

func (ctx *Context) boolState(key string) *widget.Bool {
	_, state := ctx.boolStateWithKey(key)
	return state
}

func (ctx *Context) boolStateWithKey(key string) (string, *widget.Bool) {
	key = ctx.claimKey(flowstate.KindCheckbox, key)
	if ctx.bools == nil {
		ctx.bools = make(map[string]*widget.Bool)
	}
	if b := ctx.bools[key]; b != nil {
		return key, b
	}
	b := new(widget.Bool)
	ctx.bools[key] = b
	return key, b
}

func (ctx *Context) checkboxState(key string) *checkboxState {
	if ctx.checkboxes == nil {
		ctx.checkboxes = make(map[string]*checkboxState)
	}
	if state := ctx.checkboxes[key]; state != nil {
		return state
	}
	state := new(checkboxState)
	ctx.checkboxes[key] = state
	return state
}

func (ctx *Context) switchState(key string) *switchState {
	key = ctx.claimKey(flowstate.KindSwitch, key)
	if ctx.switches == nil {
		ctx.switches = make(map[string]*switchState)
	}
	if state := ctx.switches[key]; state != nil {
		return state
	}
	state := new(switchState)
	ctx.switches[key] = state
	return state
}

func (ctx *Context) radioGroupState(key string) *radioGroupState {
	key = ctx.claimKey(flowstate.KindRadioGroup, key)
	if ctx.radioGroups == nil {
		ctx.radioGroups = make(map[string]*radioGroupState)
	}
	if state := ctx.radioGroups[key]; state != nil {
		return state
	}
	state := new(radioGroupState)
	ctx.radioGroups[key] = state
	return state
}

func (ctx *Context) progressBarState(key string) *progressBarState {
	key = ctx.claimKey(flowstate.KindProgressBar, key)
	if ctx.progressBars == nil {
		ctx.progressBars = make(map[string]*progressBarState)
	}
	if state := ctx.progressBars[key]; state != nil {
		return state
	}
	state := new(progressBarState)
	ctx.progressBars[key] = state
	return state
}

func (ctx *Context) listBoxState(key string) *listBoxState {
	key = ctx.claimKey(flowstate.KindListBox, key)
	if ctx.listBoxes == nil {
		ctx.listBoxes = make(map[string]*listBoxState)
	}
	if state := ctx.listBoxes[key]; state != nil {
		return state
	}
	state := new(listBoxState)
	ctx.listBoxes[key] = state
	return state
}

func (ctx *Context) modalState(key string) *modalState {
	key = ctx.claimKey(flowstate.KindModal, key)
	if ctx.modals == nil {
		ctx.modals = make(map[string]*modalState)
	}
	if state := ctx.modals[key]; state != nil {
		return state
	}
	state := new(modalState)
	ctx.modals[key] = state
	return state
}

func (ctx *Context) popoverState(key string) *popoverState {
	key = ctx.claimKey(flowstate.KindPopover, key)
	if ctx.popovers == nil {
		ctx.popovers = make(map[string]*popoverState)
	}
	if state := ctx.popovers[key]; state != nil {
		return state
	}
	state := new(popoverState)
	ctx.popovers[key] = state
	return state
}

func (ctx *Context) hasVisiblePopover(key string) bool {
	if ctx.popovers == nil {
		return false
	}
	state := ctx.popovers[key]
	return state != nil && state.visible()
}

func (ctx *Context) hasVisibleModal(key string) bool {
	if ctx.modals == nil {
		return false
	}
	state := ctx.modals[key]
	return state != nil && state.visible()
}

func (ctx *Context) listState(key string) *layout.List {
	key = ctx.claimKey(flowstate.KindList, key)
	if ctx.lists == nil {
		ctx.lists = make(map[string]*layout.List)
	}
	if l := ctx.lists[key]; l != nil {
		return l
	}
	l := &layout.List{Axis: layout.Vertical}
	ctx.lists[key] = l
	return l
}

func (ctx *Context) scrollState(key string) *layout.List {
	key = ctx.claimKey(flowstate.KindScroll, key)
	if ctx.scrolls == nil {
		ctx.scrolls = make(map[string]*layout.List)
	}
	if s := ctx.scrolls[key]; s != nil {
		return s
	}
	s := &layout.List{Axis: layout.Vertical}
	ctx.scrolls[key] = s
	return s
}

func (ctx *Context) pushKey(key string) func() {
	return ctx.keys.Push(key)
}

func (ctx *Context) claimKey(kind flowstate.Kind, key string) string {
	return ctx.keys.Claim(kind, key)
}

func (ctx *Context) fullKey(key string) string {
	return ctx.keys.FullKey(key)
}
