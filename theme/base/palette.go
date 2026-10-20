package base

import (
	"image/color"
	"math"

	"github.com/qianniancn/flowui/theme"
)

func rgb(value uint32) color.NRGBA {
	return color.NRGBA{R: byte(value >> 16), G: byte(value >> 8), B: byte(value), A: 255}
}

// Default palette: Slate surfaces and Indigo/Red/Green/Amber accents.
// Only the roles we consume are stored here; custom brand colors continue to
// use the seed-derived palette.
type accentScale struct {
	soft, softHover, edge, hover, text uint32
}

var accentScales = map[uint32][2]accentScale{
	0x3e63dd: {
		{soft: 0xedf2fe, softHover: 0xe1e9ff, edge: 0x8da4ef, hover: 0x3358d4, text: 0x3a5bc7},
		{soft: 0x182449, softHover: 0x1d2e62, edge: 0x435db1, hover: 0x5472e4, text: 0x9eb1ff},
	},
	0xe5484d: {
		{soft: 0xfeebec, softHover: 0xffdbdc, edge: 0xeb8e90, hover: 0xdc3e42, text: 0xce2c31},
		{soft: 0x3b1219, softHover: 0x500f1c, edge: 0xb54548, hover: 0xec5d5e, text: 0xff9592},
	},
	0x30a46c: {
		{soft: 0xe6f6eb, softHover: 0xd6f1df, edge: 0x5bb98b, hover: 0x2b9a66, text: 0x218358},
		{soft: 0x132d21, softHover: 0x113b29, edge: 0x2f7c57, hover: 0x33b074, text: 0x3dd68c},
	},
	0xffc53d: {
		{soft: 0xfff7c2, softHover: 0xffee9c, edge: 0xe2a336, hover: 0xffba18, text: 0xab6400},
		{soft: 0x302008, softHover: 0x3f2700, edge: 0x8f6424, hover: 0xffd60a, text: 0xffca16},
	},
}

func slate(dark bool) [12]color.NRGBA {
	values := [12]uint32{
		0xfcfcfd, 0xf9f9fb, 0xf0f0f3, 0xe8e8ec, 0xe0e1e6, 0xd9d9e0,
		0xcdced6, 0xb9bbc6, 0x8b8d98, 0x80838d, 0x60646c, 0x1c2024,
	}
	if dark {
		values = [12]uint32{
			0x111113, 0x18191b, 0x212225, 0x272a2d, 0x2e3135, 0x363a3f,
			0x43484e, 0x5a6169, 0x696e77, 0x777b84, 0xb0b4ba, 0xedeef0,
		}
	}
	var colors [12]color.NRGBA
	for i, value := range values {
		colors[i] = rgb(value)
	}
	return colors
}

// A supplied Neutral adds a subtle cast without changing surface contrast.
func semanticColors(sem theme.Semantic, seed theme.Seed, dark bool) theme.Semantic {
	gray := slate(dark)
	canvas, panel, field := gray[0], rgb(0xffffff), rgb(0xffffff)
	hover, pressed := gray[2], gray[3]
	foreground, muted, border, separator := gray[11], gray[10], gray[6], gray[4]
	shadow := color.NRGBA{A: 50}
	if dark {
		panel = gray[1]
		field = theme.Composite(panel, rgb(0), .25)
		shadow.A = 160
	}
	if seed.Neutral != (color.NRGBA{}) {
		for _, surface := range []*color.NRGBA{&canvas, &panel, &field, &hover, &pressed} {
			*surface = theme.Composite(*surface, seed.Neutral, .03)
		}
	}
	sem.Background, sem.Foreground = canvas, foreground
	sem.Surface, sem.SurfaceForeground = panel, foreground
	sem.SurfaceSecondary, sem.SurfaceSecondaryForeground = field, foreground
	sem.SurfaceTertiary, sem.SurfaceTertiaryForeground = hover, foreground
	sem.SurfaceHover, sem.SurfaceRaised, sem.SurfacePressed = hover, hover, pressed
	sem.Overlay, sem.OverlayForeground = panel, foreground
	sem.Muted, sem.MutedForeground = hover, muted
	sem.Border, sem.Separator = border, separator
	sem.FieldBackground, sem.FieldHover, sem.FieldFocus = field, field, field
	// Placeholder text is a translucent ink over the field face rather than a
	// solid ramp rung, so it recedes on every surface while staying readable.
	sem.FieldForeground, sem.FieldPlaceholder = foreground, theme.Composite(field, foreground, .48)
	sem.Segment, sem.SegmentForeground = panel, foreground
	sem.ControlSurface, sem.ControlSurfaceHover = hover, pressed
	sem.ControlOutline, sem.ControlOutlineForeground = border, foreground
	// Disabled surfaces blend a few percent of ink (light) or paper (dark) over
	// the panel instead of stepping the whole background down to gray-2, and the
	// label keeps gray-a11-level contrast rather than fading to gray-7.
	sem.DisabledBackground, sem.DisabledForeground, sem.DisabledBorder = disabledSurface(panel, dark), gray[8], separator
	sem.SurfaceShadow, sem.OverlayShadow = color.NRGBA{}, shadow
	sem.PrimaryGradient = nil

	for tone, fill := range map[theme.Tone]color.NRGBA{
		theme.TonePrimary: seed.Primary, theme.ToneSecondary: seed.Primary,
		theme.ToneDanger: seed.Danger, theme.ToneSuccess: seed.Success,
		theme.ToneWarning: seed.Warning,
	} {
		if fill == (color.NRGBA{}) {
			fill = sem.Tone(tone).Fill
		}
		fill.A = 255
		sem.Tones[tone] = toneColors(fill, panel, foreground, dark)
	}
	sem.Tones[theme.ToneDefault] = theme.TonePair{
		Fill: hover, OnFill: foreground, Hover: pressed, Pressed: gray[4],
		Soft: hover, SoftHover: pressed, OnSoft: foreground,
	}
	sem.Tones[theme.ToneHover] = theme.TonePair{
		Fill: hover, OnFill: foreground, Hover: pressed, Pressed: gray[4],
		Soft: hover, SoftHover: pressed, OnSoft: foreground,
	}
	sem.Ring = sem.Tone(theme.TonePrimary).Fill
	sem.Selection = theme.Composite(field, sem.Ring, .3)
	return sem
}

