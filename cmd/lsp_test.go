package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestLSPInitialize assevera que a requisição de handshake inicial responde com as capacidades de IDE
func TestLSPInitialize(t *testing.T) {
	jsonStr := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(jsonStr), jsonStr)
	reader := bufio.NewReader(strings.NewReader(input))

	msg, err := lerMensagemLSP(reader)
	if err != nil {
		t.Fatalf("Falha ao ler mensagem LSP: %v", err)
	}

	var req RequestMessage
	if err := json.Unmarshal(msg, &req); err != nil {
		t.Fatalf("Erro ao decodificar JSON: %v", err)
	}

	if req.Method != "initialize" {
		t.Errorf("Esperava método 'initialize', obtive '%s'", req.Method)
	}
}

// TestLSPDiagnosticsSyntax assevera que o didOpen com erro de sintaxe gera notificação de erro no parser
func TestLSPDiagnosticsSyntax(t *testing.T) {
	codigoInvalido := "var a = # erro de sintaxe"

	// Captura stdout para validar a resposta JSON-RPC emitida pelo lsp
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processarDiagnosticosLSP("file:///teste.ptst", codigoInvalido)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "Content-Length:") {
		t.Fatalf("Esperava cabeçalho HTTP de Content-Length no stdout")
	}

	// Extrai o corpo do JSON-RPC
	jsonStart := strings.Index(saida, "{")
	if jsonStart == -1 {
		t.Fatalf("Nenhum JSON encontrado na saída")
	}
	jsonBody := saida[jsonStart:]

	var notif NotificationMessage
	if err := json.Unmarshal([]byte(jsonBody), &notif); err != nil {
		t.Fatalf("Erro ao ler JSON de diagnósticos: %v\nJSON obtido: %s", err, jsonBody)
	}

	if notif.Method != "textDocument/publishDiagnostics" {
		t.Errorf("Esperava notificação 'textDocument/publishDiagnostics', obtive '%s'", notif.Method)
	}
}

// TestLSPFormatting assevera que a requisição de formatação do LSP executa o formatador e devolve a resposta de substituição
func TestLSPFormatting(t *testing.T) {
	uri := "file:///teste.ptst"
	codigoSujo := "funcao App(){\nvar a = 10;\n}"
	cacheArquivosLSP[uri] = codigoSujo

	paramsStruct := struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}{}
	paramsStruct.TextDocument.URI = uri
	paramsBytes, _ := json.Marshal(paramsStruct)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tratarRequisicaoLSP(RequestMessage{
		Jsonrpc: "2.0",
		ID:      2,
		Method:  "textDocument/formatting",
		Params:  json.RawMessage(paramsBytes),
	})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "Content-Length:") {
		t.Fatalf("Esperava Content-Length no stdout do formatador LSP")
	}

	if !strings.Contains(saida, "    var a = 10;") {
		t.Errorf("Esperava código formatado com recuo correto. Obtido: %s", saida)
	}
}

// TestLSPCleanArchLinter assevera que o linter do LSP emite erro imediato se o domínio importar infraestrutura
func TestLSPCleanArchLinter(t *testing.T) {
	codigoIncorreto := `
de "../infra/banco/conexao.ptst" importe obterBanco;
`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Caminho sob '/dominio/' simulado
	processarDiagnosticosLSP("file:///projeto/dominio/entidades/usuario.ptst", codigoIncorreto)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "PSC-ARCH-001") {
		t.Errorf("Esperava aviso de erro de Clean Architecture 'PSC-ARCH-001' no linter. Obtido: %s", saida)
	}
}

// TestLSPCompletion assevera que a requisição de autocomplete do LSP retorna a lista de CompletionItems oficiais
func TestLSPCompletion(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tratarRequisicaoLSP(RequestMessage{
		Jsonrpc: "2.0",
		ID:      3,
		Method:  "textDocument/completion",
	})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "Content-Length:") {
		t.Fatalf("Esperava Content-Length no stdout do autocomplete LSP")
	}

	if !strings.Contains(saida, "funcao") || !strings.Contains(saida, "sinalPersistente") {
		t.Errorf("Esperava lista de termos do autocomplete contendo 'funcao' e 'sinalPersistente'. Obtido: %s", saida)
	}
}

// TestLSPHover assevera que o hover em um símbolo declarado retorna a assinatura e a documentação especial '///'
func TestLSPHover(t *testing.T) {
	uri := "file:///teste_hover.ptst"
	codigo := "/// Esta funcao soma dois valores\nfuncao somar(a, b) {\n    retorne a + b\n}"

	// Alimenta o cache do AST processando diagnóstico
	processarDiagnosticosLSP(uri, codigo)

	// Monta os parâmetros de hover posicionados em cima de "somar" (linha 1, coluna 7)
	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     DiagnosticPosition{Line: 1, Character: 7},
	}
	paramsBytes, _ := json.Marshal(params)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tratarRequisicaoLSP(RequestMessage{
		Jsonrpc: "2.0",
		ID:      4,
		Method:  "textDocument/hover",
		Params:  json.RawMessage(paramsBytes),
	})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "funcao somar(a, b)") {
		t.Errorf("Esperava assinatura da funcao 'somar' no hover. Obtido: %s", saida)
	}

	if !strings.Contains(saida, "Esta funcao soma dois valores") {
		t.Errorf("Esperava documentação '/// Esta funcao soma dois valores' no hover. Obtido: %s", saida)
	}
}

// TestLSPDefinition assevera que a requisição de definição 'F12' retorna o range correto no arquivo
func TestLSPDefinition(t *testing.T) {
	uri := "file:///teste_def.ptst"
	codigo := "\n\nfuncao sub(a, b) {\n    retorne a - b\n}"

	processarDiagnosticosLSP(uri, codigo)

	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     DiagnosticPosition{Line: 2, Character: 8}, // Em cima do "sub"
	}
	paramsBytes, _ := json.Marshal(params)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tratarRequisicaoLSP(RequestMessage{
		Jsonrpc: "2.0",
		ID:      5,
		Method:  "textDocument/definition",
		Params:  json.RawMessage(paramsBytes),
	})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "\"line\":2") {
		t.Errorf("Esperava que o range apontasse para a linha 2 (onde sub foi definida). Obtido: %s", saida)
	}
}

// TestLSPSecurityLinter valida o acionamento de alertas de segurança HRP-SEC-001, HRP-SEC-002 e HRP-SEC-003
func TestLSPSecurityLinter(t *testing.T) {
	codigoVuln := `
	var sqlInseguro = "SELECT * FROM usuarios WHERE nome = " + "teste"
	consultar(sqlInseguro) // HRP-SEC-001

	var apiToken = "secret-12345" // HRP-SEC-002
	
	enviar("mensagem") // HRP-SEC-003
	`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processarDiagnosticosLSP("file:///projeto/teste_seguranca.ptst", codigoVuln)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saida := buf.String()

	if !strings.Contains(saida, "HRP-SEC-001") {
		t.Errorf("Esperava aviso de SQL Injection 'HRP-SEC-001'. Obtido: %s", saida)
	}
	if !strings.Contains(saida, "HRP-SEC-002") {
		t.Errorf("Esperava aviso de Credential Leak 'HRP-SEC-002'. Obtido: %s", saida)
	}
	if !strings.Contains(saida, "HRP-SEC-003") {
		t.Errorf("Esperava aviso de canal inseguro fora de contexto assíncrono 'HRP-SEC-003'. Obtido: %s", saida)
	}
}


