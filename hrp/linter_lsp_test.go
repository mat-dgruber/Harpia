package hrp_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// LSPRangePosition mapeia uma coordenada de posição compatível com o protocolo LSP (Language Server Protocol).
type LSPRangePosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// LSPRange representa a faixa de coordenadas espaciais de um erro (início e fim) no protocolo LSP.
type LSPRange struct {
	Start LSPRangePosition `json:"start"`
	End   LSPRangePosition `json:"end"`
}

// LSPDiagnostic define a estrutura padrão de diagnóstico enviada pelo linter quando o formato JSON é solicitado.
type LSPDiagnostic struct {
	Range    LSPRange `json:"range"`
	Severity int      `json:"severity"`
	Code     string   `json:"code"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

// TestLinterLSPDiagnosticsJSON realiza um teste de roundtrip executando a ferramenta CLI 'checar'
// com a flag '--formato=json' para garantir que os diagnósticos de linter/AST sejam gerados
// no formato JSON estrito compatível com a extensão oficial do VS Code e outros clientes LSP.
func TestLinterLSPDiagnosticsJSON(t *testing.T) {
	dir, err := os.MkdirTemp("", "Harpia_linter_lsp_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Código com erro semântico de variável não declarada
	codigoErrado := `
	b = 20 # b não foi declarado
	`


	caminhoArquivo := filepath.Join(dir, "teste.hrp")
	if err := os.WriteFile(caminhoArquivo, []byte(codigoErrado), 0644); err != nil {
		t.Fatal(err)
	}

	// Executa a checagem diretamente chamando o binário compilado
	// Isso faz o roundtrip completo testando flags do Cobra
	execPath, err := filepath.Abs("../Harpia")
	if err != nil {
		t.Fatal(err)
	}

	// ponytail: evita falhas se o binário de produção ainda não foi compilado no root
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		t.Skipf("Ignorando teste de linter LSP: binário '%s' não encontrado. Compile com 'go build -o Harpia main.go' primeiro.", execPath)
	}

	cmdRun := exec.Command(execPath, "checar", caminhoArquivo, "--formato=json")
	cmdRun.Dir = dir
	output, _ := cmdRun.CombinedOutput()

	var diagnostics []LSPDiagnostic
	if err := json.Unmarshal(output, &diagnostics); err != nil {
		t.Fatalf("Erro ao decodificar JSON do LSP Diagnostics: %v\nSaída obtida: %s", err, string(output))
	}

	if len(diagnostics) != 1 {
		t.Fatalf("Esperava exatamente 1 diagnóstico de erro, mas obtive: %d", len(diagnostics))
	}

	diag := diagnostics[0]
	if diag.Code != "HRP-0005" {
		t.Errorf("Código de erro incorreto. Esperava 'HRP-0005', obtive '%s'", diag.Code)
	}

	if diag.Range.Start.Line != 1 && diag.Range.Start.Line != 2 {
		t.Errorf("Linha do erro incorreta. Esperava 1 ou 2, obtive %d", diag.Range.Start.Line)
	}

}
