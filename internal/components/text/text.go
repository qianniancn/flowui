package text

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/interact"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

const stateSlotSelectableText = "selectable-text"

type selectableTextState struct {
	selectable widget.Selectable
	pressTag   struct{ _ byte }
}

type Widget struct {
	text            string
	key             string
	selectable      bool
	defaults        flowstyle.Style
	customStyle     flowstyle.Style
	size            unit.Sp
	hasSize         bool
	color           color.NRGBA
	hasColor        bool
	font            font.Font
	hasTypeface     bool
	hasStyle        bool
	hasWeight       bool
	alignment       giotext.Alignment
	hasAlignment    bool
	maxLines        int
	hasMaxLines     bool
	truncator       string
	hasTruncator    bool
	wrapPolicy      giotext.WrapPolicy
	hasWrapPolicy   bool
	lineHeight      unit.Sp
	hasLineHeight   bool
	lineHeightScale float32
	hasHeightScale  bool
}

func New(text string) Widget {
	return Widget{text: text}
}

func Selectable(key, text string) Widget {
	return Widget{key: key, text: text, selectable: true}
}

// WithDefaults supplies component-owned text defaults without exposing them
// as user instance properties.
func WithDefaults(value Widget, defaults flowstyle.Style) Widget {
	value.defaults = flowstyle.Join(value.defaults, defaults)
	return value
}

// ResolveStyleStatic exposes the resolved declaration to sibling components
// that need to measure composed text.
func ResolveStyleStatic(ctx *frame.Context, value Widget) flowstyle.ResolvedStyle {
	return value.resolveStyleStatic(ctx, flowstyle.StyleState{})
}

func (t Widget) Size(sp float32) Widget {
	t.size = unit.Sp(sp)
	t.hasSize = true
	return t
}

func (t Widget) Color(c color.NRGBA) Widget {
	t.color = c
	t.hasColor = true
	return t
}

func (t Widget) Weight(w font.Weight) Widget {
	t.font.Weight = w
	t.hasWeight = true
	return t
}

func (t Widget) FontStyle(style font.Style) Widget {
	t.font.Style = style
	t.hasStyle = true
	return t
}

func (t Widget) Style(value flowstyle.Style) Widget {
	t.customStyle = value
	return t
}

func (t Widget) Typeface(typeface font.Typeface) Widget {
	t.font.Typeface = typeface
	t.hasTypeface = true
	return t
}

func (t Widget) Font(value font.Font) Widget {
	t.font = value
	t.hasTypeface = true
	t.hasStyle = true
	t.hasWeight = true
	return t
}

func (t Widget) Align(alignment giotext.Alignment) Widget {
	t.alignment = alignment
	t.hasAlignment = true
	return t
}

func (t Widget) MaxLines(lines int) Widget {
	t.maxLines = max(lines, 0)
	t.hasMaxLines = true
	return t
}

func (t Widget) Truncator(value string) Widget {
	t.truncator = value
	t.hasTruncator = true
	return t
}

func (t Widget) Wrap(policy giotext.WrapPolicy) Widget {
	t.wrapPolicy = policy
	t.hasWrapPolicy = true
	return t
}

func (t Widget) LineHeight(sp float32) Widget {
	t.lineHeight = unit.Sp(max(sp, 0))
	t.hasLineHeight = true
	return t
}

func (t Widget) LineHeightScale(scale float32) Widget {
	t.lineHeightScale = max(scale, 0)
	t.hasHeightScale = true
	return t
}

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if t.selectable {
		key := frame.ClaimKey(ctx, state.KindSelectableText, t.key)
		stateValue := frame.UseState[selectableTextState](ctx, key, stateSlotSelectableText)
		resolved := t.resolveStyle(ctx, gtx, key, flowstyle.StyleState{Focused: gtx.Focused(&stateValue.selectable)})
		return layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
			label := t.labelStyle(ctx, resolved)
			preserveSelectableFocus(ctx, gtx, &stateValue.pressTag)
			label.State = &stateValue.selectable
			semantic.LabelOp(t.text).Add(gtx.Ops)
			dims := label.Layout(gtx)
			registerSelectablePress(gtx, dims, &stateValue.pressTag)
			return dims
		}))
	}
	resolved := t.resolveLayoutStyle(ctx, gtx)
	return layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return t.labelStyle(ctx, resolved).Layout(gtx)
	}))
}

// Measure lays out text without adding paint or input operations to gtx.
func Measure(ctx *frame.Context, gtx layout.Context, value Widget) layout.Dimensions {
	measureGtx := gtx
	measureGtx.Ops = new(op.Ops)
	resolved := value.resolveStyleStatic(ctx, flowstyle.StyleState{})
	return value.labelStyle(ctx, resolved).Layout(measureGtx)
}

func (t Widget) resolveLayoutStyle(ctx *frame.Context, gtx layout.Context) flowstyle.ResolvedStyle {
	resolved := t.resolveStyleStatic(ctx, flowstyle.StyleState{})
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	key := frame.ClaimKey(ctx, state.KindStyle, "text")
	return styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
}

