package titlebar

import (
	"image"
	"runtime"

	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotTitleBar = "title-bar"

var titleBarActions = [...]system.Action{
	system.ActionMinimize,
	system.ActionMaximize,
	system.ActionClose,
}

type controlGlyph uint8

const (
	controlGlyphNone controlGlyph = iota
	controlGlyphMinimize
	controlGlyphMaximize
	controlGlyphRestore
	controlGlyphClose
)

type titleBarState struct {
	decorations widget.Decorations
}

// Widget lays out either client-side window decorations or an in-content
// application header.
type Widget struct {
	override *theme.TitleBarOverride
	key      string
	title    string
	leading  core.Widget
	menu     core.Widget
	center   core.Widget
	trailing core.Widget
	// Keep the title widget alive across frames so its text Provider cache is
	// reused instead of being rebuilt for every TitleBar layout.
	titleWidget       *text.Widget
	showMinimize      bool
	showMaximize      bool
	showClose         bool
	onClose           func()
	clientDecorations bool
	visualCache       *titleBarVisualCache
}

func New(key, title string, menu core.Widget) Widget {
	titleWidget := text.New(title).MaxLines(1)
	return Widget{
		key:               key,
		title:             title,
		menu:              menu,
		titleWidget:       &titleWidget,
		showMinimize:      true,
		showMaximize:      true,
		showClose:         true,
		clientDecorations: true,
		visualCache:       &titleBarVisualCache{},
	}
}

// NewPlatform creates client-side decorations on Windows and a content header
// without window controls on platforms that use native decorations.
func NewPlatform(key, title string, menu core.Widget) Widget {
	w := New(key, title, menu)
	w.clientDecorations = Supported()
	return w
}

// Supported reports whether the current platform uses FlowUI client-side
// window decorations.
func Supported() bool {
	return runtime.GOOS == "windows"
}

// Leading sets content before the application menu, such as an app icon.
func (w Widget) Leading(value core.Widget) Widget {
	w.leading = value
	return w
}

// Menu replaces the application menu supplied to New.
func (w Widget) Menu(value core.Widget) Widget {
	w.menu = value
	return w
}

// Center replaces the default title with interactive center content.
func (w Widget) Center(value core.Widget) Widget {
	w.center = value
	return w
}

// Trailing sets application controls at the end of the application header,
// before window controls when client-side decorations are active.
func (w Widget) Trailing(value core.Widget) Widget {
	w.trailing = value
	return w
}

// ShowMinimize controls whether the Windows minimize button is present.
func (w Widget) ShowMinimize(visible bool) Widget {
	w.showMinimize = visible
	return w
}

// ShowMaximize controls whether the Windows maximize or restore button is
// present.
func (w Widget) ShowMaximize(visible bool) Widget {
	w.showMaximize = visible
	return w
}

// ShowClose controls whether the Windows close button is present.
func (w Widget) ShowClose(visible bool) Widget {
	w.showClose = visible
	return w
}

// OnClose handles client-side close requests instead of performing the native
// close action.
func (w Widget) OnClose(fn func()) Widget {
	w.onClose = fn
	return w
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.TitleBarOverride) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindTitleBar, w.key)
	state := core.UseState[titleBarState](ctx, key, stateSlotTitleBar)
	state.decorations.Maximized = ctx.WindowState().Mode == core.Maximized
	w.update(ctx, gtx, state)

	resolved := w.resolveStyle(ctx, gtx, key, !gtx.Enabled(), override)
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		dims := w.layoutContent(ctx, gtx, key, state, resolved)
		layoutTitleBarSeparator(ctx, gtx, dims.Size, resolved.separator)
		return dims
	}))
}

func (w Widget) layoutContent(ctx *core.Context, gtx layout.Context, key string, state *titleBarState, resolved titleBarResolvedStyle) layout.Dimensions {
	children := make([]core.Widget, 0, 6)
	if w.leading != nil {
		children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return layoutSlot(ctx, gtx, w.leading, resolved.leading)
		}))
	}
	if w.menu != nil {
		children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return w.menu.Layout(ctx, gtx)
		}))
	}
	children = append(children, layoutui.Expanded(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutCenterRegion(ctx, gtx, state, resolved)
	})))
	if w.trailing != nil {
		children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return layoutSlot(ctx, gtx, w.trailing, resolved.trailing)
		}))
	}
	if w.clientDecorations && w.visibleActions() != 0 {
		children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return w.layoutControls(ctx, gtx, key, state, resolved.visual)
		}))
	}
	return layoutui.TrackedFlex(ctx, gtx, layout.Horizontal, 0, layout.Middle, children...)
}

