package hrp

import (
	"math"
	"testing"

	"github.com/mat-dgruber/Harpia/compartilhado"
)

func TestTiposR3(t *testing.T) {
	t.Run("R2-HRP-03: Inteiro M__mod__ com zero", func(t *testing.T) {
		res, err := Inteiro(1).M__mod__(Inteiro(0))
		if err == nil {
			t.Fatalf("Esperava erro de divisão por zero, obteve res=%v", res)
		}
	})

	t.Run("R3-HRP-01: Texto M__mod__ com tipo invalido sem panic", func(t *testing.T) {
		res, err := Texto("%i").M__mod__(Texto("abc"))
		if err == nil {
			t.Fatalf("Esperava erro ao formatar string 'abc' com %%i, obteve res=%v", res)
		}
	})

	t.Run("R3-HRP-02: StringParaInt limites e sinais invalidos", func(t *testing.T) {
		val, err := compartilhado.StringParaInt("-9223372036854775808")
		if err != nil {
			t.Fatalf("Erro inesperado para MinInt64: %v", err)
		}
		if val != math.MinInt64 {
			t.Fatalf("Esperava math.MinInt64 (%d), obteve %d", math.MinInt64, val)
		}

		if _, err := compartilhado.StringParaInt("--5"); err == nil {
			t.Fatalf("Esperava erro para '--5'")
		}
		if _, err := compartilhado.StringParaInt("-+5"); err == nil {
			t.Fatalf("Esperava erro para '-+5'")
		}
	})

	t.Run("R3-HRP-02: NewInteiro com string invalida e TipoInteiro.Nova vazio", func(t *testing.T) {
		_, err := NewInteiro("abc")
		if err == nil {
			t.Fatalf("Esperava erro de conversão para NewInteiro('abc')")
		}

		obj, err := TipoInteiro.Nova(Tupla{})
		if err != nil {
			t.Fatalf("Erro inesperado em TipoInteiro.Nova(Tupla{}): %v", err)
		}
		if i, ok := obj.(Inteiro); !ok || i != Inteiro(0) {
			t.Fatalf("Esperava Inteiro(0), obteve %v (%T)", obj, obj)
		}
	})

	t.Run("R3-HRP-03: NewBytes com tipo incompativel e TipoBytes.Nova vazio", func(t *testing.T) {
		obj, err := NewBytes(123)
		if err == nil {
			t.Fatalf("Esperava erro para NewBytes(123), obteve obj=%v", obj)
		}
		if obj != nil {
			t.Fatalf("Esperava obj == nil quando NewBytes falha, obteve %v", obj)
		}

		objVazio, err := TipoBytes.Nova(Tupla{})
		if err != nil {
			t.Fatalf("Erro inesperado em TipoBytes.Nova(Tupla{}): %v", err)
		}
		if b, ok := objVazio.(*Bytes); !ok || b == nil {
			t.Fatalf("Esperava *Bytes, obteve %v (%T)", objVazio, objVazio)
		}
	})

	t.Run("PERF-02: Texto M__multiplica__", func(t *testing.T) {
		res, err := Texto("ab").M__multiplica__(Inteiro(5))
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if res != Texto("ababababab") {
			t.Fatalf("Esperava 'ababababab', obteve %v", res)
		}

		resZero, err := Texto("ab").M__multiplica__(Inteiro(0))
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if resZero != Texto("") {
			t.Fatalf("Esperava '', obteve %v", resZero)
		}

		_, errMax := Texto("a").M__multiplica__(Inteiro((64 << 20) + 1))
		if errMax == nil {
			t.Fatalf("Esperava erro de estouro de memória (>64MB), mas obteve sucesso")
		}
	})
}
