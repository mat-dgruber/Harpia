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
