package theme

import "image/color"

// CodeColors holds the syntax-highlighting palette.
//
// It is a theme value, not a component visual contract: the roles are a fixed
// color table with no geometry or state, so a design system overrides them by
// assigning Theme.Code rather than by implementing a provider method. A zero
// CodeColors means "not customized"; CodeColors() then derives the roles from
// the palette and semantic layer.
type CodeColors struct {
	// Background is the block surface the code sits on.
	Background  color.NRGBA
	Plain       color.NRGBA
	Keyword     color.NRGBA
	Type        color.NRGBA
	Function    color.NRGBA
	String      color.NRGBA
	Number      color.NRGBA
	Comment     color.NRGBA
	Operator    color.NRGBA
	Punctuation color.NRGBA
	Builtin     color.NRGBA
	LineNumber  color.NRGBA
}

// IsZero reports whether no role was customized.
func (c CodeColors) IsZero() bool { return c == CodeColors{} }

// CodeColors returns the syntax palette this theme should render with. Roles the
// theme did not customize come from the derived table, so a caller may override
// a single role and keep the rest.
func (t *Theme) CodeColors() CodeColors {
	if t == nil {
		return CodeColors{}
	}
	effective := defaultCodeColors(t)
	custom := t.Code
	overrideColor(&effective.Background, custom.Background)
	overrideColor(&effective.Plain, custom.Plain)
	overrideColor(&effective.Keyword, custom.Keyword)
	overrideColor(&effective.Type, custom.Type)
	overrideColor(&effective.Function, custom.Function)
	overrideColor(&effective.String, custom.String)
	overrideColor(&effective.Number, custom.Number)
	overrideColor(&effective.Comment, custom.Comment)
	overrideColor(&effective.Operator, custom.Operator)
	overrideColor(&effective.Punctuation, custom.Punctuation)
	overrideColor(&effective.Builtin, custom.Builtin)
	overrideColor(&effective.LineNumber, custom.LineNumber)
	return effective
}

// overrideColor applies a customized role. A zero color means "not customized"
// rather than "transparent", so every role stays readable by default.
func overrideColor(target *color.NRGBA, custom color.NRGBA) {
	if custom.A != 0 {
		*target = custom
	}
}

// codeTint is how much of the accent hue the code surface carries. It is small
// enough to read as a neutral panel and large enough to separate the block from
// the card it sits on.
const codeTint = 6.0

// codeBackground tints the tertiary surface with the accent hue, so a code block
// reads as its own panel and follows the brand hue instead of being pure grey.
func codeBackground(palette Palette, semantic Semantic) color.NRGBA {
	return Mix(semantic.SurfaceTertiary, palette.Accent, codeTint)
}

// defaultCodeColors maps the existing palette and semantic roles onto syntax
// roles. Comment and plain text follow the text colors so code matches prose;
// the accent and status ramps carry the token colors.
func defaultCodeColors(t *Theme) CodeColors {
	palette := t.Palette
	semantic := t.Semantic()
	lineNumber := semantic.MutedForeground
	lineNumber.A = uint8(float32(lineNumber.A) * .65)
	return CodeColors{
		Background:  codeBackground(palette, semantic),
		Plain:       semantic.Foreground,
		Keyword:     palette.Accent,
		Type:        palette.Accent,
		Function:    palette.AccentHover,
		String:      palette.Success,
		Number:      palette.Warning,
		Comment:     semantic.MutedForeground,
		Operator:    semantic.Foreground,
		Punctuation: semantic.Foreground,
		Builtin:     palette.Danger,
		LineNumber:  lineNumber,
	}
}
