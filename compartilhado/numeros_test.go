package compartilhado

import (
	"testing"
)

func TestStringParaIntBase(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado int64
		comErro  bool
	}{
		{"123", 123, false},
		{"-123", -123, false},
		{"+123", 123, false},
		{"0x1A", 26, false},
		{"-0x1A", -26, false},
		{"+0x1A", 26, false},
		{"0b1010", 10, false},
		{"-0b1010", -10, false},
		{"0o755", 493, false},
		{"-0o755", -493, false},
		{"invalido", 0, true},
	}

	for _, c := range casos {
		val, err := StringParaIntBase(c.entrada)
		if c.comErro {
			if err == nil {
				t.Errorf("esperava erro para entrada %q", c.entrada)
			}
		} else {
			if err != nil || val != c.esperado {
				t.Errorf("StringParaIntBase(%q) = %d (err=%v), esperava %d", c.entrada, val, err, c.esperado)
			}
		}
	}
}

func TestFormatacaoNumerica(t *testing.T) {
	if IntParaString(42) != "42" {
		t.Errorf("IntParaString falhou")
	}
	if DecParaString(3.14159) != "3.14159" {
		t.Errorf("DecParaString falhou")
	}
}
