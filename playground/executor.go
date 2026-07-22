package playground

import (
	"fmt"
	"os"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Executor é o componente responsável pelo gerenciamento de compilação rápida, isolamento de escopo
// e avaliação sintática (interpretação) das entradas de código fornecidas no playground.
type Executor struct {
	// Contexto armazena a instância da máquina virtual e as opções compartilhadas de execução.
	Contexto *hrp.Contexto

	// Modulo é um escopo dinâmico isolado de execução simulando um arquivo físico virtualizado (<playground>).
	// Isso permite que o usuário defina funções, variáveis e classes em uma linha e elas permaneçam
	// vivas e acessíveis na linha seguinte do console interativo.
	Modulo *hrp.Modulo
}

// NovoExecutor é o construtor padrão da estrutura Executor.
//
// Parâmetros:
//   - ctx: Ponteiro para o contexto corrente de runtime do Harpia (*hrp.Contexto).
//
// Retorna:
//   - *Executor: Estrutura pronta para orquestrar e interpretar código sob escopo persistente.
//
// Detalhes:
// Ele inicializa um novo módulo virtualizado sob o nome especial "<playground>" e configura o seu respectivo escopo.
// Quaisquer variáveis e métodos declarados no console REPL serão anexados ao escopo persistente deste módulo.
func NovoExecutor(ctx *hrp.Contexto) *Executor {
	exec := new(Executor)
	exec.Contexto = ctx
	exec.Modulo, _ = ctx.InicializarModulo(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Arquivo: "<playground>",
		},
	})

	return exec
}

// ExecutarCodigo realiza o ciclo completo de interpretação e saída visual de uma entrada de código no REPL:
//
// Parâmetros:
//   - codigo: String contendo o script ou expressão Harpia digitado pelo programador.
//
// Fluxo Interno:
//  1. Recebe a string de código e compila dinamicamente para uma Árvore de Sintaxe Abstrata (AST)
//     referenciando a origem virtual "<playground>";
//  2. Avalia a AST no ambiente da VM sob o escopo isolado e persistente do módulo virtual do playground;
//  3. Converte o Objeto Go de retorno resultante em sua representação textual correspondente no Harpia;
//  4. Imprime o resultado final diretamente no terminal padrão do usuário.
//
// Em caso de qualquer falha léxica, sintática ou de tempo de execução, a exceção é interceptada,
// formatada de forma legível e impressa na saída padrão de erros (Stderr), sem derrubar ou encerrar o REPL.
func (e *Executor) ExecutarCodigo(codigo string) {
	ast, err := e.Contexto.StringParaAst(codigo, "<playground>")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var resultado, texto hrp.Objeto

	if resultado, err = e.Contexto.AvaliarAst(ast, e.Modulo.Escopo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if texto, err = hrp.NewTexto(resultado); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(texto)
}

// RegistrarMetodo permite injetar métodos auxiliares e funções embutidas personalizadas
// diretamente no escopo global acessível pelo console REPL.
//
// Parâmetros:
//   - metodo: Instância de método estruturado do Harpia (*hrp.Metodo) contendo a assinatura e lógica Go nativa.
//
// Retorna:
//   - error: Nil em caso de sucesso, ou erro de definição caso o símbolo colida ou seja inválido.
//
// Aplicação:
// É utilizado pelo playground para expor comandos de utilidade geral do console (como `sair()`).
func (e *Executor) RegistrarMetodo(metodo *hrp.Metodo) error {
	return e.Modulo.Escopo.DefinirSimbolo(
		hrp.NewVarSimbolo(
			metodo.Nome,
			metodo,
		),
	)
}

// Terminar encerra as atividades da VM associada, liberando recursos e caches retidos pelo interpretador.
func (e *Executor) Terminar() {
	e.Contexto.Terminar()
}
