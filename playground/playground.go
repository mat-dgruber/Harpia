package playground

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/peterh/liner"
)

// Cores ANSI de acordo com BRAND_GUIDELINES.md
const (
	CorOlhoHarpia   = "\033[1;33m"                                // Olho da Harpia em Negrito (#F2A900)
	CorNevoManha    = "\033[3;90m"                                // Névoa da Manhã em Itálico (#8C9B9E)
	CorLinha        = "\033[90m"                                  // Cinza para linhas
	CorTerracota    = "\033[38;2;212;93;52m"                      // Terracota (#D45D34)
	CorBarraAtalhos = "\033[48;2;23;30;38m\033[38;2;243;246;244m" // Fundo Rio Profundo e texto em Penagem Branca
	CorReset        = "\033[0m"
)

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
// O histórico é salvo em um arquivo oculto chamado `.historico_harpia` no diretório Home do usuário.
// Se 'escrita' for verdadeiro, abre o arquivo em modo append/create. Caso contrário, abre em modo somente leitura.
func ArquivoHistorico(escrita bool) (arquivo *os.File) {
	caminho := path.Join(homeDirectory(), ".historico_harpia")

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

// Inicializa configura, orquestra e dispara o loop de eventos REPL principal (TUI) do playground do Harpia.
//
// O fluxo operacional é composto por:
//  1. Exibir o banner informativo contendo a versão e dados de build;
//  2. Instanciar e preparar o Executor da VM, injetando dinamicamente a função embutida 'sair()' no escopo local;
//  3. Inicializar a biblioteca de leitura de console Liner (que oferece suporte nativo a histórico de digitação,
//     atalhos de terminal e setas direcionais);
//  4. Ler o histórico de comandos persistido no disco a partir de `~/.historico_harpia`;
//  5. Rodar o loop iterativo principal, coletando linhas do terminal e analisando o fechamento de blocos;
//  6. Ao fechar o bloco de código, envia o acumulado para processamento pela VM via 'ExecutarCodigo';
//  7. Em caso de encerramento do console (por digitação de `sair()` ou interrupção via sinal como Ctrl+D),
//     o defer garante a escrita de histórico acumulado de volta ao disco de forma persistente.
func Inicializa(ctx *hrp.Contexto, version, datetime, commit string) {
	caminho := path.Join(homeDirectory(), ".historico_harpia")
	arquivo, _ := os.OpenFile(caminho, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)

	finalizou := false
	finalizar := func() {
		fmt.Printf("Saindo...")
		finalizou = true
	}

	exec := NovoExecutor(ctx)

	// Injeta a função nativa sair() no REPL de forma simples e amigável.
	// Quando chamada pelo usuário no terminal, dispara a finalização do loop de forma graciosa.
	exec.RegistrarMetodo(hrp.NewMetodoOuPanic("sair", func(_ hrp.Objeto, args hrp.Objeto) (hrp.Objeto, error) {
		finalizar()
		return nil, nil
	}, ""))

	so := runtime.GOOS
	arch := runtime.GOARCH

	// Tenta descobrir o nome do projeto ativo a partir do diretório atual
	projetoAtivo := "global"
	wd, err := os.Getwd()
	if err == nil {
		projetoAtivo = filepath.Base(wd)
	}

	templateBanner := `
 {CorOlhoHarpia}🦅  HARPIA v{version}{CorReset}  •  Linguagem de Programação 100% Brasileira
 {CorLinha}──────────────────────────────────────────────────────────────{CorReset}
 {CorNevoManha}"O código bonito é como uma boa bossa nova: simples e fluido."{CorReset}
 {CorLinha}──────────────────────────────────────────────────────────────{CorReset}
 Sistema: {so} ({arch})  |  Projeto ativo: {projetoAtivo}

 {CorTerracota}› ajuda{CorReset}   — Exibe instruções e comandos úteis do REPL
 {CorTerracota}› sair(){CorReset}   — Encerra a sessão interativa do Harpia

 {CorBarraAtalhos} [ajuda] Ajuda  │  [sair()] Sair  │  [ctrl+d]/[ctrl+c] Sair {CorReset}
 {CorLinha}──────────────────────────────────────────────────────────────{CorReset}`

	r := strings.NewReplacer(
		"{CorOlhoHarpia}", CorOlhoHarpia,
		"{CorReset}", CorReset,
		"{CorLinha}", CorLinha,
		"{CorNevoManha}", CorNevoManha,
		"{CorTerracota}", CorTerracota,
		"{CorBarraAtalhos}", CorBarraAtalhos,
		"{version}", version,
		"{so}", so,
		"{arch}", arch,
		"{projetoAtivo}", projetoAtivo,
	)

	bannerFormatado := r.Replace(templateBanner)
	fmt.Println(strings.Trim(bannerFormatado, "\n"))

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
			if err == liner.ErrPromptAborted || err == io.EOF {
				fmt.Println("\nSaindo...")
				break
			}
			fmt.Fprintln(os.Stderr, err)
			break
		}

		if len(codigo) < 1 {
			fmt.Println("Entrada vazia")
			continue
		}

		cmdLimpo := strings.TrimSpace(codigo)

		// Comando interativo: limpar tela
		if cmdLimpo == "limpar" || cmdLimpo == "clear" {
			fmt.Print("\033[H\033[2J")
			continue
		}

		// Comando interativo: escopo de variáveis
		if cmdLimpo == "escopo" || cmdLimpo == "vars" || cmdLimpo == "simbolos" {
			temSimbolo := false
			fmt.Println("Símbolos e variáveis ativos declarados nesta sessão:")
			for nome, simbolo := range exec.Modulo.Escopo.Simbolos {
				// Oculta funções internas ou o comando 'sair'
				if nome == "sair" {
					continue
				}
				temSimbolo = true
				txtVal, _ := hrp.NewTexto(simbolo.Valor)
				fmt.Printf("  var %s = %s\n", nome, txtVal)
			}
			if !temSimbolo {
				fmt.Println("  (Nenhuma variável declarada nesta sessão ainda)")
			}
			continue
		}

		// Comando interativo: ajuda detalhada
		if strings.HasPrefix(cmdLimpo, "ajuda ") || strings.HasPrefix(cmdLimpo, "help ") {
			partes := strings.Fields(cmdLimpo)
			if len(partes) > 1 {
				funcaoAlvo := partes[1]
				documentacaoNativa := map[string]string{
					"imprimir": `Função: imprimir(...) ou imprima(...)
Descrição: Exibe um ou mais valores na tela, separados por espaço e terminados por quebra de linha.
Exemplo:
  imprimir("Olá,", "Mundo!")
  imprimir(10, Verdadeiro, Nulo)`,
					"imprima": `Função: imprima(...)
Descrição: Alias para 'imprimir(...)'. Exibe valores na tela.`,
					"leia": `Função: leia()
Descrição: Aguarda e lê uma linha de texto digitada pelo usuário no teclado (entrada padrão).
Retorno: O texto digitado (como uma string/texto).
Exemplo:
  imprimir("Qual o seu nome?");
  var nome = leia();
  imprimir("Olá,", nome);`,
					"tamanho": `Função: tamanho(objeto)
Descrição: Retorna a quantidade de elementos de uma lista, tupla, mapa ou o número de caracteres de um texto.
Retorno: Um número inteiro representando a contagem.
Exemplo:
  tamanho("Harpia")      # Retorna 6
  tamanho([1, 2, 3, 4])  # Retorna 4`,
					"tipo": `Função: tipo(objeto)
Descrição: Obtém a classe de tipo estrutural do objeto informado.
Retorno: A classe do objeto (ex: texto, inteiro, decimal, booleano, etc.).
Exemplo:
  tipo("Olá")  # Retorna <classe 'texto'>
  tipo(42)     # Retorna <classe 'inteiro'>`,
					"sequencia": `Função: sequencia(limite) ou sequencia(inicio, fim, passo)
Descrição: Cria uma lista de números inteiros de forma progressiva, útil para laços de repetição 'para'.
Retorno: Uma lista de inteiros.
Exemplo:
  sequencia(5)        # Retorna [0, 1, 2, 3, 4]
  sequencia(2, 8, 2)  # Retorna [2, 4, 6]`,
					"sinal": `Função: sinal(valorInicial)
Descrição: Primitiva central de reatividade do Harpia. Cria um estado reativo que expõe uma função getter de leitura e uma setter de escrita.
Retorno: Uma lista contendo [getter, setter].
Exemplo:
  var s = sinal(42);
  var ler = s[0];
  var set = s[1];
  imprimir(ler()); # Mostra 42
  set(100);
  imprimir(ler()); # Mostra 100`,
				}

				doc, existe := documentacaoNativa[funcaoAlvo]
				if existe {
					fmt.Println(doc)
				} else {
					fmt.Printf("Sem ajuda detalhada disponível para '%s'. Tente: imprimir, leia, tamanho, tipo, sequencia, sinal\n", funcaoAlvo)
				}
				continue
			}
		}

		if cmdLimpo == "ajuda" || cmdLimpo == "help" {
			fmt.Println("Comandos úteis do interpretador interativo:")
			fmt.Println("  ajuda / help         - Mostra esta mensagem de ajuda")
			fmt.Println("  ajuda <funcao>       - Mostra ajuda detalhada sobre uma função (ex: ajuda sinal)")
			fmt.Println("  escopo / vars        - Lista as variáveis e símbolos ativos declarados")
			fmt.Println("  limpar / clear       - Limpa a tela do console")
			fmt.Println("  sair()               - Sai do terminal interativo")
			fmt.Println("\nFunções nativas embutidas no escopo global do Harpia:")
			fmt.Println("  imprimir(...)  - Imprime valores ou variáveis no console (alias: imprima)")
			fmt.Println("  leia()         - Aguarda e lê uma linha digitada no teclado")
			fmt.Println("  tamanho(x)     - Retorna o comprimento de um texto, lista ou mapa")
			fmt.Println("  tipo(x)        - Retorna a classe estrutural/tipo de um objeto")
			fmt.Println("  sequencia(n)   - Gera uma lista numérica sequencial de 0 até n-1")
			fmt.Println("  sinal(valor)   - Cria um sinal de estado reativo, retornando [getter, setter]")
			fmt.Println("\nExemplo prático de reatividade no REPL:")
			fmt.Println("  var s = sinal(42); var get = s[0]; var set = s[1];")
			fmt.Println("  imprimir(get()); set(100); imprimir(get());")
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
