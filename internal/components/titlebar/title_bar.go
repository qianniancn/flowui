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
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
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
	theme func(*theme.Theme)
	key   string
	title string
	menu  frame.Widget
}

func New(key, title string, menu frame.Widget) Widget {
	return Widget{key: key, title: title, menu: menu}
}

func (w Widget) Theme(fn func(*theme.Theme)) Widget {
	w.theme = fn
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, w.theme); restore != nil {
		defer restore()
	}
	key := frame.ClaimKey(ctx, stateutil.KindTitleBar, w.key)
	state := frame.UseState[titleBarState](ctx, key, stateSlotTitleBar)
	state.decorations.Maximized = ctx.WindowState().Mode == frame.Maximized
	w.update(ctx, gtx, state)

	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.TitleBar
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, max(gtx.Dp(tokens.Height), 1)))
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)
	drawRoot(rootGtx, size, activeTheme)

	background := activeTheme.Palette.SurfaceSecondary
	foreground := activeTheme.Palette.SurfaceSecondaryForeground
	restore := frame.PushColors(ctx, foreground, background)
	defer restore()

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(rootGtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if w.menu == nil {
				return layout.Dimensions{}
			}
			return layout.Inset{Left: tokens.PaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return w.menu.Layout(ctx, gtx)
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return state.decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return text.New(w.title).
						Size(float32(tokens.TitleTextSize)).
						Color(foreground).
						MaxLines(1).
						Layout(ctx, gtx)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.layoutControls(ctx, gtx, state)
		}),
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
	style := controlStyleFor(frame.ActiveTheme(ctx), clickable, closeAction)
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

func controlStyleFor(activeTheme *theme.Theme, clickable *widget.Clickable, closeAction bool) controlStyle {
	style := controlStyle{
		foreground: activeTheme.Palette.SurfaceSecondaryForeground,
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

func drawRoot(gtx layout.Context, size image.Point, activeTheme *theme.Theme) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	paint.FillShape(gtx.Ops, activeTheme.Palette.SurfaceSecondary, clip.Rect{Max: size}.Op())
	width := min(max(gtx.Dp(activeTheme.Components.TitleBar.BorderWidth), 1), size.Y)
	rect := image.Rect(0, size.Y-width, size.X, size.Y)
	paint.FillShape(gtx.Ops, activeTheme.Palette.SeparatorColor(), clip.Rect(rect).Op())
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
