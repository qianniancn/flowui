package text

import (
	"image/color"
	"testing"
)

func TestTextTransparentColor(t *testing.T) {
	text := New("hidden").Color(color.NRGBA{})
	if !text.hasColor {
		t.Fatal("transparent color should still count as explicit")
	}
}
