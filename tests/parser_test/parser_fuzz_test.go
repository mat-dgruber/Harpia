package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

func FuzzParser(f *testing.F) {
	sementes := []string{
		"var x = 10 + 20 * 3",
		"se x > 5 { imprimir(x) } senao { imprimir(0) }",
		"var [a, b, ...resto] = [1, 2, 3]",
		"const { nome, idade } = usuario",
		"x ??= 10; y ||= Falso",
		"tente { risco() } capture (e) { log(e) } finalmente { limpa() }",
	}

	for _, semente := range sementes {
		f.Add([]byte(semente))
	}

	f.Fuzz(func(t *testing.T, dados []byte) {
		p := parser.NewParserFromString(string(dados), "fuzz.hrp")
		// Must never panic or hang
		_, _ = p.Parse()
	})
}
