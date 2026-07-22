package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// ponytail: comando 'servir' que atua como Dev Server ultra-leve.
// Em vez de usar complexos watchers fsnotify e sockets de HMR que dão leaks de memória,
// ele re-compila o projeto sob demanda do browser de forma instantânea (on-request build, <5ms)
// e serve os assets estáticos de forma integrada.
func comandoServir() *cobra.Command {
	var porta int
	var entrada string
	var saidaTemp string

	servir := &cobra.Command{
		Use:   "servir",
		Short: "Inicia o Dev Server de desenvolvimento integrado com compilação sob demanda",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				entrada = args[0]
			}

			// ponytail: autodetecta o arquivo de entrada se nenhum for especificado!
			// Procura main.hrp, index.hrp, app.hrp no diretório atual ou na pasta web/
			if entrada == "" {
				candidatos := []string{
					"main.hrp", "index.hrp", "app.hrp",
					filepath.Join("web", "main.hrp"),
					filepath.Join("web", "index.hrp"),
					filepath.Join("web", "rotas", "index.hrp"),
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

			saidaTemp = filepath.Join(os.TempDir(), "Harpia_servir_dist")
			_ = os.MkdirAll(saidaTemp, 0755)

			var erroBuildAtual string
			var mutexErro sync.RWMutex

			// Função de build instantâneo com pré-checagem semântica
			// ponytail: compilarParaWeb é função pura (sem os.Exit), segura para chamar num servidor.
			fazerBuild := func() error {
				err := compilarParaWeb(OpcsCompilarWeb{
					Entrada:     entrada,
					Saida:       saidaTemp,
					PularLinter: false,
				})
				mutexErro.Lock()
				if err != nil {
					erroBuildAtual = err.Error()
					fmt.Fprintf(os.Stderr, "❌ Erro no build: %v\n", err)
				} else {
					erroBuildAtual = ""
				}
				mutexErro.Unlock()
				return err
			}

			// Helper de spinner animado no terminal
			executarComSpinner := func(mensagem string, acao func() error) error {
				frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
				parar := make(chan bool)
				var err error

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

				err = acao()
				close(parar)
				return err
			}


			// Build inicial instantâneo
			tInicio := time.Now()
			errBuild := executarComSpinner("Compilando projeto de desenvolvimento", fazerBuild)
			if errBuild != nil {
				fmt.Fprintf(os.Stderr, "❌ Erro de compilação/sintaxe: %v\n", errBuild)
			} else {
				fmt.Printf("⚡ Build inicial concluído em %v (%s)\n", time.Since(tInicio).Round(time.Millisecond), entrada)
			}

			// Pré-checagem semântica assíncrona para não atrasar o subimento do servidor
			go func() {
				if errosChecagem := ExecutarChecagemSilenciosa(entrada); errosChecagem > 0 {
					fmt.Fprintf(os.Stderr, "⚠️ %d aviso(s) detectado(s) na checagem estática.\n", errosChecagem)
				}
			}()

			// Gerenciamento síncrono de clientes do Server-Sent Events (SSE)
			var clientes []chan bool
			var mutex sync.Mutex

			// Watcher de arquivos leve com polling recorrente
			go func() {
				diretorioParaVer, _ := os.Getwd()
				if entryDir := filepath.Dir(entrada); entryDir != "" && entryDir != "." {
					diretorioParaVer = entryDir
				}
				ultimoCheck := time.Now()


				for {
					time.Sleep(500 * time.Millisecond)
					mudou := false

					_ = filepath.Walk(diretorioParaVer, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return nil
						}
						if strings.Contains(path, "Harpia_servir_dist") || strings.Contains(path, "dist") {
							return nil
						}
						if !info.IsDir() && info.ModTime().After(ultimoCheck) {
							if strings.HasSuffix(path, ".hrp") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".css") {
								mudou = true
								ultimoCheck = info.ModTime()
							}
						}
						return nil
					})

					if mudou {
						tRebuild := time.Now()
						errRebuild := executarComSpinner("Recompilando alterações", fazerBuild)
						if errRebuild == nil {
							fmt.Printf("⚡ Recompilado com sucesso em %v. Recarregando navegador...\n", time.Since(tRebuild).Round(time.Millisecond))
							mutex.Lock()
							for _, ch := range clientes {
								select {
								case ch <- true:
								default:
								}
							}
							mutex.Unlock()
						} else {
							fmt.Fprintf(os.Stderr, "❌ Erro no build: %v\n", errRebuild)
						}
					}
				}
			}()


			// Rota especial do SSE de Hot-Reload
			http.HandleFunc("/hot-reload", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", "*")

				flusher, ok := w.(http.Flusher)
				if !ok {
					http.Error(w, "Streaming não suportado", http.StatusInternalServerError)
					return
				}

				ch := make(chan bool, 1)
				mutex.Lock()
				clientes = append(clientes, ch)
				mutex.Unlock()

				defer func() {
					mutex.Lock()
					for i, c := range clientes {
						if c == ch {
							clientes = append(clientes[:i], clientes[i+1:]...)
							break
						}
					}
					mutex.Unlock()
					close(ch)
				}()

				fmt.Fprintf(w, "data: conectado\n\n")
				flusher.Flush()

				for {
					select {
					case <-r.Context().Done():
						return
					case <-ch:
						fmt.Fprintf(w, "data: recarregar\n\n")
						flusher.Flush()
					case <-time.After(10 * time.Second):
						fmt.Fprintf(w, "data: ping\n\n")
						flusher.Flush()
					}
				}
			})

			// Handler HTTP integrado para servir arquivos e injetar hot-reload dinâmico
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				caminho := r.URL.Path

				if caminho == "/" {
					caminho = "/index.html"
				}

				p := filepath.Join(saidaTemp, caminho)
				rel, err := filepath.Rel(saidaTemp, p)
				if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
					http.NotFound(w, r)
					return
				}

				if _, err := os.Stat(p); os.IsNotExist(err) {
					http.NotFound(w, r)
					return
				}

				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")

				// Injeta o script cliente de SSE de forma transparente em tempo de transmissão
				if caminho == "/index.html" {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")

					mutexErro.RLock()
					errStr := erroBuildAtual
					mutexErro.RUnlock()

					if errStr != "" {
						htmlErro := fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>Erro de Compilação — Harpia Dev Server</title>
  <style>
    body { background: #12131C; color: #F87171; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; padding: 40px; margin: 0; }
    .card { background: #1E1E2E; border: 1px solid #EF4444; border-radius: 12px; padding: 24px; box-shadow: 0 10px 30px rgba(0,0,0,0.5); max-width: 900px; margin: 0 auto; }
    h1 { color: #FCA5A5; font-size: 20px; margin-top: 0; display: flex; align-items: center; gap: 10px; }
    pre { background: #0F0F17; color: #F3F4F6; padding: 16px; border-radius: 8px; overflow-x: auto; white-space: pre-wrap; font-size: 14px; line-height: 1.5; border: 1px solid #374151; }
    .footer { margin-top: 20px; color: #9CA3AF; font-size: 13px; text-align: center; }
  </style>
</head>
<body>
  <div class="card">
    <h1>⚠️ Harpia — Erro de Compilação / Checagem</h1>
    <p style="color: #D1D5DB;">O navegador recarregará automaticamente assim que o erro for corrigido no código:</p>
    <pre>%s</pre>
  </div>
  <div class="footer">Dev Server Harpia ativo com Hot-Reload automático.</div>
  <script>
    const sse = new EventSource('/hot-reload');
    sse.onmessage = (e) => {
      if (e.data === 'recarregar') {
        console.log('⚡ Código corrigido! Recarregando...');
        window.location.reload();
      }
    };
  </script>
</body>
</html>`, strings.ReplaceAll(errStr, "<", "&lt;"))
						w.Write([]byte(htmlErro))
						return
					}

					conteudo, err := os.ReadFile(p)
					if err == nil {
						htmlStr := string(conteudo)
						scriptSSE := `
<script>
  (function() {
    const sse = new EventSource('/hot-reload');
    sse.onmessage = (e) => {
      if (e.data === 'recarregar') {
        console.log('🔄 Hot-Reload ativado! Recarregando página...');
        window.location.reload();
      }
    };
    sse.onerror = () => console.log('Sem conexão com Dev Server de Hot-Reload.');
  })();
</script>
`
						if strings.Contains(htmlStr, "</body>") {
							htmlStr = strings.Replace(htmlStr, "</body>", scriptSSE+"</body>", 1)
						} else {
							htmlStr += scriptSSE
						}
						w.Write([]byte(htmlStr))
						return
					}
				}

				http.ServeFile(w, r, p)
			})

			url := fmt.Sprintf("http://localhost:%d", porta)
			fmt.Printf("\n🚀 Harpia Dev Server rodando em %s\n", url)
			fmt.Println("✨ Hot-Reload ativo — edições em .hrp, .html ou .css atualizam o navegador instantaneamente.")
			fmt.Println("   Pressione Ctrl+C para encerrar.")
			fmt.Println()

			server := &http.Server{
				Addr: fmt.Sprintf(":%d", porta),
			}

			if err := server.ListenAndServe(); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Erro ao iniciar o servidor na porta %d: %v\n", porta, err)
				os.Exit(1)
			}

		},
	}

	servir.Flags().IntVarP(&porta, "porta", "p", 3000, "Porta de escuta do Dev Server local")
	servir.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .hrp de entrada principal")
	return servir
}
