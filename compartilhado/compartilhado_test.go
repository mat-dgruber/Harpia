package compartilhado

import (
	"testing"
)

func TestContemApenasAlfaNum(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado bool
	}{
		{"123", true},
		{"abc", true},
		{"usuario123", true},
		{"a1", true},
		{"api_v1", false}, // underline nao e alfa/num
		{"", false},
		{"hello world", false},
	}

	for _, c := range casos {
		res := ContemApenasAlfaNum(c.entrada)
		if res != c.esperado {
			t.Errorf("ContemApenasAlfaNum(%q) = %v, esperava %v", c.entrada, res, c.esperado)
		}
	}
}

func TestEIdentificadorValido(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado bool
	}{
		{"_usuario", true},
		{"usuario123", true},
		{"contador", true},
		{"123usuario", false}, // nao pode começar com numero
		{"", false},
		{"var-nome", false}, // hifen invalido
	}

	for _, c := range casos {
		res := EIdentificadorValido(c.entrada)
		if res != c.esperado {
			t.Errorf("EIdentificadorValido(%q) = %v, esperava %v", c.entrada, res, c.esperado)
		}
	}
}
