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

	"github.com/mat-dgruber/Harpia/ptst"
	"github.com/spf13/cobra"
)

func comandoCompilar() *cobra.Command {
	var alvo string
	var entrada string
	var saida string
	var estrito bool
	var otimizarAssets bool

	compilar := &cobra.Command{
		Use:   "compilar",
		Short: "Compila/Transpila o código Harpia para a plataforma Web",
		Run: func(cmd *cobra.Command, args []string) {
			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao obter diretório atual:", err)
				os.Exit(1)
			}

			// Se tiver argumento posicional, usa como arquivo de entrada
			if len(args) > 0 {
				entrada = args[0]
			}

			if entrada == "" {
				fmt.Fprintln(os.Stderr, "erro: arquivo de entrada não especificado. Use o argumento posicional ou --entrada.")
				os.Exit(1)
			}

			if alvo != "web" && alvo != "nativo" {
				fmt.Fprintf(os.Stderr, "erro: alvo de compilação '%s' não suportado. Alvos suportados: web, nativo\n", alvo)
				os.Exit(1)
			}

			if alvo == "nativo" {
				ctx := ptst.NewContexto(ptst.OpcsContexto{CaminhosPadrao: []string{cur}})
				defer ctx.Terminar()

				_, ast, err := ctx.TransformarEmAst(entrada, false, cur)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro de compilação/sintaxe: %v\n", err)
					os.Exit(1)
				}

				transpiler := &TranspilerNative{}
				goCode := transpiler.GenerateFullCode(ast)

				tmpGoFile := filepath.Join(cur, "main_aot.go")
				err = os.WriteFile(tmpGoFile, []byte(goCode), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "erro ao gravar arquivo Go temporário: %v\n", err)
					os.Exit(1)
				}
				defer os.Remove(tmpGoFile)

				saidaBin := saida
				if saidaBin == "dist" {
					baseName := filepath.Base(entrada)
					saidaBin = strings.TrimSuffix(baseName, filepath.Ext(baseName))
				}

				fmt.Printf("Compilando binário nativo '%s'...\n", saidaBin)

				cmdBuild := exec.Command("go", "build", "-o", saidaBin, tmpGoFile)
				cmdBuild.Stdout = os.Stdout
				cmdBuild.Stderr = os.Stderr
				if errBuild := cmdBuild.Run(); errBuild != nil {
					fmt.Fprintf(os.Stderr, "erro ao executar go build: %v\n", errBuild)
					os.Exit(1)
				}

				fmt.Println("🚀 Compilação AOT concluída com sucesso!")
				return
			}

			ctx := ptst.NewContexto(ptst.OpcsContexto{CaminhosPadrao: []string{cur}})
			defer ctx.Terminar()

			// Transforma o código do arquivo de entrada em AST
			_, ast, err := ctx.TransformarEmAst(entrada, false, cur)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro de compilação/sintaxe: %v\n", err)
				os.Exit(1)
			}

			// Executa o transpiler
			transpiler := &TranspilerWeb{
				Estrito:       estrito,
				DiretorioBase: filepath.Dir(entrada),
			}
			jsOutput := transpiler.Transpile(ast)

			// ponytail: garante visibilidade do componente raiz no bundle final.
			// Caso o usuário tenha declarado `funcao MeuApp(...) { ... }` no main.ptst sem
			// marcar como exportado, prefixamos a declaração com `export` para que o
			// `<script>` de bootstrap no index.html consiga importá-lo.
			// Deve rodar ANTES do bloco de rotas para que o `if len(rotas) > 0` abaixo
			// possa sobrescrever com sua própria definição de export-rotas.
			if !strings.Contains(jsOutput, "export function MeuApp") && strings.Contains(jsOutput, "function MeuApp(") {
				jsOutput = stringsReplaceOnce(jsOutput, "function MeuApp(", "export function MeuApp(")
			}

			// Detecção de diretório de rotas para roteamento SPA baseado em arquivos
			entryDir := filepath.Dir(entrada)
			rotasDir := ""
			rotasPaths := []string{
				filepath.Join(entryDir, "rotas"),
				filepath.Join(entryDir, "web", "rotas"),
				filepath.Join(cur, "rotas"),
				filepath.Join(cur, "web", "rotas"),
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
					if err != nil || info.IsDir() || (ext != ".hrp" && ext != ".ptst") {
						return nil
					}
					// Carrega e transpila o arquivo de rota
					_, routeAst, err := ctx.TransformarEmAst(path, false, cur)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Erro de sintaxe na rota %s: %v\n", path, err)
						return nil
					}
					routeJs := transpiler.Transpile(routeAst)

					rel, _ := filepath.Rel(rotasDir, path)
					baseName := rel[:len(rel)-len(filepath.Ext(rel))]

					// Gera nome de componente sanitizado
					componentName := ""
					parts := strings.Split(baseName, string(filepath.Separator))
					for _, part := range parts {
						if len(part) > 0 {
							componentName += strings.ToUpper(part[:1]) + part[1:]
						}
					}

					// Determina o caminho da rota do navegador
					routePath := "/" + filepath.ToSlash(baseName)
					if routePath == "/index" {
						routePath = "/"
					} else if strings.HasSuffix(routePath, "/index") {
						routePath = routePath[:len(routePath)-6]
					}

					// Embrulha o código transpiliado em uma função reativa
					jsOutput += fmt.Sprintf("\nfunction Rota_%s() {\n%s\n}\n", componentName, routeJs)
					rotas = append(rotas, RotaInfo{Caminho: routePath, Nome: componentName})
					return nil
				})
			}

			// Se houver rotas configuradas, gera o mapeamento e componente MeuApp automaticamente
			if len(rotas) > 0 {
				var routeMappings []string
				for _, r := range rotas {
					routeMappings = append(routeMappings, fmt.Sprintf("'%s': Rota_%s", r.Caminho, r.Nome))
				}
				jsOutput += fmt.Sprintf("\nexport function MeuApp() {\n  return roteador({\n    %s\n  });\n}\n", strings.Join(routeMappings, ",\n    "))
			}

			// Garante a existência da pasta de saída
			err = os.MkdirAll(saida, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao criar diretório de saída '%s': %v\n", saida, err)
				os.Exit(1)
			}

			// 1. Escreve o app.js
			appPath := filepath.Join(saida, "app.js")
			err = os.WriteFile(appPath, []byte("import { h, sinal, efeito, derivado, armazem, montar, navegar, roteador } from './runtime-web.js';\n\n"+jsOutput), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao gravar app.js: %v\n", err)
				os.Exit(1)
			}

			// 2. Escreve os estilos.css
			estiloPath := filepath.Join(saida, "estilos.css")
			var cssContent string
			for _, styleBlock := range transpiler.Styles {
				cssContent += styleBlock + "\n\n"
			}

			// ponytail: integra classes utilitárias PT sob demanda baseadas no que foi detectado no JS
			utilCSS := extraiEGerarCssUtilitarios(jsOutput)
			if utilCSS != "" {
				cssContent += utilCSS + "\n"
			}

			err = os.WriteFile(estiloPath, []byte(cssContent), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao gravar estilos.css: %v\n", err)
				os.Exit(1)
			}

			// 3. Copia o runtime-web.js a partir da stdlib
			// ponytail: lê o arquivo físico da stdlib para não depender de pacotes externos ou embed
			runtimeSrcPath := filepath.Join(cur, "stdlib", "web", "runtime-web.js")
			runtimeDestPath := filepath.Join(saida, "runtime-web.js")

			err = copiarArquivo(runtimeSrcPath, runtimeDestPath)
			if err != nil {
				// Fallback caso executado fora da raiz do repo
				fmt.Printf("Aviso: stdlib runtime física não encontrada em %s. Gerando cópia padrão...\n", runtimeSrcPath)
				fallbackRuntime := `// Runtime fallback enxuto
export function h(t, p, ...c) { return { tag: t, props: p || {}, children: c.flat(Infinity) }; }
export function sinal(v) { let s = new Set(); return [() => v, (n) => { v = n; s.forEach(fn => fn()); }]; }
export function montar(app, el) { el.innerHTML = app().tag; }
export function navegar(d) { console.log('navegando para', d); }
export function roteador(r) { return () => h('div', {}, 'Roteador fallback'); }`
				os.WriteFile(runtimeDestPath, []byte(fallbackRuntime), 0644)
			}

			// 4. Cria um index.html básico para servir o app
			indexPath := filepath.Join(saida, "index.html")
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
    <script type="module">
        import { montar } from './runtime-web.js';
        import { MeuApp } from './app.js';

        // Inicializa a montagem no container #app
        if (typeof MeuApp === 'function') {
            montar(MeuApp, document.getElementById('app'));
        } else {
            console.warn("Componente 'MeuApp' não encontrado ou não exportado no arquivo principal.");
        }
    </script>
