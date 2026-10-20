package text

import giotext "gioui.org/text"

// Alignment controls text alignment within its layout width. It is the same
// value type Widget.Align accepts.
type Alignment = giotext.Alignment

// WrapPolicy controls where text may wrap. It is the same value type
// Widget.Wrap accepts.
type WrapPolicy = giotext.WrapPolicy

const (
	// AlignStart aligns text to the start of the layout width.
	AlignStart = giotext.Start
	// AlignEnd aligns text to the end of the layout width.
	AlignEnd = giotext.End
	// AlignCenter centers text within the layout width.
	AlignCenter = giotext.Middle
)

const (
	// WrapHeuristically wraps lines heuristically.
	WrapHeuristically = giotext.WrapHeuristically
	// WrapWords wraps only at word boundaries.
	WrapWords = giotext.WrapWords
	// WrapGraphemes wraps at grapheme clusters.
	WrapGraphemes = giotext.WrapGraphemes
)
