package titlebar

import (
	"image"
	"image/color"
	"runtime"

	"gioui.org/f32"
	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
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
	key               string
	title             string
	leading           frame.Widget
	menu              frame.Widget
	center            frame.Widget
	trailing          frame.Widget
	showMinimize      bool
	showMaximize      bool
	showClose         bool
	onClose           func()
	clientDecorations bool
	customStyle       flowstyle.Style
}

func New(key, title string, menu frame.Widget) Widget {
	return Widget{
		key:               key,
		title:             title,
		menu:              menu,
		showMinimize:      true,
		showMaximize:      true,
		showClose:         true,
		clientDecorations: true,
	}
}

// NewPlatform creates client-side decorations on Windows and a content header
// without window controls on platforms that use native decorations.
func NewPlatform(key, title string, menu frame.Widget) Widget {
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
func (w Widget) Leading(value frame.Widget) Widget {
	w.leading = value
	return w
}

// Menu replaces the application menu supplied to New.
func (w Widget) Menu(value frame.Widget) Widget {
	w.menu = value
	return w
}

// Center replaces the default title with interactive center content.
func (w Widget) Center(value frame.Widget) Widget {
	w.center = value
	return w
}

// Trailing sets application controls at the end of the application header,
// before window controls when client-side decorations are active.
func (w Widget) Trailing(value frame.Widget) Widget {
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

func (w Widget) Style(value flowstyle.Style) Widget {
	w.customStyle = value
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, stateutil.KindTitleBar, w.key)
	state := frame.UseState[titleBarState](ctx, key, stateSlotTitleBar)
	state.decorations.Maximized = ctx.WindowState().Mode == frame.Maximized
	w.update(ctx, gtx, state)

	resolved := w.resolveStyle(ctx, gtx, key, !gtx.Enabled())
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		dims := w.layoutContent(ctx, gtx, key, state, resolved)
		layoutTitleBarSeparator(ctx, gtx, dims.Size, resolved.separator)
		return dims
	}))
}

func (w Widget) layoutContent(ctx *frame.Context, gtx layout.Context, key string, state *titleBarState, resolved titleBarResolvedStyle) layout.Dimensions {
	children := make([]frame.Widget, 0, 6)
	if w.leading != nil {
		children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return layoutSlot(ctx, gtx, w.leading, resolved.leading)
		}))
	}
	if w.menu != nil {
		children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return w.menu.Layout(ctx, gtx)
		}))
	}
	children = append(children, layoutui.Expanded(frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutCenterRegion(ctx, gtx, state, resolved)
	})))
	if w.trailing != nil {
		children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return layoutSlot(ctx, gtx, w.trailing, resolved.trailing)
		}))
	}
	if w.clientDecorations && w.visibleActions() != 0 {
		children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return w.layoutControls(ctx, gtx, key, state)
		}))
	}
	return layoutui.LayoutTrackedFlex(ctx, gtx, layout.Horizontal, 0, layout.Middle, children...)
}

func (w Widget) layoutCenterRegion(ctx *frame.Context, gtx layout.Context, state *titleBarState, resolved titleBarResolvedStyle) layout.Dimensions {
	if !w.clientDecorations && w.center == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	if w.center == nil && w.title == "" {
		return w.layoutCenterSpacer(gtx, state)
	}
	children := make([]frame.Widget, 0, 3)
	children = append(children, layoutui.Expanded(frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutCenterSpacer(gtx, state)
	})))
	if w.center != nil {
		children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return layoutSlot(ctx, gtx, w.center, resolved.center)
		}))
	} else {
		children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return state.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layoutTitle(ctx, gtx, w.title, resolved.title)
				})
			})
		}))
	}
	children = append(children, layoutui.Expanded(frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutCenterSpacer(gtx, state)
	})))
	return layoutui.LayoutTrackedFlex(ctx, gtx, layout.Horizontal, 0, layout.Middle, children...)
}

