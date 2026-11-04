package matematica

import (
	"errors"
	"math"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestMatRaizCasosEspeciais(t *testing.T) {
	// raiz(4, 0) -> erro índice zero
	_, err := met_mat_raiz(nil, hrp.Tupla{hrp.Decimal(4), hrp.Decimal(0)})
	if err == nil {
		t.Fatal("esperava erro para raiz com índice zero")
	}
	var erroHrp *hrp.Erro
	if !errors.As(err, &erroHrp) || erroHrp.Tipo() != hrp.ValorErro {
		t.Fatalf("esperava ValorErro para índice zero, obteve: %v", err)
	}

	// raiz(-4, 2) -> erro radicando negativo com índice par
	_, err = met_mat_raiz(nil, hrp.Tupla{hrp.Decimal(-4), hrp.Decimal(2)})
	if err == nil {
		t.Fatal("esperava erro para raiz quadrada de número negativo")
	}
	if !errors.As(err, &erroHrp) || erroHrp.Tipo() != hrp.ValorErro {
		t.Fatalf("esperava ValorErro para raiz par de negativo, obteve: %v", err)
	}

	// raiz(-8, 3) -> deve retornar -2
	res, err := met_mat_raiz(nil, hrp.Tupla{hrp.Decimal(-8), hrp.Decimal(3)})
	if err != nil {
		t.Fatalf("erro inesperado para raiz(-8, 3): %v", err)
	}
	dec, ok := res.(hrp.Decimal)
	if !ok || math.Abs(float64(dec)-(-2.0)) > 1e-9 {
		t.Fatalf("esperava -2.0, obteve %v", res)
	}
}

func TestMatLogCasosInvalidos(t *testing.T) {
	// log(0)
	_, err := met_mat_log(nil, hrp.Tupla{hrp.Decimal(0)})
	if err == nil {
		t.Fatal("esperava erro para log(0)")
	}
	var erroHrp *hrp.Erro
	if !errors.As(err, &erroHrp) || erroHrp.Tipo() != hrp.ValorErro {
		t.Fatalf("esperava ValorErro para log(0), obteve: %v", err)
	}

	// log(-1)
	_, err = met_mat_log(nil, hrp.Tupla{hrp.Decimal(-1)})
	if err == nil {
		t.Fatal("esperava erro para log(-1)")
	}
	if !errors.As(err, &erroHrp) || erroHrp.Tipo() != hrp.ValorErro {
		t.Fatalf("esperava ValorErro para log(-1), obteve: %v", err)
	}
}

func TestMatPisoTetoNaNInf(t *testing.T) {
	nan := hrp.Decimal(math.NaN())
	inf := hrp.Decimal(math.Inf(1))

	if _, err := met_mat_piso(nil, hrp.Tupla{nan}); err == nil {
		t.Fatal("esperava erro em piso(NaN)")
	}
	if _, err := met_mat_piso(nil, hrp.Tupla{inf}); err == nil {
		t.Fatal("esperava erro em piso(Inf)")
	}

	if _, err := met_mat_teto(nil, hrp.Tupla{nan}); err == nil {
		t.Fatal("esperava erro em teto(NaN)")
	}
	if _, err := met_mat_teto(nil, hrp.Tupla{inf}); err == nil {
		t.Fatal("esperava erro em teto(Inf)")
	}
}
