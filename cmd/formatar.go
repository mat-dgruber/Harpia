package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// comandoFormatar inicializa o comando 'portuscript formatar'
func comandoFormatar() *cobra.Command {
	var escrever bool
	cmdFormatar := &cobra.Command{
		Use:   "formatar [arquivo.hrp]",
		Short: "Formata a identação e os blocos de um arquivo Harpia",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			caminho := args[0]
			conteudo, err := os.ReadFile(caminho)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler arquivo %s: %v\n", caminho, err)
				os.Exit(1)
			}

			formatado := FormatarCodigoPortuscript(string(conteudo))

			if escrever {
				err = os.WriteFile(caminho, []byte(formatado), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao salvar arquivo formatado %s: %v\n", caminho, err)
					os.Exit(1)
				}
				fmt.Printf("Arquivo '%s' formatado e salvo com sucesso.\n", caminho)
			} else {
				fmt.Print(formatado)
			}
		},
	}
	cmdFormatar.Flags().BoolVarP(&escrever, "escrever", "w", false, "Salva as alterações de volta no arquivo original")
	return cmdFormatar
}

// FormatarCodigoPortuscript formata recuos de blocos baseando-se em contagens de delimitadores
func FormatarCodigoPortuscript(codigo string) string {
	linhas := strings.Split(codigo, "\n")
	var res []string
	nivel := 0
	linhaAnteriorVazia := false

	for _, linhaRaw := range linhas {
		linha := strings.TrimSpace(linhaRaw)
		if linha == "" {
			if !linhaAnteriorVazia {
				res = append(res, "")
				linhaAnteriorVazia = true
			}
			continue
		}
		linhaAnteriorVazia = false

		fechamentos := strings.Count(linha, "}") + strings.Count(linha, "]") + strings.Count(linha, ")")
		aberturas := strings.Count(linha, "{") + strings.Count(linha, "[") + strings.Count(linha, "(")

		// Se a linha começar com fechamento, reduz o nível imediatamente
		if strings.HasPrefix(linha, "}") || strings.HasPrefix(linha, "]") || strings.HasPrefix(linha, ")") {
			nivel = maximo(0, nivel-1)
		} else if fechamentos > aberturas {
			nivel = maximo(0, nivel-(fechamentos-aberturas))
		}

		// Adiciona recuo de 4 espaços correspondente
		recuo := strings.Repeat("    ", nivel)
		res = append(res, recuo+linha)

		// Incrementa recuo se abriu mais blocos
		if aberturas > fechamentos {
			nivel += aberturas - fechamentos
		} else if strings.HasSuffix(linha, "{") || strings.HasSuffix(linha, "[") || strings.HasSuffix(linha, "(") {
			nivel++
		}
	}

	return strings.Join(res, "\n")
}

func maximo(a, b int) int {
	if a > b {
		return a
	}
	return b
}
