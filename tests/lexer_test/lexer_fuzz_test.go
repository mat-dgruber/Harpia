package lexer_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/lexer"
)

func FuzzLexer(f *testing.F) {
	seeds := []string{
		`var x = 10; const PI = 3.1415;`,
		`0x1F 0b1010 0o755 1_000_000 3.14_15 1e10 3.14e-5`,
		"`linha1\nlinha2 \\n raw string`",
		"/// Doc comment line\n/** Doc comment block */",
		"var x = 1;\r\nconst y = 2;",
		"texto = \"Olá, mundo! 🐍🚀 ⭐\"",
		`/* nested /* comment */ test */`,
		"0x + 1 !verdadeiro `unclosed raw",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		lex := lexer.NewLexer(input)
		count := 0
		maxTokens := len(input)*4 + 100
		for {
			tok := lex.PróximoToken()
			count++
			if tok.Tipo == lexer.TokenFimDeArquivo {
				break
			}
			if count > maxTokens {
				t.Fatalf("Possible infinite loop detected for input %q: exceeded %d tokens", input, maxTokens)
			}
		}
	})
}
