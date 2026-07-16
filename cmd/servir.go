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

			if entrada == "" {
				fmt.Fprintln(os.Stderr, "erro: arquivo de entrada não especificado. Use o argumento posicional ou --entrada.")
				os.Exit(1)
			}

			saidaTemp = filepath.Join(os.TempDir(), "portuscript_servir_dist")
			_ = os.MkdirAll(saidaTemp, 0755)

			// Função de build instantâneo
			fazerBuild := func() error {
				buildCmd := comandoCompilar()
				buildCmd.SetArgs([]string{"--entrada", entrada, "--saida", saidaTemp})
				return buildCmd.Execute()
			}

			// Build inicial
			fmt.Printf("Iniciando build de desenvolvimento de '%s'...\n", entrada)
			if err := fazerBuild(); err != nil {
				fmt.Fprintf(os.Stderr, "Falha inicial no build: %v\n", err)
			}

			// Gerenciamento síncrono de clientes do Server-Sent Events (SSE)
			var clientes []chan bool
			var mutex sync.Mutex

			// Watcher de arquivos leve com polling recorrente
			go func() {
				diretorioParaVer := filepath.Dir(entrada)
				ultimoCheck := time.Now()

				for {
					time.Sleep(500 * time.Millisecond)
					mudou := false

					_ = filepath.Walk(diretorioParaVer, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return nil
						}
						if strings.Contains(path, "portuscript_servir_dist") || strings.Contains(path, "dist") {
							return nil
						}
						if !info.IsDir() && info.ModTime().After(ultimoCheck) {
							if strings.HasSuffix(path, ".hrp") || strings.HasSuffix(path, ".ptst") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".css") {
								mudou = true
								ultimoCheck = info.ModTime()
							}
						}
						return nil
					})

					if mudou {
						fmt.Println("\n🔄 Alteração detectada no disco! Recompilando...")
						if err := fazerBuild(); err == nil {
							fmt.Println("⚡ Build concluído com sucesso. Recarregando navegador...")
							mutex.Lock()
							for _, ch := range clientes {
								select {
								case ch <- true:
								default:
								}
							}
							mutex.Unlock()
						} else {
							fmt.Fprintf(os.Stderr, "Erro no build: %v\n", err)
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

				if caminho == "/" || caminho == "/index.html" || caminho == "/app.js" || caminho == "/estilos.css" {
					_ = fazerBuild()
				}

				if caminho == "/" {
					caminho = "/index.html"
				}

				p := filepath.Join(saidaTemp, caminho)
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
			fmt.Printf("\n📢 Dev Server rodando em %s\n", url)
			fmt.Println("👉 Altere seus arquivos .ptst, .html ou .css e o navegador recarregará automaticamente!")

			server := &http.Server{
				Addr:         fmt.Sprintf(":%d", porta),
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
			}

			if err := server.ListenAndServe(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao iniciar o servidor: %v\n", err)
				os.Exit(1)
			}
		},
	}

	servir.Flags().IntVarP(&porta, "porta", "p", 3000, "Porta de escuta do Dev Server local")
	servir.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .hrp de entrada principal")
	return servir
}
