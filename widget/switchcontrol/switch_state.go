package switchcontrol

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotSwitch = "switch"

func switchStateFor(ctx *core.Context, key string) *switchState {
	key = core.ClaimKey(ctx, core.KindSwitch, key)
	return core.UseState[switchState](ctx, key, stateSlotSwitch)
}

type switchState struct {
	disclosure disclosure.Binding[bool]
	checked    bool
	value      widget.Bool
	selected   animation.FloatTransition
	focus      animation.FocusAnimation
	styleCache VisualCache
}

// switchDisclosureCfg builds a disclosure.Config from the widget's checked fields.
func switchDisclosureCfg(widget Widget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasChecked,
		Value:      widget.checked,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultChecked,
		OnChange:   widget.onChange,
	}
}

func (s *switchState) currentChecked(widget Widget) bool {
	s.checked = s.disclosure.Current(switchDisclosureCfg(widget))
	return s.checked
}

func (s *switchState) bind(widget Widget) {
	s.disclosure.Bind(switchDisclosureCfg(widget))
}

func (s *switchState) requestChecked(widget Widget, checked bool) bool {
	s.checked, _ = s.disclosure.Request(switchDisclosureCfg(widget), checked)
	return s.checked
}

func (s *switchState) selection(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.selected.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *switchState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}
