package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
	"github.com/spf13/cobra"
)

// ImportRel representa a relação estática de importação mapeada entre camadas
type ImportRel struct {
	De      string
	Para    string
	Arquivo string
}

// comandoDiagramar inicializa o comando 'Harpia diagramar'
// comandoDiagramar inicializa `harpia diagramar`.
//
// O subcomando é responsável por varrer o projeto em busca de imports relativos
// e mapeá-los em um grafo arquitetural, sinalizando visualmente violações da
// Clean Architecture (camada `dominio` importando `infra` ou `web`, e camada
// `infra` importando `web`).
//
// Suporta 3 formatos de saída:
//   - `mermaid` (padrão): emite apenas o código-fonte textual do grafo;
//   - `html`: gera arquivo standalone com header, botões de download SVG e
//     área de alertas;
//   - `svg`: gera o HTML com dica para baixar o SVG vetorial ao abrir.
func comandoDiagramar() *cobra.Command {
	var formato string
	var saida string
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

			codigoMermaid := gerarMermaid(rels)

			if formato == "mermaid" {
				fmt.Println("\n=== DIAGRAMA MERMAID ===")
				fmt.Println(codigoMermaid)
				if saida != "" {
					os.WriteFile(saida, []byte(codigoMermaid), 0644)
					fmt.Printf("Diagrama salvo com sucesso em: %s\n", saida)
				}
			} else if formato == "html" || formato == "svg" {
				alertas := gerarAlertasHTML(errs)
				htmlFinal := strings.Replace(templateHTMLDiagrama, "{{MERMAID_CODE}}", codigoMermaid, 1)
				htmlFinal = strings.Replace(htmlFinal, "{{ALERTS_MARKUP}}", alertas, 1)

				destino := "diagrama.html"
				if saida != "" {
					destino = saida
				}

				err := os.WriteFile(destino, []byte(htmlFinal), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao gravar diagrama HTML: %v\n", err)
					return
				}
				fmt.Printf("\n📊 Diagrama interativo de arquitetura gerado com sucesso em: %s\n", destino)
				if formato == "svg" {
					fmt.Println("👉 Dica: Abra o arquivo HTML gerado no seu navegador e clique em 'Baixar SVG' para obter a imagem vetorial.")
				}
			}
		},
	}
	cmdDiag.Flags().StringVarP(&formato, "formato", "f", "mermaid", "Formato de saída do diagrama (mermaid, html, svg)")
	cmdDiag.Flags().StringVarP(&saida, "saida", "s", "", "Caminho do arquivo para salvar a saída")
	return cmdDiag
}

