package highlight

import (
	"strings"
	"testing"
)

func joined(tokens []Token) string {
	var builder strings.Builder
	for _, token := range tokens {
		builder.WriteString(token.Text)
	}
	return builder.String()
}

func TestTokenizeGoKeepsSourceAndClassifiesSyntax(t *testing.T) {
	source := "package main\n\nfunc main() { println(42) }\n"
	tokens := New().Tokenize("go", source)
	if got := joined(tokens); got != source {
		t.Fatalf("tokens = %q, want source", got)
	}
	kinds := map[Kind]bool{}
	for _, token := range tokens {
		kinds[token.Kind] = true
	}
	if !kinds[KindKeyword] || !kinds[KindNumber] {
		t.Fatalf("Go token kinds = %v, want keyword and number", kinds)
	}
}

func TestUnknownLanguageFallsBackToPlain(t *testing.T) {
	tokens := New().Tokenize("not-a-language", "hello\nworld")
	if len(tokens) != 1 || tokens[0].Kind != KindPlain || tokens[0].Text != "hello\nworld" {
		t.Fatalf("unknown language tokens = %#v", tokens)
	}
}

func TestTokenizePreservesCRLF(t *testing.T) {
	source := "package main\r\n\r\nfunc main() {}\r\n"
	tokens := New().Tokenize("go", source)
	if got := joined(tokens); got != source {
		t.Fatalf("tokens = %q, want source verbatim", got)
	}
}

func TestKnownLanguagesProduceTokens(t *testing.T) {
	tests := map[string]string{
		"python":     "def greet(name):\n    return f\"hello {name}\"\n",
		"typescript": "const value: number = 42;\n",
		"rust":       "fn main() { println!(\"hi\"); }\n",
		"json":       `{"ready": true}`,
	}
	for lang, source := range tests {
		t.Run(lang, func(t *testing.T) {
			tokens := New().Tokenize(lang, source)
			if got := joined(tokens); got != source {
				t.Fatalf("tokens = %q, want source", got)
			}
			colored := false
			for _, token := range tokens {
				if token.Kind != KindPlain {
					colored = true
				}
			}
			if !colored {
				t.Fatalf("%s produced no colored tokens", lang)
			}
		})
	}
}
