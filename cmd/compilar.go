package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/natanfeitosa/portuscript/ptst"
	"github.com/spf13/cobra"
)

func comandoCompilar() *cobra.Command {
	var alvo string
	var entrada string
	var saida string

	compilar := &cobra.Command{
		Use:   "compilar",
		Short: "Compila/Transpila o código Portuscript para a plataforma Web",
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

			if alvo != "web" {
				fmt.Fprintf(os.Stderr, "erro: alvo de compilação '%s' não suportado. Alvos suportados: web\n", alvo)
				os.Exit(1)
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
			transpiler := &TranspilerWeb{}
			jsOutput := transpiler.Transpile(ast)

			// Garante a existência da pasta de saída
			err = os.MkdirAll(saida, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao criar diretório de saída '%s': %v\n", saida, err)
				os.Exit(1)
			}

			// 1. Escreve o app.js
			appPath := filepath.Join(saida, "app.js")
			err = os.WriteFile(appPath, []byte("import { h, sinal, efeito, derivado, armazem, montar } from './runtime-web.js';\n\n"+jsOutput), 0644)
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
export function montar(app, el) { el.innerHTML = app().tag; }`
				os.WriteFile(runtimeDestPath, []byte(fallbackRuntime), 0644)
			}

			// 4. Cria um index.html básico para servir o app
			indexPath := filepath.Join(saida, "index.html")
			htmlTemplate := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Portuscript App</title>
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

			fmt.Println("🚀 Compilação concluída com sucesso!")
			fmt.Printf("Arquivos gerados em '%s/':\n", saida)
			fmt.Println("  - index.html   (ponto de entrada web)")
			fmt.Println("  - runtime-web.js (motor Virtual DOM & reatividade)")
			fmt.Println("  - app.js       (lógica de negócios transpilada)")
			fmt.Println("  - estilos.css  (folhas de estilo unificadas)")
		},
	}

	compilar.Flags().StringVarP(&alvo, "alvo", "a", "web", "Alvo da compilação (web)")
	compilar.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .ptst principal de entrada")
	compilar.Flags().StringVarP(&saida, "saida", "s", "dist", "Diretório de destino da compilação")
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
