// Package highlight classifies source text into colored token runs.
//
// The package is deliberately ignorant of widgets and themes: a Highlighter
// returns Tokens with a Kind, and the caller maps a Kind to a color. That keeps
// tokenization reusable across the code block, inline code, logs, or diffs.
//
// New returns a Chroma v2 backed Highlighter. The package is internal because
// the rendered token taxonomy is an implementation detail of widget/code.
package highlight

// Kind classifies one token for color mapping.
type Kind uint8

const (
	// KindPlain is text with no special meaning: identifiers, whitespace, and
	// anything the tokenizer does not classify.
	KindPlain Kind = iota
	KindKeyword
	KindString
	KindNumber
	KindComment
	KindOperator
	KindPunctuation
	KindType
	KindFunction
	KindBuiltin
)

// Token is a contiguous run of source text with one Kind.
type Token struct {
	Text string
	Kind Kind
}

// Highlighter tokenizes source text for one language tag.
//
// Tokenize must not fail: an empty or unknown language returns the source as a
// single KindPlain token, so callers always render something.
type Highlighter interface {
	Tokenize(lang, src string) []Token
}
