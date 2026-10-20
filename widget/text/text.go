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
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotSelectableText = "selectable-text"

type selectableTextState struct {
	selectable     widget.Selectable
	pressTag       struct{ _ byte }
	pendingEdit    core.SelectableEditAction
	hasPendingEdit bool
}

type Widget struct {
	text            string
	key             string
	selectable      bool
	defaults        style.Style
	customStyle     style.Style
	size            unit.Sp
	hasSize         bool
	color           color.NRGBA
	hasColor        bool
	selectionColor  color.NRGBA
	hasSelection    bool
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
	resolvedText    *style.TextStyle
	visualCache     *textVisualCache
}

func New(text string) Widget {
	return Widget{text: text, visualCache: &textVisualCache{}}
}

func Selectable(key, text string) Widget {
	return Widget{key: key, text: text, selectable: true, visualCache: &textVisualCache{}}
}

// WithDefaults supplies component-owned text defaults without exposing them
// as user instance properties.
func WithDefaults(value Widget, defaults style.Style) Widget {
	value.defaults = style.Join(value.defaults, defaults)
	return value
}

// WithResolvedText supplies an already-resolved, read-only text declaration.
// It is intended for compound widgets that resolve a shared text part once
// and render many labels with it.
func WithResolvedText(value Widget, resolved *style.TextStyle) Widget {
	value.resolvedText = resolved
	return value
}

// ResolveStatic exposes the resolved declaration to sibling components
// that need to measure composed text.
func ResolveStatic(ctx *core.Context, value Widget) style.ResolvedStyle {
	return value.resolveMeasureStyle(ctx)
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

// SelectionColor overrides the background painted behind a selection on
// Selectable text. It has no effect on plain text, which cannot be selected.
// The default is the active theme's Semantic().Selection.
func (t Widget) SelectionColor(c color.NRGBA) Widget {
	t.selectionColor = c
	t.hasSelection = true
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

// Style replaces the text's instance style with the given declarations joined
// in order.
func (t Widget) Style(values ...style.Style) Widget {
	t.customStyle = style.Join(values...)
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

func (t Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if t.selectable {
		key := core.ClaimKey(ctx, core.KindSelectableText, t.key)
		stateValue := core.UseState[selectableTextState](ctx, key, stateSlotSelectableText)
		visual := t.visualWithCache(ctx)
		resolved := t.resolveStyle(ctx, gtx, key, style.State{Focused: gtx.Focused(&stateValue.selectable)})
		stateValue.takePendingEdit(gtx, &stateValue.selectable)
		field := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
				label := t.labelStyle(ctx, resolved)
				preserveSelectableFocus(ctx, gtx, &stateValue.pressTag)
				label.State = &stateValue.selectable
				semantic.LabelOp(t.text).Add(gtx.Ops)
				dims := label.Layout(gtx)
				underlineLabel(gtx, visual, resolved, dims)
				registerSelectablePress(gtx, dims, &stateValue.pressTag)
				return dims
			}))
		})
		if core.SelectableTextMenu != nil {
			queue := func(action core.SelectableEditAction) {
				stateValue.pendingEdit = action
				stateValue.hasPendingEdit = true
			}
			return core.SelectableTextMenu(ctx, gtx, key, field, &stateValue.selectable, queue)
		}
		return field(ctx, gtx)
	}
	visual := t.visualForResolvedText(ctx)
	resolved := t.resolveLayoutStyle(ctx, gtx)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		dims := t.labelStyleWithVisual(ctx, resolved, visual).Layout(gtx)
		underlineLabel(gtx, visual, resolved, dims)
		return dims
	}))
}

// Measure lays out text without adding paint or input operations to gtx.
func Measure(ctx *core.Context, gtx layout.Context, value Widget) layout.Dimensions {
	measureGtx := gtx
	resolved := value.resolveMeasureStyle(ctx)
	label := value.labelStyleWithVisual(ctx, resolved, value.visualForResolvedText(ctx))
	key := core.TextMeasureKey{
		Text:            label.Text,
		Shaper:          label.Shaper,
		Font:            label.Font,
		TextSize:        label.TextSize,
		Alignment:       label.Alignment,
		MaxLines:        label.MaxLines,
		Truncator:       label.Truncator,
		WrapPolicy:      label.WrapPolicy,
		LineHeight:      label.LineHeight,
		LineHeightScale: label.LineHeightScale,
		Constraints:     measureGtx.Constraints,
		Metric:          measureGtx.Metric,
		Locale:          measureGtx.Locale,
		ThemeGeneration: core.ThemeGeneration(ctx),
	}
	if dimensions, ok := core.LookupTextMeasure(ctx, key); ok {
		return dimensions
	}
	// Only allocate an operation stream when a cache miss requires Gio to lay
	// out and record the label. Cache hits return before touching operations.
	measureGtx.Ops = new(op.Ops)
	dimensions := label.Layout(measureGtx)
	core.StoreTextMeasure(ctx, key, dimensions)
	return dimensions
}