func preserveSelectableFocus(ctx *frame.Context, gtx layout.Context, tag event.Tag) {
	for {
		_, ok := interact.NextPointerEvent(gtx, tag, pointer.Press)
		if !ok {
			return
		}
		frame.PreserveFocus(ctx)
	}
}

func registerSelectablePress(gtx layout.Context, dims layout.Dimensions, tag event.Tag) {
	area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	pass.Pop()
	area.Pop()
}

func (t Widget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) flowstyle.ResolvedStyle {
	return styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		textDefaultDeclaration(ctx),
		t.defaults,
		textPropertyDeclaration(t),
		t.customStyle,
	)
}

func (t Widget) resolveStyleStatic(ctx *frame.Context, state flowstyle.StyleState) flowstyle.ResolvedStyle {
	return styleruntime.ResolveStatic(
		ctx,
		state,
		textDefaultDeclaration(ctx),
		t.defaults,
		textPropertyDeclaration(t),
		t.customStyle,
	)
}

func textDefaultDeclaration(ctx *frame.Context) flowstyle.Style {
	return flowstyle.Style{}.
		FontSize(frame.ActiveTheme(ctx).Typography.BodySize).
		TextColor(flowstyle.SolidColor{Color: ctx.ForegroundColor()})

}

func textPropertyDeclaration(value Widget) flowstyle.Style {
	builder := flowstyle.Style{}
	if value.hasSize {
		builder = builder.FontSize(value.size)
	}
	if value.hasColor {
		builder = builder.TextColor(flowstyle.SolidColor{Color: value.color})
	}
	if value.hasWeight {
		builder = builder.FontWeight(int(value.font.Weight))
	}
	if value.hasTypeface {
		builder = builder.Typeface(value.font.Typeface)
	}
	if value.hasStyle {
		builder = builder.FontStyle(value.font.Style)
	}
	if value.hasAlignment {
		builder = builder.TextAlign(styleTextAlignment(value.alignment))
	}
	if value.hasMaxLines {
		builder = builder.MaxLines(value.maxLines)
	}
	if value.hasLineHeight {
		builder = builder.LineHeight(value.lineHeight)
	}
	if value.hasHeightScale {
		builder = builder.LineHeightScale(value.lineHeightScale)
	}
	if value.hasWrapPolicy {
		builder = builder.Wrap(value.wrapPolicy)
	}
	if value.hasTruncator {
		builder = builder.Truncator(value.truncator)
	}
	return builder
}

func (t Widget) labelStyle(ctx *frame.Context, resolved flowstyle.ResolvedStyle) material.LabelStyle {
	size := frame.ActiveTheme(ctx).Typography.BodySize
	if resolved.Text != nil && resolved.Text.FontSize != nil {
		size = *resolved.Text.FontSize
	}
	label := material.Label(frame.ActiveMaterial(ctx), size, t.text)
	label.SelectionColor = frame.ActiveTheme(ctx).Palette.Selection
	if resolved.Text != nil {
		if color, ok := styleColor(resolved.Text.Color); ok {
			label.Color = color
		}
		if resolved.Text.FontWeight != nil {
			label.Font.Weight = font.Weight(*resolved.Text.FontWeight)
		}
		if resolved.Text.Typeface != nil {
			label.Font.Typeface = *resolved.Text.Typeface
		}
		if resolved.Text.FontStyle != nil {
			label.Font.Style = *resolved.Text.FontStyle
		}
		if resolved.Text.Align != nil {
			label.Alignment = gioTextAlignment(*resolved.Text.Align)
		}
		if resolved.Text.MaxLines != nil {
			label.MaxLines = *resolved.Text.MaxLines
		}
		if resolved.Text.LineHeight != nil {
			label.LineHeight = *resolved.Text.LineHeight
		}
		if resolved.Text.LineHeightScale != nil {
			label.LineHeightScale = *resolved.Text.LineHeightScale
		}
		if resolved.Text.Wrap != nil {
			label.WrapPolicy = *resolved.Text.Wrap
		}
		if resolved.Text.Truncator != nil {
			label.Truncator = *resolved.Text.Truncator
		}
	}
	return label
}

func styleTextAlignment(value giotext.Alignment) flowstyle.TextAlign {
	switch value {
	case giotext.Middle:
		return flowstyle.TextAlignCenter
	case giotext.End:
		return flowstyle.TextAlignEnd
	default:
		return flowstyle.TextAlignStart
	}
}

func gioTextAlignment(value flowstyle.TextAlign) giotext.Alignment {
	switch value {
	case flowstyle.TextAlignCenter:
		return giotext.Middle
	case flowstyle.TextAlignEnd:
		return giotext.End
	default:
		return giotext.Start
	}
}

func styleColor(source flowstyle.ColorSource) (color.NRGBA, bool) {
	switch value := source.(type) {
	case flowstyle.SolidColor:
		return value.Color, true
	case *flowstyle.SolidColor:
		return value.Color, true
	default:
		return color.NRGBA{}, false
	}
}
