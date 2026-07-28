package animation

import (
	"math"
	"time"
)

// SpringConfig configures a damped spring used by Tween and layout helpers.
// Units are UI-oriented: stiffness and damping are tuned for density-independent
// values around typical screen coordinates (not SI units).
type SpringConfig struct {
	// Stiffness pulls the value toward the target. Higher is snappier.
	Stiffness float32
	// Damping resists velocity. Higher settles faster with less overshoot.
	Damping float32
	// Mass scales inertia. Must be positive; default 1.
	Mass float32
	// RestDisplacement is the |value-target| threshold for rest (default 0.5).
	RestDisplacement float32
	// RestVelocity is the |velocity| threshold for rest (default 0.5).
	RestVelocity float32
}

// DefaultSpring returns a balanced desktop spring (slight overshoot possible).
func DefaultSpring() SpringConfig {
	return SpringConfig{
		Stiffness:        170,
		Damping:          26,
		Mass:             1,
		RestDisplacement: 0.5,
		RestVelocity:     0.5,
	}
}

// SpringSnappy is a quick, low-overshoot spring for small UI motions.
func SpringSnappy() SpringConfig {
	return SpringConfig{
		Stiffness:        300,
		Damping:          30,
		Mass:             1,
		RestDisplacement: 0.35,
		RestVelocity:     0.35,
	}
}

// SpringGentle is a soft spring for larger layout moves.
func SpringGentle() SpringConfig {
	return SpringConfig{
		Stiffness:        120,
		Damping:          22,
		Mass:             1,
		RestDisplacement: 0.75,
		RestVelocity:     0.75,
	}
}

// SpringBouncy overshoots noticeably before settling.
func SpringBouncy() SpringConfig {
	return SpringConfig{
		Stiffness:        200,
		Damping:          12,
		Mass:             1,
		RestDisplacement: 0.5,
		RestVelocity:     0.5,
	}
}

func (c SpringConfig) normalized() SpringConfig {
	if c.Stiffness <= 0 || math.IsNaN(float64(c.Stiffness)) || math.IsInf(float64(c.Stiffness), 0) {
		c.Stiffness = 170
	}
	if c.Damping < 0 || math.IsNaN(float64(c.Damping)) || math.IsInf(float64(c.Damping), 0) {
		c.Damping = 26
	}
	if c.Mass <= 0 || math.IsNaN(float64(c.Mass)) || math.IsInf(float64(c.Mass), 0) {
		c.Mass = 1
	}
	if c.RestDisplacement <= 0 || math.IsNaN(float64(c.RestDisplacement)) || math.IsInf(float64(c.RestDisplacement), 0) {
		c.RestDisplacement = 0.5
	}
	if c.RestVelocity <= 0 || math.IsNaN(float64(c.RestVelocity)) || math.IsInf(float64(c.RestVelocity), 0) {
		c.RestVelocity = 0.5
	}
	return c
}

// springState integrates a damped harmonic oscillator toward target.
type springState struct {
	value    float32
	velocity float32
	target   float32
	ready    bool
	config   SpringConfig
}

func (s *springState) snap(target float32) {
	s.value = target
	s.velocity = 0
	s.target = target
	s.ready = true
}

// step advances the spring by dt and reports whether it is still moving.
func (s *springState) step(dt time.Duration, config SpringConfig) bool {
	config = config.normalized()
	s.config = config
	if dt <= 0 {
		return !s.atRest(config)
	}
	// Cap huge frame gaps so a tab-out does not explode the integrator.
	seconds := float32(dt.Seconds())
	if seconds > 1.0/15.0 {
		seconds = 1.0 / 15.0
	}
	// Semi-implicit Euler with small substeps for stability.
	const maxStep = float32(1.0 / 120.0)
	remaining := seconds
	for remaining > 0 {
		step := remaining
		if step > maxStep {
			step = maxStep
		}
		remaining -= step
		force := -config.Stiffness*(s.value-s.target) - config.Damping*s.velocity
		accel := force / config.Mass
		s.velocity += accel * step
		s.value += s.velocity * step
		if math.IsNaN(float64(s.value)) || math.IsInf(float64(s.value), 0) {
			s.snap(s.target)
			return false
		}
	}
	if s.atRest(config) {
		s.value = s.target
		s.velocity = 0
		return false
	}
	return true
}

func (s *springState) atRest(config SpringConfig) bool {
	return absFloat(s.value-s.target) <= config.RestDisplacement &&
		absFloat(s.velocity) <= config.RestVelocity
}

func absFloat(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}
