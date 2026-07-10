package render

import "image/color"

func DisabledColor(c color.NRGBA) color.NRGBA {
	c.A = byte(uint16(c.A) / 2)
	return c
}
