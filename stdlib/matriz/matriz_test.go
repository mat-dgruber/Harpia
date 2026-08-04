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
