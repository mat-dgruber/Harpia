package playground

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/natanfeitosa/portuscript/ptst"
	"github.com/peterh/liner"
)

// banner é o texto de boas-vindas impresso na inicialização do playground.
// Exibe dinamicamente informações de versão, data de compilação e commit do build.
const banner = `
Bem vindos ao Portuscript v%s.

(%s) [%s]
`

// homeDirectory resolve e retorna de forma resiliente o caminho absoluto da pasta Home do usuário atual.
//
// Tenta primeiro utilizar o utilitário nativo de sistema 'user.Current()' para recuperar de forma segura.
// Em caso de falhas ou ambientes com permissões isoladas, recorre ao fallback da variável de ambiente "$HOME".
func homeDirectory() string {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}

// ArquivoHistorico gerencia de maneira simplificada a abertura de fluxo de leitura ou escrita do histórico de comandos.
//
// O histórico é salvo em um arquivo oculto chamado `.historico_portuscript` no diretório Home do usuário.
// Se 'escrita' for verdadeiro, abre o arquivo em modo append/create. Caso contrário, abre em modo somente leitura.
func ArquivoHistorico(escrita bool) (arquivo *os.File) {
	caminho := path.Join(homeDirectory(), ".historico_portuscript")

	if escrita {
		arquivo, _ = os.OpenFile(caminho, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		// if err != nil {
		// 	return err
		// }
		return
	}

	arquivo, _ = os.Open(caminho)
	defer arquivo.Close()
	return
}

// Inicializa configura, orquestra e dispara o loop de eventos REPL principal (TUI) do playground do Portuscript.
//
// O fluxo operacional é composto por:
//  1. Exibir o banner informativo contendo a versão e dados de build;
//  2. Instanciar e preparar o Executor da VM, injetando dinamicamente a função embutida 'sair()' no escopo local;
//  3. Inicializar a biblioteca de leitura de console Liner (que oferece suporte nativo a histórico de digitação,
//     atalhos de terminal e setas direcionais);
//  4. Ler o histórico de comandos persistido no disco a partir de `~/.historico_portuscript`;
//  5. Rodar o loop iterativo principal, coletando linhas do terminal e analisando o fechamento de blocos;
//  6. Ao fechar o bloco de código, envia o acumulado para processamento pela VM via 'ExecutarCodigo';
//  7. Em caso de encerramento do console (por digitação de `sair()` ou interrupção via sinal como Ctrl+D),
//     o defer garante a escrita de histórico acumulado de volta ao disco de forma persistente.
func Inicializa(ctx *ptst.Contexto, version, datetime, commit string) {
	caminho := path.Join(homeDirectory(), ".historico_portuscript")
	arquivo, _ := os.OpenFile(caminho, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)

	finalizou := false
	finalizar := func() {
		fmt.Printf("Saindo...")
		finalizou = true
	}

	exec := NovoExecutor(ctx)

	// Injeta a função nativa sair() no REPL de forma simples e amigável.
	// Quando chamada pelo usuário no terminal, dispara a finalização do loop de forma graciosa.
	exec.RegistrarMetodo(ptst.NewMetodoOuPanic("sair", func(_ ptst.Objeto, args ptst.Objeto) (ptst.Objeto, error) {
		finalizar()
		return nil, nil
	}, ""))

	fmt.Println(fmt.Sprintf(strings.Trim(banner, " \n"), version, datetime, commit))

	line := liner.NewLiner()
	line.ReadHistory(arquivo)

	defer func() {
		line.Close()
		arquivo.Close()
		// exec.Terminar()
	}()

	estado := NewEstado()

	// Loop iterativo de leitura de comandos.
	for !finalizou {
		codigo, err := line.Prompt(string(estado.Indicador))
		if err != nil {
			finalizar()
			fmt.Fprintln(os.Stderr, err)
		}

		if len(codigo) < 1 {
			fmt.Println("Entrada vazia")
			continue
		}

		line.AppendHistory(codigo)
		estado.RecalcularEstado(codigo)

		// Se o estado não estiver pendente de fechar blocos em uma nova linha,
		// envia o buffer para o executor e zera o acumulado.
		if !estado.Continua {
			exec.ExecutarCodigo(estado.Codigo)
			estado.Codigo = ""
		}
	}

	line.WriteHistory(arquivo)
}
