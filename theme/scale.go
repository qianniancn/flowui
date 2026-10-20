package theme

import "image/color"

// Step names one rung of a ColorScale. The ten rungs follow the conventional
// 50-900 naming shared by common design systems.
type Step uint8

const (
	Step50 Step = iota
	Step100
	Step200
	Step300
	Step400
	Step500
	Step600
	Step700
	Step800
	Step900
)

// ColorScale is one hue's ten-step ramp, ordered light to dark. Step500 is the
// mid rung and is where a seed lands.
type ColorScale struct {
	Steps [10]color.NRGBA
}

// At returns the colour at a step. An out-of-range step yields the mid rung so
// a bad index degrades to a plausible colour instead of transparent.
func (c ColorScale) At(s Step) color.NRGBA {
	if int(s) >= len(c.Steps) {
		return c.Steps[Step500]
	}
	return c.Steps[s]
}

// Scales is the v2 colour surface: one ramp per semantic hue. Every tone reads
// from here, so a style whose second hue is implicit can point Secondary at the
// same ramp as Primary.
type Scales struct {
	Primary   ColorScale
	Secondary ColorScale
	Neutral   ColorScale
	Danger    ColorScale
	Success   ColorScale
	Warning   ColorScale
	Info      ColorScale
}
