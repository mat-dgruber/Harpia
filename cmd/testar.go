package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/natanfeitosa/portuscript/parser"
	"github.com/natanfeitosa/portuscript/ptst"
	_ "github.com/natanfeitosa/portuscript/stdlib"
	"github.com/spf13/cobra"
)

func comandoTestar() *cobra.Command {
	var htmlReport bool
	testar := &cobra.Command{
		Use:   "testar [caminho]",
		Short: "Executa os blocos de teste 'testar' nativos no arquivo ou diretório",
		Run: func(cmd *cobra.Command, args []string) {
			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao obter o diretório atual:", err)
				os.Exit(1)
			}

			caminho := "."
			if len(args) > 0 {
				caminho = args[0]
			}

			caminhosAbs, err := encontrarArquivosTeste(caminho)
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao listar arquivos de teste:", err)
				os.Exit(1)
			}

			totalTestes := 0
			testesFalhos := 0

			coberturaLinhas := make(map[string]map[int]bool)
			arquivosCodigos := make(map[string]string)

			for _, arq := range caminhosAbs {
				ctx := ptst.NewContexto(ptst.OpcsContexto{CaminhosPadrao: []string{cur}})

				// Carrega o arquivo e parseia em AST
				_, ast, err := ctx.TransformarEmAst(arq, false, cur)
				if err != nil {
					fmt.Printf("Erro de sintaxe no arquivo %s: %v\n", arq, err)
					continue
				}

				// Filtra declarações normais e de testes para evitar execuções duplicadas
				var testes []*parser.DeclTeste
				var declsNormais []parser.BaseNode

				if prog, ok := ast.(*parser.Programa); ok {
					arquivosCodigos[prog.Arquivo] = prog.Codigo
					for _, decl := range prog.Declaracoes {
						if tNode, ok := decl.(*parser.DeclTeste); ok {
							testes = append(testes, tNode)
						} else {
							declsNormais = append(declsNormais, decl)
						}
					}
				}

				if len(testes) == 0 {
					ctx.Terminar()
					continue
				}

				// Inicializa o módulo mestre avaliando apenas as importações e funções normais
				modulo, err := ctx.InicializarModulo(&ptst.ModuloImpl{
					Info: ptst.ModuloInfo{Arquivo: arq},
					Ast:  &parser.Bloco{Declaracoes: declsNormais},
				})

				if err != nil {
					fmt.Printf("Erro ao inicializar definições de %s: %v\n", arq, err)
					ctx.Terminar()
					continue
				}

				fmt.Printf("Rodando testes em %s:\n", filepath.Base(arq))

				for _, tNode := range testes {
					totalTestes++
					// Cria um escopo isolado que herda as variáveis e funções globais do módulo
					escopo := modulo.Escopo.NewEscopo()

					interpret := &ptst.Interpretador{
						Ast:      tNode.Corpo,
						Contexto: ctx,
						Escopo:   escopo,
					}

					if prog, ok := ast.(*parser.Programa); ok {
						interpret.Arquivo = prog.Arquivo
						interpret.Codigo = prog.Codigo
						interpret.Posicoes = prog.Posicoes
					}

					_, err := interpret.Inicializa()
					if err != nil {
						testesFalhos++
						fmt.Printf("  ❌ [FALHOU] %s\n", tNode.Nome)
						fmt.Fprintln(os.Stderr, err) // Exibe o erro de traceback sem interromper abruptamente a suite
					} else {
						fmt.Printf("  ✅ [PASSOU] %s\n", tNode.Nome)
					}
				}

				// ponytail: acumula cobertura de linhas visitadas
				for f, lMap := range ctx.LinhasExecutadas {
					if coberturaLinhas[f] == nil {
						coberturaLinhas[f] = make(map[int]bool)
					}
					for l, exec := range lMap {
						if exec {
							coberturaLinhas[f][l] = true
						}
					}
				}

				ctx.Terminar()
			}

			fmt.Printf("\n--- Relatório de Testes ---\n")
			fmt.Printf("Total de testes executados: %d\n", totalTestes)
			fmt.Printf("Passaram: %d\n", totalTestes-testesFalhos)
			fmt.Printf("Falharam: %d\n", testesFalhos)

			if htmlReport && len(arquivosCodigos) > 0 {
				gerarRelatorioCoberturaHTML(coberturaLinhas, arquivosCodigos)
			}

			if testesFalhos > 0 {
				os.Exit(1)
			}
		},
	}

	testar.Flags().BoolVar(&htmlReport, "html", false, "Gera relatório visual de cobertura em cobertura.html")
	return testar
}

