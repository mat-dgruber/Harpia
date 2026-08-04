//go:build !js && !wasm

package cmd

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
	"github.com/spf13/cobra"
)

func gerarImportHeaderWeb(relPath string) string {
	depth := strings.Count(relPath, "/")
	prefix := "./" + strings.Repeat("../", depth)
	return "import { h, sinal, sinalDebounce, sinalPersistente, efeito, derivado, armazem, montar, navegar, roteador, usarFormulario, usarConsulta, usarTema, usarArrastar, usarNotificacao, NotificacaoToast, Portal, BarraDeProgresso, Link, Aguardar, Escolha, Caso, Padrao, recurso, requisitar, sanitizarHtml, sanitizarSqlArgumento, Animado, mola, escalonar, animar, linhaTempo } from '" + prefix + "runtime-web.js';\n\n"
}

// OpcsCompilarWeb agrupa as opções para a função compilarParaWeb.
type OpcsCompilarWeb struct {
	Entrada        string
	Saida          string
	Estrito        bool
	OtimizarAssets bool
	PularLinter    bool
}

// compilarParaWeb executa a transpilação Harpia→JS sem nenhum os.Exit.
// Retorna erro em qualquer falha — seguro para ser chamado dentro de um servidor HTTP.
func compilarParaWeb(opts OpcsCompilarWeb) error {
	cur, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("erro ao obter diretório atual: %w", err)
	}

	if opts.Entrada == "" {
		return fmt.Errorf("arquivo de entrada não especificado")
	}

	if !opts.PularLinter {
		if conteudoLint, errL := os.ReadFile(opts.Entrada); errL == nil {
			if prog, errP := parser.NewParserFromString(string(conteudoLint), opts.Entrada).Parse(); errP == nil && prog != nil {

				linter := &Linter{
					Posicoes:          prog.Posicoes,
					DiretorioAtual:    filepath.Dir(opts.Entrada),
					ArquivosVisitados: map[string]bool{filepath.Clean(opts.Entrada): true},
				}
				linter.Escopo = &EscopoLinter{Variaveis: make(map[string]bool), Consts: make(map[string]bool)}
				linter.Checar(prog)
				var errosFatais []LinterError
				for _, e := range linter.Erros {
					if e.Severity == 1 {
						errosFatais = append(errosFatais, e)
					}
				}
				if len(errosFatais) > 0 {
					var msgs []string
					for _, e := range errosFatais {
						msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Message))
					}
					return fmt.Errorf("HRP-LINT-002: %d erro(s) de linter em %s:\n%s",
						len(errosFatais), opts.Entrada, strings.Join(msgs, "\n"))
				}
			}
		}
	}

	conteudoEntrada, err := os.ReadFile(opts.Entrada)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo '%s': %w", opts.Entrada, err)
	}

	ast, err := parser.NewParserFromString(string(conteudoEntrada), opts.Entrada).Parse()
	if err != nil {
		return fmt.Errorf("erro de compilação/sintaxe: %w", err)
	}

	transpiler := &TranspilerWeb{
		Estrito:          opts.Estrito,
		DiretorioBase:    filepath.Dir(opts.Entrada),
		DiretorioProjeto: filepath.Dir(opts.Entrada),
	}
	jsOutput := transpiler.Transpile(ast)

	// ponytail: garante visibilidade do componente raiz no bundle final.
	if strings.Contains(jsOutput, "function MeuApp(") && !strings.Contains(jsOutput, "export function MeuApp(") {
		jsOutput = stringsReplaceOnce(jsOutput, "function MeuApp(", "export function MeuApp(")
	}

	entryDir, _ := filepath.Abs(filepath.Dir(opts.Entrada))
	projDir := entryDir
	for {
		if _, err := os.Stat(filepath.Join(projDir, "web")); err == nil {
			break
		}
		parent := filepath.Dir(projDir)
		if parent == projDir || parent == "." || parent == "/" {
			projDir = entryDir
			break
		}
		projDir = parent
	}

	rotasDir := ""
	rotasPaths := []string{
		filepath.Join(projDir, "rotas"),
		filepath.Join(projDir, "web", "rotas"),
	}
	for _, p := range rotasPaths {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			rotasDir = p
			break
		}
	}

	type RotaInfo struct {
		Caminho string
		Nome    string
	}
	var rotas []RotaInfo

	if rotasDir != "" {
		err = filepath.Walk(rotasDir, func(path string, info os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			if err != nil || info.IsDir() || ext != ".hrp" {
				return nil
			}

			if filepath.Base(path) == "rotas.hrp" || strings.HasSuffix(path, ".estilo.hrp") || strings.HasSuffix(path, ".estilos.hrp") {
				return nil
			}

			if filepath.Base(path) == "layout.hrp" {
				conteudoLayout, err := os.ReadFile(path)
				if err == nil {
					layoutAst, err := parser.NewParserFromString(string(conteudoLayout), path).Parse()
					if err == nil {
						layoutJs := transpiler.Transpile(layoutAst)
						// ponytail: extrai imports e constantes de estilo para o topo do bundle
						var layoutImports []string
						var layoutBodyLines []string
						for _, line := range strings.Split(layoutJs, "\n") {
							trimmed := strings.TrimSpace(line)
							if (strings.HasPrefix(trimmed, "import ") && strings.HasSuffix(trimmed, ";")) ||
								(strings.HasPrefix(trimmed, "const ") && strings.Contains(trimmed, " = \"") && strings.HasSuffix(trimmed, ";")) {
								layoutImports = append(layoutImports, line)
							} else {
								layoutBodyLines = append(layoutBodyLines, line)
							}
						}
						for _, imp := range layoutImports {
							if strings.HasPrefix(strings.TrimSpace(imp), "import ") {
								imp = corrigirCaminhoImport(imp, path, projDir)
							}
							if !importJaExiste(jsOutput, imp) {
								jsOutput = imp + "\n" + jsOutput
							}
						}
						jsOutput += fmt.Sprintf("\nfunction LayoutGlobal(props) {\n%s\n}\n", strings.Join(layoutBodyLines, "\n"))
					}
				}
				return nil
			}

			conteudoRota, err := os.ReadFile(path)

			if err != nil {
				return nil
			}
			routeAst, err := parser.NewParserFromString(string(conteudoRota), path).Parse()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro de sintaxe na rota %s: %v\n", path, err)
				return nil
			}
			routeJs := transpiler.Transpile(routeAst)

			// ponytail: extrai imports do corpo da rota e coloca no topo do bundle
			// (imports JS não podem ficar dentro de corpos de função)
			var routeImports []string
			var routeBodyLines []string
			for _, line := range strings.Split(routeJs, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "import ") && strings.HasSuffix(trimmed, ";") {
					routeImports = append(routeImports, line)
				} else if strings.HasPrefix(trimmed, "const ") && strings.Contains(trimmed, " = \"") && strings.HasSuffix(trimmed, ";") {
					// ponytail: constantes de estilo (ex: const SidebarContainer = "SidebarContainer") também vão pro topo
					routeImports = append(routeImports, line)
				} else {
					routeBodyLines = append(routeBodyLines, line)
				}
			}
			routeJs = strings.Join(routeBodyLines, "\n")
			for _, imp := range routeImports {
				if strings.HasPrefix(strings.TrimSpace(imp), "import ") {
					imp = corrigirCaminhoImport(imp, path, projDir)
				}
				// evita duplicatas de import no bundle
				if !importJaExiste(jsOutput, imp) {
					jsOutput = imp + "\n" + jsOutput
				}
			}

			rel, _ := filepath.Rel(rotasDir, path)
			baseName := rel[:len(rel)-len(filepath.Ext(rel))]

			componentName := ""
			parts := strings.Split(baseName, string(filepath.Separator))
			for _, part := range parts {
				if len(part) > 0 {
					componentName += strings.ToUpper(part[:1]) + part[1:]
				}
			}

			routePath := "/" + filepath.ToSlash(baseName)
			if routePath == "/index" {
				routePath = "/"
			} else if strings.HasSuffix(routePath, "/index") {
				routePath = routePath[:len(routePath)-6]
			}

			cleanCompName := strings.NewReplacer("[", "", "]", "", "-", "_", " ", "_", ".", "_").Replace(componentName)

			// Se o código da rota já contiver declaração de função (ex: export function RotaX() ou function RotaX())
			routeJsClean := strings.ReplaceAll(routeJs, "export function ", "function ")
			routeJsClean = strings.ReplaceAll(routeJsClean, "export default function ", "function ")

			if strings.Contains(routeJsClean, "function ") {
				var funcDeclared string
				for _, l := range strings.Split(routeJsClean, "\n") {
					tr := strings.TrimSpace(l)
					if strings.HasPrefix(tr, "function ") {
						parts := strings.Fields(tr)
						if len(parts) >= 2 {
							fnName := parts[1]
							if idx := strings.Index(fnName, "("); idx != -1 {
								fnName = fnName[:idx]
							}
							if fnName != "" {
								funcDeclared = fnName
								break
							}
						}
					}
				}

				if funcDeclared != "" && jaDeclaradoOuImportado(jsOutput, funcDeclared) {
					// Função já importada ou declarada no bundle principal (ex: via main.hrp)
					rotaWrapper := "Rota_" + cleanCompName
					if !jaDeclaradoOuImportado(jsOutput, rotaWrapper) {
						jsOutput += fmt.Sprintf("\nfunction %s(props) {\n  return %s(props);\n}\n", rotaWrapper, funcDeclared)
					}
				} else {
					jsOutput += "\n" + routeJsClean + "\n"
					rotaWrapper := "Rota_" + cleanCompName
					if funcDeclared != "" {
						if funcDeclared != rotaWrapper && !jaDeclaradoOuImportado(jsOutput, rotaWrapper) {
							jsOutput += fmt.Sprintf("\nfunction %s(props) {\n  return %s(props);\n}\n", rotaWrapper, funcDeclared)
						}
					} else if !jaDeclaradoOuImportado(jsOutput, rotaWrapper) {
						jsOutput += fmt.Sprintf("\nfunction %s(props) {\n  return typeof Rota%s === 'function' ? Rota%s(props) : null;\n}\n", rotaWrapper, cleanCompName, cleanCompName)
					}
				}
			} else {
				rotaWrapper := "Rota_" + cleanCompName
				if !jaDeclaradoOuImportado(jsOutput, rotaWrapper) {
					jsOutput += fmt.Sprintf("\nfunction %s(props) {\n%s\n}\n", rotaWrapper, routeJsClean)
				}
			}

			// Converte /usuarios/[id] em /usuarios/:id
			paramRoutePath := routePath

			if strings.Contains(paramRoutePath, "[") && strings.Contains(paramRoutePath, "]") {
				for {
					start := strings.Index(paramRoutePath, "[")
					end := strings.Index(paramRoutePath, "]")
					if start == -1 || end == -1 || end < start {
						break
					}
					paramName := paramRoutePath[start+1 : end]
					paramRoutePath = paramRoutePath[:start] + ":" + paramName + paramRoutePath[end+1:]
				}
			}

			rotas = append(rotas, RotaInfo{Caminho: paramRoutePath, Nome: cleanCompName})
			return nil
		})
	}

	// ponytail: garante exportação automática de funções raiz principais para novos projetos
	for _, fnName := range []string{"Aplicacao", "MeuApp", "App", "Main"} {
		if strings.Contains(jsOutput, "function "+fnName) && !strings.Contains(jsOutput, "export function "+fnName) && !strings.Contains(jsOutput, "export {"+fnName) {
			jsOutput = strings.ReplaceAll(jsOutput, "function "+fnName, "export function "+fnName)
		}
	}

	if len(rotas) > 0 {
		if !jaDeclaradoOuImportado(jsOutput, "MeuApp") {
			var routeMappings []string
			for _, r := range rotas {
				routeMappings = append(routeMappings, fmt.Sprintf("'%s': Rota_%s", r.Caminho, r.Nome))
			}
			if strings.Contains(jsOutput, "function LayoutGlobal") {
				jsOutput += fmt.Sprintf("\nexport function MeuApp() {\n  return h(LayoutGlobal, {}, roteador({\n    %s\n  }));\n}\n", strings.Join(routeMappings, ",\n    "))
			} else {
				jsOutput += fmt.Sprintf("\nexport function MeuApp() {\n  return roteador({\n    %s\n  });\n}\n", strings.Join(routeMappings, ",\n    "))
			}
		}
	} else if !jaDeclaradoOuImportado(jsOutput, "MeuApp") {
		if strings.Contains(jsOutput, "export function RotaIndex") || strings.Contains(jsOutput, "function RotaIndex") {
			jsOutput += "\nexport function MeuApp() { return RotaIndex(); }\n"
		} else if strings.Contains(jsOutput, "export function App") || strings.Contains(jsOutput, "function App") {
			jsOutput += "\nexport function MeuApp() { return typeof App === 'function' ? App() : null; }\n"
		} else {
			jsOutput += "\nexport function MeuApp() { return typeof renderizarApp === 'function' ? renderizarApp() : null; }\n"
		}
	}

	if err = os.MkdirAll(opts.Saida, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída '%s': %w", opts.Saida, err)
	}

	absEntrada, _ := filepath.Abs(opts.Entrada)
	webDir := filepath.Join(projDir, "web")
	if fi, err := os.Stat(webDir); err == nil && fi.IsDir() {
		_ = filepath.Walk(webDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".hrp" {
				return nil
			}
			if path == absEntrada || path == filepath.Join(projDir, "servidor.hrp") {
				return nil
			}

			conteudoFile, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			fileAst, err := parser.NewParserFromString(string(conteudoFile), path).Parse()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Aviso: Erro ao obter AST para %s: %v\n", path, err)
				return nil
			}

			subTranspiler := &TranspilerWeb{
				Estrito:          opts.Estrito,
				DiretorioBase:    filepath.Dir(path),
				DiretorioProjeto: projDir,
			}
			subJs := subTranspiler.Transpile(fileAst)
			transpiler.Styles = append(transpiler.Styles, subTranspiler.Styles...)
			rel, _ := filepath.Rel(projDir, path)

			destJs := filepath.Join(opts.Saida, strings.TrimSuffix(rel, ".hrp")+".js")
			_ = os.MkdirAll(filepath.Dir(destJs), 0755)
			importHeaderSub := gerarImportHeaderWeb(rel)
			_ = os.WriteFile(destJs, []byte(importHeaderSub+subJs), 0644)

			// Copia também sem o prefixo web/ para resolver imports relativos como ./../componentes/navbar.js
			if strings.HasPrefix(rel, "web/") {
				relSemWeb := strings.TrimPrefix(rel, "web/")
				destJsSemWeb := filepath.Join(opts.Saida, strings.TrimSuffix(relSemWeb, ".hrp")+".js")
				_ = os.MkdirAll(filepath.Dir(destJsSemWeb), 0755)
				importHeaderSubSemWeb := gerarImportHeaderWeb(relSemWeb)
				_ = os.WriteFile(destJsSemWeb, []byte(importHeaderSubSemWeb+subJs), 0644)
			}
			return nil
		})
	}

	relEntrada, _ := filepath.Rel(projDir, absEntrada)
	destEntradaJs := filepath.Join(opts.Saida, strings.TrimSuffix(relEntrada, ".hrp")+".js")
	_ = os.MkdirAll(filepath.Dir(destEntradaJs), 0755)
	importHeaderEntrada := gerarImportHeaderWeb(relEntrada)
	_ = os.WriteFile(destEntradaJs, []byte(importHeaderEntrada+jsOutput), 0644)

	if strings.HasPrefix(relEntrada, "web/") {
		relSemWeb := strings.TrimPrefix(relEntrada, "web/")
		destEntradaJsSemWeb := filepath.Join(opts.Saida, strings.TrimSuffix(relSemWeb, ".hrp")+".js")
		_ = os.MkdirAll(filepath.Dir(destEntradaJsSemWeb), 0755)
		importHeaderEntradaSemWeb := gerarImportHeaderWeb(relSemWeb)
		_ = os.WriteFile(destEntradaJsSemWeb, []byte(importHeaderEntradaSemWeb+jsOutput), 0644)
	}

	appPath := filepath.Join(opts.Saida, "app.js")
	importHeader := gerarImportHeaderWeb("")
	if err = os.WriteFile(appPath, []byte(importHeader+jsOutput), 0644); err != nil {
		return fmt.Errorf("erro ao gravar app.js: %w", err)
	}

	estiloPath := filepath.Join(opts.Saida, "estilos.css")
	var cssContent string
	for _, styleBlock := range transpiler.Styles {
		cssContent += styleBlock + "\n\n"
	}
	utilCSS := extraiEGerarCssUtilitarios(jsOutput)
	if utilCSS != "" {
		cssContent += utilCSS + "\n"
	}
	if err = os.WriteFile(estiloPath, []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("erro ao gravar estilos.css: %w", err)
	}

	// Copia o runtime-web.js a partir da stdlib
	runtimeSrcPath := filepath.Join(cur, "stdlib", "web", "runtime-web.js")
	if _, errS := os.Stat(runtimeSrcPath); os.IsNotExist(errS) {
		// Tenta encontrar em projDir, em diretórios pai do monorepo ou próximo ao executável
		candidatos := []string{
			filepath.Join(projDir, "stdlib", "web", "runtime-web.js"),
			filepath.Join(filepath.Dir(projDir), "Harpia", "stdlib", "web", "runtime-web.js"),
			filepath.Join(filepath.Dir(filepath.Dir(projDir)), "Harpia", "stdlib", "web", "runtime-web.js"),
		}
		if exePath, errE := os.Executable(); errE == nil {
			candidatos = append(candidatos, filepath.Join(filepath.Dir(exePath), "stdlib", "web", "runtime-web.js"))
			candidatos = append(candidatos, filepath.Join(filepath.Dir(exePath), "..", "Harpia", "stdlib", "web", "runtime-web.js"))
		}
		for _, cand := range candidatos {
			if _, errC := os.Stat(cand); errC == nil {
				runtimeSrcPath = cand
				break
			}
		}
	}
	runtimeDestPath := filepath.Join(opts.Saida, "runtime-web.js")
	if err = copiarArquivo(runtimeSrcPath, runtimeDestPath); err != nil {
		fmt.Printf("Aviso: stdlib runtime física não encontrada em %s. Gerando cópia padrão...\n", runtimeSrcPath)
		_ = os.WriteFile(runtimeDestPath, []byte(FallbackRuntimeWebJS), 0644)
	}
	webSubDir := filepath.Join(opts.Saida, "web")
	if _, errStat := os.Stat(webSubDir); errStat == nil {
		_ = copiarArquivo(runtimeDestPath, filepath.Join(webSubDir, "runtime-web.js"))
	}

	indexPath := filepath.Join(opts.Saida, "index.html")
	customIndexPath := filepath.Join(entryDir, "index.html")
	if _, errC := os.Stat(customIndexPath); os.IsNotExist(errC) {
		if _, errP := os.Stat(filepath.Join(projDir, "index.html")); errP == nil {
			customIndexPath = filepath.Join(projDir, "index.html")
		}
	}
	var htmlContent string
	if conteudoCustom, errC := os.ReadFile(customIndexPath); errC == nil {
		htmlContent = string(conteudoCustom)
		if !strings.Contains(htmlContent, "estilos.css") && strings.Contains(htmlContent, "</head>") {
			htmlContent = strings.Replace(htmlContent, "</head>", "    <link rel=\"stylesheet\" href=\"/estilos.css\">\n</head>", 1)
		}
		if !strings.Contains(htmlContent, "runtime-web.js") {
			scriptBlock := `
    <script>
      window.tamanho = function(valor) {
        if (valor === null || valor === undefined) return 0;
        if (typeof valor.length === 'number') return valor.length;
        if (valor instanceof Map || valor instanceof Set) return valor.size;
        if (typeof valor === 'object') return Object.keys(valor).length;
        return 0;
      };
      window.inteiro = function(v) {
        if (v === null || v === undefined) return 0;
        let n = parseInt(v, 10);
        return isNaN(n) ? 0 : n;
      };
      window.decimal = function(v) {
        if (v === null || v === undefined) return 0.0;
        let n = parseFloat(v);
        return isNaN(n) ? 0.0 : n;
      };
      window.texto = function(v) {
        if (v === null || v === undefined) return "";
        return String(v);
      };
    </script>
    <script type="module">
      import { h, montar } from '/runtime-web.js';
      import * as AppExports from '/app.js';

      const RootComponent = AppExports.MeuApp || AppExports.Aplicacao || AppExports.App || AppExports.Main || AppExports.default || Object.values(AppExports).find(v => typeof v === 'function');

      if (typeof RootComponent === 'function') {
        montar(RootComponent, document.getElementById('app') || 'app');
      } else {
        console.warn('[Harpia Web] Nenhum componente raiz exportado (MeuApp, Aplicacao, App) foi encontrado em app.js.');
      }
    </script>`
			if strings.Contains(htmlContent, "</body>") {
				htmlContent = strings.Replace(htmlContent, "</body>", scriptBlock+"\n</body>", 1)
			} else if strings.Contains(htmlContent, "</html>") {
				htmlContent = strings.Replace(htmlContent, "</html>", scriptBlock+"\n</html>", 1)
			} else {
				htmlContent += scriptBlock
			}
		}
	} else {
		htmlContent = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Harpia App</title>
    <link rel="stylesheet" href="/estilos.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { margin: 0; padding: 0; background-color: #0A0F0D; color: #F3F6F4; font-family: Outfit, sans-serif; }
    </style>
    <!--
      Dica de DX: Caso use bare modules como '@backend/modulo', configure um Import Map
      para mapear os identificadores para os proxies Javascript gerados, por exemplo:
      <script type="importmap">
        {
          "imports": {
            "@backend/filmes": "/web/backend/filmes.js"
          }
        }
      </script>
    -->
</head>
<body class="CorpoGlobal">
    <div id="app"></div>
    <script>
      window.tamanho = function(valor) {
        if (valor === null || valor === undefined) return 0;
        if (typeof valor.length === 'number') return valor.length;
        if (valor instanceof Map || valor instanceof Set) return valor.size;
        if (typeof valor === 'object') return Object.keys(valor).length;
        return 0;
      };
      window.inteiro = function(v) {
        if (v === null || v === undefined) return 0;
        let n = parseInt(v, 10);
        return isNaN(n) ? 0 : n;
      };
      window.decimal = function(v) {
        if (v === null || v === undefined) return 0.0;
        let n = parseFloat(v);
        return isNaN(n) ? 0.0 : n;
      };
      window.texto = function(v) {
        if (v === null || v === undefined) return "";
        return String(v);
      };
    </script>
    <script type="module">
      import { h, montar } from '/runtime-web.js';
      import * as AppExports from '/app.js';

      const RootComponent = AppExports.MeuApp || AppExports.Aplicacao || AppExports.App || AppExports.Main || AppExports.default || Object.values(AppExports).find(v => typeof v === 'function');

      if (typeof RootComponent === 'function') {
        montar(RootComponent, document.getElementById('app') || 'app');
      } else {
        console.warn('[Harpia Web] Nenhum componente raiz exportado (MeuApp, Aplicacao, App) foi encontrado em app.js.');
      }
    </script>
</body>
</html>`
	}

	if err = os.WriteFile(indexPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("erro ao gravar index.html: %w", err)
	}

	if err = otimizarECopiarAssets(entryDir, opts.Saida, opts.OtimizarAssets); err != nil {
		fmt.Fprintf(os.Stderr, "aviso ao processar assets de imagens: %v\n", err)
	}

	// Gera configurações de SPA Fallback para servidores de estáticos (npx serve, Vercel, Netlify)
	_ = os.WriteFile(filepath.Join(opts.Saida, "serve.json"), []byte(`{
  "rewrites": [
    { "source": "**", "destination": "/index.html" }
  ]
}`), 0644)

	_ = os.WriteFile(filepath.Join(opts.Saida, "vercel.json"), []byte(`{
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}`), 0644)

	_ = os.WriteFile(filepath.Join(opts.Saida, "_redirects"), []byte(`/*    /index.html   200`), 0644)

	return nil
}

func comandoCompilar() *cobra.Command {
	var alvo string
	var entrada string
	var saida string
	var estrito bool
	var otimizarAssets bool
	var pularLinter bool
	var osTarget string
	var archTarget string
	var ajuda bool

	compilar := &cobra.Command{
		Use:           "compilar",
		Aliases:       []string{"construir"},
		Short:         "Compila/Transpila o código Harpia para a plataforma Web",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ajuda {
				return cmd.Help()
			}

			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			cur, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("erro ao obter diretório atual: %w", err)
			}

			if len(args) > 0 {
				entrada = args[0]
			}

			// Autodetecta o arquivo de entrada se nenhum for especificado
			if entrada == "" {
				candidatos := []string{
					"main.hrp", "index.hrp", "app.hrp",
					filepath.Join("web", "main.hrp"),
					filepath.Join("web", "index.hrp"),
				}
				for _, cand := range candidatos {
					if _, err := os.Stat(cand); err == nil {
						entrada = cand
						break
					}
				}
			}

			if entrada == "" {
				return fmt.Errorf("erro: nenhum arquivo de entrada especificado e nenhum arquivo padrão (main.hrp, index.hrp) encontrado.")
			}

			if saida == "" {
				saida = "dist"
			}

			if alvo == "" {
				alvo = "web"
			}

			if alvo != "web" && alvo != "nativo" && alvo != "wasm" && alvo != "wasi" {
				return fmt.Errorf("erro: alvo de compilação '%s' não suportado. Alvos suportados: web, nativo, wasm, wasi", alvo)
			}

			// Helper de spinner animado
			executarComSpinner := func(mensagem string, acao func() error) error {
				frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
				parar := make(chan bool)
				var errExec error

				go func() {
					i := 0
					for {
						select {
						case <-parar:
							fmt.Fprint(out, "\r\033[K")
							return
						default:
							fmt.Fprintf(out, "\r%s %s...", frames[i%len(frames)], mensagem)
							i++
							time.Sleep(80 * time.Millisecond)
						}
					}
				}()

				errExec = acao()
				close(parar)
				return errExec
			}

			// Alvos nativos/wasm usam o caminho AOT separado
			if alvo == "nativo" || alvo == "wasm" || alvo == "wasi" {
				ctx := hrp.NewContexto(hrp.OpcsContexto{CaminhosPadrao: []string{cur}})
				defer ctx.Terminar()

				_, ast, err := ctx.TransformarEmAst(entrada, false, cur)
				if err != nil {
					return fmt.Errorf("Erro de compilação/sintaxe: %w", err)
				}
				if prog, ok := ast.(*parser.Programa); ok {
					ast = Otimizar(prog)
				}
				transpiler := &TranspilerNative{}
				goCode := transpiler.GenerateFullCode(ast)

				tmpGoFile := filepath.Join(cur, "main_aot.go")
				if err = os.WriteFile(tmpGoFile, []byte(goCode), 0644); err != nil {
					return fmt.Errorf("erro ao gravar arquivo Go temporário: %v", err)
				}
				defer os.Remove(tmpGoFile)

				saidaBin := saida
				if saidaBin == "" || saidaBin == "dist" {
					saidaBin = "app"
				}
				cmdBuild := exec.Command("go", "build", "-o", saidaBin, tmpGoFile)
				cmdBuild.Stdout = out
				cmdBuild.Stderr = errOut
				switch alvo {
				case "wasm":
					cmdBuild.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
				case "wasi":
					cmdBuild.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
				}

				if osTarget != "" {
					if cmdBuild.Env == nil {
						cmdBuild.Env = os.Environ()
					}
					cmdBuild.Env = append(cmdBuild.Env, "GOOS="+osTarget)
				}
				if archTarget != "" {
					if cmdBuild.Env == nil {
						cmdBuild.Env = os.Environ()
					}
					cmdBuild.Env = append(cmdBuild.Env, "GOARCH="+archTarget)
				}

				if errBuild := cmdBuild.Run(); errBuild != nil {
					return fmt.Errorf("erro ao executar go build: %w", errBuild)
				}
				fmt.Fprintf(out, "🚀 Compilação AOT (%s) concluída com sucesso!\n", alvo)
				return nil
			}

			// Alvo web: usa a função pura compilarParaWeb com spinner
			tInicio := time.Now()
			errComp := executarComSpinner("Compilando projeto Harpia para Web ("+entrada+")", func() error {
				return compilarParaWeb(OpcsCompilarWeb{
					Entrada:        entrada,
					Saida:          saida,
					Estrito:        estrito,
					OtimizarAssets: otimizarAssets,
					PularLinter:    pularLinter,
				})
			})

			if errComp != nil {
				return fmt.Errorf("❌ Erro na compilação: %w", errComp)
			}

			fmt.Fprintf(out, "🚀 Compilação Web concluída com sucesso em %v!\n", time.Since(tInicio).Round(time.Millisecond))
			fmt.Fprintf(out, "📦 Artefatos gerados em '%s/':\n", saida)
			fmt.Fprintln(out, "  - index.html     (Ponto de entrada HTML)")
			fmt.Fprintln(out, "  - runtime-web.js (Engine de Reatividade & Virtual DOM)")
			fmt.Fprintln(out, "  - app.js         (Código de aplicação transpilado)")
			fmt.Fprintln(out, "  - estilos.css    (Estilos unificados e utilitários)")
			fmt.Fprintln(out)
			return nil
		},
	}

	compilar.Flags().StringVarP(&alvo, "alvo", "a", "web", "Alvo da compilação (web, nativo)")
	compilar.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .hrp principal de entrada")
	compilar.Flags().StringVarP(&saida, "saida", "s", "dist", "Diretório de destino da compilação")
	compilar.Flags().BoolVar(&estrito, "estrito", false, "Ativa anotações JSDoc para tipagem estática")
	compilar.Flags().BoolVar(&otimizarAssets, "otimizar-assets", false, "Otimiza e comprime imagens de assets (PNG, JPG, JPEG) para a saída")
	compilar.Flags().BoolVar(&pularLinter, "pular-linter", false, "Pula a checagem do linter antes de compilar")
	compilar.Flags().StringVar(&osTarget, "os", "", "Sistema operacional alvo para compilação nativa (darwin, linux, windows)")
	compilar.Flags().StringVar(&archTarget, "arch", "", "Arquitetura de CPU alvo para compilação nativa (amd64, arm64, 386)")
	compilar.Flags().BoolVar(&ajuda, "ajuda", false, "Exibe a ajuda do comando")
	return compilar
}

func copiarArquivo(origem, destino string) error {
	in, err := os.Open(origem)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

// ponytail: helper mínimo para fazer apenas a PRIMEIRA substituição (em vez de ReplaceAll) sem
// precisar import a stdlib regexp. Substituir por strings.Replace com n=2 não tem a mesma semântica
// pois trocaria todos os matches a partir do segundo. A função abaixo preserva o primeiro match
// e mantém o resto intacto.
func stringsReplaceOnce(s, old, newStr string) string {
	i := strings.Index(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + newStr + s[i+len(old):]
}

// ponytail: otimizador e copiador estático de assets de imagens nativo
func otimizarECopiarAssets(srcDir, destDir string, otimizar bool) error {
	if _, err := os.Stat(srcDir); err != nil {
		return nil
	}

	extensions := map[string]bool{
		".png":         true,
		".jpg":         true,
		".jpeg":        true,
		".gif":         true,
		".svg":         true,
		".webp":        true,
		".ico":         true,
		".webmanifest": true,
		".manifest":    true,
		".json":        true,
		".woff":        true,
		".woff2":       true,
		".ttf":         true,
		".otf":         true,
		".eot":         true,
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == "dist" || name == "pt_modulos" || name == "node_modules" || name == "infra" || name == "dominio" || name == "testes" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !extensions[ext] {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		destPath := filepath.Join(destDir, rel)
		if err = os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if otimizar && (ext == ".jpg" || ext == ".jpeg" || ext == ".png") {
			_ = otimizarImagemFisica(path, destPath, ext)
			gerarWebPSeDisponivel(path, destPath)
		} else {
			_ = copiarArquivo(path, destPath)
		}

		if strings.HasPrefix(rel, "web/") {
			relSemWeb := strings.TrimPrefix(rel, "web/")
			destSemWeb := filepath.Join(destDir, relSemWeb)
			_ = os.MkdirAll(filepath.Dir(destSemWeb), 0755)
			_ = copiarArquivo(path, destSemWeb)
		}

		baseLower := strings.ToLower(filepath.Base(path))
		if baseLower == "favicon.ico" || baseLower == "favicon.svg" || baseLower == "apple-touch-icon.png" {
			destRaiz := filepath.Join(destDir, baseLower)
			_ = copiarArquivo(path, destRaiz)

			destImg := filepath.Join(destDir, "assets", "img", baseLower)
			_ = os.MkdirAll(filepath.Dir(destImg), 0755)
			_ = copiarArquivo(path, destImg)

			destFavicon := filepath.Join(destDir, "assets", "favicon", baseLower)
			_ = os.MkdirAll(filepath.Dir(destFavicon), 0755)
			_ = copiarArquivo(path, destFavicon)
		}

		return nil
	})
}

// ponytail: se 'cwebp' (Google) estiver disponível, gera arquivo .webp ao lado do destino.
// Falha silenciosa — WebP é otimização opcional e não bloqueia o build.
func gerarWebPSeDisponivel(origem, destino string) {
	cwebp, err := exec.LookPath("cwebp")
	if err != nil {
		return
	}
	destinoWebp := destino + ".webp"
	cmd := exec.Command(cwebp, "-q", "75", origem, "-o", destinoWebp)
	if err := cmd.Run(); err != nil {
		os.Remove(destinoWebp)
	}
}

// ponytail: codificador/decodificador de imagens nativo com compressão síncrona
func otimizarImagemFisica(origem, destino, ext string) error {
	file, err := os.Open(origem)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	out, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer out.Close()

	if ext == ".png" {
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		return encoder.Encode(out, img)
	}

	return jpeg.Encode(out, img, &jpeg.Options{Quality: 75})
}

// ponytail: corrige caminhos relativos de importações extraídas de rotas/layouts para que fiquem corretos e únicos no bundle final app.js
func corrigirCaminhoImport(line string, sourceFile string, projDir string) string {
	var quoteChar string
	startIdx := strings.Index(line, "from \"")
	if startIdx != -1 {
		startIdx += 6
		quoteChar = "\""
	} else {
		startIdx = strings.Index(line, "from '")
		if startIdx != -1 {
			startIdx += 6
			quoteChar = "'"
		}
	}
	if startIdx == -1 {
		return line
	}
	endIdx := strings.Index(line[startIdx:], quoteChar)
	if endIdx == -1 {
		return line
	}
	endIdx += startIdx

	importPath := line[startIdx:endIdx]
	if !strings.HasPrefix(importPath, ".") || importPath == "./runtime-web.js" {
		return line
	}

	dir := filepath.Dir(sourceFile)
	absTarget := filepath.Clean(filepath.Join(dir, importPath))

	rel, err := filepath.Rel(projDir, absTarget)
	if err != nil {
		return line
	}

	relClean := strings.TrimPrefix(rel, "web/")

	relClean = filepath.ToSlash(relClean)

	if !strings.HasPrefix(relClean, ".") {
		relClean = "./" + relClean
	}

	return line[:startIdx] + relClean + line[endIdx:]
}

func jaDeclaradoOuImportado(jsCode, nome string) bool {
	if nome == "" {
		return false
	}
	padroes := []string{
		"function " + nome + "(",
		"function " + nome + " (",
		"export function " + nome + "(",
		"export function " + nome + " (",
		"const " + nome + " =",
		"let " + nome + " =",
		"var " + nome + " =",
		"export let " + nome + " =",
		"export const " + nome + " =",
	}
	for _, p := range padroes {
		if strings.Contains(jsCode, p) {
			return true
		}
	}
	for _, line := range strings.Split(jsCode, "\n") {
		tr := strings.TrimSpace(line)
		if strings.HasPrefix(tr, "import ") && strings.Contains(tr, "{") && strings.Contains(tr, "}") {
			if idxStart := strings.Index(tr, "{"); idxStart != -1 {
				if idxEnd := strings.Index(tr, "}"); idxEnd > idxStart {
					for _, item := range strings.Split(tr[idxStart+1:idxEnd], ",") {
						if strings.TrimSpace(item) == nome {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func importJaExiste(jsOutput string, imp string) bool {
	impTrim := strings.TrimSpace(imp)
	if strings.HasPrefix(impTrim, "import ") {
		if idxStart := strings.Index(impTrim, "{"); idxStart != -1 {
			if idxEnd := strings.Index(impTrim, "}"); idxEnd > idxStart {
				simbolos := strings.Split(impTrim[idxStart+1:idxEnd], ",")
				todosJaExistem := true
				for _, s := range simbolos {
					sym := strings.TrimSpace(s)
					if sym != "" && !jaDeclaradoOuImportado(jsOutput, sym) {
						todosJaExistem = false
						break
					}
				}
				if todosJaExistem {
					return true
				}
			}
		}
	} else if strings.HasPrefix(impTrim, "const ") {
		parts := strings.Fields(impTrim)
		if len(parts) >= 2 {
			constName := parts[1]
			if jaDeclaradoOuImportado(jsOutput, constName) {
				return true
			}
		}
	}
	return strings.Contains(jsOutput, imp)
}
