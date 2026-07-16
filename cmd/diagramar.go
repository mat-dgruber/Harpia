package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natanfeitosa/portuscript/parser"
	"github.com/natanfeitosa/portuscript/ptst"
	"github.com/spf13/cobra"
)

// ImportRel representa a relação estática de importação mapeada entre camadas
type ImportRel struct {
	De      string
	Para    string
	Arquivo string
}

// comandoDiagramar inicializa o comando 'portuscript diagramar'
func comandoDiagramar() *cobra.Command {
	var formato string
	cmdDiag := &cobra.Command{
		Use:   "diagramar [diretorio]",
		Short: "Mapeia as relações de importações e valida regras da Clean Architecture",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}

			fmt.Printf("Mapeando relações de dependência em: %s...\n", dir)
			rels, errs := analisarDependencias(dir)

			if len(errs) > 0 {
				fmt.Fprintln(os.Stderr, "\n🚨 Violações de Clean Architecture detectadas:")
				for _, errStr := range errs {
					fmt.Fprintf(os.Stderr, "  - %s\n", errStr)
				}
			} else {
				fmt.Println("\n✅ Nenhuma violação de Clean Architecture detectada!")
			}

			if formato == "mermaid" {
				fmt.Println("\n=== DIAGRAMA MERMAID ===")
				fmt.Println(gerarMermaid(rels))
			}
		},
	}
	cmdDiag.Flags().StringVarP(&formato, "formato", "f", "mermaid", "Formato de saída do diagrama (ex: mermaid)")
	return cmdDiag
}

// analisarDependencias varre recursivamente e encontra importações e violações arquiteturais
func analisarDependencias(raizDir string) ([]ImportRel, []string) {
	var rels []ImportRel
	var violacoes []string

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	filepath.Walk(raizDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".ptst") {
			return nil
		}

		conteudo, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		ast, err := ctx.StringParaAst(string(conteudo), path)
		if err != nil {
			return nil
		}

		importes := encontrarImportes(ast)

		for _, imp := range importes {
			caminhoDest := imp.Caminho.Valor
			if len(caminhoDest) >= 2 {
				caminhoDest = caminhoDest[1 : len(caminhoDest)-1]
			}

			if !strings.HasPrefix(caminhoDest, ".") && !strings.HasPrefix(caminhoDest, "/") {
				continue
			}

			dirOrigem := filepath.Dir(path)
			caminhoResolvido := filepath.Clean(filepath.Join(dirOrigem, caminhoDest))

			camadaOrigem := detectarCamada(path)
			camadaDest := detectarCamada(caminhoResolvido)

			if camadaOrigem != "" && camadaDest != "" && camadaOrigem != camadaDest {
				rels = append(rels, ImportRel{
					De:      camadaOrigem,
					Para:    camadaDest,
					Arquivo: path,
				})

				// Regras de Dependência da Clean Architecture:
				// 1. Dominio não pode importar de Infra ou Web (camadas externas)
				if camadaOrigem == "dominio" && (camadaDest == "infra" || camadaDest == "web") {
					violacoes = append(violacoes, fmt.Sprintf("camada '%s' em '%s' importando de '%s' (%s)", camadaOrigem, path, camadaDest, caminhoResolvido))
				}
				// 2. Infra não pode importar de Web
				if camadaOrigem == "infra" && camadaDest == "web" {
					violacoes = append(violacoes, fmt.Sprintf("camada '%s' em '%s' importando de '%s' (%s)", camadaOrigem, path, camadaDest, caminhoResolvido))
				}
			}
		}

		return nil
	})

	return rels, violacoes
}

// encontrarImportes extrai os nós de importação do topo da AST (Programa)
func encontrarImportes(node parser.BaseNode) []*parser.ImporteDe {
	var importes []*parser.ImporteDe
	if prog, ok := node.(*parser.Programa); ok {
		for _, decl := range prog.Declaracoes {
			if imp, ok := decl.(*parser.ImporteDe); ok {
				importes = append(importes, imp)
			}
		}
	}
	return importes
}

// detectarCamada identifica o pertencimento a domínio, infraestrutura ou interface com base em convenção de nomes
func detectarCamada(path string) string {
	cleanPath := filepath.ToSlash(path)
	if strings.Contains(cleanPath, "/dominio/") {
		return "dominio"
	}
	if strings.Contains(cleanPath, "/infra/") {
		return "infra"
	}
	if strings.Contains(cleanPath, "/web/") {
		return "web"
	}
	return ""
}

// gerarMermaid gera o código textual do gráfico Mermaid
func gerarMermaid(rels []ImportRel) string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	set := make(map[string]bool)
	for _, rel := range rels {
		key := fmt.Sprintf("  %s --> %s\n", rel.De, rel.Para)
		if !set[key] {
			set[key] = true
			sb.WriteString(key)
		}
	}

	if len(set) == 0 {
		sb.WriteString("  dominio\n  infra\n  web\n")
	}

	return sb.String()
}
