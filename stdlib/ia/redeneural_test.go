package ia

import (
	"math"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestSigmoideESimilaridade(t *testing.T) {
	sig, err := met_ia_sigmoide(nil, hrp.Tupla{hrp.Decimal(0)})
	if err != nil {
		t.Fatalf("Erro no sigmoide: %v", err)
	}
	if math.Abs(float64(sig.(hrp.Decimal))-0.5) > 1e-6 {
		t.Errorf("Sigmoide de 0 deveria ser 0.5, obtido: %f", sig)
	}

	v1 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(1), hrp.Decimal(0)}}
	v2 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(1), hrp.Decimal(0)}}
	cos, err := met_ia_similaridadeCosseno(nil, hrp.Tupla{v1, v2})
	if err != nil {
		t.Fatalf("Erro na similaridadeCosseno: %v", err)
	}
	if math.Abs(float64(cos.(hrp.Decimal))-1.0) > 1e-6 {
		t.Errorf("Similaridade cosseno de vetores iguais deveria ser 1.0, obtido: %f", cos)
	}
}
