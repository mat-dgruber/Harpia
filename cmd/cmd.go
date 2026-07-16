// Package cmd agrupa, estrutura e instala os subcomandos que compõem a CLI (Interface de Linha de Comando) do Portuscript.
//
// O pacote atua como o ponto central de integração entre a inicialização do executável (definido no arquivo principal main.go)
// e a biblioteca de gerenciamento de comandos Cobra (github.com/spf13/cobra).
//
// Responsabilidades do pacote cmd:
//
//   - Expor as variáveis globais de build (Commit, Datetime, Version), preenchidas
//     via injeção de linker (ldflags) pelo GoReleaser no pipeline de CI/CD. Estas variáveis são
//     propagadas para recursos internos como o atualizador e o console interativo (playground).
//
//   - Fornecer a função InstalarComandos como o ponto único de montagem de toda a árvore de comandos
//     e subcomandos suportados pelo utilitário de terminal do Portuscript.
//
// Filosofia de Design da CLI:
// Toda a comunicação com o desenvolvedor ocorre prioritariamente em Português (PT-BR). As descrições curtas,
// longas, exemplos, mensagens de erro e aliases (apelidos de comandos) refletem essa escolha de design,
// oferecendo uma experiência linguística fluida e coerente do início ao fim.
package cmd

import "github.com/spf13/cobra"

// Variáveis de build injetadas dinamicamente durante a etapa de compilação.
//
// Por padrão, são declaradas com valores de fallback ("sentinelas") para viabilizar
// o desenvolvimento local (ex: rodar utilizando `go run main.go`).
// Durante o build oficial de distribuição, essas strings são substituídas via flags de ligação:
//
//	go build -ldflags "-X 'github.com/natanfeitosa/portuscript/cmd.Version=1.0.0' -X 'github.com/natanfeitosa/portuscript/cmd.Commit=abcdef' -X 'github.com/natanfeitosa/portuscript/cmd.Datetime=2026-07-14T00:00:00Z'"
var (
	// Commit armazena a hash curta (ou completa) do commit do Git no qual este build foi gerado.
	// É útil para auditoria e rastreamento exato do código-fonte em execução no cliente.
	Commit string = "-"

	// Datetime armazena o carimbo de data e hora (geralmente formato ISO-8601) que representa
	// o momento exato em que o binário foi compilado no servidor de build.
	Datetime string = "0000-00-00T00:00:00"

	// Version armazena a representação semântica oficial da versão (SemVer, ex: "0.2.1") deste binário.
	// O valor padrão "dev" é substituído pela tag correspondente do Git durante o processo de empacotamento.
	// É utilizado pelo comando 'atualize' para validar a existência de versões mais novas.
	Version string = "dev"
)

// InstalarComandos é a porta de entrada pública do pacote para montagem e registro dos subcomandos CLI.
//
// Esta função encapsula o acoplamento do Cobra e registra na instância raiz (raiz) do console
// os comandos secundários disponíveis:
//   - `atualize`: Busca e instala novas releases binárias a partir do GitHub.
//   - `executar` (exec): Interpreta arquivos Portuscript (.pt) ou pequenos trechos passados diretamente via terminal.
//   - `testar`: Varre o diretório em busca de arquivos `.ptst`/`.pt` e executa os blocos `teste`/blocos nativos.
//   - `checar`: Linter estático — analisa a AST em busca de erros semânticos sem executar (escopo, redeclaração, constantes).
//
// Centralizar o registro de subcomandos nesta função simplifica consideravelmente a manutenção e legibilidade
// do projeto, pois funciona como um índice declarativo dos comandos suportados pela CLI.
//
// O parâmetro 'raiz' é a estrutura base *cobra.Command inicializada no ponto de entrada (main.go).
func InstalarComandos(raiz *cobra.Command) {
	raiz.AddCommand(comandoAtualize())
	raiz.AddCommand(comandoExecutar())
	raiz.AddCommand(comandoTestar())
	raiz.AddCommand(comandoChecar())
	raiz.AddCommand(comandoErroCLI())
	raiz.AddCommand(comandoCompilar())
	raiz.AddCommand(comandoServir())
	raiz.AddCommand(comandoNovo())
	raiz.AddCommand(comandoCrie())
	raiz.AddCommand(comandoLsp())
	raiz.AddCommand(comandoPlayground())
	raiz.AddCommand(comandoFormatar())
	raiz.AddCommand(comandoInstalar())
	raiz.AddCommand(comandoDiagramar())
	raiz.AddCommand(comandoTui())
	raiz.AddCommand(comandoDoc())
	raiz.AddCommand(comandoEmpacotar())
	raiz.AddCommand(comandoStressar())
	raiz.AddCommand(comandoDepurar())
}
