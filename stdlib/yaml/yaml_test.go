package yaml

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestYamlAnalisarESerializar(t *testing.T) {
	yamlStr := "nome: Harpia\nversao: 1\n"
	obj, err := met_yaml_analisar(nil, hrp.Tupla{hrp.Texto(yamlStr)})
	if err != nil {
		t.Fatalf("Erro ao analisar YAML: %v", err)
	}

	mapa, ok := obj.(hrp.Mapa)
	if !ok {
		t.Fatalf("Esperava Mapa, obteve %T", obj)
	}

	if mapa["nome"] != hrp.Texto("Harpia") {
		t.Errorf("Esperava nome Harpia, obteve %v", mapa["nome"])
	}

	ser, err := met_yaml_serializar(nil, hrp.Tupla{mapa})
	if err != nil {
		t.Fatalf("Erro ao serializar YAML: %v", err)
	}
	if ser == nil {
		t.Errorf("Esperava texto serializado não nulo")
	}
}
