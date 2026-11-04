package stdlib

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// funcaoDoModulo importa `nome` do módulo dentro do contexto e devolve a função.
func funcaoDoModulo(t *testing.T, ctx *hrp.Contexto, modulo, nome string) hrp.Objeto {
	t.Helper()
	mod, err := hrp.ExecutarString(ctx, "de \""+modulo+"\" importe "+nome)
	if err != nil {
		t.Fatalf("importar %s.%s: %v", modulo, nome, err)
	}
	fn, err := mod.Escopo.ObterValor(nome)
	if err != nil {
		t.Fatal(err)
	}
	return fn
}

func esperaNegado(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "Acesso Negado") {
		t.Fatalf("esperava 'Acesso Negado', veio %v", err)
	}
}

// Um contexto sem sandbox criado depois (que vira o "ativo" global) não pode liberar
// o acesso de um script que nasceu em sandbox.
func TestSandboxNaoDependeDoContextoAtivo(t *testing.T) {
	bloqueado := hrp.NewContexto(hrp.OpcsContexto{BloquearArquivos: true, BloquearRede: true})
	defer bloqueado.Terminar()
	executar := funcaoDoModulo(t, bloqueado, "sistema", "executar")
	requisitar := funcaoDoModulo(t, bloqueado, "http", "requisitar")

	livre := hrp.NewContexto(hrp.OpcsContexto{})
	defer livre.Terminar()

	_, err := hrp.Chamar(executar, hrp.Tupla{hrp.Texto("echo"), hrp.Tupla{hrp.Texto("x")}})
	esperaNegado(t, err)
	_, err = hrp.Chamar(requisitar, hrp.Tupla{hrp.Texto("GET"), hrp.Texto("http://127.0.0.1:1")})
	esperaNegado(t, err)
}

// Sem nenhum contexto para consultar, a operação é negada (fail-closed).
func TestExigirPermissaoSemContextoNega(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	ctx.Terminar() // limpa o contexto ativo global
	if hrp.ObterContextoAtivo() != nil {
		t.Skip("outro contexto ativo no processo")
	}
	esperaNegado(t, hrp.ExigirPermissao(nil, hrp.RecursoArquivos))
	esperaNegado(t, hrp.ExigirPermissao(hrp.Nulo, hrp.RecursoRede))
}