func (w Widget) layoutCenterSpacer(gtx layout.Context, state *titleBarState) layout.Dimensions {
	if !w.clientDecorations {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	return layoutMoveSpacer(gtx, state)
}

func layoutSlot(ctx *frame.Context, gtx layout.Context, child frame.Widget, resolved flowstyle.ResolvedStyle) layout.Dimensions {
	gtx.Constraints.Min.Y = 0
	return layoutui.LayoutResolved(ctx, gtx, resolved, child)
}

func layoutMoveSpacer(gtx layout.Context, state *titleBarState) layout.Dimensions {
	return state.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
}

func layoutTitle(ctx *frame.Context, gtx layout.Context, value string, style flowstyle.ResolvedStyle) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, style,
		text.New(value).
			MaxLines(1).
			Style(flowstyle.TextDeclaration(style.Text)),
	)
}

func (w Widget) update(ctx *frame.Context, gtx layout.Context, state *titleBarState) {
	if !w.clientDecorations {
		return
	}
	var active [len(titleBarActions)]int
	for index, action := range titleBarActions {
		if !w.actionVisible(action) {
			continue
		}
		active[index] = stateutil.ActivePresses(state.decorations.Clickable(action).History())
	}
	actions := state.decorations.Update(gtx) & w.visibleActions()
	for index, action := range titleBarActions {
		if !w.actionVisible(action) {
			continue
		}
		clickable := state.decorations.Clickable(action)
		frame.FocusOnPress(ctx, clickable, clickable.History(), active[index])
	}
	if actions&system.ActionClose != 0 {
		if w.onClose != nil {
			w.onClose()
			actions &^= system.ActionClose
		} else if frame.RequestWindowClose(ctx) {
			actions &^= system.ActionClose
		}
	}
	if actions == 0 {
		return
	}
	frame.PerformWindowActions(ctx, actions)
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

func (w Widget) layoutControls(ctx *frame.Context, gtx layout.Context, key string, state *titleBarState) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(titleBarActions))
	for _, action := range titleBarActions {
		if !w.actionVisible(action) {
			continue
		}
		action := action
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.layoutControl(ctx, gtx, key, state, action)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (w Widget) layoutControl(ctx *frame.Context, gtx layout.Context, key string, state *titleBarState, action system.Action) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.TitleBar
	clickable := state.decorations.Clickable(action)
	closeAction := action == system.ActionClose
	focusVisible := frame.FocusVisible(ctx, clickable, gtx.Focused(clickable))
	label := controlLabel(ctx, action, state.decorations.Maximized)
	glyph := controlGlyphFor(action, state.decorations.Maximized)
	styleState := flowstyle.StyleState{
		Hovered:      clickable.Hovered(),
		Pressed:      clickable.Pressed(),
		Focused:      gtx.Focused(clickable),
		FocusVisible: focusVisible,
	}
	resolved := w.resolveControlStyle(ctx, gtx, key, action, closeAction, styleState)
	foreground := resolvedControlForeground(resolved, ctx.ForegroundColor())
	return layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.Button.Add(gtx.Ops)
			semantic.LabelOp(label).Add(gtx.Ops)
			drawControl(gtx, size, glyph, foreground, focusVisible, frame.ActiveTheme(ctx).Palette.Focus, tokens)
			return layout.Dimensions{Size: size}
		})
	}))
}

func resolvedControlForeground(resolved flowstyle.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Text == nil {
		return fallback
	}
	if foreground, ok := styleruntime.Color(resolved.Text.Color); ok {
		return foreground
	}
	return fallback
}

type titleBarResolvedStyle struct {
	root      flowstyle.ResolvedStyle
	leading   flowstyle.ResolvedStyle
	title     flowstyle.ResolvedStyle
	center    flowstyle.ResolvedStyle
	trailing  flowstyle.ResolvedStyle
	separator flowstyle.ResolvedStyle
}

