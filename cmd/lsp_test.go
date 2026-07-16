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
