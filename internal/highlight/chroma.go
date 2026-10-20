package highlight

import (
	"strings"
	"sync"

	chromalib "github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Chroma tokenizes source with Chroma v2 and maps Chroma token types onto
// FlowUI's syntax roles. It caches the lexer for each language tag.
type Chroma struct {
	mu     sync.Mutex
	lexers map[string]chromalib.Lexer
}

// New creates a Chroma-backed highlighter.
func New() *Chroma {
	return &Chroma{lexers: map[string]chromalib.Lexer{}}
}

// Tokenize implements Highlighter. Unknown languages and lexer failures fall
// back to one plain token, matching the package contract.
func (c *Chroma) Tokenize(lang, source string) (tokens []Token) {
	lexer := c.lexer(lang)
	if lexer == nil {
		return []Token{{Text: source, Kind: KindPlain}}
	}

	// Chroma may panic inside a lexer iterator; a code block must still paint.
	defer func() {
		if recover() != nil {
			tokens = []Token{{Text: source, Kind: KindPlain}}
		}
	}()

	iterator, err := lexer.Tokenise(&chromalib.TokeniseOptions{
		State:    "root",
		EnsureLF: false,
	}, source)
	if err != nil {
		return []Token{{Text: source, Kind: KindPlain}}
	}
	for _, token := range iterator.Tokens() {
		if token.Value == "" {
			continue
		}
		kind := kindOf(token.Type)
		if n := len(tokens); n > 0 && tokens[n-1].Kind == kind {
			tokens[n-1].Text += token.Value
			continue
		}
		tokens = append(tokens, Token{Text: token.Value, Kind: kind})
	}
	if len(tokens) == 0 && source != "" {
		return []Token{{Text: source, Kind: KindPlain}}
	}
	return tokens
}

func (c *Chroma) lexer(lang string) chromalib.Lexer {
	key := strings.ToLower(strings.TrimSpace(lang))
	if key == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lexers == nil {
		c.lexers = map[string]chromalib.Lexer{}
	}
	if lexer, ok := c.lexers[key]; ok {
		return lexer
	}
	lexer := lexers.Get(key)
	if lexer != nil {
		lexer = chromalib.Coalesce(lexer)
	}
	c.lexers[key] = lexer
	return lexer
}

// kindOf collapses Chroma's token taxonomy onto the FlowUI syntax palette.
func kindOf(token chromalib.TokenType) Kind {
	switch {
	case token.InCategory(chromalib.Comment):
		return KindComment
	case token == chromalib.KeywordType:
		return KindType
	case token.InCategory(chromalib.Keyword):
		return KindKeyword
	case token.InSubCategory(chromalib.LiteralString):
		return KindString
	case token.InSubCategory(chromalib.LiteralNumber):
		return KindNumber
	case token == chromalib.NameBuiltin, token == chromalib.NameBuiltinPseudo:
		return KindBuiltin
	case token == chromalib.NameFunction:
		return KindFunction
	case token == chromalib.NameClass,
		token == chromalib.NameNamespace,
		token == chromalib.NameException:
		return KindType
	case token.InCategory(chromalib.Operator):
		return KindOperator
	case token.InCategory(chromalib.Punctuation):
		return KindPunctuation
	default:
		return KindPlain
	}
}

var _ Highlighter = (*Chroma)(nil)
