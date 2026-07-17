package ffi

import (
	"testing"

	"github.com/mat-dgruber/Harpia/ptst"
)

func TestFFIMockSoma(t *testing.T) {
	libObj, err := met_ffi_abrir(nil, ptst.Tupla{ptst.Texto("libmath_mock.so")})
	if err != nil {
		t.Fatalf("Erro ao abrir lib: %v", err)
	}

	lib := libObj.(*Biblioteca)
	obtMetodo, errAtt := lib.M__obtem_attributo__("obterFuncao")
	if errAtt != nil {
		t.Fatalf("Erro ao obter método obterFuncao: %v", errAtt)
	}

	funcObj, errCall := ptst.Chamar(obtMetodo, ptst.Tupla{ptst.Texto("soma")})
	if errCall != nil {
		t.Fatalf("Erro ao chamar obterFuncao: %v", errCall)
	}

	somaFunc := funcObj.(*FuncaoFFI)
	res, errSoma := ptst.Chamar(somaFunc, ptst.Tupla{ptst.Decimal(10.5), ptst.Decimal(20.3)})
	if errSoma != nil {
		t.Fatalf("Erro ao chamar função ffi soma: %v", errSoma)
	}

	if float64(res.(ptst.Decimal)) != 30.8 {
		t.Errorf("Resultado soma incorreto, obteve: %v", res)
	}
}
