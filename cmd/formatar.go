//go:build !js && !wasm

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func GerarUnifiedDiff(caminho, orig, form string, ui *UI) string {
	linhasOrig := strings.Split(orig, "\n")
	linhasForm := strings.Split(form, "\n")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n", caminho))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", caminho))

	max := len(linhasOrig)
	if len(linhasForm) > max {
		max = len(linhasForm)
	}

	for i := 0; i < max; i++ {
		lO := ""
		lF := ""
		if i < len(linhasOrig) {
			lO = linhasOrig[i]
		}
		if i < len(linhasForm) {
			lF = linhasForm[i]
		}

		if lO != lF {
			if i < len(linhasOrig) {
				sb.WriteString(ui.Vermelho("- " + lO + "\n"))
			}
			if i < len(linhasForm) {
				sb.WriteString(ui.Verde("+ " + lF + "\n"))
			}
		} else {
			sb.WriteString("  " + lO + "\n")
		}
	}
	return sb.String()
}

// comandoFormatar inicializa o comando 'harpia formatar'
func comandoFormatar() *cobra.Command {
	var escrever bool
	var verificar bool
	var diff bool

	ui := &UI{}

	cmdFormatar := &cobra.Command{
		Use:           "formatar [arquivo.hrp | diretorio]",
		Short:         "Formata a indentação e o estilo visual de arquivos Harpia",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			var arquivos []string
			for _, arg := range args {
				if arg == "./..." || arg == "..." {
					arg = "."
				}
				info, err := os.Stat(arg)
				if err != nil {
					fmt.Fprintf(errOut, "Erro ao acessar caminho '%s': %v\n", arg, err)
					return err
				}
				if info.IsDir() {
					_ = filepath.Walk(arg, func(path string, f os.FileInfo, err error) error {
						if err == nil && !f.IsDir() && (strings.HasSuffix(path, ".hrp") || strings.HasSuffix(path, ".pt")) {
							arquivos = append(arquivos, path)
						}
						return nil
					})
				} else {
					arquivos = append(arquivos, arg)
				}
			}

			desformatados := 0
			for _, caminho := range arquivos {
				conteudo, err := os.ReadFile(caminho)
				if err != nil {
					fmt.Fprintf(errOut, "Erro ao ler arquivo %s: %v\n", caminho, err)
					continue
				}

				origStr := string(conteudo)
				formatado := FormatarCodigoHarpia(origStr)

				if origStr != formatado {
					desformatados++
					if diff {
						fmt.Fprint(out, GerarUnifiedDiff(caminho, origStr, formatado, ui))
					} else if verificar {
						fmt.Fprintf(errOut, "HRP-FMT-001: %s precisa de formatação.\n", caminho)
					}
				}

				if escrever && origStr != formatado {
					err = os.WriteFile(caminho, []byte(formatado), 0644)
					if err != nil {
						fmt.Fprintf(errOut, "Erro ao salvar arquivo %s: %v\n", caminho, err)
					} else {
						fmt.Fprintf(out, "Formatado: %s\n", caminho)
					}
				} else if !verificar && !diff && !escrever {
					fmt.Fprint(out, formatado)
				}
			}

			if verificar && desformatados > 0 {
				fmt.Fprintf(errOut, "\n%d arquivo(s) precisam de formatação. Execute 'harpia formatar -w [caminho]'.\n", desformatados)
				return fmt.Errorf("%d arquivo(s) desformatados", desformatados)
			}

			return nil
		},
	}

	cmdFormatar.Flags().BoolVarP(&escrever, "escrever", "w", false, "Salva as alterações de volta nos arquivos originais")
	cmdFormatar.Flags().BoolVar(&verificar, "verificar", false, "Sai com erro se algum arquivo não estiver formatado (para CI)")
	cmdFormatar.Flags().BoolVar(&diff, "diff", false, "Exibe as diferenças (diff) entre o arquivo atual e o formatado")
	return cmdFormatar
}

// FormatarCodigoHarpia formata recuos de blocos, remove trailing spaces e higieniza espaçamento
func FormatarCodigoHarpia(codigo string) string {
	linhas := strings.Split(codigo, "\n")
	var res []string
	nivel := 0
	linhaAnteriorVazia := false

	for _, linhaRaw := range linhas {
		// Remove espaços em branco do final da linha (trailing spaces)
		linha := strings.TrimRight(linhaRaw, " \t\r")
		linhaTrimmed := strings.TrimSpace(linha)

		if linhaTrimmed == "" {
			if !linhaAnteriorVazia {
				res = append(res, "")
				linhaAnteriorVazia = true
			}
			continue
		}
		linhaAnteriorVazia = false

		fechamentos := strings.Count(linhaTrimmed, "}") + strings.Count(linhaTrimmed, "]") + strings.Count(linhaTrimmed, ")")
		aberturas := strings.Count(linhaTrimmed, "{") + strings.Count(linhaTrimmed, "[") + strings.Count(linhaTrimmed, "(")

		// Se a linha começar com fechamento, reduz o nível imediatamente
		if strings.HasPrefix(linhaTrimmed, "}") || strings.HasPrefix(linhaTrimmed, "]") || strings.HasPrefix(linhaTrimmed, ")") {
			nivel = maximo(0, nivel-1)
		} else if fechamentos > aberturas {
			nivel = maximo(0, nivel-(fechamentos-aberturas))
		}

		// Adiciona recuo de 4 espaços correspondente
		recuo := strings.Repeat("    ", nivel)
		res = append(res, recuo+linhaTrimmed)

		// Incrementa recuo se abriu mais blocos
		if aberturas > fechamentos {
			nivel += aberturas - fechamentos
		} else if strings.HasSuffix(linhaTrimmed, "{") || strings.HasSuffix(linhaTrimmed, "[") || strings.HasSuffix(linhaTrimmed, "(") {
			nivel++
		}
	}

	resultado := strings.Join(res, "\n")
	if !strings.HasSuffix(resultado, "\n") {
		resultado += "\n"
	}
	return resultado
}

// maximo é o substituto minimalista para a função nativa `max()` (Go 1.21+),
// garantindo compatibilidade com versões anteriores do toolchain.
// Mantida como helper local para evitar adicionar uma versão de linguagem mínima ao build.
func maximo(a, b int) int {
	if a > b {
		return a
	}
	return b
}