// analisarDependencias varre recursivamente e encontra importações e violações arquiteturais
func analisarDependencias(raizDir string) ([]ImportRel, []string) {
	var rels []ImportRel
	var violacoes []string

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	filepath.Walk(raizDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".hrp") {
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

// gerarMermaid gera o código textual do gráfico Mermaid com colorização dinâmica de violações
func gerarMermaid(rels []ImportRel) string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	set := make(map[string]bool)
	linkIndex := 0
	var linkStyles []string

	for _, rel := range rels {
		key := fmt.Sprintf("  %s --> %s\n", rel.De, rel.Para)
		if !set[key] {
			set[key] = true
			sb.WriteString(key)

			// Detecta violações arquiteturais e escolhe cor correspondente
			isViolacao := false
			if rel.De == "dominio" && (rel.Para == "infra" || rel.Para == "web") {
				isViolacao = true
			} else if rel.De == "infra" && rel.Para == "web" {
				isViolacao = true
			}

			if isViolacao {
				linkStyles = append(linkStyles, fmt.Sprintf("  linkStyle %d stroke:#ff3333,stroke-width:3px;", linkIndex))
			} else {
				linkStyles = append(linkStyles, fmt.Sprintf("  linkStyle %d stroke:#33ff33,stroke-width:2px;", linkIndex))
			}
			linkIndex++
		}
	}

	if len(set) == 0 {
		sb.WriteString("  dominio\n  infra\n  web\n")
	} else {
		for _, style := range linkStyles {
			sb.WriteString(style + "\n")
		}
	}

	return sb.String()
}

func gerarAlertasHTML(errs []string) string {
	if len(errs) == 0 {
		return "<div class=\"sucesso\"><b>✅ Sucesso:</b> Nenhuma violação de Clean Architecture detectada no seu projeto! Tudo limpo e organizado.</div>"
	}

	var sb strings.Builder
	sb.WriteString("<div class=\"violacao\">")
	sb.WriteString("<b>🚨 Violações de Clean Architecture detectadas:</b>")
	sb.WriteString("<ul style=\"margin: 10px 0 0 0; padding-left: 20px;\">")
	for _, errStr := range errs {
		sb.WriteString(fmt.Sprintf("<li style=\"margin-bottom: 5px;\">%s</li>", errStr))
	}
	sb.WriteString("</ul>")
	sb.WriteString("</div>")
	return sb.String()
}

// ponytail: template html minimalista offline-friendly, renderiza o Mermaid via CDN
const templateHTMLDiagrama = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Harpia — Diagrama de Arquitetura</title>
  <script type="module">
    import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
    mermaid.initialize({ startOnLoad: true, theme: 'dark' });

    window.baixarSVG = function() {
      const svgEl = document.querySelector('.mermaid svg');
      if (!svgEl) return;
      const svgString = new XMLSerializer().serializeToString(svgEl);
      const blob = new Blob([svgString], {type: 'image/svg+xml;charset=utf-8'});
      const url = URL.createObjectURL(blob);
      const downloadLink = document.createElement('a');
      downloadLink.href = url;
      downloadLink.download = 'diagrama-arquitetura.svg';
      document.body.appendChild(downloadLink);
      downloadLink.click();
      document.body.removeChild(downloadLink);
    };
  </script>
  <style>
    body { background: #121212; color: #e0e0e0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 100vh; margin: 0; padding: 20px; box-sizing: border-box; }
    h1 { color: #ffffff; margin-bottom: 5px; font-size: 24px; }
    p { color: #888888; margin-top: 0; margin-bottom: 25px; font-size: 14px; }
    .container { background: #1e1e1e; border: 1px solid #333; border-radius: 8px; padding: 30px; width: 90%; max-width: 1000px; box-shadow: 0 4px 20px rgba(0,0,0,0.5); display: flex; flex-direction: column; align-items: center; }
    .mermaid { width: 100%; margin: 20px 0; background: #161616; padding: 20px; border-radius: 6px; border: 1px solid #252525; display: flex; justify-content: center; }
    .botoes { display: flex; gap: 10px; margin-top: 15px; }
    button { background: #007acc; color: white; border: none; padding: 10px 20px; border-radius: 4px; font-size: 14px; cursor: pointer; transition: background 0.2s; font-weight: bold; }
    button:hover { background: #0062a3; }
    .alertas { width: 100%; max-width: 1000px; margin-top: 20px; text-align: left; }
    .violacao { background: #3c1e1e; border: 1px solid #ff3333; color: #ff9999; padding: 15px; border-radius: 6px; margin-bottom: 10px; font-size: 14px; }
    .sucesso { background: #1e3c1e; border: 1px solid #33ff33; color: #99ff99; padding: 15px; border-radius: 6px; margin-bottom: 10px; font-size: 14px; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Diagrama de Arquitetura do Harpia</h1>
    <p>Visualização interativa das relações de imports entre as camadas de domínio, infraestrutura e web.</p>

    <pre class="mermaid">
{{MERMAID_CODE}}
    </pre>

    <div class="botoes">
      <button onclick="window.baixarSVG()">Baixar SVG</button>
    </div>
  </div>

  <div class="alertas">
    {{ALERTS_MARKUP}}
  </div>
</body>
</html>
`
