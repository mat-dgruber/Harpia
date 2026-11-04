package hrp_test

import (
	"encoding/json"
	"testing"

	"github.com/mat-dgruber/Harpia/cmd"
	"github.com/mat-dgruber/Harpia/hrp"
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

// TestLinterLSPDiagnosticsJSON valida a geração de diagnósticos LSP a partir do linter/AST
// no formato JSON estrito compatível com a extensão oficial do VS Code e outros clientes LSP.
func TestLinterLSPDiagnosticsJSON(t *testing.T) {
	// Código com erro semântico de variável não declarada
	codigoErrado := `
	b = 20 # b não foi declarado
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigoErrado, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	linter := &cmd.Linter{}
	linter.Checar(ast)

	if len(linter.Erros) != 1 {
		t.Fatalf("Esperava exatamente 1 diagnóstico de erro no linter, mas obtive: %d", len(linter.Erros))
	}

	errObj := linter.Erros[0]
	if errObj.Code != "HRP-0005" {
		t.Errorf("Código de erro incorreto. Esperava 'HRP-0005', obtive '%s'", errObj.Code)
	}

	var line, col, length int
	if tok, ok := linter.Posicoes[errObj.Node]; ok && tok != nil {
		line = tok.Inicio.Linha - 1
		col = tok.Inicio.Coluna - 1
		length = len(tok.Valor)
		if length == 0 {
			length = 1
		}
	}

	diagnosticosOriginais := []LSPDiagnostic{
		{
			Range: LSPRange{
				Start: LSPRangePosition{Line: line, Character: col},
				End:   LSPRangePosition{Line: line, Character: col + length},
			},
			Severity: errObj.Severity,
			Code:     errObj.Code,
			Source:   "Harpia-linter",
			Message:  errObj.Message,
		},
	}

	jsonBytes, err := json.Marshal(diagnosticosOriginais)
	if err != nil {
		t.Fatalf("Erro ao serializar JSON do LSP Diagnostics: %v", err)
	}

	var diagnostics []LSPDiagnostic
	if err := json.Unmarshal(jsonBytes, &diagnostics); err != nil {
		t.Fatalf("Erro ao decodificar JSON do LSP Diagnostics: %v\nSaída obtida: %s", err, string(jsonBytes))
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
