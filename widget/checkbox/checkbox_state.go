package checkbox

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotCheckbox = "checkbox"

func checkboxStateFor(ctx *core.Context, key string) *checkboxState {
	return core.UseState[checkboxState](ctx, key, stateSlotCheckbox)
}

type checkboxState struct {
	disclosure disclosure.Binding[bool]
	checked    bool
	SelectionAnimation
	focus      animation.FocusAnimation
	styleCache VisualCache
}

// checkboxDisclosureCfg builds a disclosure.Config from the widget's checked fields.
func checkboxDisclosureCfg(widget Widget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasChecked,
		Value:      widget.checked,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultChecked,
		OnChange:   widget.onChange,
	}
}

func (s *checkboxState) currentChecked(widget Widget) bool {
	s.checked = s.disclosure.Current(checkboxDisclosureCfg(widget))
	return s.checked
}

func (s *checkboxState) bind(widget Widget) {
	s.disclosure.Bind(checkboxDisclosureCfg(widget))
}

func (s *checkboxState) requestChecked(widget Widget, checked bool) bool {
	s.checked, _ = s.disclosure.Request(checkboxDisclosureCfg(widget), checked)
	return s.checked
}

// SelectionAnimation drives the shared checkbox fill and check-path progress.
type SelectionAnimation struct {
	transition animation.FloatTransition
}

func (s *checkboxState) selection(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.SelectionAnimation.Progress(gtx, checked, duration, motions...)
}

func (s *SelectionAnimation) Progress(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *checkboxState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}
