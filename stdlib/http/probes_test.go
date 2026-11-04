package http

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestProbesLivezAndReadyz(t *testing.T) {
	obj, err := TipoServidor.Nova(nil)
	if err != nil {
		t.Fatalf("erro ao criar Servidor: %v", err)
	}
	srv := obj.(*Servidor)

	escutarMetodo, err := srv.M__obtem_attributo__("escutar")
	if err != nil {
		t.Fatalf("erro ao obter metodo escutar: %v", err)
	}
	// Registra rotas padrao (/livez e /readyz) sem bloquear na porta 0
	_, _ = hrp.Chamar(escutarMetodo, hrp.Tupla{hrp.Texto("0"), hrp.Falso})
	if srv.server != nil {
		_ = srv.server.Close()
	}

	// 1. Probe /livez deve responder 200
	livezHandler := srv.rotas["GET"]["/livez"]
	if livezHandler == nil {
		t.Fatal("handler /livez nao registrado")
	}
	resLivez := &Resposta{Status: http.StatusOK, Cabecalho: hrp.NewMapaVazio()}
	if _, err := hrp.Chamar(livezHandler, hrp.Tupla{&Requisicao{}, resLivez}); err != nil {
		t.Fatalf("erro ao chamar /livez: %v", err)
	}
	if int(resLivez.Status) != http.StatusOK {
		t.Fatalf("esperado 200 em /livez, obtido %d", resLivez.Status)
	}

	// 2. Probe /readyz com infraestrutura disponivel deve responder 200
	readyzHandler := srv.rotas["GET"]["/readyz"]
	if readyzHandler == nil {
		t.Fatal("handler /readyz nao registrado")
	}
	resReadyOk := &Resposta{Status: http.StatusOK, Cabecalho: hrp.NewMapaVazio()}
	if _, err := hrp.Chamar(readyzHandler, hrp.Tupla{&Requisicao{}, resReadyOk}); err != nil {
		t.Fatalf("erro ao chamar /readyz disponivel: %v", err)
	}
	if int(resReadyOk.Status) != http.StatusOK {
		t.Fatalf("esperado 200 em /readyz disponivel, obtido %d", resReadyOk.Status)
	}

	// 3. Probe /readyz com infraestrutura indisponivel deve responder 503 sem expor erro interno
	srv.msgsErrosUltimo = fmt.Errorf("falha de conexao com banco sqlite")
	resReadyFail := &Resposta{Status: http.StatusOK, Cabecalho: hrp.NewMapaVazio()}
	if _, err := hrp.Chamar(readyzHandler, hrp.Tupla{&Requisicao{}, resReadyFail}); err != nil {
		t.Fatalf("erro ao chamar /readyz indisponivel: %v", err)
	}
	if int(resReadyFail.Status) != http.StatusServiceUnavailable {
		t.Fatalf("esperado 503 em /readyz indisponivel, obtido %d", resReadyFail.Status)
	}
	if string(resReadyFail.Corpo) != `{"status":"not_ready"}` {
		t.Fatalf("corpo inesperado em /readyz 503: %s", resReadyFail.Corpo)
	}
}
