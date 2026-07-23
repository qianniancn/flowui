package colorpicker

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/input"
	"github.com/qianniancn/FlowUI/internal/components/label"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

const stateSlotColorField = "color-field"

type ColorFieldWidget struct {
	key          string
	value        color.NRGBA
	label        string
	description  string
	errorMessage string
	onChange     func(color.NRGBA)
	variant      input.InputVariant
	disabled     bool
	invalid      bool
	required     bool
	fullWidth    bool
	alpha        bool
	swatch       bool
	customStyle  flowstyle.Style
}

type colorFieldState struct {
	text        string
	syncedColor color.NRGBA
	syncedAlpha bool
	ready       bool
	valid       bool
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

func (field ColorFieldWidget) Variant(variant input.InputVariant) ColorFieldWidget {
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

func (field ColorFieldWidget) Style(value flowstyle.Style) ColorFieldWidget {
	field.customStyle = value
	return field
}

func (field ColorFieldWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, state.KindColorField, field.key)
	fieldState := frame.UseState[colorFieldState](ctx, key, stateSlotColorField)
	fieldState.sync(field.value, field.alpha)

	restore := frame.PushKey(ctx, field.key)
	defer restore()
	inputKey := "input"
	invalid := field.invalid || !fieldState.valid
	control := input.Input(inputKey, fieldState.text).
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
	group := input.InputGroup(control).
		Variant(field.variant).
		Disabled(field.disabled).
		Invalid(invalid)
	if field.fullWidth {
		group = group.FullWidth()
	}
	if field.swatch {
		group = group.
			Prefix(ColorSwatch(field.value).Size(ColorSwatchExtraSmall)).
			PrefixPadding(10, 8)
	}

	children := make([]layout.FlexChild, 0, 3)
	if field.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label.Label(field.label).
				For(inputKey).
				Required(field.required).
				Disabled(field.disabled).
				Invalid(invalid).
				Layout(ctx, gtx)
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return group.Layout(ctx, gtx)
	}))
	if invalid && field.errorMessage != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(field.errorMessage).
				Size(float32(frame.ActiveTheme(ctx).Components.Description.TextSize)).
				Color(frame.ActiveTheme(ctx).Palette.Danger).
				Layout(ctx, gtx)
		}))
	} else if field.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return description.Description(field.description).
				For(inputKey).
				Disabled(field.disabled).
				Layout(ctx, gtx)
		}))
	}
	return layoutui.LayoutStyled(ctx, gtx, key, flowstyle.StyleState{
		Disabled: field.disabled,
		Invalid:  invalid,
	}, field.customStyle, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
			Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.ColorField.Gap),
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

func (field ColorFieldWidget) resolvedLabel(ctx *frame.Context) string {
	if field.label != "" {
		return field.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "颜色"
	}
	return "Color"
}
