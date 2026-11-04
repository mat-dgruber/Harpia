package http

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestServirAppSiblingTraversalEXSS(t *testing.T) {
	baseDir := t.TempDir()
	distDir := filepath.Join(baseDir, "dist")
	distSecretDir := filepath.Join(baseDir, "dist_secret")

	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatalf("erro ao criar dist: %v", err)
	}
	if err := os.MkdirAll(distSecretDir, 0755); err != nil {
		t.Fatalf("erro ao criar dist_secret: %v", err)
	}

	indexHtml := "<!DOCTYPE html><html><head><title>Harpia App</title></head><body><div id=\"app\"></div></body></html>"
	if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte(indexHtml), 0644); err != nil {
		t.Fatalf("erro ao escrever index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(distSecretDir, "key.txt"), []byte("SEGREDO"), 0644); err != nil {
		t.Fatalf("erro ao escrever key.txt: %v", err)
	}

	meta := hrp.Mapa{"titulo": hrp.Texto("</title><script>alert(1)</script>")}
	handler := ServirAppHandler(distDir, hrp.Nulo, meta)

	// 1. Testar sibling directory traversal
	reqTraversal := &Requisicao{
		Caminho: hrp.Texto("../dist_secret/key.txt"),
	}
	resTraversal := &Resposta{
		Cabecalho: hrp.NewMapaVazio(),
	}
	_, err := hrp.Chamar(handler, hrp.Tupla{reqTraversal, resTraversal})
	if err != nil {
		t.Fatalf("erro ao chamar handler para traversal: %v", err)
	}
	if strings.Contains(string(resTraversal.Corpo), "SEGREDO") {
		t.Errorf("vulnerabilidade Sibling Path Traversal: retornou conteúdo secreto")
	}
	if resTraversal.Status != 403 {
		t.Errorf("esperava status 403 para sibling traversal, obteve %d", resTraversal.Status)
	}

	// 2. Testar escape XSS no título
	reqXSS := &Requisicao{
		Caminho: hrp.Texto("/"),
	}
	resXSS := &Resposta{
		Cabecalho: hrp.NewMapaVazio(),
	}
	_, err = hrp.Chamar(handler, hrp.Tupla{reqXSS, resXSS})
	if err != nil {
		t.Fatalf("erro ao chamar handler para XSS: %v", err)
	}
	corpoHtml := string(resXSS.Corpo)
	if strings.Contains(corpoHtml, "<script>alert(1)</script>") {
		t.Errorf("vulnerabilidade XSS: tag <script> crua encontrada no HTML gerado")
	}
	if !strings.Contains(corpoHtml, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("esperava HTML escapado no título")
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
	srv := &Servidor{
		rotas: map[string]map[string]hrp.Objeto{
			"GET": {
				"/usuarios": hrp.NewMetodoOuPanic("listar", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
					return hrp.Nulo, nil
				}, "Listar todos os usuarios"),
			},
		},
		rotasMeta: map[string]map[string]*RotaMeta{
			"GET": {
				"/usuarios": {
					Descricao: "Listar todos os usuarios",
					Tags:      []string{"Usuarios"},
				},
			},
		},
		metricas: make(map[string]int64),
	}

	resOpenAPI, err := met_gerar_openapi(nil, hrp.Tupla{srv})
	if err != nil {
		t.Fatalf("erro ao gerar openapi: %v", err)
	}
	specStr := string(resOpenAPI.(hrp.Texto))
	var openapiJSON map[string]interface{}
	if err := json.Unmarshal([]byte(specStr), &openapiJSON); err != nil {
		t.Fatalf("JSON OpenAPI 3.0 gerado é inválido: %v", err)
	}
	if openapiJSON["openapi"] != "3.0.0" {
		t.Errorf("esperava versao 3.0.0 no OpenAPI, obteve: %v", openapiJSON["openapi"])
	}
	paths, ok := openapiJSON["paths"].(map[string]interface{})
	if !ok || paths["/usuarios"] == nil {
		t.Errorf("esperava rota '/usuarios' em paths do OpenAPI, obteve: %v", openapiJSON["paths"])
	}

	if !strings.Contains(htmlDocsTemplate, "<!DOCTYPE html>") || !strings.Contains(htmlDocsTemplate, "Harpia API Console") {
		t.Errorf("htmlDocsTemplate inválido ou não contém cabeçalho da API Console")
	}
}