func (w Widget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, disabled bool) titleBarResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.TitleBar
	state := flowstyle.StyleState{Disabled: disabled}
	defaults := flowstyle.Style{}.
		Height(tokens.Height).
		PaddingLeft(tokens.PaddingX).
		Background(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceSecondary}).
		TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceSecondaryForeground}).
		Part(flowstyle.PartPrefix, flowstyle.Style{}.PaddingRight(tokens.LeadingGap)).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontSize(tokens.TitleTextSize)).
		Part(flowstyle.PartIndicator, flowstyle.Style{}.
			Height(tokens.BorderWidth).
			Background(flowstyle.SolidColor{Color: activeTheme.Palette.SeparatorColor()}))

	root := styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults,
		flowstyle.Style{},
		flowstyle.Style{},
		w.customStyle,
	)
	return titleBarResolvedStyle{
		root:      root,
		leading:   styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartPrefix, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
		title:     styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartLabel, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
		center:    styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartContent, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
		trailing:  styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartSuffix, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
		separator: styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartIndicator, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
	}
}

func (w Widget) resolveControlStyle(
	ctx *frame.Context,
	gtx layout.Context,
	key string,
	action system.Action,
	closeAction bool,
	state flowstyle.StyleState,
) flowstyle.ResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.TitleBar
	control := flowstyle.Style{}.
		Width(tokens.ControlWidth).
		FillHeight()
	if closeAction {
		control = control.
			When(flowstyle.Hovered, flowstyle.Style{}.
				Background(flowstyle.SolidColor{Color: tokens.CloseHover}).
				TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.AccentForeground})).
			When(flowstyle.Pressed, flowstyle.Style{}.
				Background(flowstyle.SolidColor{Color: tokens.ClosePressed}).
				TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.AccentForeground}))
	} else {
		control = control.
			When(flowstyle.Hovered, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceTertiary})).
			When(flowstyle.Pressed, flowstyle.Style{}.Background(flowstyle.SolidColor{
				Color: theme.ColorOr(tokens.ControlPressed, activeTheme.Palette.SurfacePressed),
			}))
	}
	defaults := flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: ctx.ForegroundColor()})
	component := flowstyle.Style{}.
		Part(flowstyle.PartItem, control)
	controlKey := frame.DerivedKey(ctx, key, "window-control:"+action.String())
	return styleruntime.ResolvePart(ctx, gtx, controlKey, flowstyle.PartItem, state, defaults, component, flowstyle.Style{}, w.customStyle)
}

func layoutTitleBarSeparator(ctx *frame.Context, gtx layout.Context, size image.Point, style flowstyle.ResolvedStyle) {
	if style.Box == nil || style.Box.Height == nil || size.X <= 0 || size.Y <= 0 {
		return
	}
	height := min(max(gtx.Dp(*style.Box.Height), 0), size.Y)
	if height == 0 {
		return
	}
	separatorGtx := gtx
	separatorGtx.Constraints = layout.Exact(image.Pt(size.X, height))
	offset := op.Offset(image.Pt(0, size.Y-height)).Push(gtx.Ops)
	layoutui.LayoutResolved(ctx, separatorGtx, style, nil)
	offset.Pop()
}

