package switches

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotSwitch = "switch"

func switchStateFor(ctx *frame.Context, key string) *switchState {
	key = frame.ClaimKey(ctx, state.KindSwitch, key)
	return frame.UseState[switchState](ctx, key, stateSlotSwitch)
}

type switchState struct {
	disclosure disclosure.Binding[bool]
	checked    bool
	value      widget.Bool
	selected   animation.FloatTransition
	focus      state.FocusAnimation
}

// switchDisclosureCfg builds a disclosure.Config from the widget's checked fields.
func switchDisclosureCfg(widget SwitchWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasChecked,
		Value:      widget.checked,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultChecked,
		OnChange:   widget.onChange,
	}
}

func (s *switchState) currentChecked(widget SwitchWidget) bool {
	s.checked = s.disclosure.Current(switchDisclosureCfg(widget))
	return s.checked
}

func (s *switchState) bind(widget SwitchWidget) {
	s.disclosure.Bind(switchDisclosureCfg(widget))
}

func (s *switchState) requestChecked(widget SwitchWidget, checked bool) bool {
	s.checked, _ = s.disclosure.Request(switchDisclosureCfg(widget), checked)
	return s.checked
}

func (s *switchState) selection(gtx layout.Context, checked bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.selected.Value(gtx, target, switchSelectDuration, animation.EaseSmoothstep, motions...)
}

func (s *switchState) focusOpacity(gtx layout.Context, focused bool, motions ...theme.MotionTheme) float32 {
	return s.focus.Opacity(gtx, focused, motions...)
}