</body>
</html>`
			err = os.WriteFile(indexPath, []byte(htmlTemplate), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao gravar index.html: %v\n", err)
				os.Exit(1)
			}

			// ponytail: otimização e cópia estática síncrona de assets de imagens
			err = otimizarECopiarAssets(entryDir, saida, otimizarAssets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "aviso ao processar assets de imagens: %v\n", err)
			}

			fmt.Println("🚀 Compilação concluída com sucesso!")
			fmt.Printf("Arquivos gerados em '%s/':\n", saida)
			fmt.Println("  - index.html   (ponto de entrada web)")
			fmt.Println("  - runtime-web.js (motor Virtual DOM & reatividade)")
			fmt.Println("  - app.js       (lógica de negócios transpilada)")
			fmt.Println("  - estilos.css  (folhas de estilo unificadas)")
		},
	}

	compilar.Flags().StringVarP(&alvo, "alvo", "a", "web", "Alvo da compilação (web, nativo)")
	compilar.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .hrp principal de entrada")
	compilar.Flags().StringVarP(&saida, "saida", "s", "dist", "Diretório de destino da compilação")
	compilar.Flags().BoolVar(&estrito, "estrito", false, "Ativa anotações JSDoc para tipagem estática")
	compilar.Flags().BoolVar(&otimizarAssets, "otimizar-assets", false, "Otimiza e comprime imagens de assets (PNG, JPG, JPEG) para a saída")
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
			if name == "dist" || name == "pt_modulos" || name == "node_modules" || strings.HasPrefix(name, ".") {
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
				return nil
			}
		}

		return copiarArquivo(path, destPath)
	})
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
