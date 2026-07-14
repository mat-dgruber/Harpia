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
				ctx.Terminar()
			}

			fmt.Printf("\n--- Relatório de Testes ---\n")
			fmt.Printf("Total de testes executados: %d\n", totalTestes)
			fmt.Printf("Passaram: %d\n", totalTestes-testesFalhos)
			fmt.Printf("Falharam: %d\n", testesFalhos)

			if testesFalhos > 0 {
				os.Exit(1)
			}
		},
	}

	return testar
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
