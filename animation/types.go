package animation

import (
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// Context is the per-frame context used by stateful animations.
//
// The alias keeps animation helpers within the public API without exposing an
// internal package in their signatures.
type Context = core.Context

// Widget is a composable FlowUI node accepted by AnimateLayout.
type Widget = core.Widget

// MotionTheme controls whether an animation runs and how its duration scales.
type MotionTheme = theme.MotionTheme
