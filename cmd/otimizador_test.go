package cmd

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
)

func TestOtimizador_DCE_VariavelNaoReferenciada(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	var usada = 10
	var inutil = 20
	var soma = usada + 5
	`

	astNode, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("erro ao gerar AST: %v", err)
	}

	prog := astNode.(*parser.Programa)
	opt := Otimizar(prog)

	transpiler := &TranspilerNative{}
	goCode := transpiler.GenerateFullCode(opt)

	t.Logf("CÓDIGO GERADO: %s", goCode)

	if !strings.Contains(goCode, "usada") {
		t.Errorf("variável 'usada' deveria estar presente")
	}

	if strings.Contains(goCode, "inutil") {
		t.Errorf("variável 'inutil' deveria ter sido removida pelo DCE")
	}
}

func TestOtimizador_BranchesConstantes(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	var resultado = 0
	se (Falso) {
		resultado = 1
	} senao {
		resultado = 2
	}
	`

	astNode, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("erro ao gerar AST: %v", err)
	}

	prog := astNode.(*parser.Programa)
	opt := Otimizar(prog)

	transpiler := &TranspilerNative{}
	goCode := transpiler.GenerateFullCode(opt)

	t.Logf("CÓDIGO GERADO BRANCHES: %s", goCode)

	if strings.Contains(goCode, "var_resultado = hrp.Inteiro(1)") {
		t.Errorf("bloco 'se (Falso)' morto não deveria ter sido transpilado")
	}

	if !strings.Contains(goCode, "var_resultado = v_2") {
		t.Errorf("bloco senao alternativo deveria ter sido promovido")
	}
}
