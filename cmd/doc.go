package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// DocElement representa um elemento documentado extraído do código
type DocElement struct {
	Tipo       string   // "funcao", "classe", "constante", "variavel"
	Nome       string   
	Assinatura string   
	Descricao  []string 
}

// comandoDoc inicializa o comando 'Harpia doc'
func comandoDoc() *cobra.Command {
	var formato string
	var saida string

	cmdDoc := &cobra.Command{
		Use:   "doc [arquivo.hrp]",
		Short: "Gera documentação automática a partir de comentários especiais '///'",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			caminho := args[0]
			elementos, err := extrairDocumentacao(caminho)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler arquivo %s: %v\n", caminho, err)
				os.Exit(1)
			}

			if len(elementos) == 0 {
				fmt.Println("Nenhum comentário de documentação '///' encontrado no arquivo.")
				return
			}

			var doc string
			if formato == "html" {
				doc = gerarDocHTML(caminho, elementos)
			} else {
				doc = gerarDocMarkdown(caminho, elementos)
			}

			if saida != "" {
				err = os.WriteFile(saida, []byte(doc), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao gravar documentação em %s: %v\n", saida, err)
					os.Exit(1)
				}
				fmt.Printf("Documentação salva com sucesso em '%s'\n", saida)
			} else {
				fmt.Print(doc)
			}
		},
	}

	cmdDoc.Flags().StringVarP(&formato, "formato", "f", "markdown", "Formato da documentação (markdown ou html)")
	cmdDoc.Flags().StringVarP(&saida, "saida", "s", "", "Caminho do arquivo de saída para salvar a documentação")
	return cmdDoc
}

func extrairDocumentacao(caminho string) ([]DocElement, error) {
	file, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var elementos []DocElement
	var docsAcumuladas []string

	scanner := bufio.NewScanner(file)

	// Regex para capturar definições de funções, classes, constantes e variáveis
	reFuncao := regexp.MustCompile(`(?:exportar\s+)?(?:assincrono\s+)?funcao\s+([a-zA-Z0-9_]+)\s*\((.*)\)`)
	reClasse := regexp.MustCompile(`(?:exportar\s+)?classe\s+([a-zA-Z0-9_]+)(?:\s+estende\s+([a-zA-Z0-9_]+))?`)
	reConstante := regexp.MustCompile(`(?:exportar\s+)?constante\s+([a-zA-Z0-9_]+)`)
	reVar := regexp.MustCompile(`(?:exportar\s+)?var\s+([a-zA-Z0-9_]+)`)

	for scanner.Scan() {
		linhaRaw := scanner.Text()
		linha := strings.TrimSpace(linhaRaw)

		// Se for comentário especial de documentação
		if strings.HasPrefix(linha, "///") {
			textoComentario := strings.TrimPrefix(linha, "///")
			docsAcumuladas = append(docsAcumuladas, strings.TrimSpace(textoComentario))
			continue
		}

		// Se a linha for vazia, apenas ignora mas mantém as documentações acumuladas
		if linha == "" {
			continue
		}

		// Se contiver código e tivermos documentações acumuladas, tenta associar
		if len(docsAcumuladas) > 0 {
			if match := reFuncao.FindStringSubmatch(linha); len(match) > 0 {
				elementos = append(elementos, DocElement{
					Tipo:       "funcao",
					Nome:       match[1],
					Assinatura: fmt.Sprintf("funcao %s(%s)", match[1], match[2]),
					Descricao:  docsAcumuladas,
				})
				docsAcumuladas = nil
			} else if match := reClasse.FindStringSubmatch(linha); len(match) > 0 {
				heranca := ""
				if match[2] != "" {
					heranca = " estende " + match[2]
				}
				elementos = append(elementos, DocElement{
					Tipo:       "classe",
					Nome:       match[1],
					Assinatura: fmt.Sprintf("classe %s%s", match[1], heranca),
					Descricao:  docsAcumuladas,
				})
				docsAcumuladas = nil
			} else if match := reConstante.FindStringSubmatch(linha); len(match) > 0 {
				elementos = append(elementos, DocElement{
					Tipo:       "constante",
					Nome:       match[1],
					Assinatura: fmt.Sprintf("constante %s", match[1]),
					Descricao:  docsAcumuladas,
				})
				docsAcumuladas = nil
			} else if match := reVar.FindStringSubmatch(linha); len(match) > 0 {
				elementos = append(elementos, DocElement{
					Tipo:       "variavel",
					Nome:       match[1],
					Assinatura: fmt.Sprintf("var %s", match[1]),
					Descricao:  docsAcumuladas,
				})
				docsAcumuladas = nil
			} else {
				// Qualquer outro código com comentário acumulado reseta a fila se não for uma declaração válida
				docsAcumuladas = nil
			}
		}
	}

	return elementos, scanner.Err()
}

func gerarDocMarkdown(arquivo string, elementos []DocElement) string {
	var sb strings.Builder
	nomeBase := filepath.Base(arquivo)
	sb.WriteString(fmt.Sprintf("# 📖 Documentação do Módulo `%s`\n\n", nomeBase))
	sb.WriteString("Esta documentação foi gerada automaticamente a partir dos comentários especiais `///` do código-fonte.\n\n---\n\n")

	for _, el := range elementos {
		sb.WriteString(fmt.Sprintf("## 🏷️ %s `%s`\n\n", strings.Title(el.Tipo), el.Nome))
		sb.WriteString(fmt.Sprintf("```Harpia\n%s\n```\n\n", el.Assinatura))
		sb.WriteString("### Descrição\n")
		for _, desc := range el.Descricao {
			sb.WriteString(desc + "\n")
		}
		sb.WriteString("\n---\n\n")
	}

	return sb.String()
}

func gerarDocHTML(arquivo string, elementos []DocElement) string {
	var sb strings.Builder
	nomeBase := filepath.Base(arquivo)
	sb.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Documentação de %s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; max-width: 800px; margin: 40px auto; padding: 0 20px; color: #333; background: #fafafa; }
        h1 { color: #1e3a8a; border-bottom: 2px solid #e2e8f0; padding-bottom: 10px; }
        h2 { color: #2563eb; margin-top: 40px; border-bottom: 1px solid #f1f5f9; padding-bottom: 5px; }
        code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-size: 0.95em; }
        pre { background: #1e293b; color: #f8fafc; padding: 15px; border-radius: 8px; overflow-x: auto; }
        pre code { background: none; color: inherit; padding: 0; }
        .tag { display: inline-block; background: #dbeafe; color: #1e40af; font-size: 0.8em; font-weight: bold; padding: 2px 8px; border-radius: 9999px; margin-bottom: 10px; text-transform: uppercase; }
        .meta { color: #64748b; font-size: 0.9em; margin-bottom: 30px; }
    </style>
</head>
<body>
    <h1>📖 Documentação de %s</h1>
    <p class="meta">Gerado automaticamente a partir do código fonte original.</p>
`, nomeBase, nomeBase))

	for _, el := range elementos {
		sb.WriteString(fmt.Sprintf(`    <div class="elemento">
        <h2>%s</h2>
        <span class="tag">%s</span>
        <pre><code>%s</code></pre>
        <p>%s</p>
    </div>
`, el.Nome, el.Tipo, el.Assinatura, strings.Join(el.Descricao, "<br>")))
	}

	sb.WriteString(`</body>
</html>`)
	return sb.String()
}
