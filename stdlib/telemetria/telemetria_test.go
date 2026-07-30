package telemetria

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestTracerESpan(t *testing.T) {
	tracerObj, err := met_novo_tracer(nil, hrp.Tupla{hrp.Texto("meu-servico")})
	if err != nil {
		t.Fatalf("Erro ao criar tracer: %v", err)
	}

	tracer := tracerObj.(*Tracer)
	spanMetodo, errSpan := tracer.M__obtem_attributo__("iniciar_span")
	if errSpan != nil {
		t.Fatalf("Erro ao obter iniciar_span: %v", errSpan)
	}

	spanObj, errCall := hrp.Chamar(spanMetodo, hrp.Tupla{hrp.Texto("processar_pedido")})
	if errCall != nil {
		t.Fatalf("Erro ao chamar iniciar_span: %v", errCall)
	}

	span := spanObj.(*Span)
	if span.Servico != "meu-servico" || span.Nome != "processar_pedido" {
		t.Errorf("Dados do span incorretos: %+v", span)
	}

	finMetodo, errFin := span.M__obtem_attributo__("finalizar")
	if errFin != nil {
		t.Fatalf("Erro ao obter finalizar: %v", errFin)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_, errFinCall := hrp.Chamar(finMetodo, hrp.Tupla{hrp.Texto("OK")})

	w.Close()
	os.Stdout = oldStdout

	if errFinCall != nil {
		t.Fatalf("Erro ao chamar finalizar: %v", errFinCall)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"service":"meu-servico"`) || !strings.Contains(output, `"status":"OK"`) {
		t.Errorf("Saída do span incorreta, obtive:\n%s", output)
	}
}

func TestMetrica(t *testing.T) {
	metricaObj, err := met_nova_metrica(nil, hrp.Tupla{hrp.Texto("requisicoes"), hrp.Texto("contador")})
	if err != nil {
		t.Fatalf("Erro ao criar metrica: %v", err)
	}

	metrica := metricaObj.(*Metrica)
	regMetodo, errReg := metrica.M__obtem_attributo__("registrar")
	if errReg != nil {
		t.Fatalf("Erro ao obter registrar: %v", errReg)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_, errCall := hrp.Chamar(regMetodo, hrp.Tupla{hrp.Decimal(10.5), hrp.Texto("GET /api")})

	w.Close()
	os.Stdout = oldStdout

	if errCall != nil {
		t.Fatalf("Erro ao chamar registrar: %v", errCall)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"tag": "GET /api"`) || !strings.Contains(output, `"metric": "requisicoes"`) {
		t.Errorf("Saída da metrica incorreta, obtive:\n%s", output)
	}
}