func (w Widget) layoutCenterRegion(ctx *core.Context, gtx layout.Context, state *titleBarState, resolved titleBarResolvedStyle) layout.Dimensions {
	if !w.clientDecorations && w.center == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	if w.center == nil && w.title == "" {
		return w.layoutCenterSpacer(gtx, state)
	}
	children := make([]core.Widget, 0, 3)
	children = append(children, layoutui.Expanded(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutCenterSpacer(gtx, state)
	})))
	if w.center != nil {
		children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return layoutSlot(ctx, gtx, w.center, resolved.center)
		}))
	} else {
		children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return state.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layoutTitleWidget(ctx, gtx, w.titleWidget, resolved.title)
				})
			})
		}))
	}
	children = append(children, layoutui.Expanded(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutCenterSpacer(gtx, state)
	})))
	return layoutui.TrackedFlex(ctx, gtx, layout.Horizontal, 0, layout.Middle, children...)
}

func (w Widget) layoutCenterSpacer(gtx layout.Context, state *titleBarState) layout.Dimensions {
	if !w.clientDecorations {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	return layoutMoveSpacer(gtx, state)
}

func layoutSlot(ctx *core.Context, gtx layout.Context, child core.Widget, resolved style.ResolvedStyle) layout.Dimensions {
	gtx.Constraints.Min.Y = 0
	return layoutui.Resolved(ctx, gtx, resolved, child)
}

func layoutMoveSpacer(gtx layout.Context, state *titleBarState) layout.Dimensions {
	return state.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
}

func layoutTitle(ctx *core.Context, gtx layout.Context, value string, resolved style.ResolvedStyle) layout.Dimensions {
	title := text.New(value).MaxLines(1)
	return layoutTitleWidget(ctx, gtx, &title, resolved)
}

func layoutTitleWidget(ctx *core.Context, gtx layout.Context, value *text.Widget, resolved style.ResolvedStyle) layout.Dimensions {
	if value == nil {
		return layout.Dimensions{}
	}
	title := text.WithResolvedText(*value, resolved.Text)
	return layoutui.Resolved(ctx, gtx, resolved,
		title,
	)
}

func (w Widget) update(ctx *core.Context, gtx layout.Context, state *titleBarState) {
	if !w.clientDecorations {
		return
	}
	var active [len(titleBarActions)]int
	for index, action := range titleBarActions {
		if !w.actionVisible(action) {
			continue
		}
		active[index] = core.ActivePresses(state.decorations.Clickable(action).History())
	}
	actions := state.decorations.Update(gtx) & w.visibleActions()
	for index, action := range titleBarActions {
		if !w.actionVisible(action) {
			continue
		}
		clickable := state.decorations.Clickable(action)
		core.FocusOnPress(ctx, clickable, clickable.History(), active[index])
	}
	if actions&system.ActionClose != 0 {
		if w.onClose != nil {
			w.onClose()
			actions &^= system.ActionClose
		} else if core.RequestWindowClose(ctx) {
			actions &^= system.ActionClose
		}
	}
	if actions == 0 {
		return
	}
	core.PerformWindowActions(ctx, actions)
}

func (w Widget) visibleActions() system.Action {
	var actions system.Action
	if w.showMinimize {
		actions |= system.ActionMinimize
	}
	if w.showMaximize {
		actions |= system.ActionMaximize | system.ActionUnmaximize
	}
	if w.showClose {
		actions |= system.ActionClose
	}
	return actions
}

func (w Widget) actionVisible(action system.Action) bool {
	switch action {
	case system.ActionMinimize:
		return w.showMinimize
	case system.ActionMaximize, system.ActionUnmaximize:
		return w.showMaximize
	case system.ActionClose:
		return w.showClose
	default:
		return false
	}
}

func (w Widget) layoutControls(ctx *core.Context, gtx layout.Context, key string, state *titleBarState, visual theme.TitleBarVisual) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(titleBarActions))
	for _, action := range titleBarActions {
		if !w.actionVisible(action) {
			continue
		}
		action := action
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.layoutControl(ctx, gtx, key, state, action, visual)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (w Widget) layoutControl(ctx *core.Context, gtx layout.Context, key string, state *titleBarState, action system.Action, visual theme.TitleBarVisual) layout.Dimensions {
	clickable := state.decorations.Clickable(action)
	closeAction := action == system.ActionClose
	focusVisible := core.FocusVisible(ctx, clickable, gtx.Focused(clickable))
	label := controlLabel(ctx, action, state.decorations.Maximized)
	glyph := controlGlyphFor(action, state.decorations.Maximized)
	styleState := style.State{
		Hovered:      clickable.Hovered(),
		Pressed:      clickable.Pressed(),
		Focused:      gtx.Focused(clickable),
		FocusVisible: focusVisible,
	}
	controlStyle := visual.ControlStyle
	if closeAction {
		controlStyle = visual.CloseControlStyle
	}
	resolved := computed.ResolvePart(ctx, gtx, core.DerivedKey(ctx, key, "window-control:"+action.String()), style.PartItem, styleState, visual.Defaults, controlStyle, style.Style{}, visual.Override)
	foreground, _ := visual.ForegroundColor(resolved)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.Button.Add(gtx.Ops)
			semantic.LabelOp(label).Add(gtx.Ops)
			visual.Control(gtx, theme.TitleBarControlRenderRequest{
				Size: size, Glyph: theme.TitleBarControlGlyph(glyph), Foreground: foreground,
				Focus: visual.Focus, FocusVisible: focusVisible,
			})
			return layout.Dimensions{Size: size}
		})
	}))
}

