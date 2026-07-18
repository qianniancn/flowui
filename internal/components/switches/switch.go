package switches

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

type SwitchWidget struct {
	key         string
	checked     bool
	label       string
	description string
	onChange    func(bool)
	disabled    bool
	invalid     bool
	size        SwitchSize
	labelBefore bool
	thumb       func(bool) frame.Widget
}

type SwitchSize int

const (
	SwitchMedium SwitchSize = iota
	SwitchSmall
	SwitchLarge
)

const (
	switchSelectDuration = 300 * time.Millisecond
	switchFocusDuration  = 100 * time.Millisecond
)

func Switch(key string, checked bool, label string) SwitchWidget {
	return SwitchWidget{
		key:     key,
		checked: checked,
		label:   label,
	}
}

func (s SwitchWidget) OnChange(fn func(bool)) SwitchWidget {
	s.onChange = fn
	return s
}

func (s SwitchWidget) Description(description string) SwitchWidget {
	s.description = description
	return s
}

func (s SwitchWidget) Disabled(disabled bool) SwitchWidget {
	s.disabled = disabled
	return s
}

func (s SwitchWidget) Invalid(invalid bool) SwitchWidget {
	s.invalid = invalid
	return s
}

func (s SwitchWidget) Size(size SwitchSize) SwitchWidget {
	s.size = size
	return s
}

func (s SwitchWidget) LabelBefore() SwitchWidget {
	s.labelBefore = true
	return s
}

func (s SwitchWidget) Thumb(content func(checked bool) frame.Widget) SwitchWidget {
	s.thumb = content
	return s
}

func (s SwitchWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	componentState := switchStateFor(ctx, s.key)
	key := frame.FullKey(ctx, s.key)
	if s.description != "" {
		frame.PrepareFieldDescription(ctx, key, s.description)
	}
	componentState.value.Value = s.checked
	animGtx := gtx
	disabled := s.disabled || !gtx.Enabled()
	presses := state.ActivePresses(componentState.value.History())
	if disabled {
		gtx = gtx.Disabled()
	}
	dims := componentState.value.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Switch.Add(gtx.Ops)
		if s.label != "" {
			semantic.LabelOp(s.label).Add(gtx.Ops)
		}
		description := s.description
		if description == "" {
			description = frame.FieldDescription(ctx, key)
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}

		if !disabled {
			frame.FocusOnPress(ctx, &componentState.value, componentState.value.History(), presses)
		}
		focusVisible := frame.FocusVisible(ctx, &componentState.value, gtx.Focused(&componentState.value))
		style := switchStyleFor(frame.ActiveTheme(ctx), componentState.value.Hovered(), componentState.value.Pressed(), disabled, s.invalid)
		motion := frame.ActiveTheme(ctx).Motion
		style.selected = componentState.selection(animGtx, componentState.value.Value, motion)
		style.focus = componentState.focusOpacity(animGtx, focusVisible && !disabled, motion)
		return s.layoutContent(ctx, gtx, style, switchSizeStyleFor(frame.ActiveTheme(ctx), s.size), componentState.value.Value)
	})
	if !disabled && componentState.value.Value != s.checked && s.onChange != nil {
		s.onChange(componentState.value.Value)
	}
	return dims
}

type SwitchGroupWidget struct {
	children   []frame.Widget
	horizontal bool
}

func SwitchGroup(children ...frame.Widget) SwitchGroupWidget {
	return SwitchGroupWidget{children: children}
}

func (g SwitchGroupWidget) Horizontal() SwitchGroupWidget {
	g.horizontal = true
	return g
}

func (g SwitchGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.SwitchGroup
	children := make([]layout.Widget, 0, len(g.children))
	for _, child := range g.children {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		})
	}
	return layoutui.LayoutItems(ctx, gtx, g.horizontal, gtx.Dp(theme.HorizontalGap), gtx.Dp(theme.VerticalGap), children)
}