// disabledSurface is the disabled face: gray-a2 in light, a matching
// translucent paper blend in dark. It stays one step from the panel instead of
// turning the whole control into a gray block.
func disabledSurface(panel color.NRGBA, dark bool) color.NRGBA {
	if dark {
		return theme.Composite(panel, rgb(0xffffff), .06)
	}
	return theme.Composite(panel, rgb(0), .04)
}

func toneColors(fill, panel, foreground color.NRGBA, dark bool) theme.TonePair {
	label := onFill(fill)
	target, alpha := rgb(0), .12
	if label == rgb(0) {
		target = rgb(0xffffff)
	}
	if dark {
		alpha = .18
	}
	pair := theme.TonePair{
		Fill: fill, OnFill: label,
		Hover: theme.Mix(fill, target, 6), Pressed: theme.Mix(fill, target, 12),
		Soft: theme.Composite(panel, fill, alpha), SoftHover: theme.Composite(panel, fill, alpha+.08),
		OnSoft: foreground,
	}
	if scale, ok := accentScaleFor(fill, dark); ok {
		pair.Hover = rgb(scale.hover)
		pair.Pressed = theme.Mix(pair.Hover, target, 8)
		pair.Soft, pair.SoftHover, pair.OnSoft = rgb(scale.soft), rgb(scale.softHover), rgb(scale.text)
		// Some solid states fall just below AA with a fixed label. Keep
		// the source hue, nudging only those states toward the resting fill.
		pair.Hover = legibleState(pair.Hover, fill, label, minLabelContrast)
		pair.Pressed = legibleState(pair.Pressed, fill, label, minLabelContrast)
		// Soft labels are ordinary body text on a light tint, so they keep the
		// normal-text threshold.
		pair.OnSoft = legibleState(pair.OnSoft, foreground, pair.Soft, 4.5)
		pair.OnSoft = legibleState(pair.OnSoft, foreground, pair.SoftHover, 4.5)
	}
	return pair
}

func accentScaleFor(fill color.NRGBA, dark bool) (accentScale, bool) {
	value := uint32(fill.R)<<16 | uint32(fill.G)<<8 | uint32(fill.B)
	scales, ok := accentScales[value]
	if dark {
		return scales[1], ok
	}
	return scales[0], ok
}

// minLabelContrast is the contrast a solid tone's label must clear. Solid
// controls carry short, usually medium-weight labels, so the WCAG large-text
// threshold applies; mid-dark fills (red-9, green-9) pair with white at
// similar ratios instead of the slightly higher scoring black.
const minLabelContrast = 3.0

func legibleState(state, fill, label color.NRGBA, minimum float64) color.NRGBA {
	for amount := float64(0); amount <= 100; amount++ {
		candidate := theme.Mix(state, fill, amount)
		a, b := luminance(candidate), luminance(label)
		if (max(a, b)+.05)/(min(a, b)+.05) >= minimum {
			return candidate
		}
	}
	return fill
}

// onFill picks the label for a solid fill: white unless the fill is too light
// to carry it. Picking the mathematically higher score instead would put black
// on the red and green fills, which reads as a tinted warning rather than an
// action.
func onFill(fill color.NRGBA) color.NRGBA {
	white := rgb(0xffffff)
	if theme.ContrastRatio(fill, white) >= minLabelContrast {
		return white
	}
	return rgb(0)
}

func luminance(fill color.NRGBA) float64 {
	linear := func(channel byte) float64 {
		v := float64(channel) / 255
		if v <= .04045 {
			return v / 12.92
		}
		return math.Pow((v+.055)/1.055, 2.4)
	}
	return .2126*linear(fill.R) + .7152*linear(fill.G) + .0722*linear(fill.B)
}
