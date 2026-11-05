package parser

import (
	"testing"
	"time"
)

// parseComPrazo falha o teste se o parser não terminar em 2s (travamentos encontrados por fuzz).
func parseComPrazo(t *testing.T, src string) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		_, err := NewParserFromString(src, "t.hrp").Parse()
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatalf("parser travou em %q", src)
		return nil
	}
}

func TestParserNaoTrava(t *testing.T) {
	for _, src := range []string{"!", "!ar", "se !a { }", "funcao f(", "funcao f(a", "funcao f(a,", "funcao f(a b) {}"} {
		parseComPrazo(t, src)
	}
}

func TestNegacaoExclamacao(t *testing.T) {
	if err := parseComPrazo(t, "var a = Falso\nse !a { }\n"); err != nil {
		t.Fatalf("'!a' deveria ser válido: %v", err)
	}
}

func TestParametrosMalformadosDaoErro(t *testing.T) {
	for _, src := range []string{"funcao f(a", "funcao f(a,", "funcao f(a b) {}"} {
		if err := parseComPrazo(t, src); err == nil {
			t.Errorf("esperava erro de sintaxe para %q", src)
		}
	}
}

func FuzzParser(f *testing.F) {
	for _, s := range []string{
		"var a = 1",
		"funcao f(a, b = 2) { retorne a + b }",
		"se !x { } senao { }",
		"<div classe=\"x\">{a}</div>",
		"tente { } capture (e) { }",
		"a |> b",
		"importar { a } de \"b\"",
		"classe A { construtor() {} }",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		done := make(chan struct{})
		go func() {
			defer close(done)
			NewParserFromString(s, "f.hrp").Parse()
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("parser travou em %q", s)
		}
	})
}

// '}' fecha a instrução como a quebra de linha: blocos de uma linha com atribuição são válidos.
func TestFechaChavesTerminaInstrucao(t *testing.T) {
	for _, src := range []string{
		"funcao f() { var a = 1 }",
		"se Verdadeiro { a = 2 }",
		"funcao h(req, res) { res.corpo = \"oi\" }",
		"enquanto Falso { x += 1 }",
		"funcao g() { retorne 1 }",
	} {
		if err := parseComPrazo(t, src); err != nil {
			t.Errorf("%q deveria ser válido: %v", src, err)
		}
	}
}

// Encadeamento fluente com o ponto no começo da linha (estilo do QueryBuilder na documentação).
func TestEncadeamentoComPontoNoInicioDaLinha(t *testing.T) {
	src := "var r = bd.tabela(\"u\")\n    .onde(\"id\", \"==\", 1)\n    .obterUm()\nimprimir(r)\n"
	prog, err := NewParserFromString(src, "t.hrp").Parse()
	if err != nil {
		t.Fatalf("encadeamento multilinha deveria ser válido: %v", err)
	}
	if n := len(prog.Declaracoes); n != 2 {
		t.Fatalf("esperava 2 instruções (var + imprimir), veio %d", n)
	}
}
