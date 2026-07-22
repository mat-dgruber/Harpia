package cmd

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestLinterShadowingWarning(t *testing.T) {
	codigo := `
	var x = 10;
	se (Verdadeiro) {
		var x = 20;
		imprimir(x);
	}
	imprimir(x);
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	linter := &Linter{}
	linter.Checar(ast)

	// Filter out unused variable warnings for this shadowing test
	var errorsFiltered []LinterError
	for _, e := range linter.Erros {
		if e.Code != "HRP-0006" {
			errorsFiltered = append(errorsFiltered, e)
		}
	}

	if len(errorsFiltered) != 1 {
		t.Fatalf("Esperava exatamente 1 ocorrência, obteve %d", len(errorsFiltered))
	}

	aviso := errorsFiltered[0]
	if aviso.Severity != 2 {
		t.Errorf("Esperava Severity=2 (Aviso), obteve %d", aviso.Severity)
	}

	if aviso.Code != "HRP-0002" {
		t.Errorf("Esperava Code=PSC-0002, obteve %s", aviso.Code)
	}
}

func TestLinterShadowingErroMesmoEscopo(t *testing.T) {
	codigo := `
	var x = 10;
	var x = 20;
	imprimir(x);
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	linter := &Linter{}
	linter.Checar(ast)

	var errorsFiltered []LinterError
	for _, e := range linter.Erros {
		if e.Code != "HRP-0006" {
			errorsFiltered = append(errorsFiltered, e)
		}
	}

	if len(errorsFiltered) != 1 {
		t.Fatalf("Esperava exatamente 1 erro, obteve %d", len(errorsFiltered))
	}

	erro := errorsFiltered[0]
	if erro.Severity != 1 {
		t.Errorf("Esperava Severity=1 (Erro), obteve %d", erro.Severity)
	}
}
