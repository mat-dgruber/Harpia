//go:build !js && !wasm

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completarCmd declara o subcomando CLI `harpia completar`.
//
// Encapsula a geração de scripts de autocompletar para múltiplos shells
// (bash, zsh, fish, PowerShell) usando exclusivamente as APIs públicas do Cobra,
// reaproveitando 100% da lógica nativa de introspecção de sub-comandos, flags e aliases.
// O cabeçalho Long está em português e a saída do `--help` segue a mesma direção.
var completarCmd = &cobra.Command{
	Use:     "completar [bash|zsh|fish|powershell]",
	Aliases: []string{"autocompletar", "completion"},
	Short:   "Gera o script de autocompletar para o shell informado",
	Long: `Gera o script de autocompletar para ` + "harpia" + ` para o shell informado.

Cada sub-comando aceita um argumento obrigatório indicando o shell alvo:
  bash        gera o script para Bash (utilize 'source <(harpia completar bash)')
  zsh         gera o script para Zsh (inclui '#compdef harpia' no cabeçalho)
  fish        gera o script para Fish (utilize 'harpia completar fish | source')
  powershell  gera o script para PowerShell (registre via 'harpia completar powershell | Out-String | Invoke-Expression')

Exemplos:
  harpia completar bash > ~/.local/share/bash-completion/completions/harpia
  source <(harpia completar bash)
`,
	// silencio erros internos para impressão controlada de mensagens PT-BR.
	SilenceErrors: true,
	SilenceUsage:  true,
	ValidArgs:     []string{"bash", "zsh", "fish", "powershell"},
	Args:          cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := cmd.Root()
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return root.GenBashCompletion(out)
		case "zsh":
			return root.GenZshCompletion(out)
		case "fish":
			return root.GenFishCompletion(out, false)
		case "powershell":
			return root.GenPowerShellCompletion(out)
		}
		return fmt.Errorf("shell não suportado: %s", args[0])
	},
}

// comandoCompletar é o ponto de entrada público, espelhando o padrão usado pelos demais sub-comandos do pacote.
func comandoCompletar() *cobra.Command {
	return completarCmd
}

// ErroEmPortugolpadroniza a mensagem de erro para `fmt.Fprintln(os.Stderr)`
// mantendo paridade visual com o restante da CLI em português.
func init() {
	// ponytail: redireciona mensagens de erro do Cobra em PT-BR via hook local;
	// add caso a CLI ganhe internacionalização multi-idioma (hoje é só PT-BR).
	if completarCmd.SilenceUsage {
		completarCmd.SetUsageFunc(func(c *cobra.Command) error {
			fmt.Fprintf(os.Stderr, "Uso: %s\n", c.UseLine())
			return nil
		})
	}
}
