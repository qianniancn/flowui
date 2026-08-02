package ui

import (
	"gioui.org/font"
	"gioui.org/font/opentype"
)

// ParseFontCollection loads a TTF, OTF, or TTC font file for use in Theme.Fonts.
// The returned faces keep the font data in memory and can be shared by themes.
func ParseFontCollection(data []byte) ([]font.FontFace, error) {
	return opentype.ParseCollection(data)
}
