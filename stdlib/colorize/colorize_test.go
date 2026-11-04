package colorize

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestConverteRGBValidacao(t *testing.T) {
	// Valores válidos
	res, err := met_color_converteRGB(nil, hrp.Tupla{hrp.Inteiro(255), hrp.Inteiro(128), hrp.Inteiro(0)})
	if err != nil {
		t.Fatalf("Erro ao converter RGB válido: %v", err)
	}
	if string(res.(hrp.Texto)) == "" {
		t.Errorf("Esperava string ANSI não vazia")
	}

	// Valor negativo
	_, err = met_color_converteRGB(nil, hrp.Tupla{hrp.Inteiro(-1), hrp.Inteiro(128), hrp.Inteiro(0)})
	if err == nil {
		t.Errorf("Esperava erro para RGB negativo")
	}

	// Valor acima de 255
	_, err = met_color_converteRGB(nil, hrp.Tupla{hrp.Inteiro(256), hrp.Inteiro(128), hrp.Inteiro(0)})
	if err == nil {
		t.Errorf("Esperava erro para RGB > 255")
	}
}

func TestImprimacRetornoNulo(t *testing.T) {
	res, err := met_color_imprimac(nil, hrp.Tupla{hrp.Texto("teste")})
	if err != nil {
		t.Fatalf("Erro ao executar imprimac: %v", err)
	}
	if res != hrp.Nulo {
		t.Errorf("Esperava hrp.Nulo no retorno de imprimac, obteve %v", res)
	}
}
