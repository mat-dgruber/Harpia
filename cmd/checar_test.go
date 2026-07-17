package cmd

import (
	"testing"

	"github.com/mat-dgruber/Harpia/ptst"
)

func TestLinterShadowingWarning(t *testing.T) {
	codigo := `
	var x = 10;
	se (Verdadeiro) {
		var x = 20;
	}
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	linter := &Linter{}
	linter.Checar(ast)

	if len(linter.Erros) != 1 {
		t.Fatalf("Esperava exatamente 1 ocorrência, obteve %d", len(linter.Erros))
	}

	aviso := linter.Erros[0]
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
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	linter := &Linter{}
	linter.Checar(ast)

	if len(linter.Erros) != 1 {
		t.Fatalf("Esperava exatamente 1 erro, obteve %d", len(linter.Erros))
	}

	erro := linter.Erros[0]
	if erro.Severity != 1 {
		t.Errorf("Esperava Severity=1 (Erro), obteve %d", erro.Severity)
	}
}
