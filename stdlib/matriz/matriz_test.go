package matriz

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestProdutoEscalar(t *testing.T) {
	v1 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(1), hrp.Decimal(2), hrp.Decimal(3)}}
	v2 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(4), hrp.Decimal(5), hrp.Decimal(6)}}

	res, err := met_matriz_produtoEscalar(nil, hrp.Tupla{v1, v2})
	if err != nil {
		t.Fatalf("Erro inesperado no produtoEscalar: %v", err)
	}
	dec := float64(res.(hrp.Decimal))
	// 1*4 + 2*5 + 3*6 = 4 + 10 + 18 = 32
	if dec != 32.0 {
		t.Errorf("Esperava 32.0, obteve %f", dec)
	}
}

func TestMatrizCriarDimensoesInvalidas(t *testing.T) {
	_, err := met_matriz_criar(nil, hrp.Tupla{hrp.Decimal(-1), hrp.Decimal(2)})
	if err == nil {
		t.Fatal("esperava erro ao criar matriz com linha negativa")
	}

	_, err = met_matriz_criar(nil, hrp.Tupla{hrp.Decimal(2), hrp.Decimal(0)})
	if err == nil {
		t.Fatal("esperava erro ao criar matriz com coluna zero")
	}
}

func TestMatrizMultiplicarMatrizIrregular(t *testing.T) {
	// m1 irregular: [[1, 2], [3]]
	l1 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(1), hrp.Decimal(2)}}
	l2 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(3)}}
	m1 := &hrp.Lista{Itens: hrp.Tupla{l1, l2}}

	m2 := &hrp.Lista{Itens: hrp.Tupla{
		&hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(1), hrp.Decimal(2)}},
		&hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(3), hrp.Decimal(4)}},
	}}

	_, err := met_matriz_multiplicar(nil, hrp.Tupla{m1, m2})
	if err == nil {
		t.Fatal("esperava erro ao multiplicar com matriz irregular")
	}
}
