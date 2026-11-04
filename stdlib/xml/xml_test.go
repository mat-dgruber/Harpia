package xml

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestXmlAnalisarESerializar(t *testing.T) {
	xmlStr := "<config><nome>Harpia</nome></config>"
	obj, err := met_xml_analisar(nil, hrp.Tupla{hrp.Texto(xmlStr)})
	if err != nil {
		t.Fatalf("Erro ao analisar XML: %v", err)
	}

	mapa, ok := obj.(hrp.Mapa)
	if !ok {
		t.Fatalf("Esperava Mapa, obteve %T", obj)
	}

	cfg, ok := mapa["config"].(hrp.Mapa)
	if !ok {
		t.Fatalf("Esperava Mapa para config, obteve %T", mapa["config"])
	}

	if cfg["nome"] != hrp.Texto("Harpia") {
		t.Errorf("Esperava nome Harpia, obteve %v", cfg["nome"])
	}

	ser, err := met_xml_serializar(nil, hrp.Tupla{mapa, hrp.Texto("root")})
	if err != nil {
		t.Fatalf("Erro ao serializar XML: %v", err)
	}
	serStr := string(ser.(hrp.Texto))
	if !strings.Contains(serStr, "<root>") || !strings.Contains(serStr, "</root>") {
		t.Errorf("Esperava tags root na serialização, obteve: %s", serStr)
	}
}