func gerarRelatorioCoberturaHTML(cobertura map[string]map[int]bool, codigos map[string]string) {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="pt-BR" class="h-full bg-slate-950 text-slate-100">
<head>
    <meta charset="UTF-8">
    <title>Portuscript — Relatório de Cobertura de Testes</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="p-8 max-w-6xl mx-auto">
    <header class="mb-8 border-b border-slate-800 pb-6">
        <h1 class="text-3xl font-black text-blue-500">Relatório de Cobertura de Testes</h1>
        <p class="text-slate-400 mt-1">Garantia e integridade de código do ecossistema Portuscript</p>
    </header>
    <div class="space-y-8">
`)

	for arq, codigo := range codigos {
		executadas := cobertura[arq]
		if executadas == nil {
			executadas = make(map[int]bool)
		}

		linhas := strings.Split(codigo, "\n")
		totalExecutaveis := 0
		totalCobertas := 0

		linhasExecutaveis := make(map[int]bool)
		for idx, linhaRaw := range linhas {
			lNo := idx + 1
			linha := strings.TrimSpace(linhaRaw)
			if linha != "" && !strings.HasPrefix(linha, "#") && !strings.HasPrefix(linha, "//") && linha != "}" && linha != "{" {
				linhasExecutaveis[lNo] = true
				totalExecutaveis++
				if executadas[lNo] {
					totalCobertas++
				}
			}
		}

		pct := 100.0
		if totalExecutaveis > 0 {
			pct = (float64(totalCobertas) / float64(totalExecutaveis)) * 100.0
		}

		sb.WriteString(fmt.Sprintf(`
        <section class="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
            <div class="bg-slate-800/50 px-6 py-4 border-b border-slate-800 flex justify-between items-center">
                <div>
                    <h2 class="text-lg font-bold text-slate-200">%s</h2>
                    <p class="text-xs text-slate-400 mt-0.5">%s</p>
                </div>
                <div class="text-right">
                    <span class="text-2xl font-black %s">%.1f%%</span>
                    <p class="text-xs text-slate-400 mt-0.5">%d de %d linhas cobertas</p>
                </div>
            </div>
            <div class="p-6 overflow-x-auto font-mono text-xs leading-relaxed select-none">
                <table class="w-full border-collapse">
`, filepath.Base(arq), arq, corPorPorcentagem(pct), pct, totalCobertas, totalExecutaveis))

		for idx, linha := range linhas {
			lNo := idx + 1
			classeFundo := "hover:bg-slate-800/30"
			statusCor := "text-slate-600"

			if linhasExecutaveis[lNo] {
				if executadas[lNo] {
					classeFundo = "bg-green-950/20 hover:bg-green-950/30 text-green-300 border-l-2 border-green-500"
					statusCor = "text-green-500 font-bold"
				} else {
					classeFundo = "bg-red-950/20 hover:bg-red-950/30 text-red-300 border-l-2 border-red-500"
					statusCor = "text-red-500 font-bold"
				}
			}

			linhaEscapada := strings.ReplaceAll(linha, "<", "&lt;")
			linhaEscapada = strings.ReplaceAll(linhaEscapada, ">", "&gt;")

			sb.WriteString(fmt.Sprintf(`
                    <tr class="%s">
                        <td class="w-12 text-right pr-4 text-slate-500 select-none border-r border-slate-800/40">%d</td>
                        <td class="w-8 text-center select-none %s">%s</td>
                        <td class="pl-4 whitespace-pre">%s</td>
                    </tr>
`, classeFundo, lNo, statusCor, indicadorStatus(linhasExecutaveis[lNo], executadas[lNo]), linhaEscapada))
		}

		sb.WriteString(`
                </table>
            </div>
        </section>
`)
	}

	sb.WriteString(`
    </div>
</body>
</html>
`)

	err := os.WriteFile("cobertura.html", []byte(sb.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gravar relatório cobertura.html: %v\n", err)
	} else {
		fmt.Println("\n✅ Relatório estético de cobertura gerado com sucesso em: ./cobertura.html")
	}
}

func corPorPorcentagem(pct float64) string {
	if pct >= 90 {
		return "text-green-400"
	}
	if pct >= 70 {
		return "text-yellow-400"
	}
	return "text-red-400"
}

func indicadorStatus(executavel, coberta bool) string {
	if !executavel {
		return " "
	}
	if coberta {
		return "✓"
	}
	return "✗"
}

func encontrarArquivosTeste(raiz string) ([]string, error) {
	var arquivos []string
	info, err := os.Stat(raiz)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{raiz}, nil
	}

	err = filepath.WalkDir(raiz, func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(caminho, ".ptst") || strings.HasSuffix(caminho, ".pt")) {
			arquivos = append(arquivos, caminho)
		}
		return nil
	})

	return arquivos, err
}
