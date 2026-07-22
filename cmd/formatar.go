package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// comandoFormatar inicializa o comando 'harpia formatar'
func comandoFormatar() *cobra.Command {
	var escrever bool
	var verificar bool
	cmdFormatar := &cobra.Command{
		Use:   "formatar [arquivo.hrp | diretorio]",
		Short: "Formata a indentação e o estilo visual de arquivos Harpia",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var arquivos []string
			for _, arg := range args {
				if arg == "./..." || arg == "..." {
					arg = "."
				}
				info, err := os.Stat(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao acessar caminho '%s': %v\n", arg, err)
					os.Exit(1)
				}
				if info.IsDir() {
					filepath.Walk(arg, func(path string, f os.FileInfo, err error) error {
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
					fmt.Fprintf(os.Stderr, "Erro ao ler arquivo %s: %v\n", caminho, err)
					continue
				}

				formatado := FormatarCodigoHarpia(string(conteudo))

				if verificar {
					if string(conteudo) != formatado {
						fmt.Fprintf(os.Stderr, "HRP-FMT-001: %s não está formatado.\n", caminho)
						desformatados++
					}
					continue
				}

				if escrever {
					if string(conteudo) != formatado {
						err = os.WriteFile(caminho, []byte(formatado), 0644)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Erro ao salvar arquivo %s: %v\n", caminho, err)
						} else {
							fmt.Printf("Formatado: %s\n", caminho)
						}
					}
				} else {
					fmt.Print(formatado)
				}
			}

			if verificar && desformatados > 0 {
				fmt.Fprintf(os.Stderr, "\n%d arquivo(s) precisam de formatação. Execute 'harpia formatar -w [caminho]'.\n", desformatados)
				os.Exit(1)
			}
		},
	}
	cmdFormatar.Flags().BoolVarP(&escrever, "escrever", "w", false, "Salva as alterações de volta nos arquivos originais")
	cmdFormatar.Flags().BoolVar(&verificar, "verificar", false, "Sai com código 1 se algum arquivo não estiver formatado (para CI)")
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
