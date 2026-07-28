package field

import (
	"image/color"
	"math"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type Variant uint8

const (
	Primary Variant = iota
	Secondary
)

const TransitionDuration = 150 * time.Millisecond

type DeclarationOptions struct {
	TargetPart          flowstyle.Part
	Height              unit.Dp
	MinHeight           unit.Dp
	Radius              unit.Dp
	PaddingX            unit.Dp
	PaddingY            unit.Dp
	TextSize            unit.Sp
	LineHeight          unit.Sp
	FocusRingWidth      unit.Dp
	InvalidOutlineWidth unit.Dp
	ShadowColor         color.NRGBA
	ShadowOpacity       float32
	ShadowStrength      float32
	FillWidth           bool
}

// DefaultDeclaration is shared by text inputs and the field face of compound
// controls. TargetPart selects whether the declaration styles the root or an
// internal part such as PartContent.
func DefaultDeclaration(activeTheme *theme.Theme, variant Variant, options DeclarationOptions) flowstyle.Style {
	background := flowstyle.TokenFieldBackground
	hover := flowstyle.TokenFieldHover
	focused := flowstyle.TokenFieldFocus
	shadowOpacity := options.ShadowOpacity
	if variant == Secondary {
		background = flowstyle.TokenDefault
		hover = flowstyle.TokenDefaultHover
		focused = flowstyle.TokenDefault
		shadowOpacity = 0
	}

	transparent := flowstyle.RGBA(0)
	content := flowstyle.Style{}.
		Background(background).
		TextColor(flowstyle.TokenFieldForeground).
		Radius(options.Radius).
		Overflow(flowstyle.OverflowHidden).
		Cursor(pointer.CursorText).
		Opacity(1).
		Outline(options.FocusRingWidth, 0, transparent).
		Transition(flowstyle.PropBackgroundColor, TransitionDuration).
		Transition(flowstyle.PropOutlineColor, TransitionDuration).
		Transition(flowstyle.PropOpacity, TransitionDuration).
		When(flowstyle.Hovered, flowstyle.Style{}.Background(hover)).
		When(flowstyle.Focused, flowstyle.Style{}.
			Background(focused).
			Outline(options.FocusRingWidth, 0, flowstyle.TokenFocus)).
		When(flowstyle.Invalid, flowstyle.Style{}.
			Background(focused).
			Outline(options.InvalidOutlineWidth, 0, flowstyle.TokenDanger)).
		When(flowstyle.All(flowstyle.Invalid, flowstyle.Focused), flowstyle.Style{}.
			Outline(options.FocusRingWidth, 0, flowstyle.TokenDanger)).
		When(flowstyle.Disabled, flowstyle.Style{}.Opacity(activeTheme.DisabledOpacityValue()))
	if options.Height > 0 {
		content = content.Height(options.Height)
	}
	if options.MinHeight > 0 {
		content = content.MinHeight(options.MinHeight)
	}
	if options.PaddingX != 0 {
		content = content.PaddingX(options.PaddingX)
	}
	if options.PaddingY != 0 {
		content = content.PaddingY(options.PaddingY)
	}
	if options.TextSize > 0 {
		content = content.FontSize(options.TextSize)
	}
	if options.LineHeight > 0 {
		content = content.LineHeight(options.LineHeight)
	}
	if options.FillWidth {
		content = content.FillWidth()
	}
	content = addShadows(content, activeTheme, options.ShadowColor, shadowOpacity, options.ShadowStrength)

	return flowstyle.Style{}.
		Part(options.TargetPart, content).
		Part(flowstyle.PartPlaceholder, flowstyle.Style{}.TextColor(flowstyle.TokenFieldPlaceholder)).
		Part(flowstyle.PartSelection, flowstyle.Style{}.Background(flowstyle.TokenSelection))

}

type Colors struct {
	Foreground  color.NRGBA
	Placeholder color.NRGBA
	Selection   color.NRGBA
}

type Resolved struct {
	Content flowstyle.ResolvedStyle
	Colors  Colors
}

func Resolve(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState, variant Variant, options DeclarationOptions, custom flowstyle.Style) Resolved {
	options.TargetPart = flowstyle.PartContent
	activeTheme := frame.ActiveTheme(ctx)
	defaults, variantStyle, size := fieldStyleDeclarations(activeTheme, variant, options)
	content := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartContent, state, defaults, variantStyle, size, custom)
	placeholder := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartPlaceholder, state, defaults, variantStyle, size, custom)
	selection := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartSelection, state, defaults, variantStyle, size, custom)
	return Resolved{
		Content: content,
		Colors:  ResolvedColors(content, placeholder, selection, activeTheme),
	}
}

// fieldStyleDeclarations places primary chrome in defaults and secondary chrome
// in the variant slot (Button four-slot protocol).
func fieldStyleDeclarations(activeTheme *theme.Theme, variant Variant, options DeclarationOptions) (defaults, variantStyle, size flowstyle.Style) {
	switch variant {
	case Secondary:
		return flowstyle.Style{}, DefaultDeclaration(activeTheme, Secondary, options), flowstyle.Style{}
	default:
		return DefaultDeclaration(activeTheme, Primary, options), flowstyle.Style{}, flowstyle.Style{}
	}
}

func ResolvedColors(content, placeholder, selection flowstyle.ResolvedStyle, activeTheme *theme.Theme) Colors {
	result := Colors{
		Foreground:  activeTheme.Palette.FieldForegroundColor(),
		Placeholder: activeTheme.Palette.FieldPlaceholderColor(),
		Selection:   activeTheme.Palette.Selection,
	}
	if content.Text != nil {
		if value, ok := styleruntime.Color(content.Text.Color); ok {
			result.Foreground = value
		}
	}
	if placeholder.Text != nil {
		if value, ok := styleruntime.Color(placeholder.Text.Color); ok {
			result.Placeholder = value
		}
	}
	if selection.Paint != nil {
		if brush, ok := styleruntime.Brush(selection.Paint.Background); ok {
			result.Selection = brush.ColorAt(.5)
		}
	}
	return result
}

func addShadows(declaration flowstyle.Style, activeTheme *theme.Theme, shadowColor color.NRGBA, opacity, strength float32) flowstyle.Style {
	shadowColor = theme.ColorOr(shadowColor, activeTheme.Palette.SurfaceShadow)
	shadow := render.ThemeShadow(activeTheme.Shadows.Control, shadowColor, opacity)
	if !(strength > 0) || math.IsInf(float64(strength), 0) {
		return declaration
	}
	for _, layer := range shadow.Layers {
		layer.Color.A = uint8(min(float32(layer.Color.A)*strength, 255) + .5)
		declaration = declaration.BoxShadow(
			unit.Dp(layer.OffsetX), unit.Dp(layer.OffsetY), unit.Dp(layer.Blur), unit.Dp(layer.Spread),
			flowstyle.SolidColor{Color: layer.Color},
		)
	}
	return declaration
}
