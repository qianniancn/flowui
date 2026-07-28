package switches

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type SwitchWidget struct {
	key            string
	checked        bool
	hasChecked     bool
	defaultChecked bool
	hasDefault     bool
	label          string
	description    string
	onChange       func(bool)
	disabled       bool
	invalid        bool
	size           SwitchSize
	labelBefore    bool
	thumb          func(bool) frame.Widget
	customStyle    flowstyle.Style
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
		key:        key,
		checked:    checked,
		hasChecked: true,
		label:      label,
	}
}

func (s SwitchWidget) OnChange(fn func(bool)) SwitchWidget {
	s.onChange = fn
	return s
}

func (s SwitchWidget) Checked(checked bool) SwitchWidget {
	s.checked = checked
	s.hasChecked = true
	return s
}

func (s SwitchWidget) DefaultChecked(checked bool) SwitchWidget {
	s.defaultChecked = checked
	s.hasDefault = true
	s.hasChecked = false
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

func (s SwitchWidget) Style(value flowstyle.Style) SwitchWidget {
	s.customStyle = value
	return s
}

func (s SwitchWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	componentState := switchStateFor(ctx, s.key)
	key := frame.FullKey(ctx, s.key)
	if s.description != "" {
		frame.PrepareFieldDescription(ctx, key, s.description)
	}

	// Bind disclosure state
	componentState.bind(s)
	checked := componentState.currentChecked(s)

	componentState.value.Value = checked
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
		styleState := flowstyle.StyleState{
			Hovered:      componentState.value.Hovered(),
			Pressed:      componentState.value.Pressed(),
			Focused:      gtx.Focused(&componentState.value),
			FocusVisible: focusVisible,
			Disabled:     disabled,
			Selected:     componentState.value.Value,
			Checked:      componentState.value.Value,
			Invalid:      s.invalid,
		}
		style, size := s.resolveStyle(ctx, animGtx, key, styleState)
		motion := frame.ActiveTheme(ctx).Motion
		style.selected = componentState.selection(animGtx, componentState.value.Value, motion)
		style.focus = componentState.focusOpacity(animGtx, focusVisible && !disabled, motion)
		return layoutui.LayoutStyled(ctx, gtx, key, styleState, s.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return s.layoutContent(ctx, gtx, style, size, componentState.value.Value)
		}))
	})
	if !disabled && componentState.value.Value != checked {
		componentState.requestChecked(s, componentState.value.Value)
	}
	return dims
}

type SwitchGroupWidget struct {
	children    []frame.Widget
	horizontal  bool
	customStyle flowstyle.Style
}

func SwitchGroup(children ...frame.Widget) SwitchGroupWidget {
	return SwitchGroupWidget{children: children}
}

func (g SwitchGroupWidget) Horizontal() SwitchGroupWidget {
	g.horizontal = true
	return g
}

func (g SwitchGroupWidget) Style(value flowstyle.Style) SwitchGroupWidget {
	g.customStyle = value
	return g
}

func (g SwitchGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutui.Box(frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		theme := frame.ActiveTheme(ctx).Components.SwitchGroup
		children := make([]layout.Widget, 0, len(g.children))
		for _, child := range g.children {
			children = append(children, func(gtx layout.Context) layout.Dimensions {
				return child.Layout(ctx, gtx)
			})
		}
		return layoutui.LayoutItems(ctx, gtx, g.horizontal, gtx.Dp(theme.HorizontalGap), gtx.Dp(theme.VerticalGap), children)
	})).Style(g.customStyle).Layout(ctx, gtx)
}
