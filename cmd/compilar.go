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
	if !strings.Contains(jsOutput, "export function MeuApp") {
		if strings.Contains(jsOutput, "function MeuApp(") {
			jsOutput = stringsReplaceOnce(jsOutput, "function MeuApp(", "export function MeuApp(")
		} else if strings.Contains(jsOutput, "export function RotaIndex") || strings.Contains(jsOutput, "function RotaIndex") {
			jsOutput += "\nexport function MeuApp() { return RotaIndex(); }\n"
		} else {
			jsOutput += "\nexport function MeuApp() { return typeof renderizarApp === 'function' ? renderizarApp() : null; }\n"
		}
	}

	entryDir, _ := filepath.Abs(filepath.Dir(opts.Entrada))
	rotasDir := ""
	rotasPaths := []string{
		filepath.Join(entryDir, "rotas"),
		filepath.Join(entryDir, "web", "rotas"),
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

			if filepath.Base(path) == "rotas.hrp" {
				return nil
			}

			if filepath.Base(path) == "layout.hrp" {
				conteudoLayout, err := os.ReadFile(path)
				if err == nil {
					layoutAst, err := parser.NewParserFromString(string(conteudoLayout), path).Parse()
					if err == nil {
						layoutJs := transpiler.Transpile(layoutAst)
						jsOutput += fmt.Sprintf("\nfunction LayoutGlobal(props) {\n%s\n}\n", layoutJs)
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

			cleanCompName := strings.NewReplacer("[", "", "]", "").Replace(componentName)
			jsOutput += fmt.Sprintf("\nfunction Rota_%s(props) {\n%s\n}\n", cleanCompName, routeJs)

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


	if len(rotas) > 0 {
		var routeMappings []string
		for _, r := range rotas {
			routeMappings = append(routeMappings, fmt.Sprintf("'%s': Rota_%s", r.Caminho, r.Nome))
		}
		jsOutput += fmt.Sprintf("\nexport function MeuApp() {\n  return roteador({\n    %s\n  });\n}\n", strings.Join(routeMappings, ",\n    "))
	}

	if err = os.MkdirAll(opts.Saida, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída '%s': %w", opts.Saida, err)
	}

	appPath := filepath.Join(opts.Saida, "app.js")
	if err = os.WriteFile(appPath, []byte("import { h, sinal, efeito, derivado, armazem, montar, navegar, roteador } from './runtime-web.js';\n\n"+jsOutput), 0644); err != nil {
		return fmt.Errorf("erro ao gravar app.js: %w", err)
	}

	absEntrada, _ := filepath.Abs(opts.Entrada)
	webDir := filepath.Join(entryDir, "web")
	if fi, err := os.Stat(webDir); err == nil && fi.IsDir() {
		_ = filepath.Walk(webDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".hrp" {
				return nil
			}
			if path == absEntrada || filepath.Base(path) == "servidor.hrp" {
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
				DiretorioProjeto: entryDir,
			}
			subJs := subTranspiler.Transpile(fileAst)
			transpiler.Styles = append(transpiler.Styles, subTranspiler.Styles...)
			jsOutput += "\n" + subJs
			rel, _ := filepath.Rel(entryDir, path)


			destJs := filepath.Join(opts.Saida, strings.TrimSuffix(rel, ".hrp")+".js")
			_ = os.MkdirAll(filepath.Dir(destJs), 0755)
			return os.WriteFile(destJs, []byte(subJs), 0644)
		})
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
	runtimeDestPath := filepath.Join(opts.Saida, "runtime-web.js")
	if err = copiarArquivo(runtimeSrcPath, runtimeDestPath); err != nil {
		fmt.Printf("Aviso: stdlib runtime física não encontrada em %s. Gerando cópia padrão...\n", runtimeSrcPath)
		_ = os.WriteFile(runtimeDestPath, []byte(FallbackRuntimeWebJS), 0644)
	}

	indexPath := filepath.Join(opts.Saida, "index.html")
	htmlTemplate := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Harpia App</title>
    <link rel="stylesheet" href="estilos.css">
</head>
<body>
    <div id="app"></div>
    <script type="module" src="./app.js"></script>
</body>
</html>`

	if err = os.WriteFile(indexPath, []byte(htmlTemplate), 0644); err != nil {
		return fmt.Errorf("erro ao gravar index.html: %w", err)
	}

	if err = otimizarECopiarAssets(entryDir, opts.Saida, opts.OtimizarAssets); err != nil {
		fmt.Fprintf(os.Stderr, "aviso ao processar assets de imagens: %v\n", err)
	}

	return nil
}

func comandoCompilar() *cobra.Command {
	var alvo string
	var entrada string
	var saida string
	var estrito bool
	var otimizarAssets bool
	var pularLinter bool

	compilar := &cobra.Command{
		Use:   "compilar",
		Short: "Compila/Transpila o código Harpia para a plataforma Web",
		Run: func(cmd *cobra.Command, args []string) {
			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao obter diretório atual:", err)
				os.Exit(1)
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
				fmt.Fprintln(os.Stderr, "erro: nenhum arquivo de entrada especificado e nenhum arquivo padrão (main.hrp, index.hrp) encontrado.")
				os.Exit(1)
			}

			if saida == "" {
				saida = "dist"
			}

			if alvo == "" {
				alvo = "web"
			}

			if alvo != "web" && alvo != "nativo" && alvo != "wasm" && alvo != "wasi" {
				fmt.Fprintf(os.Stderr, "erro: alvo de compilação '%s' não suportado. Alvos suportados: web, nativo, wasm, wasi\n", alvo)
				os.Exit(1)
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
							fmt.Print("\r\033[K")
							return
						default:
							fmt.Printf("\r%s %s...", frames[i%len(frames)], mensagem)
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
					fmt.Fprintf(os.Stderr, "Erro de compilação/sintaxe: %v\n", err)
					os.Exit(1)
				}
				if prog, ok := ast.(*parser.Programa); ok {
					ast = Otimizar(prog)
				}
				transpiler := &TranspilerNative{}
				goCode := transpiler.GenerateFullCode(ast)

				tmpGoFile := filepath.Join(cur, "main_aot.go")
				if err = os.WriteFile(tmpGoFile, []byte(goCode), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "erro ao gravar arquivo Go temporário: %v\n", err)
					os.Exit(1)
				}
				defer os.Remove(tmpGoFile)

				saidaBin := saida
				if saidaBin == "" || saidaBin == "dist" {
					saidaBin = "app"
				}
				cmdBuild := exec.Command("go", "build", "-o", saidaBin, tmpGoFile)
				cmdBuild.Stdout = os.Stdout
				cmdBuild.Stderr = os.Stderr
				if alvo == "wasm" {
					cmdBuild.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
				} else if alvo == "wasi" {
					cmdBuild.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
				}
				if errBuild := cmdBuild.Run(); errBuild != nil {
					fmt.Fprintf(os.Stderr, "erro ao executar go build: %v\n", errBuild)
					os.Exit(1)
				}
				fmt.Printf("🚀 Compilação AOT (%s) concluída com sucesso!\n", alvo)
				return
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
				fmt.Fprintf(os.Stderr, "❌ Erro na compilação: %v\n", errComp)
				os.Exit(1)
			}

			fmt.Printf("🚀 Compilação Web concluída com sucesso em %v!\n", time.Since(tInicio).Round(time.Millisecond))
			fmt.Printf("📦 Artefatos gerados em '%s/':\n", saida)
			fmt.Println("  - index.html     (Ponto de entrada HTML)")
			fmt.Println("  - runtime-web.js (Engine de Reatividade & Virtual DOM)")
			fmt.Println("  - app.js         (Código de aplicação transpilado)")
			fmt.Println("  - estilos.css    (Estilos unificados e utilitários)")
			fmt.Println()
		},
	}

	compilar.Flags().StringVarP(&alvo, "alvo", "a", "web", "Alvo da compilação (web, nativo)")
	compilar.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .hrp principal de entrada")
	compilar.Flags().StringVarP(&saida, "saida", "s", "dist", "Diretório de destino da compilação")
	compilar.Flags().BoolVar(&estrito, "estrito", false, "Ativa anotações JSDoc para tipagem estática")
	compilar.Flags().BoolVar(&otimizarAssets, "otimizar-assets", false, "Otimiza e comprime imagens de assets (PNG, JPG, JPEG) para a saída")
	compilar.Flags().BoolVar(&pularLinter, "pular-linter", false, "Pula a checagem do linter antes de compilar")
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
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".svg":  true,
		".webp": true,
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
		err = os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return err
		}

		if otimizar && (ext == ".jpg" || ext == ".jpeg" || ext == ".png") {
			err = otimizarImagemFisica(path, destPath, ext)
			if err == nil {
				gerarWebPSeDisponivel(path, destPath)
				return nil
			}
		}

		return copiarArquivo(path, destPath)
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
