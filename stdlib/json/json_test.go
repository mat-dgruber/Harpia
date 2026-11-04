package json

import (
	"math"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestJSONAnalisarESerializar(t *testing.T) {
	jsonStr := `{"nome":"teste","idade":42,"pi":3.14,"grande":10000000000000000}`
	res, err := met_json_analisar(nil, hrp.Tupla{hrp.Texto(jsonStr)})
	if err != nil {
		t.Fatalf("Erro ao analisar JSON: %v", err)
	}

	mapa, ok := res.(hrp.Mapa)
	if !ok {
		t.Fatalf("Esperava Mapa, obteve %T", res)
	}

	if mapa["idade"] != hrp.Inteiro(42) {
		t.Errorf("Esperava Inteiro 42, obteve %v", mapa["idade"])
	}

	if _, ok := mapa["pi"].(hrp.Decimal); !ok {
		t.Errorf("Esperava Decimal para pi, obteve %T", mapa["pi"])
	}

	// 10000000000000000 excede o limite seguro de inteiros float64 (53 bits, 9007199254740992), deve virar Decimal
	if _, ok := mapa["grande"].(hrp.Decimal); !ok {
		t.Errorf("Esperava Decimal para valor acima do limite seguro, obteve %T", mapa["grande"])
	}

	// Serializar de volta
	resSer, err := met_json_serializar(nil, hrp.Tupla{mapa})
	if err != nil {
		t.Fatalf("Erro ao serializar: %v", err)
	}
	if _, ok := resSer.(hrp.Texto); !ok {
		t.Errorf("Esperava Texto, obteve %T", resSer)
	}
}

func TestJSONConverteNaN(t *testing.T) {
	obj := converteParahrp(math.NaN())
	if dec, ok := obj.(hrp.Decimal); !ok || !math.IsNaN(float64(dec)) {
		t.Errorf("Esperava Decimal NaN, obteve %v", obj)
	}
}
