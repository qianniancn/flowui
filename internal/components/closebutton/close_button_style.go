package closebutton

import (
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type closeButtonWidget struct {
	background   color.NRGBA
	foreground   color.NRGBA
	radius       unit.Dp
	padding      unit.Dp
	iconSize     unit.Dp
	pressedScale float32
}

type closeButtonResolvedStyle struct {
	root flowstyle.ResolvedStyle
	icon flowstyle.ResolvedStyle
}

func closeButtonPressedScale(pressedScale float32) float32 {
	if pressedScale <= 0 || pressedScale > 1 {
		return 0.93
	}
	return pressedScale
}

// resolveStyle uses the four-slot protocol. CloseButton has no variant/size
// layers, so those slots stay empty (defaults + instance only).
func (b CloseButtonWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) closeButtonResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := closeButtonDefaultDeclaration(activeTheme)
	root := styleruntime.Resolve(ctx, gtx, key, state, defaults, flowstyle.Style{}, flowstyle.Style{}, b.customStyle)
	icon := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartIcon, state, defaults, flowstyle.Style{}, flowstyle.Style{}, b.customStyle)
	return closeButtonResolvedStyle{root: root, icon: icon}
}

func closeButtonDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	base := closeButtonWidgetFor(activeTheme, false, false)
	hovered := closeButtonWidgetFor(activeTheme, true, false)
	disabled := closeButtonWidgetFor(activeTheme, false, true)
	component := activeTheme.Components.CloseButton
	return flowstyle.Style{}.
		Width(component.Size).
		AspectRatio(1).
		Padding(base.padding).
		Background(flowstyle.SolidColor{Color: base.background}).
		TextColor(flowstyle.SolidColor{Color: base.foreground}).
		Radius(base.radius).
		Outline(component.FocusRingWidth, 0, flowstyle.WithAlpha(flowstyle.TokenFocus, 0)).
		Opacity(1).
		Scale(1, 1).
		Cursor(pointer.CursorPointer).
		Part(flowstyle.PartIcon, flowstyle.Style{}.Width(component.IconSize).Height(component.IconSize)).
		Transition(flowstyle.PropBackgroundColor, closeButtonColorDuration).
		Transition(flowstyle.PropOutlineColor, closeButtonColorDuration).
		Transition(flowstyle.PropTransform, closeButtonScaleDuration).
		When(flowstyle.Hovered,
			flowstyle.Style{}.Background(flowstyle.SolidColor{Color: hovered.background}),
		).
		When(flowstyle.Pressed,
			flowstyle.Style{}.Scale(closeButtonPressedScale(base.pressedScale), closeButtonPressedScale(base.pressedScale)),
		).
		When(flowstyle.All(flowstyle.FocusVisible, flowstyle.Not(flowstyle.Disabled)),
			flowstyle.Style{}.Outline(component.FocusRingWidth, 0, flowstyle.TokenFocus),
		).
		When(flowstyle.Disabled,
			flowstyle.Style{}.
				Background(flowstyle.SolidColor{Color: disabled.background}).
				TextColor(flowstyle.SolidColor{Color: disabled.foreground}).
				Cursor(pointer.CursorDefault),
		)

}

func closeButtonWidgetFor(activeTheme *theme.Theme, hovered, disabled bool) closeButtonWidget {
	component := activeTheme.Components.CloseButton
	background := activeTheme.Palette.SurfaceRaised
	if hovered && !disabled {
		background = activeTheme.Palette.SurfacePressed
	}
	foreground := activeTheme.Palette.MutedForeground
	if disabled {
		background = activeTheme.DisabledColor(background)
		foreground = activeTheme.DisabledColor(foreground)
	}
	return closeButtonWidget{
		background:   background,
		foreground:   foreground,
		radius:       component.Radius,
		padding:      component.Padding,
		iconSize:     component.IconSize,
		pressedScale: component.PressedScale,
	}
}
