//go:build !js && !wasm

package cmd

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
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

func TestLinterJSXComponentSignature(t *testing.T) {
	codigo := `
	funcao MeuComponente(nome, idade) {
		retorne <div />;
	}
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	linter := &Linter{}
	linter.Checar(ast)

	var errosJsx []LinterError
	for _, e := range linter.Erros {
		if e.Code == "HRP-JSX-001" {
			errosJsx = append(errosJsx, e)
		}
	}

	if len(errosJsx) != 1 {
		t.Fatalf("Esperava erro de assinatura de componente JSX HRP-JSX-001, mas obteve %d", len(errosJsx))
	}

	erro := errosJsx[0]
	if erro.Severity != 2 {
		t.Errorf("Esperava Severity=2 (Aviso), obteve %d", erro.Severity)
	}
}

func TestChecarArquiteturaEstritaECPF(t *testing.T) {
	// 1. checarArquitetura com aspas no caminho
	impComAspas := &parser.ImporteDe{
		Caminho: &parser.TextoLiteral{Valor: "\"../web/tela.hrp\""},
	}
	prog := &parser.Programa{
		Arquivo:     "infra/servicos/api.hrp",
		Declaracoes: []parser.BaseNode{impComAspas},
	}
	errosArch := checarArquitetura("infra/servicos/api.hrp", prog)
	if len(errosArch) == 0 {
		t.Errorf("Esperava violação HRP-ARCH-002 para importação da camada web dentro de infra com caminho com aspas")
	}

	// 2. gerarCPF validação matemática
	for i := 0; i < 50; i++ {
		cpf := gerarCPF()
		if len(cpf) != 11 {
			t.Fatalf("Esperava CPF com 11 dígitos, obteve %d (%s)", len(cpf), cpf)
		}
		digs := make([]int, 11)
		for idx, ch := range cpf {
			if ch < '0' || ch > '9' {
				t.Fatalf("CPF contém caractere não numérico: %s", cpf)
			}
			digs[idx] = int(ch - '0')
		}

		soma1 := 0
		for idx := 0; idx < 9; idx++ {
			soma1 += digs[idx] * (10 - idx)
		}
		dv1 := (soma1 * 10) % 11
		if dv1 == 10 {
			dv1 = 0
		}
		if digs[9] != dv1 {
			t.Errorf("Dígito verificador 1 inválido no CPF %s: esperava %d, obteve %d", cpf, dv1, digs[9])
		}

		soma2 := 0
		for idx := 0; idx < 9; idx++ {
			soma2 += digs[idx] * (11 - idx)
		}
		soma2 += digs[9] * 2
		dv2 := (soma2 * 10) % 11
		if dv2 == 10 {
			dv2 = 0
		}
		if digs[10] != dv2 {
			t.Errorf("Dígito verificador 2 inválido no CPF %s: esperava %d, obteve %d", cpf, dv2, digs[10])
		}
	}
}

func TestGerarCPF(t *testing.T) {
	for i := 0; i < 30; i++ {
		cpf := gerarCPF()
		if len(cpf) != 11 {
			t.Fatalf("Esperava 11 dígitos, obteve %d", len(cpf))
		}
	}
}
