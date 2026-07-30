package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

func BenchmarkLexer(b *testing.B) {
	codigo := "var [a, b, ...resto] = [1, 2, 3]; escolha x { caso 1 { retorne 10 } }"
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		l := lexer.NovoLexer(codigo, "bench.hrp")
		_ = l.ObterTokens()
	}
}

func BenchmarkParser(b *testing.B) {
	codigo := "var [a, b, ...resto] = [1, 2, 3]; escolha x { caso 1 { retorne 10 } }"
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		p := parser.NewParserFromString(codigo, "bench.hrp")
		_, _ = p.Parse()
	}
}