type titleBarResolvedStyle struct {
	root      style.ResolvedStyle
	leading   style.ResolvedStyle
	title     style.ResolvedStyle
	center    style.ResolvedStyle
	trailing  style.ResolvedStyle
	separator style.ResolvedStyle
	visual    theme.TitleBarVisual
}

func (w Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, disabled bool, override *theme.TitleBarOverride) titleBarResolvedStyle {
	visual := w.visualWithCache(ctx, override)
	state := style.State{Disabled: disabled}
	root := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		visual.Defaults,
		style.Style{},
		style.Style{},
		visual.Override,
	)
	return titleBarResolvedStyle{
		root:      root,
		leading:   computed.ResolvePart(ctx, gtx, key, style.PartPrefix, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		title:     computed.ResolvePart(ctx, gtx, key, style.PartLabel, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		center:    computed.ResolvePart(ctx, gtx, key, style.PartContent, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		trailing:  computed.ResolvePart(ctx, gtx, key, style.PartSuffix, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		separator: computed.ResolvePart(ctx, gtx, key, style.PartIndicator, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override),
		visual:    visual,
	}
}

func layoutTitleBarSeparator(ctx *core.Context, gtx layout.Context, size image.Point, resolved style.ResolvedStyle) {
	if resolved.Box == nil || resolved.Box.Height == nil || size.X <= 0 || size.Y <= 0 {
		return
	}
	height := min(max(gtx.Dp(*resolved.Box.Height), 0), size.Y)
	if height == 0 {
		return
	}
	separatorGtx := gtx
	separatorGtx.Constraints = layout.Exact(image.Pt(size.X, height))
	offset := op.Offset(image.Pt(0, size.Y-height)).Push(gtx.Ops)
	layoutui.Resolved(ctx, separatorGtx, resolved, nil)
	offset.Pop()
}

func controlGlyphFor(action system.Action, maximized bool) controlGlyph {
	switch action {
	case system.ActionMinimize:
		return controlGlyphMinimize
	case system.ActionMaximize:
		if maximized {
			return controlGlyphRestore
		}
		return controlGlyphMaximize
	case system.ActionClose:
		return controlGlyphClose
	default:
		return controlGlyphNone
	}
}

func controlLabel(ctx *core.Context, action system.Action, maximized bool) string {
	switch action {
	case system.ActionMinimize:
		return ctx.Text(i18n.TitleBarMinimize)
	case system.ActionMaximize:
		if maximized {
			return ctx.Text(i18n.TitleBarRestore)
		}
		return ctx.Text(i18n.TitleBarMaximize)
	case system.ActionClose:
		return ctx.Text(i18n.TitleBarClose)
	default:
		return ""
	}
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.TitleBarOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
