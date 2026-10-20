package switchcontrol

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	override       *theme.SwitchOverride
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
	thumb          func(bool) core.Widget
}

type SwitchSize = theme.SwitchSize

const (
	SwitchMedium = theme.SwitchMedium
	SwitchSmall  = theme.SwitchSmall
	SwitchLarge  = theme.SwitchLarge
)

// New creates an uncontrolled switch that starts off. Use Checked for controlled
// state and DefaultChecked for an uncontrolled starting value.
func New(label string) Widget {
	return Widget{
		key:   core.AutoKey(),
		label: label,
	}
}

// Key sets an explicit identity for the switch state.
func (s Widget) Key(key string) Widget {
	s.key = key
	return s
}

func (s Widget) OnChange(fn func(bool)) Widget {
	s.onChange = fn
	return s
}

func (s Widget) Checked(checked bool) Widget {
	s.checked = checked
	s.hasChecked = true
	return s
}

func (s Widget) DefaultChecked(checked bool) Widget {
	s.defaultChecked = checked
	s.hasDefault = true
	s.hasChecked = false
	return s
}

func (s Widget) Description(description string) Widget {
	s.description = description
	return s
}

func (s Widget) Disabled(disabled bool) Widget {
	s.disabled = disabled
	return s
}

func (s Widget) Invalid(invalid bool) Widget {
	s.invalid = invalid
	return s
}

func (s Widget) Size(size SwitchSize) Widget {
	s.size = size
	return s
}

func (s Widget) LabelBefore() Widget {
	s.labelBefore = true
	return s
}

func (s Widget) Thumb(content func(checked bool) core.Widget) Widget {
	s.thumb = content
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return s.layoutWithOverride(ctx, gtx, s.override)
}

func (s Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.SwitchOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	componentState := switchStateFor(ctx, s.key)
	key := core.FullKey(ctx, s.key)
	if s.description != "" {
		core.PrepareFieldDescription(ctx, key, s.description)
	}

	// Bind disclosure state
	componentState.bind(s)
	checked := componentState.currentChecked(s)

	componentState.value.Value = checked
	animGtx := gtx
	disabled := s.disabled || !gtx.Enabled()
	presses := core.ActivePresses(componentState.value.History())
	if disabled {
		gtx = gtx.Disabled()
	}
	dims := componentState.value.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Switch.Add(gtx.Ops)
		semantic.SelectedOp(checked).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		if s.label != "" {
			semantic.LabelOp(s.label).Add(gtx.Ops)
		}
		description := s.description
		if description == "" {
			description = core.FieldDescription(ctx, key)
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}

		if !disabled {
			core.FocusOnPress(ctx, &componentState.value, componentState.value.History(), presses)
		}
		focusVisible := core.FocusVisible(ctx, &componentState.value, gtx.Focused(&componentState.value))
		styleState := style.State{
			Hovered:      componentState.value.Hovered(),
			Pressed:      componentState.value.Pressed(),
			Focused:      gtx.Focused(&componentState.value),
			FocusVisible: focusVisible,
			Disabled:     disabled,
			Selected:     componentState.value.Value,
			Checked:      componentState.value.Value,
			Invalid:      s.invalid,
		}
		resolved := s.resolveStyle(ctx, animGtx, key, styleState, override, &componentState.styleCache)
		motion := resolved.visual.Motion
		resolved.selected = componentState.selection(animGtx, componentState.value.Value, resolved.visual.SelectionDuration, motion)
		resolved.focus = componentState.focusOpacity(animGtx, focusVisible && !disabled, resolved.visual.FocusDuration, motion)
		return layoutui.Styled(ctx, gtx, key, styleState, resolved.visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return s.layoutContent(ctx, gtx, resolved, componentState.value.Value)
		}))
	})
	if !disabled && componentState.value.Value != checked {
		componentState.requestChecked(s, componentState.value.Value)
	}
	return dims
}

type GroupWidget struct {
	override   *theme.SwitchOverride
	children   []core.Widget
	horizontal bool
}

func SwitchGroup(children ...core.Widget) GroupWidget {
	return GroupWidget{children: children}
}

func (g GroupWidget) Horizontal() GroupWidget {
	g.horizontal = true
	return g
}

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return g.layoutWithOverride(ctx, gtx, g.override)
}

func (g GroupWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.SwitchOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	activeTheme := core.ActiveTheme(ctx)
	visual := theme.ResolveVisual(activeTheme, theme.SwitchGroupRequest{Override: override},
		theme.ComponentVisualProvider.SwitchGroup, theme.ValidateSwitchGroupVisual)
	return layoutui.Box(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		children := make([]layout.Widget, 0, len(g.children))
		for _, child := range g.children {
			children = append(children, func(gtx layout.Context) layout.Dimensions {
				return child.Layout(ctx, gtx)
			})
		}
		return layoutui.Items(ctx, gtx, g.horizontal, gtx.Dp(visual.Metrics.HorizontalGap), gtx.Dp(visual.Metrics.VerticalGap), children)
	})).Style(visual.Override).Layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.SwitchOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w GroupWidget) Style(values ...style.Style) GroupWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.SwitchOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
