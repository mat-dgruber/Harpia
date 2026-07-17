package lexer_test

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/lexer"
)

func benchSource(n int) string {
	linha := `var x123 = 1234 + 5678; `
	return strings.Repeat(linha, n)
}

func BenchmarkLexer1k(b *testing.B) {
	src := benchSource(20) // ~1k
	b.ResetTimer()
	for b.Loop() {
		lex := lexer.NewLexer(src)
		for {
			tk := lex.ProximoToken()
			if tk.Tipo == lexer.TokenFimDeArquivo {
				break
			}
		}
	}
}

func BenchmarkLexer10k(b *testing.B) {
	src := benchSource(200) // ~10k
	b.ResetTimer()
	for b.Loop() {
		lex := lexer.NewLexer(src)
		for {
			tk := lex.ProximoToken()
			if tk.Tipo == lexer.TokenFimDeArquivo {
				break
			}
		}
	}
}