func drawControl(gtx layout.Context, size image.Point, glyph controlGlyph, foreground color.NRGBA, focusVisible bool, focus color.NRGBA, tokens theme.TitleBarTheme) {
	drawControlGlyph(gtx, size, glyph, foreground, tokens)
	if !focusVisible || focus.A == 0 {
		return
	}
	width := max(gtx.Dp(tokens.FocusRingWidth), 1)
	ring := image.Rectangle{Max: size}.Inset(width)
	if ring.Empty() {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(ring, 0).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, focus)
	stroke.Pop()
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

func drawControlGlyph(gtx layout.Context, size image.Point, glyph controlGlyph, foreground color.NRGBA, tokens theme.TitleBarTheme) {
	if glyph == controlGlyphNone || foreground.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	glyphSize := min(max(gtx.Dp(tokens.IconSize), 1), min(size.X, size.Y))
	stroke := controlGlyphStroke(gtx.Metric, tokens.IconStrokeWidth, glyphSize)
	left := (float32(size.X)-float32(glyphSize))/2 + stroke
	top := (float32(size.Y)-float32(glyphSize))/2 + stroke
	inner := controlGlyphBounds{
		left:   left,
		top:    top,
		right:  left + float32(glyphSize) - stroke*2,
		bottom: top + float32(glyphSize) - stroke*2,
	}
	if inner.empty() {
		return
	}

	switch glyph {
	case controlGlyphMinimize:
		y := (inner.top + inner.bottom) / 2
		drawControlLine(gtx.Ops, f32.Pt(inner.left, y), f32.Pt(inner.right, y), stroke, foreground)
	case controlGlyphMaximize:
		drawControlOutline(gtx.Ops, inner, stroke, foreground)
	case controlGlyphRestore:
		drawRestoreGlyph(gtx.Ops, inner, stroke, foreground)
	case controlGlyphClose:
		drawControlLine(gtx.Ops, f32.Pt(inner.left, inner.top), f32.Pt(inner.right, inner.bottom), stroke, foreground)
		drawControlLine(gtx.Ops, f32.Pt(inner.right, inner.top), f32.Pt(inner.left, inner.bottom), stroke, foreground)
	}
}

type controlGlyphBounds struct {
	left, top, right, bottom float32
}

func (b controlGlyphBounds) empty() bool {
	return b.right <= b.left || b.bottom <= b.top
}

func controlGlyphStroke(metric unit.Metric, width unit.Dp, glyphSize int) float32 {
	if width <= 0 {
		width = 1
	}
	scale := metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	stroke := max(float32(width)*scale, 1)
	return min(stroke, max(float32(glyphSize)/3, 1))
}

func drawRestoreGlyph(ops *op.Ops, bounds controlGlyphBounds, stroke float32, foreground color.NRGBA) {
	offset := stroke * 2
	side := min(bounds.right-bounds.left, bounds.bottom-bounds.top) - offset
	if side <= stroke*2 {
		drawControlOutline(ops, bounds, stroke, foreground)
		return
	}
	back := controlGlyphBounds{
		left: bounds.left + offset, top: bounds.top,
		right: bounds.left + offset + side, bottom: bounds.top + side,
	}
	front := controlGlyphBounds{
		left: bounds.left, top: bounds.top + offset,
		right: bounds.left + side, bottom: bounds.top + offset + side,
	}

	drawControlLine(ops, f32.Pt(back.left, back.top), f32.Pt(back.right, back.top), stroke, foreground)
	drawControlLine(ops, f32.Pt(back.right, back.top), f32.Pt(back.right, back.bottom), stroke, foreground)
	drawControlLine(ops, f32.Pt(back.left, back.top), f32.Pt(back.left, front.top), stroke, foreground)
	drawControlLine(ops, f32.Pt(front.right, back.bottom), f32.Pt(back.right, back.bottom), stroke, foreground)
	drawControlOutline(ops, front, stroke, foreground)
}

func drawControlOutline(ops *op.Ops, bounds controlGlyphBounds, stroke float32, foreground color.NRGBA) {
	if bounds.empty() || stroke <= 0 {
		return
	}
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(f32.Pt(bounds.left, bounds.top))
	path.LineTo(f32.Pt(bounds.right, bounds.top))
	path.LineTo(f32.Pt(bounds.right, bounds.bottom))
	path.LineTo(f32.Pt(bounds.left, bounds.bottom))
	path.Close()
	paint.FillShape(ops, foreground, clip.Stroke{Path: path.End(), Width: stroke}.Op())
}

func drawControlLine(ops *op.Ops, from, to f32.Point, stroke float32, foreground color.NRGBA) {
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(from)
	path.LineTo(to)
	paint.FillShape(ops, foreground, clip.Stroke{Path: path.End(), Width: stroke}.Op())
}

func controlLabel(ctx *frame.Context, action system.Action, maximized bool) string {
	chinese := frame.ActiveLanguage(ctx) == locale.LanguageChinese
	switch action {
	case system.ActionMinimize:
		if chinese {
			return "最小化"
		}
		return "Minimize"
	case system.ActionMaximize:
		if maximized {
			if chinese {
				return "还原"
			}
			return "Restore"
		}
		if chinese {
			return "最大化"
		}
		return "Maximize"
	case system.ActionClose:
		if chinese {
			return "关闭"
		}
		return "Close"
	default:
		return ""
	}
}
