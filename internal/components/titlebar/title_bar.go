package titlebar

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

const stateSlotTitleBar = "title-bar"

var titleBarActions = [...]system.Action{
	system.ActionMinimize,
	system.ActionMaximize,
	system.ActionClose,
}

type titleBarState struct {
	decorations widget.Decorations
}

// Widget provides client-side window decorations with an application menu.
type Widget struct {
	key         string
	title       string
	menu        frame.Widget
	customStyle flowstyle.Style
}

func New(key, title string, menu frame.Widget) Widget {
	return Widget{key: key, title: title, menu: menu}
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
		dims := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if w.menu == nil {
					return layout.Dimensions{}
				}
				return w.menu.Layout(ctx, gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return state.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layoutTitle(ctx, gtx, w.title, resolved.title)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.layoutControls(ctx, gtx, state)
			}),
		)
		layoutTitleBarSeparator(ctx, gtx, dims.Size, resolved.separator)
		return dims
	}))
}

func layoutTitle(ctx *frame.Context, gtx layout.Context, value string, style flowstyle.ResolvedStyle) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, style,
		text.New(value).
			MaxLines(1).
			Style(flowstyle.TextDeclaration(style.Text)),
	)
}

func (w Widget) update(ctx *frame.Context, gtx layout.Context, state *titleBarState) {
	var active [len(titleBarActions)]int
	for index, action := range titleBarActions {
		active[index] = stateutil.ActivePresses(state.decorations.Clickable(action).History())
	}
	actions := state.decorations.Update(gtx)
	for index, action := range titleBarActions {
		clickable := state.decorations.Clickable(action)
		frame.FocusOnPress(ctx, clickable, clickable.History(), active[index])
	}
	if actions == 0 {
		return
	}
	frame.PerformWindowActions(ctx, actions)
}

func (w Widget) layoutControls(ctx *frame.Context, gtx layout.Context, state *titleBarState) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.layoutControl(ctx, gtx, state, system.ActionMinimize)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.layoutControl(ctx, gtx, state, system.ActionMaximize)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.layoutControl(ctx, gtx, state, system.ActionClose)
		}),
	)
}

func (w Widget) layoutControl(ctx *frame.Context, gtx layout.Context, state *titleBarState, action system.Action) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.TitleBar
	size := gtx.Constraints.Constrain(image.Pt(gtx.Dp(tokens.ControlWidth), gtx.Dp(tokens.Height)))
	gtx.Constraints = layout.Exact(size)
	clickable := state.decorations.Clickable(action)
	closeAction := action == system.ActionClose
	style := controlStyleFor(frame.ActiveTheme(ctx), clickable, closeAction, ctx.ForegroundColor())
	focusVisible := frame.FocusVisible(ctx, clickable, gtx.Focused(clickable))
	label := controlLabel(ctx, action, state.decorations.Maximized)
	data := controlIcon(action, state.decorations.Maximized)

	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(label).Add(gtx.Ops)
		drawControl(gtx, size, data, style, focusVisible, tokens)
		return layout.Dimensions{Size: size}
	})
}

type controlStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	focus      color.NRGBA
}

func controlStyleFor(activeTheme *theme.Theme, clickable *widget.Clickable, closeAction bool, foreground color.NRGBA) controlStyle {
	style := controlStyle{
		foreground: foreground,
		focus:      activeTheme.Palette.Focus,
	}
	if closeAction && clickable.Hovered() {
		style.background = activeTheme.Components.TitleBar.CloseHover
		style.foreground = activeTheme.Palette.AccentForeground
		if clickable.Pressed() {
			style.background = activeTheme.Components.TitleBar.ClosePressed
		}
		return style
	}
	if clickable.Hovered() {
		style.background = activeTheme.Palette.SurfaceTertiary
	}
	if clickable.Pressed() {
		style.background = theme.ColorOr(activeTheme.Components.TitleBar.ControlPressed, activeTheme.Palette.SurfacePressed)
	}
	return style
}

type titleBarResolvedStyle struct {
	root      flowstyle.ResolvedStyle
	title     flowstyle.ResolvedStyle
	separator flowstyle.ResolvedStyle
}

func (w Widget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, disabled bool) titleBarResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.TitleBar
	state := flowstyle.StyleState{Disabled: disabled}
	defaults := flowstyle.Style{}.
		Height(tokens.Height).
		PaddingX(tokens.PaddingX).
		Background(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceSecondary}).
		TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceSecondaryForeground}).
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
		title:     styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartLabel, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
		separator: styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartIndicator, state, defaults, flowstyle.Style{}, flowstyle.Style{}, w.customStyle),
	}
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

func drawControl(gtx layout.Context, size image.Point, data []byte, style controlStyle, focusVisible bool, tokens theme.TitleBarTheme) {
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.Rect{Max: size}.Op())
	}
	iconSize := min(max(gtx.Dp(tokens.IconSize), 1), min(size.X, size.Y))
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
	offset := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
	icon.Layout(data, iconGtx, style.foreground)
	offset.Pop()
	if !focusVisible || style.focus.A == 0 {
		return
	}
	width := max(gtx.Dp(tokens.FocusRingWidth), 1)
	ring := image.Rectangle{Max: size}.Inset(width)
	if ring.Empty() {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(ring, 0).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, style.focus)
	stroke.Pop()
}

func controlIcon(action system.Action, maximized bool) []byte {
	switch action {
	case system.ActionMinimize:
		return lucide.Minus.IconVG()
	case system.ActionMaximize:
		if maximized {
			return lucide.Copy.IconVG()
		}
		return lucide.Square.IconVG()
	case system.ActionClose:
		return lucide.X.IconVG()
	default:
		return nil
	}
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
