package checkbox

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotCheckbox = "checkbox"

func checkboxStateFor(ctx *frame.Context, key string) *checkboxState {
	return frame.UseState[checkboxState](ctx, key, stateSlotCheckbox)
}

type checkboxState struct {
	disclosure disclosure.Binding[bool]
	checked    bool
	SelectionAnimation
	focus state.FocusAnimation
}

// checkboxDisclosureCfg builds a disclosure.Config from the widget's checked fields.
func checkboxDisclosureCfg(widget CheckboxWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasChecked,
		Value:      widget.checked,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultChecked,
		OnChange:   widget.onChange,
	}
}

func (s *checkboxState) currentChecked(widget CheckboxWidget) bool {
	s.checked = s.disclosure.Current(checkboxDisclosureCfg(widget))
	return s.checked
}

func (s *checkboxState) bind(widget CheckboxWidget) {
	s.disclosure.Bind(checkboxDisclosureCfg(widget))
}

func (s *checkboxState) requestChecked(widget CheckboxWidget, checked bool) bool {
	s.checked, _ = s.disclosure.Request(checkboxDisclosureCfg(widget), checked)
	return s.checked
}

// SelectionAnimation drives the shared checkbox fill and check-path progress.
type SelectionAnimation struct {
	transition animation.FloatTransition
}

func (s *checkboxState) selection(gtx layout.Context, checked bool, motions ...theme.MotionTheme) float32 {
	return s.SelectionAnimation.Progress(gtx, checked, motions...)
}

func (s *SelectionAnimation) Progress(gtx layout.Context, checked bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.transition.Value(gtx, target, checkboxSelectDuration, animation.EaseSmoothstep, motions...)
}

func (s *checkboxState) focusOpacity(gtx layout.Context, focused bool, motions ...theme.MotionTheme) float32 {
	return s.focus.Opacity(gtx, focused, motions...)
}
