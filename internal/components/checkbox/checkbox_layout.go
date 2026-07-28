package checkbox

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/components/description"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func (c CheckboxWidget) layoutContent(ctx *frame.Context, gtx layout.Context, options ControlOptions, indicatorState IndicatorState, style checkboxResolvedStyle) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return c.layoutRow(ctx, gtx, options, indicatorState, style.label)
	}
	message := c.supportingText()
	if message == "" {
		return row(gtx)
	}
	tokens := frame.ActiveTheme(ctx).Components.Checkbox
	return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(tokens.DescriptionGap)}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: tokens.DescriptionIndent}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layoutui.LayoutResolved(ctx, gtx, style.description, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
					return description.Description(message).
						For(c.key).
						Disabled(options.Disabled).
						Style(flowstyle.TextDeclaration(style.description.Text)).
						Layout(ctx, gtx)
				}))
			})
		}),
	)
}

func (c CheckboxWidget) layoutRow(ctx *frame.Context, gtx layout.Context, options ControlOptions, indicatorState IndicatorState, labelStyle flowstyle.ResolvedStyle) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return DrawControl(ctx, gtx, options)
	}
	if c.label == "" {
		return control(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(frame.ActiveTheme(ctx).Components.Checkbox.LabelGap),
	}.Layout(gtx,
		layout.Rigid(control),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, labelStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
				children := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return text.New(c.label).Layout(ctx, gtx)
				})}
				if c.required {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						color := checkboxSupportingColor(frame.ActiveTheme(ctx), frame.ActiveTheme(ctx).Palette.Danger, indicatorState.Disabled)
						return text.New("*").Color(color).Layout(ctx, gtx)
					}))
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline, Gap: gtx.Dp(frame.ActiveTheme(ctx).Components.Label.RequiredMarkOffset)}.Layout(gtx, children...)
			}))
		}),
	)
}