func (t Widget) resolveMeasureStyle(ctx *core.Context) style.ResolvedStyle {
	if t.resolvedText != nil {
		return style.ResolvedStyle{Text: t.resolvedText}
	}
	return t.resolveStyleStatic(ctx, style.State{})
}

// visualForResolvedText avoids a provider lookup for the common
// compound-widget path where the parent already supplied the immutable text
// declaration and no selectable or underline behavior needs visual metadata.
func (t Widget) visualForResolvedText(ctx *core.Context) theme.TextVisual {
	if t.resolvedText == nil || (t.resolvedText.Underline != nil && *t.resolvedText.Underline) {
		return t.visualWithCache(ctx)
	}
	return theme.TextVisual{}
}

// Measure reports this text widget's bounds without registering paint or
// pointer operations.
func (t Widget) Measure(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return Measure(ctx, gtx, t)
}

func (t Widget) resolveLayoutStyle(ctx *core.Context, gtx layout.Context) style.ResolvedStyle {
	if t.resolvedText != nil {
		return style.ResolvedStyle{Text: t.resolvedText}
	}
	resolved := t.resolveStyleStatic(ctx, style.State{})
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	key := core.ClaimKey(ctx, core.KindStyle, "text")
	return computed.ApplyTransitions(ctx, gtx, key, resolved)
}

func preserveSelectableFocus(ctx *core.Context, gtx layout.Context, tag event.Tag) {
	for {
		_, ok := core.NextPointerEvent(gtx, tag, pointer.Press)
		if !ok {
			return
		}
		core.PreserveFocus(ctx)
	}
}

func registerSelectablePress(gtx layout.Context, dims layout.Dimensions, tag event.Tag) {
	area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	pass.Pop()
	area.Pop()
}

func (t Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State) style.ResolvedStyle {
	visual := t.visualWithCache(ctx)
	return computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		visual.Defaults,
		t.defaults,
		textPropertyDeclaration(t),
		t.customStyle,
	)
}

func (t Widget) resolveStyleStatic(ctx *core.Context, state style.State) style.ResolvedStyle {
	visual := t.visualWithCache(ctx)
	return computed.ResolveStatic(
		ctx,
		state,
		visual.Defaults,
		t.defaults,
		textPropertyDeclaration(t),
		t.customStyle,
	)
}

func textPropertyDeclaration(value Widget) style.Style {
	builder := style.Style{}
	if value.hasSize {
		builder = builder.FontSize(value.size)
	}
	if value.hasColor {
		builder = builder.TextColor(style.Solid(value.color))
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

func (t Widget) labelStyle(ctx *core.Context, resolved style.ResolvedStyle) material.LabelStyle {
	return t.labelStyleWithVisual(ctx, resolved, t.visualWithCache(ctx))
}

func (t Widget) labelStyleWithVisual(ctx *core.Context, resolved style.ResolvedStyle, visual theme.TextVisual) material.LabelStyle {
	size := unit.Sp(0)
	if resolved.Text != nil && resolved.Text.FontSize != nil {
		size = *resolved.Text.FontSize
	}
	label := material.Label(core.ActiveMaterial(ctx), size, t.text)
	label.SelectionColor = visual.Selection
	if t.hasSelection {
		label.SelectionColor = t.selectionColor
	}
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

// underlineLabel requests the provider-owned rule a text declaration asks for.
//
// Gio's label API reports a single baseline and no per-line bounds, so the
// rule lands under that baseline: exact for single-line text, and the last
// line's rule for wrapped text.
func underlineLabel(gtx layout.Context, visual theme.TextVisual, resolved style.ResolvedStyle, dims layout.Dimensions) {
	if resolved.Text == nil || resolved.Text.Underline == nil || !*resolved.Text.Underline {
		return
	}
	col, ok := styleColor(resolved.Text.Color)
	if !ok || visual.Underline == nil {
		return
	}
	visual.Underline(gtx, theme.TextUnderlineRequest{Size: dims.Size, Baseline: dims.Baseline, Color: col})
}

func styleTextAlignment(value giotext.Alignment) style.TextAlign {
	switch value {
	case giotext.Middle:
		return style.TextAlignCenter
	case giotext.End:
		return style.TextAlignEnd
	default:
		return style.TextAlignStart
	}
}

func gioTextAlignment(value style.TextAlign) giotext.Alignment {
	switch value {
	case style.TextAlignCenter:
		return giotext.Middle
	case style.TextAlignEnd:
		return giotext.End
	default:
		return giotext.Start
	}
}

func styleColor(source style.ColorSource) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		return value.Color, true
	default:
		return color.NRGBA{}, false
	}
}
