//go:build !js && !wasm

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
	_ "github.com/mat-dgruber/Harpia/stdlib"
	"github.com/spf13/cobra"
)

type ResultadoTesteJSON struct {
	Nome   string `json:"nome"`
	Status string `json:"status"`
	Erro   string `json:"erro,omitempty"`
}

type SuiteTestesJSON struct {
	Arquivo string               `json:"arquivo"`
	Testes  []ResultadoTesteJSON `json:"testes"`
}

type RelatorioTestesJSON struct {
	Total    int               `json:"total"`
	Passaram int               `json:"passaram"`
	Falharam int               `json:"falharam"`
	Suites   []SuiteTestesJSON `json:"suites"`
}

// comandoTestar cria e retorna o comando Cobra 'testar' (`harpia testar`).
// Este comando descobre recursivamente arquivos com blocos de testes 'testar "nome" { ... }',
// executa-os em escopos isolados capturando as asserções e gera relatórios de cobertura em HTML.
func comandoTestar() *cobra.Command {
	var htmlReport bool
	var formato string

	testar := &cobra.Command{
		Use:           "testar [caminho]",
		Short:         "Executa os blocos de teste 'testar' nativos no arquivo ou diretório",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(errOut, "erro ao obter o diretório atual:", err)
				return err
			}

			caminho := "."
			if len(args) > 0 {
				caminho = args[0]
			}

			caminhosAbs, err := encontrarArquivosTeste(caminho)
			if err != nil {
				fmt.Fprintln(errOut, "erro ao listar arquivos de teste:", err)
				return err
			}

			totalTestes := 0
			testesFalhos := 0

			coberturaLinhas := make(map[string]map[int]bool)
			arquivosCodigos := make(map[string]string)
			var reportJSON RelatorioTestesJSON

			for _, arq := range caminhosAbs {
				ctx := hrp.NewContexto(hrp.OpcsContexto{CaminhosPadrao: []string{cur}})

				// Carrega o arquivo e parseia em AST
				_, ast, err := ctx.TransformarEmAst(arq, false, cur)
				if err != nil {
					if formato == "json" {
						reportJSON.Suites = append(reportJSON.Suites, SuiteTestesJSON{
							Arquivo: arq,
							Testes: []ResultadoTesteJSON{
								{Nome: "sintaxe", Status: "falhou", Erro: err.Error()},
							},
						})
					} else {
						fmt.Fprintf(out, "Erro de sintaxe no arquivo %s: %v\n", arq, err)
					}
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
				modulo, err := ctx.InicializarModulo(&hrp.ModuloImpl{
					Info: hrp.ModuloInfo{Arquivo: arq},
					Ast:  &parser.Bloco{Declaracoes: declsNormais},
				})

				if err != nil {
					if formato == "json" {
						reportJSON.Suites = append(reportJSON.Suites, SuiteTestesJSON{
							Arquivo: arq,
							Testes: []ResultadoTesteJSON{
								{Nome: "inicializacao", Status: "falhou", Erro: err.Error()},
							},
						})
					} else {
						fmt.Fprintf(out, "Erro ao inicializar definições de %s: %v\n", arq, err)
					}
					ctx.Terminar()
					continue
				}

				if formato != "json" {
					fmt.Fprintf(out, "Rodando testes em %s:\n", filepath.Base(arq))
				}

				var suiteJSON SuiteTestesJSON
				suiteJSON.Arquivo = arq

				for _, tNode := range testes {
					totalTestes++
					// Cria um escopo isolado que herda as variáveis e funções globais do módulo
					escopo := modulo.Escopo.NewEscopo()

					_, err := ctx.AvaliarAst(tNode.Corpo, escopo)
					if err != nil {
						testesFalhos++
						if formato == "json" {
							suiteJSON.Testes = append(suiteJSON.Testes, ResultadoTesteJSON{
								Nome:   tNode.Nome,
								Status: "falhou",
								Erro:   err.Error(),
							})
						} else {
							fmt.Fprintf(out, "  ❌ [FALHOU] %s\n", tNode.Nome)
							fmt.Fprintln(errOut, err)
						}
					} else {
						if formato == "json" {
							suiteJSON.Testes = append(suiteJSON.Testes, ResultadoTesteJSON{
								Nome:   tNode.Nome,
								Status: "passou",
							})
						} else {
							fmt.Fprintf(out, "  ✅ [PASSOU] %s\n", tNode.Nome)
						}
					}
				}

				reportJSON.Suites = append(reportJSON.Suites, suiteJSON)

				// ponytail: acumula cobertura de linhas visitadas
				for f, lMap := range ctx.ObterLinhasExecutadas() {
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

			reportJSON.Total = totalTestes
			reportJSON.Falharam = testesFalhos
			reportJSON.Passaram = totalTestes - testesFalhos

			if formato == "json" {
				bytes, _ := json.MarshalIndent(reportJSON, "", "  ")
				fmt.Fprintln(out, string(bytes))
				if testesFalhos > 0 {
					return fmt.Errorf("%d teste(s) falharam", testesFalhos)
				}
				return nil
			}

			fmt.Fprintf(out, "\n--- Relatório de Testes ---\n")
			fmt.Fprintf(out, "Total de testes executados: %d\n", totalTestes)
			fmt.Fprintf(out, "Passaram: %d\n", totalTestes-testesFalhos)
			fmt.Fprintf(out, "Falharam: %d\n", testesFalhos)

			if htmlReport && len(arquivosCodigos) > 0 {
				gerarRelatorioCoberturaHTML(coberturaLinhas, arquivosCodigos, out, errOut)
			}

			if testesFalhos > 0 {
				return fmt.Errorf("%d teste(s) falharam", testesFalhos)
			}
			return nil
		},
	}

	testar.Flags().BoolVar(&htmlReport, "html", false, "Gera relatório visual de cobertura em cobertura.html")
	testar.Flags().StringVar(&formato, "formato", "texto", "Formato de saída: 'texto' ou 'json'")
	return testar
}

func gerarRelatorioCoberturaHTML(cobertura map[string]map[int]bool, codigos map[string]string, out, errOut io.Writer) {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="pt-BR" class="h-full bg-slate-950 text-slate-100">
<head>
    <meta charset="UTF-8">
    <title>Harpia — Relatório de Cobertura de Testes</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="p-8 max-w-6xl mx-auto">
    <header class="mb-8 border-b border-slate-800 pb-6">
        <h1 class="text-3xl font-black text-blue-500">Relatório de Cobertura de Testes</h1>
        <p class="text-slate-400 mt-1">Garantia e integridade de código do ecossistema Harpia</p>
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
		fmt.Fprintf(errOut, "Erro ao gravar relatório cobertura.html: %v\n", err)
	} else {
		fmt.Fprintln(out, "\n✅ Relatório estético de cobertura gerado com sucesso em: ./cobertura.html")
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
		if !d.IsDir() && (strings.HasSuffix(caminho, ".hrp") || strings.HasSuffix(caminho, ".pt")) {
			arquivos = append(arquivos, caminho)
		}
		return nil
	})

	return arquivos, err
}
