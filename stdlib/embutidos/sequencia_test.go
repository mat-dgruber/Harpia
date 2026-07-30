package embutidos

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestSequenciaUmArgumento(t *testing.T) {
	// sequencia(5) -> 0, 1, 2, 3, 4
	res, err := met_emb_sequencia(nil, hrp.Tupla{hrp.Inteiro(5)})
	if err != nil {
		t.Fatalf("erro em sequencia(5): %v", err)
	}

	seq, ok := res.(*SequenciaNumerica)
	if !ok {
		t.Fatalf("esperava *SequenciaNumerica")
	}

	var valores []int64
	for {
		v, err := seq.M__próximo__()
		if err != nil {
			break
		}
		valores = append(valores, int64(v.(hrp.Inteiro)))
	}

	esperados := []int64{0, 1, 2, 3, 4}
	if len(valores) != len(esperados) {
		t.Fatalf("esperava %d itens, obteve %d", len(esperados), len(valores))
	}
	for i, v := range valores {
		if v != esperados[i] {
			t.Errorf("pos %d: esperava %d, obteve %d", i, esperados[i], v)
		}
	}
}

func TestSequenciaDoisArgumentos(t *testing.T) {
	// sequencia(3, 7) -> 3, 4, 5, 6
	res, err := met_emb_sequencia(nil, hrp.Tupla{hrp.Inteiro(3), hrp.Inteiro(7)})
	if err != nil {
		t.Fatalf("erro em sequencia(3, 7): %v", err)
	}

	seq := res.(*SequenciaNumerica)
	var valores []int64
	for {
		v, err := seq.M__próximo__()
		if err != nil {
			break
		}
		valores = append(valores, int64(v.(hrp.Inteiro)))
	}

	esperados := []int64{3, 4, 5, 6}
	if len(valores) != len(esperados) {
		t.Fatalf("esperava %d itens, obteve %d", len(esperados), len(valores))
	}
	for i, v := range valores {
		if v != esperados[i] {
			t.Errorf("pos %d: esperava %d, obteve %d", i, esperados[i], v)
		}
	}
}

func TestSequenciaTresArgumentosPassoNegativo(t *testing.T) {
	// sequencia(10, 5, -2) -> 10, 8, 6
	res, err := met_emb_sequencia(nil, hrp.Tupla{hrp.Inteiro(10), hrp.Inteiro(5), hrp.Inteiro(-2)})
	if err != nil {
		t.Fatalf("erro em sequencia(10, 5, -2): %v", err)
	}

	seq := res.(*SequenciaNumerica)
	var valores []int64
	for {
		v, err := seq.M__próximo__()
		if err != nil {
			break
		}
		valores = append(valores, int64(v.(hrp.Inteiro)))
	}

	esperados := []int64{10, 8, 6}
	if len(valores) != len(esperados) {
		t.Fatalf("esperava %d itens, obteve %d", len(esperados), len(valores))
	}
	for i, v := range valores {
		if v != esperados[i] {
			t.Errorf("pos %d: esperava %d, obteve %d", i, esperados[i], v)
		}
	}
}

func TestSequenciaPassoZero(t *testing.T) {
	_, err := met_emb_sequencia(nil, hrp.Tupla{hrp.Inteiro(0), hrp.Inteiro(10), hrp.Inteiro(0)})
	if err == nil {
		t.Fatalf("esperava erro ao informar passo 0")
	}
}
