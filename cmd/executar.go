package cmd

import (
	"fmt"
	"os"

	"github.com/natanfeitosa/portuscript/playground"
	"github.com/natanfeitosa/portuscript/ptst"
	"github.com/natanfeitosa/portuscript/vm"

	// A importação blank (_) é fundamental aqui: ela força a execução da função init()
	// presente no pacote stdlib. Isso faz com que todos os módulos nativos da biblioteca
	// padrão do Portuscript (como 'embutidos', 'matematica', 'sis', etc.) se registrem de
	// forma automática no interpretador antes que qualquer código comece a rodar.
	_ "github.com/natanfeitosa/portuscript/stdlib"
	"github.com/spf13/cobra"
)

// codigo armazena o conteúdo textual de um código Portuscript fornecido inline através da flag `-c` ou `--codigo`.
//
// Esta variável é declarada no escopo do pacote porque a biblioteca Cobra necessita de uma referência
// estável na memória para preencher o valor do argumento por meio de referenciamento de ponteiro
// na inicialização do subcomando (`PersistentFlags().StringVarP`).
var codigo string
var estrito bool
var rodarNaVM bool
var profilador bool

// comandoExecutar projeta, configura e constrói a especificação do subcomando `executar` (com alias de atalho `exec`).
//
// Este comando é a alma da interface de linha de comando para a execução do interpretador, operando em três cenários distintos:
//
//  1. Modo Interativo (Playground/TUI): Ativado quando o comando é executado sem nenhum argumento posicional e sem a flag `-c`.
//     Inicia a interface de terminal TUI, atuando como um ambiente de desenvolvimento e REPL interativo de aprendizado.
//
//  2. Execução de Arquivo: Ativado quando um arquivo `.pt` é fornecido como primeiro argumento posicional.
//     O interpretador lê, analisa (parser), compila em AST e executa as instruções descritas no arquivo físico.
//
//  3. Execução Inline (Código Rápido): Ativado pela flag `-c "codigo_portuscript"`.
//     Permite testar pequenos trechos de código diretamente pelo shell do sistema sem a necessidade de criar arquivos no disco.
//
// Ordem de Avaliação e Regras de Negócio:
//   - O diretório corrente (Working Directory) do processo atual é obtido via `os.Getwd()` no momento do disparo.
//     Ele é adicionado por padrão aos caminhos de busca (`CaminhosPadrao`) do contexto do Portuscript, garantindo que
//     importações relativas de módulos (`importar modulo`) funcionem corretamente a partir de onde o usuário chamou a CLI.
//   - Um novo contexto de máquina virtual é criado (`ptst.NewContexto`). O comando registra um `defer ctx.Terminar()`
//     imediatamente após. Isso garante que a destruição controlada do contexto ocorra de forma segura, limpando caches,
//     desalocando estruturas internas da VM e executando flush de streams pendentes, prevenindo vazamentos de recursos
//     mesmo que a execução do script sofra interrupção abrupta ou pânico (panic).
//   - Se o usuário especificar um arquivo E também a flag `-c`, o interpretador executa o arquivo posicional PRIMEIRO.
//     O código passado na flag `-c` é executado LOGO EM SEGUIDA sob o MESMO contexto. Essa precedência lógica foi planejada
//     para permitir que scripts ou bibliotecas de inicialização/configuração rodem antes de um snippet de teste rápido.
//
// Tratamento de Erros:
//   - Falhas ocorridas dentro do ambiente de execução do interpretador são interceptadas e encaminhadas para `ptst.LancarErro(err)`.
//     Esta função de tratamento formata o erro em PT-BR amigável para o usuário, destacando visualmente a linha exata e a
//     posição da sintaxe ou semântica que causou a quebra do programa (traceback).
func comandoExecutar() *cobra.Command {
	executar := &cobra.Command{
		Use:     "executar [arquivo]",
		Short:   "Executa um arquivo ou algum código inline",
		Aliases: []string{"exec"},
		Run: func(cmd *cobra.Command, args []string) {
			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao obter o diretório atual:", err)
				os.Exit(1)
			}

			ctx := ptst.NewContexto(ptst.OpcsContexto{
				CaminhosPadrao: []string{cur},
				Estrito:        estrito,
			})
			defer ctx.Terminar()

			// Cenário 1: Sem arquivo e sem código inline. Inicia o playground interativo.
			if codigo == "" && len(args) == 0 {
				playground.Inicializa(ctx, Version, Datetime, Commit)
				return
			}

			// Cenário 2: Arquivo posicional recebido. Prioridade de carregamento antes do código inline.
			if len(args) > 0 {
				if rodarNaVM {
					_, ast, errAst := ctx.TransformarEmAst(args[0], false, cur)
					if errAst != nil {
						ptst.LancarErro(errAst)
						return
					}

					comp := vm.NewCompilador()
					prog, errComp := comp.Compilar(ast)
					if errComp != nil {
						ptst.LancarErro(errComp)
						return
					}

					mainModulo, errMod := ctx.ObterModulo("__main__")
					var mainEscopo *ptst.Escopo
					if errMod == nil && mainModulo != nil {
						mainEscopo = mainModulo.Escopo
					} else {
						mainEscopo = ptst.NewEscopo()
					}

					virtualMachine := vm.NewVM(ctx)
					virtualMachine.Perfil = profilador
					frame := vm.NewFrame(prog.Bytecode, prog.Constantes, mainEscopo, nil)
					_, err = virtualMachine.Executar(frame)
					if err != nil {
						ptst.LancarErro(err)
						return
					}
					if virtualMachine.Perfil {
						virtualMachine.ImprimirPerfil()
					}
				} else {
					_, err = ptst.ExecutarArquivo(ctx, "", args[0], cur, false)
					if err != nil {
						ptst.LancarErro(err)
						return
					}
				}
			}

			// Cenário 3: Flag `-c` presente. Executa o snippet textual dentro do contexto já estabelecido.
			if codigo != "" {
				_, err = ptst.ExecutarString(ctx, codigo)
				if err != nil {
					ptst.LancarErro(err)
				}
			}
		},
	}

	// Define a flag de persistência `--codigo` (e seu atalho curto `-c`).
	// Sendo uma flag persistente, garante-se que subcomandos adjacentes herdem a definição e comportamento.
	executar.PersistentFlags().StringVarP(&codigo, "codigo", "c", "", "Use para rodar um código inline.")
	executar.PersistentFlags().BoolVar(&estrito, "estrito", false, "Ativa a validação estrita de tipos em tempo de execução.")
	executar.PersistentFlags().BoolVar(&rodarNaVM, "vm", false, "Habilita a execução experimental de bytecode na Máquina Virtual de Pilha.")
	executar.PersistentFlags().BoolVar(&profilador, "perfil", false, "Ativa o monitoramento e perfilamento de tempo das instruções da VM.")
	return executar
}
