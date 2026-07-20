package ffi

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestFFIMockSoma(t *testing.T) {
	libObj, err := met_ffi_abrir(nil, hrp.Tupla{hrp.Texto("libmath_mock.so")})
	if err != nil {
		t.Fatalf("Erro ao abrir lib: %v", err)
	}

	lib := libObj.(*Biblioteca)
	obtMetodo, errAtt := lib.M__obtem_attributo__("obterFuncao")
	if errAtt != nil {
		t.Fatalf("Erro ao obter método obterFuncao: %v", errAtt)
	}

	funcObj, errCall := hrp.Chamar(obtMetodo, hrp.Tupla{hrp.Texto("soma")})
	if errCall != nil {
		t.Fatalf("Erro ao chamar obterFuncao: %v", errCall)
	}

	somaFunc := funcObj.(*FuncaoFFI)
	res, errSoma := hrp.Chamar(somaFunc, hrp.Tupla{hrp.Decimal(10.5), hrp.Decimal(20.3)})
	if errSoma != nil {
		t.Fatalf("Erro ao chamar função ffi soma: %v", errSoma)
	}

	if float64(res.(hrp.Decimal)) != 30.8 {
		t.Errorf("Resultado soma incorreto, obteve: %v", res)
	}
}
