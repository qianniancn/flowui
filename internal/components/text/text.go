package text

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
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
	size            unit.Sp
	color           color.NRGBA
	hasColor        bool
	font            font.Font
	hasTypeface     bool
	hasStyle        bool
	hasWeight       bool
	alignment       giotext.Alignment
	maxLines        int
	truncator       string
	wrapPolicy      giotext.WrapPolicy
	lineHeight      unit.Sp
	lineHeightScale float32
}

func New(text string) Widget {
	return Widget{text: text}
}

func Selectable(key, text string) Widget {
	return Widget{key: key, text: text, selectable: true}
}

func (t Widget) Size(sp float32) Widget {
	t.size = unit.Sp(sp)
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

func (t Widget) Style(style font.Style) Widget {
	t.font.Style = style
	t.hasStyle = true
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
	return t
}

func (t Widget) MaxLines(lines int) Widget {
	t.maxLines = max(lines, 0)
	return t
}

func (t Widget) Truncator(value string) Widget {
	t.truncator = value
	return t
}

func (t Widget) Wrap(policy giotext.WrapPolicy) Widget {
	t.wrapPolicy = policy
	return t
}

func (t Widget) LineHeight(sp float32) Widget {
	t.lineHeight = unit.Sp(max(sp, 0))
	return t
}

func (t Widget) LineHeightScale(scale float32) Widget {
	t.lineHeightScale = max(scale, 0)
	return t
}

func (t Widget) DefaultSize(sp float32) Widget {
	if t.size == 0 {
		t.size = unit.Sp(sp)
	}
	return t
}

func (t Widget) DefaultColor(c color.NRGBA) Widget {
	if !t.hasColor {
		t.color = c
		t.hasColor = true
	}
	return t
}

func (t Widget) DefaultWeight(w font.Weight) Widget {
	if !t.hasWeight {
		t.font.Weight = w
		t.hasWeight = true
	}
	return t
}

func (t Widget) ConfiguredSize() unit.Sp {
	return t.size
}

func (t Widget) ConfiguredColor() (color.NRGBA, bool) {
	return t.color, t.hasColor
}

func (t Widget) ConfiguredWeight() font.Weight {
	return t.font.Weight
}

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	label := t.labelStyle(ctx)
	if t.selectable {
		key := frame.ClaimKey(ctx, state.KindSelectableText, t.key)
		stateValue := frame.UseState[selectableTextState](ctx, key, stateSlotSelectableText)
		preserveSelectableFocus(ctx, gtx, &stateValue.pressTag)
		label.State = &stateValue.selectable
		semantic.LabelOp(t.text).Add(gtx.Ops)
		dims := label.Layout(gtx)
		registerSelectablePress(gtx, dims, &stateValue.pressTag)
		return dims
	}
	return label.Layout(gtx)
}

func preserveSelectableFocus(ctx *frame.Context, gtx layout.Context, tag event.Tag) {
	for {
		value, ok := gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Press})
		if !ok {
			return
		}
		if _, ok := value.(pointer.Event); ok {
			frame.PreserveFocus(ctx)
		}
	}
}

func registerSelectablePress(gtx layout.Context, dims layout.Dimensions, tag event.Tag) {
	area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	pass.Pop()
	area.Pop()
}

func (t Widget) labelStyle(ctx *frame.Context) material.LabelStyle {
	size := t.size
	if size == 0 {
		size = frame.ActiveTheme(ctx).Typography.BodySize
	}
	label := material.Label(frame.ActiveTheme(ctx).Material, size, t.text)
	label.Alignment = t.alignment
	label.MaxLines = t.maxLines
	label.Truncator = t.truncator
	label.WrapPolicy = t.wrapPolicy
	label.LineHeight = t.lineHeight
	label.LineHeightScale = t.lineHeightScale
	label.SelectionColor = frame.ActiveTheme(ctx).Palette.Selection
	if t.hasColor {
		label.Color = t.color
	} else {
		label.Color = ctx.ForegroundColor()
	}
	if t.hasTypeface {
		label.Font.Typeface = t.font.Typeface
	}
	if t.hasStyle {
		label.Font.Style = t.font.Style
	}
	if t.hasWeight {
		label.Font.Weight = t.font.Weight
	}
	return label
}
