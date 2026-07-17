package ptst

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

func TestTipagemEstaticaParsing(t *testing.T) {
	codigo := `
	var x: Inteiro = 10;
	const y: Texto = "Olá";
	funcao soma(a: Inteiro, b: Inteiro = 0): Inteiro {
		retorne a + b;
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao parsear código com tipagem: %v", err)
	}

	prog, ok := ast.(*parser.Programa)
	if !ok {
		t.Fatal("AST de nível superior não é *parser.Programa")
	}

	// Verifica var x: Inteiro
	declX, ok := prog.Declaracoes[0].(*parser.DeclVar)
	if !ok {
		t.Fatalf("Esperava *parser.DeclVar na declaração 0, obteve %T", prog.Declaracoes[0])
	}
	if declX.Nome != "x" {
		t.Errorf("Esperava Nome=%q, obteve %q", "x", declX.Nome)
	}
	if declX.Tipo != "Inteiro" {
		t.Errorf("Esperava Tipo=%q, obteve %q", "Inteiro", declX.Tipo)
	}

	// Verifica const y: Texto
	declY, ok := prog.Declaracoes[1].(*parser.DeclVar)
	if !ok {
		t.Fatalf("Esperava *parser.DeclVar na declaração 1, obteve %T", prog.Declaracoes[1])
	}
	if declY.Nome != "y" {
		t.Errorf("Esperava Nome=%q, obteve %q", "y", declY.Nome)
	}
	if declY.Tipo != "Texto" {
		t.Errorf("Esperava Tipo=%q, obteve %q", "Texto", declY.Tipo)
	}
	if !declY.Constante {
		t.Errorf("Esperava Constante=Verdadeiro na declaração 'const'")
	}

	// Verifica funcao soma
	declSoma, ok := prog.Declaracoes[2].(*parser.DeclFuncao)
	if !ok {
		t.Fatalf("Esperava *parser.DeclFuncao na declaração 2, obteve %T", prog.Declaracoes[2])
	}
	if declSoma.Nome != "soma" {
		t.Errorf("Esperava Nome=%q, obteve %q", "soma", declSoma.Nome)
	}
	if declSoma.TipoRetorno != "Inteiro" {
		t.Errorf("Esperava TipoRetorno=%q, obteve %q", "Inteiro", declSoma.TipoRetorno)
	}

	if len(declSoma.Parametros) != 2 {
		t.Fatalf("Esperava 2 parâmetros, obteve %d", len(declSoma.Parametros))
	}
	if declSoma.Parametros[0].Nome != "a" || declSoma.Parametros[0].Tipo != "Inteiro" {
		t.Errorf("Parâmetro 0 esperado ('a','Inteiro'), obteve (%q,%q)",
			declSoma.Parametros[0].Nome, declSoma.Parametros[0].Tipo)
	}
	if declSoma.Parametros[1].Nome != "b" || declSoma.Parametros[1].Tipo != "Inteiro" {
		t.Errorf("Parâmetro 1 esperado ('b','Inteiro'), obteve (%q,%q)",
			declSoma.Parametros[1].Nome, declSoma.Parametros[1].Tipo)
	}
	if declSoma.Parametros[1].Padrao == nil {
		t.Errorf("Esperava que parâmetro 'b' tivesse valor padrão")
	}
}

func TestTipagemEstaticaExecucao(t *testing.T) {
	testes := []struct {
		nome    string
		codigo  string
		estrito bool
		comErro bool
	}{
		{
			nome: "Var simples tipo correto (Inteiro)",
			codigo: `
			var x: Inteiro = 10;
			`,
			estrito: true,
			comErro: false,
		},
		{
			nome: "Var simples tipo incorreto (Inteiro = Texto)",
			codigo: `
			var x: Inteiro = "olá";
			`,
			estrito: true,
			comErro: true,
		},
		{
			nome: "Var simples tipo incorreto sem estrito (não deve falhar)",
			codigo: `
			var x: Inteiro = "olá";
			`,
			estrito: false,
			comErro: false,
		},
		{
			nome: "Reatribuicao tipo correto",
			codigo: `
			var x: Inteiro = 10;
			x = 20;
			`,
			estrito: true,
			comErro: false,
		},
		{
			nome: "Reatribuicao tipo incorreto",
			codigo: `
			var x: Inteiro = 10;
			x = "texto";
			`,
			estrito: true,
			comErro: true,
		},
		{
			nome: "Chamada de funcao tipo correto",
			codigo: `
			funcao soma(a: Inteiro, b: Inteiro): Inteiro {
				retorne a + b;
			}
			soma(10, 20);
			`,
			estrito: true,
			comErro: false,
		},
		{
			nome: "Chamada de funcao tipo incorreto no param",
			codigo: `
			funcao soma(a: Inteiro, b: Inteiro): Inteiro {
				retorne a + b;
			}
			soma(10, "20");
			`,
			estrito: true,
			comErro: true,
		},
		{
			nome: "Retorno de funcao tipo incorreto",
			codigo: `
			funcao soma(a: Inteiro, b: Inteiro): Inteiro {
				retorne "não sou inteiro";
			}
			soma(10, 20);
			`,
			estrito: true,
			comErro: true,
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			ctx := NewContexto(OpcsContexto{Estrito: tc.estrito})
			defer ctx.Terminar()

			ast, err := ctx.StringParaAst(tc.codigo, "<teste>")
			if err != nil {
				t.Fatalf("Erro de parsing: %v", err)
			}

			escopo := NewEscopo()
			_, err = (&Interpretador{Ast: ast, Contexto: ctx, Escopo: escopo}).Inicializa()

			if tc.comErro && err == nil {
				t.Errorf("Esperava que ocorresse erro de tipagem estrita, mas executou com sucesso")
			} else if !tc.comErro && err != nil {
				t.Errorf("Não esperava erro, mas obteve: %v", err)
			}
		})
	}
}

