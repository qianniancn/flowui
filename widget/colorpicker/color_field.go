package colorpicker

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotColorField = "color-field"

type ColorFieldWidget struct {
	override     *theme.ColorPickerOverride
	key          string
	value        color.NRGBA
	label        string
	description  string
	errorMessage string
	onChange     func(color.NRGBA)
	variant      input.Variant
	disabled     bool
	invalid      bool
	required     bool
	fullWidth    bool
	alpha        bool
	swatch       bool
}

type colorFieldState struct {
	text        string
	syncedColor color.NRGBA
	syncedAlpha bool
	ready       bool
	valid       bool
	visualCache colorPickerVisualCache
}

func ColorField(key string, value color.NRGBA) ColorFieldWidget {
	return ColorFieldWidget{key: key, value: value}
}

func (field ColorFieldWidget) Label(label string) ColorFieldWidget {
	field.label = label
	return field
}

func (field ColorFieldWidget) Description(description string) ColorFieldWidget {
	field.description = description
	return field
}

func (field ColorFieldWidget) ErrorMessage(message string) ColorFieldWidget {
	field.errorMessage = message
	return field
}

func (field ColorFieldWidget) OnChange(fn func(color.NRGBA)) ColorFieldWidget {
	field.onChange = fn
	return field
}

func (field ColorFieldWidget) Variant(variant input.Variant) ColorFieldWidget {
	field.variant = variant
	return field
}

func (field ColorFieldWidget) Disabled(disabled bool) ColorFieldWidget {
	field.disabled = disabled
	return field
}

func (field ColorFieldWidget) Invalid(invalid bool) ColorFieldWidget {
	field.invalid = invalid
	return field
}

func (field ColorFieldWidget) Required(required bool) ColorFieldWidget {
	field.required = required
	return field
}

func (field ColorFieldWidget) FullWidth() ColorFieldWidget {
	field.fullWidth = true
	return field
}

func (field ColorFieldWidget) Alpha(enabled bool) ColorFieldWidget {
	field.alpha = enabled
	return field
}

func (field ColorFieldWidget) Swatch(show bool) ColorFieldWidget {
	field.swatch = show
	return field
}

func (field ColorFieldWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return field.layoutWithOverride(ctx, gtx, field.override)
}

func (field ColorFieldWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ColorPickerOverride) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorField, field.key)
	fieldState := core.UseState[colorFieldState](ctx, key, stateSlotColorField)
	visual := fieldState.visualCache.resolve(core.ActiveTheme(ctx), colorPickerRequest(override))
	fieldState.sync(field.value, field.alpha)

	restore := core.PushKey(ctx, field.key)
	defer restore()
	inputKey := "input"
	invalid := field.invalid || !fieldState.valid
	control := input.New(fieldState.text).Key(inputKey).
		Placeholder("#000000").
		Label(field.resolvedLabel(ctx)).
		Variant(field.variant).
		Disabled(field.disabled).
		Invalid(invalid).
		MaxLength(func() int {
			if field.alpha {
				return 9
			}
			return 7
		}()).
		OnChange(func(value string) {
			fieldState.text = value
			parsed, ok := parseHexColor(value, field.value.A)
			fieldState.valid = ok
			if ok && field.onChange != nil && parsed != field.value {
				field.onChange(parsed)
			}
		})
	group := input.Group(control).
		Variant(field.variant).
		Disabled(field.disabled).
		Invalid(invalid)
	var groupLayout core.Widget = group
	if field.fullWidth {
		group = group.FullWidth()
		groupLayout = group
	}
	if field.swatch {
		group = group.Prefix(ColorSwatch(field.value).Size(ColorSwatchExtraSmall))
		groupLayout = (group).Style(style.Style{}.
			Part(style.PartPrefix, style.Style{}.PaddingLeft(10).PaddingRight(8)))
	}

	children := make([]layout.FlexChild, 0, 3)
	if field.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label.New(field.label).
				For(inputKey).
				Required(field.required).
				Disabled(field.disabled).
				Invalid(invalid).
				Layout(ctx, gtx)
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return groupLayout.Layout(ctx, gtx)
	}))
	if invalid && field.errorMessage != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(field.errorMessage).
				Size(float32(visual.Field.ErrorTextSize)).
				Color(visual.Field.ErrorForeground).
				Layout(ctx, gtx)
		}))
	} else if field.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return description.New(field.description).
				For(inputKey).
				Disabled(field.disabled).
				Layout(ctx, gtx)
		}))
	}
	return layoutui.Styled(ctx, gtx, key, style.State{
		Disabled: field.disabled,
		Invalid:  invalid,
	}, visual.Override, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
			Gap:  gtx.Dp(visual.Field.Gap),
		}.Layout(gtx, children...)
	}))
}

func (state *colorFieldState) sync(value color.NRGBA, alpha bool) {
	if state.ready && state.syncedColor == value && state.syncedAlpha == alpha {
		return
	}
	state.text = formatHexColor(value, alpha)
	state.syncedColor = value
	state.syncedAlpha = alpha
	state.ready = true
	state.valid = true
}

func (field ColorFieldWidget) resolvedLabel(ctx *core.Context) string {
	if field.label != "" {
		return field.label
	}
	return ctx.Text(i18n.ColorPickerColor)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ColorFieldWidget) Style(values ...style.Style) ColorFieldWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ColorPickerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
