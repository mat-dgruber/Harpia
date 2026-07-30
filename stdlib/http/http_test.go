package http

import (
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestServidorEHMAC(t *testing.T) {
	// Teste de HMAC
	chave := hrp.Texto("segredo123")
	msg := hrp.Texto("mensagem importante")

	resSig, err := met_assinar_hmac(nil, hrp.Tupla{chave, msg})
	if err != nil {
		t.Fatalf("erro ao assinar HMAC: %v", err)
	}

	sigStr := resSig.(hrp.Texto)
	if len(sigStr) == 0 {
		t.Fatalf("assinatura HMAC vazia")
	}

	// Verificar HMAC correto
	resVerif, err := met_verificar_hmac(nil, hrp.Tupla{chave, msg, sigStr})
	if err != nil {
		t.Fatalf("erro ao verificar HMAC: %v", err)
	}
	if bool(resVerif.(hrp.Booleano)) != true {
		t.Errorf("esperava HMAC valido")
	}

	// Verificar HMAC incorreto
	resVerifIncorreta, err := met_verificar_hmac(nil, hrp.Tupla{chave, hrp.Texto("mensagem alterada"), sigStr})
	if err != nil {
		t.Fatalf("erro ao verificar HMAC incorreto: %v", err)
	}
	if bool(resVerifIncorreta.(hrp.Booleano)) != false {
		t.Errorf("esperava HMAC invalido")
	}
}

func TestRoteamentoEForca(t *testing.T) {
	if f := calcularForcaRota("/usuarios/:id"); f != 15 {
		t.Errorf("forca esperada 15, obteve %d", f)
	}
	if f := calcularForcaRota("/*"); f != 0 {
		t.Errorf("forca esperada 0, obteve %d", f)
	}

	match, params := matchRoute("/api/v1/usuarios/:id", "/api/v1/usuarios/42")
	if !match {
		t.Fatalf("esperava match na rota")
	}
	if params["id"] != "42" {
		t.Errorf("esperava id=42, obteve %s", params["id"])
	}
}

func TestServidorHandler(t *testing.T) {
	srv := &Servidor{
		rotas:       make(map[string]map[string]hrp.Objeto),
		rotasMeta:   make(map[string]map[string]*RotaMeta),
		middlewares: make([]hrp.Objeto, 0),
		metricas:    make(map[string]int64),
	}

	handler := hrp.NewMetodoOuPanic("teste_handler", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		res := args[1].(*Resposta)
		res.Status = 200
		res.Corpo = hrp.Texto("Olá Harpia!")
		return hrp.Nulo, nil
	}, "Handler de teste")

	srv.registrarRota("GET", "/ping", handler, nil)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	// Simula execução básica do mux
	srv.mu.RLock()
	h := srv.rotas["GET"]["/ping"]
	srv.mu.RUnlock()

	if h == nil {
		t.Fatalf("rota /ping não registrada")
	}

	resObj := &Resposta{Status: 200, Cabecalho: hrp.NewMapaVazio(), Corpo: ""}
	reqObj := &Requisicao{Metodo: hrp.Texto("GET"), Caminho: hrp.Texto("/ping")}

	_, err := hrp.Chamar(h, hrp.Tupla{reqObj, resObj})
	if err != nil {
		t.Fatalf("erro ao chamar handler: %v", err)
	}

	if string(resObj.Corpo) != "Olá Harpia!" {
		t.Errorf("corpo da resposta incorreto: %s", string(resObj.Corpo))
	}
	_ = w
	_ = req
}

func TestServirAppPathTraversal(t *testing.T) {
	dirTemp := t.TempDir()
	req := &Requisicao{
		Caminho: hrp.Texto("../../../etc/passwd"),
	}
	res := &Resposta{
		Cabecalho: hrp.NewMapaVazio(),
	}

	handler := ServirAppHandler(dirTemp, hrp.Nulo, hrp.Nulo)
	_, err := hrp.Chamar(handler, hrp.Tupla{req, res})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if res.Status != 403 {
		t.Errorf("esperava status 403 para path traversal, obteve %d", res.Status)
	}
	if string(res.Corpo) != "Acesso negado" {
		t.Errorf("esperava corpo 'Acesso negado', obteve %s", string(res.Corpo))
	}
}

func TestHMACInvalidoArgs(t *testing.T) {
	_, err := met_assinar_hmac(nil, hrp.Tupla{hrp.Texto("chave")})
	if err == nil {
		t.Errorf("esperava erro por falta de argumentos em assinar_hmac")
	}

	_, err = met_verificar_hmac(nil, hrp.Tupla{hrp.Texto("chave"), hrp.Texto("msg")})
	if err == nil {
		t.Errorf("esperava erro por falta de argumentos em verificar_hmac")
	}
}

func TestHttpRequisitarInvalidoArgs(t *testing.T) {
	_, err := met_http_requisitar(nil, hrp.Tupla{hrp.Texto("GET")})
	if err == nil {
		t.Errorf("esperava erro por falta de argumentos em requisitar")
	}
}

func TestGerarOpenAPI(t *testing.T) {
	srv := &Servidor{
		rotas:     make(map[string]map[string]hrp.Objeto),
		rotasMeta: make(map[string]map[string]*RotaMeta),
		metricas:  make(map[string]int64),
	}

	res, err := met_gerar_openapi(nil, hrp.Tupla{srv})
	if err != nil {
		t.Fatalf("erro ao gerar openapi: %v", err)
	}

	specStr := string(res.(hrp.Texto))
	if !strings.Contains(specStr, `"openapi": "3.0.0"`) {
		t.Errorf("spec OpenAPI não contém versão 3.0.0")
	}
}

func TestAuxiliaresDocs(t *testing.T) {
	_ = fmt.Sprintf("%v", io.EOF)
}
