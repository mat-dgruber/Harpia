package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// comandoEmpacotar retorna o comando Cobra "empacotar".
// Este comando compila e empacota um script Harpia embutindo-o diretamente
// no interpretador para gerar um executável nativo autônomo (Single Binary Bundle)
// ou compila todo o interpretador para rodar de forma portátil em WebAssembly (WASM).
func comandoEmpacotar() *cobra.Command {
	var entrada string
	var saida string
	var so string
	var arq string

	empacotar := &cobra.Command{
		Use:   "empacotar",
		Short: "Empacota um script Harpia em um executável nativo autônomo (Single Binary Bundle) ou WASM",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				entrada = args[0]
			}

			isWasm := (so == "js" && arq == "wasm")

			if entrada == "" && !isWasm {
				fmt.Fprintln(os.Stderr, "erro: arquivo de entrada não especificado")
				os.Exit(1)
			}

			if saida == "" {
				if isWasm {
					saida = filepath.Join("docs", "portal", "Harpia.wasm")
				} else {
					saida = "app_compilado"
					if so == "windows" {
						saida += ".exe"
					}
				}
			}

			// Localiza a raiz do projeto (onde está main.go)
			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao obter diretório atual: %v\n", err)
				os.Exit(1)
			}

			// Procura main.go para garantir que estamos no diretório certo
			caminhoMain := filepath.Join(cur, "main.go")
			if _, err := os.Stat(caminhoMain); os.IsNotExist(err) {
				// Fallback se estiver em subdiretório
				cur = filepath.Dir(cur)
				caminhoMain = filepath.Join(cur, "main.go")
			}

			if entrada != "" {
				conteudo, err := os.ReadFile(entrada)
				if err != nil {
					fmt.Fprintf(os.Stderr, "erro ao ler arquivo de entrada: %v\n", err)
					os.Exit(1)
				}

				// Cria o arquivo de embutir temporário
				tempGoFile := filepath.Join(filepath.Dir(caminhoMain), "z_embedded.go")
				contentGo := fmt.Sprintf(`package main

func init() {
	embeddedSource = %q
}
`, string(conteudo))

				err = os.WriteFile(tempGoFile, []byte(contentGo), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "erro ao gerar arquivo temporário de build: %v\n", err)
					os.Exit(1)
				}
				defer os.Remove(tempGoFile)
			}

			// Garante que o diretório de destino do build exista
			if dirDest := filepath.Dir(saida); dirDest != "." && dirDest != "" {
				os.MkdirAll(dirDest, 0755)
			}

			// Executa go build com as variáveis de ambiente corretas para cross-compilation
			if isWasm {
				fmt.Printf("📦 Compilando interpretador Harpia para WebAssembly [%s]...\n", saida)
			} else {
				fmt.Printf("📦 Empacotando '%s' para o executável nativo '%s' [%s/%s]...\n", entrada, saida, so, arq)
			}

			buildCmd := exec.Command("go", "build", "-o", saida, ".")
			buildCmd.Dir = filepath.Dir(caminhoMain)
			buildCmd.Env = append(os.Environ(),
				"GOOS="+so,
				"GOARCH="+arq,
			)

			out, err := buildCmd.CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro durante a compilação nativa Go:\n%s\n", string(out))
				os.Exit(1)
			}

			if isWasm {
				// ponytail: copia o wasm_exec.js correspondente à versão instalada do Go no sistema (prevê caminhos antigos e novos do brew)
				gorootOut, err := exec.Command("go", "env", "GOROOT").Output()
				if err == nil {
					goroot := strings.TrimSpace(string(gorootOut))
					caminhosWasmExec := []string{
						filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),
						filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"),
					}
					for _, origemWasmExec := range caminhosWasmExec {
						if _, err := os.Stat(origemWasmExec); err == nil {
							destinoWasmExec := filepath.Join(filepath.Dir(saida), "wasm_exec.js")
							data, err := os.ReadFile(origemWasmExec)
							if err == nil {
								os.WriteFile(destinoWasmExec, data, 0644)
								fmt.Printf("✓ Carregador 'wasm_exec.js' copiado com sucesso para: %s\n", destinoWasmExec)
								break
							}
						}
					}
				}
			}

			fmt.Printf("🚀 Sucesso! Binário nativo gerado em '%s'\n", saida)
		},
	}

	empacotar.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo Harpia de entrada.")
	empacotar.Flags().StringVarP(&saida, "saida", "s", "", "Caminho de saída do binário gerado.")
	empacotar.Flags().StringVar(&so, "so", runtime.GOOS, "Sistema Operacional alvo (linux, windows, darwin, js).")
	empacotar.Flags().StringVar(&arq, "arq", runtime.GOARCH, "Arquitetura alvo (amd64, arm64, wasm).")

	return empacotar
}
