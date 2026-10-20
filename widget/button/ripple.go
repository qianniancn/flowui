package button

import (
	"time"

	"gioui.org/widget"
	"github.com/qianniancn/flowui/theme"
)

// ripple is a decorative animation. A translucent
// circle expands from the press point while its opacity fades. The widget owns
// the mechanism; the visual provider controls whether it is enabled and its duration.

const stateSlotRipple = "ripple"

type rippleState struct {
	active    []theme.ButtonRipple
	lastStart time.Time
}

func rippleDurationFor(visual theme.ButtonVisual) time.Duration {
	return theme.ResolveMotionDuration(visual.Motion, visual.RippleDuration)
}

// rippleFadeDurationFor resolves the provider-declared fade channel.
func rippleFadeDurationFor(visual theme.ButtonVisual) time.Duration {
	return theme.ResolveMotionDuration(visual.Motion, visual.RippleFadeDuration)
}

// updateRipples absorbs presses newer than the last recorded one and drops
// finished ripples. It reports whether any ripple is still animating.
func updateRipples(state *rippleState, history []widget.Press, fadeDuration time.Duration, now time.Time) bool {
	for _, press := range history {
		if press.Cancelled || !press.Start.After(state.lastStart) {
			continue
		}
		state.lastStart = press.Start
		state.active = append(state.active, theme.ButtonRipple{
			StartedAt: press.Start,
			X:         float32(press.Position.X),
			Y:         float32(press.Position.Y),
		})
	}
	kept := state.active[:0]
	animating := false
	for _, ripple := range state.active {
		if now.Sub(ripple.StartedAt) < fadeDuration {
			kept = append(kept, ripple)
			animating = true
		}
	}
	state.active = kept
	return animating
}
