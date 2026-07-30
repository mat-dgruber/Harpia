//go:build !js && !wasm

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
// e serve os assets estáticos de forma integrada.
func comandoServir() *cobra.Command {
	var porta int

	var entrada string
	var diretorio string
	var saidaTemp string

	servir := &cobra.Command{
		Use:           "servir",
		Short:         "Inicia o Dev Server de desenvolvimento integrado com compilação sob demanda",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			if len(args) > 0 {
				entrada = args[0]
			}

			// Se a flag --diretorio for informada, serve a pasta estática pré-compilada!
			if diretorio != "" {
				pAbs, err := filepath.Abs(diretorio)
				if err != nil {
					fmt.Fprintf(errOut, "erro ao resolver diretório '%s': %v\n", diretorio, err)
					return err
				}
				if _, err := os.Stat(pAbs); os.IsNotExist(err) {
					fmt.Fprintf(errOut, "erro: diretório '%s' não existe.\n", diretorio)
					return fmt.Errorf("diretório '%s' não existe", diretorio)
				}

				if _, errApp := os.Stat(filepath.Join(pAbs, "app.js")); os.IsNotExist(errApp) {
					if _, errDistApp := os.Stat(filepath.Join(pAbs, "dist", "app.js")); errDistApp == nil {
						pAbs = filepath.Join(pAbs, "dist")
					} else if _, errDistIdx := os.Stat(filepath.Join(pAbs, "dist", "index.html")); errDistIdx == nil {
						pAbs = filepath.Join(pAbs, "dist")
					}
				}

				fs := http.FileServer(http.Dir(pAbs))
				corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT, PATCH")
					if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					} else {
						w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					}
					if r.Method == "OPTIONS" {
						w.WriteHeader(http.StatusOK)
						return
					}

					p := filepath.Join(pAbs, r.URL.Path)
					info, err := os.Stat(p)
					if os.IsNotExist(err) || (err == nil && info.IsDir() && r.URL.Path != "/") {
						idxFile := filepath.Join(p, "index.html")
						if _, errIdx := os.Stat(idxFile); os.IsNotExist(errIdx) {
							if filepath.Ext(r.URL.Path) != "" {
								http.NotFound(w, r)
								return
							}
							http.ServeFile(w, r, filepath.Join(pAbs, "index.html"))
							return
						}
					}

					fs.ServeHTTP(w, r)
				})
				http.Handle("/", corsHandler)

				url := fmt.Sprintf("http://localhost:%d", porta)
				fmt.Fprintf(out, "\n🚀 Servidor Estático Harpia rodando em %s (diretório: %s)\n", url, pAbs)
				fmt.Fprintln(out, "   Pressione Ctrl+C para encerrar.")

				server := &http.Server{
					Addr: fmt.Sprintf(":%d", porta),
				}

				go func() {
					<-cmd.Context().Done()
					_ = server.Close()
				}()

				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Fprintf(errOut, "❌ Erro no servidor: %v\n", err)
					return err
				}
				return nil
			}

			if entrada == "" {
				candidatos := []string{
					"main.hrp", "index.hrp", "app.hrp",
					filepath.Join("web", "main.hrp"),
					filepath.Join("web", "index.hrp"),
					filepath.Join("web", "rotas", "index.hrp"),
					filepath.Join("web", "rotas", "rotas.hrp"),
				}
				for _, cand := range candidatos {
					if _, err := os.Stat(cand); err == nil {
						entrada = cand
						break
					}
				}
			}

			if entrada == "" {
				fmt.Fprintln(errOut, "erro: nenhum arquivo de entrada especificado e nenhum arquivo padrão (main.hrp, index.hrp) encontrado.")
				return fmt.Errorf("nenhum arquivo de entrada encontrado")
			}

			saidaTemp = filepath.Join(os.TempDir(), "Harpia_servir_dist")
			_ = os.MkdirAll(saidaTemp, 0755)

			var erroBuildAtual string
			var mutexErro sync.RWMutex

			fazerBuild := func() error {
				err := compilarParaWeb(OpcsCompilarWeb{
					Entrada:     entrada,
					Saida:       saidaTemp,
					PularLinter: false,
				})
				mutexErro.Lock()
				if err != nil {
					erroBuildAtual = err.Error()
					fmt.Fprintf(errOut, "❌ Erro no build: %v\n", err)
				} else {
					erroBuildAtual = ""
				}
				mutexErro.Unlock()
				return err
			}

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

			tInicio := time.Now()
			errBuild := executarComSpinner("Compilando projeto de desenvolvimento", fazerBuild)
			if errBuild != nil {
				fmt.Fprintf(errOut, "❌ Erro de compilação/sintaxe: %v\n", errBuild)
			} else {
				fmt.Fprintf(out, "⚡ Build inicial concluído em %v (%s)\n", time.Since(tInicio).Round(time.Millisecond), entrada)
			}

			go func() {
				if errosChecagem := ExecutarChecagemSilenciosa(entrada); errosChecagem > 0 {
					fmt.Fprintf(errOut, "⚠️ %d aviso(s) detectado(s) na checagem estática.\n", errosChecagem)
				}
			}()

			var clientes []chan bool
			var mutex sync.Mutex

			go func() {
				diretorioParaVer, _ := os.Getwd()
				if entryDir := filepath.Dir(entrada); entryDir != "" && entryDir != "." {
					diretorioParaVer = entryDir
				}
				ultimoCheck := time.Now()

				for {
					select {
					case <-cmd.Context().Done():
						return
					case <-time.After(500 * time.Millisecond):
					}

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
							fmt.Fprintf(out, "⚡ Recompilado com sucesso em %v. Recarregando navegador...\n", time.Since(tRebuild).Round(time.Millisecond))
							mutex.Lock()
							for _, ch := range clientes {
								select {
								case ch <- true:
								default:
								}
							}
							mutex.Unlock()
						} else {
							fmt.Fprintf(errOut, "❌ Erro no build: %v\n", errRebuild)
						}
					}
				}
			}()

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

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT, PATCH")
				if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				} else {
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				}

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

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

				infoP, errStat := os.Stat(p)
				eDiretorioSemIndex := false
				if errStat == nil && infoP.IsDir() && caminho != "/index.html" {
					if _, errIdx := os.Stat(filepath.Join(p, "index.html")); os.IsNotExist(errIdx) {
						eDiretorioSemIndex = true
					}
				}

				if os.IsNotExist(errStat) || eDiretorioSemIndex {
					if caminho == "/favicon.ico" || strings.Contains(caminho, "favicon") {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					base := filepath.Base(caminho)
					if !strings.Contains(base, ".") || eDiretorioSemIndex {
						caminho = "/index.html"
						p = filepath.Join(saidaTemp, "index.html")
					} else {
						http.NotFound(w, r)
						return
					}
				}

				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")

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
			fmt.Fprintf(out, "\n🚀 Harpia Dev Server rodando em %s\n", url)
			fmt.Fprintln(out, "✨ Hot-Reload ativo — edições em .hrp, .html ou .css atualizam o navegador instantaneamente.")
			fmt.Fprintln(out, "   Pressione Ctrl+C para encerrar.")
			fmt.Fprintln(out)

			server := &http.Server{
				Addr: fmt.Sprintf(":%d", porta),
			}

			go func() {
				<-cmd.Context().Done()
				_ = server.Close()
			}()

			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(errOut, "❌ Erro ao iniciar o servidor na porta %d: %v\n", porta, err)
				return err
			}
			return nil
		},
	}

	servir.Flags().IntVarP(&porta, "porta", "p", 3000, "Porta de escuta do Dev Server local")
	servir.Flags().StringVarP(&entrada, "entrada", "e", "", "Arquivo .hrp de entrada principal")
	servir.Flags().StringVarP(&diretorio, "diretorio", "d", "", "Diretório de arquivos estáticos para servir")
	return servir
}
