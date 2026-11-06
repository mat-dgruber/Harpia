package stdlib

import (
	"fmt"
	"net"
	nethttp "net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

func portaLivre(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// rodarEmSubprocesso executa o próprio teste num processo filho: antes da correção, escrita
// concorrente em Mapa era 'fatal error: concurrent map writes', que derruba o runner.
func rodarEmSubprocesso(t *testing.T, nome, env string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^"+nome+"$")
	cmd.Env = append(os.Environ(), env+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("processo filho falhou: %v\n%.3000s", err, out)
	}
}

// Handlers que mutam um mapa global não podem derrubar o servidor sob carga concorrente.
func TestHandlersConcorrentesNaoCorrompemMapa(t *testing.T) {
	if os.Getenv("HARPIA_FILHO_HTTP_MAPA") != "1" {
		rodarEmSubprocesso(t, "TestHandlersConcorrentesNaoCorrompemMapa", "HARPIA_FILHO_HTTP_MAPA")
		return
	}
	porta := portaLivre(t)
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	mod, err := hrp.ExecutarString(ctx, fmt.Sprintf(`de "http" importe Servidor
var contagem = {"x": 0}
var app = Servidor()
app.obter("/inc", funcao(req, res) { contagem["x"] = contagem["x"] + 1 })
app.escutar("%d")
`, porta))
	if err != nil {
		t.Fatal(err)
	}
	const n = 400
	vagas := make(chan struct{}, 32) // a fila de listen do macOS (128) reseta conexões acima disso
	var wg sync.WaitGroup
	var mu sync.Mutex
	falhas := map[string]int{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vagas <- struct{}{}
			defer func() { <-vagas }()
			resp, err := nethttp.Get(fmt.Sprintf("http://127.0.0.1:%d/inc", porta))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				falhas[err.Error()]++
				return
			}
			resp.Body.Close()
			if resp.StatusCode != nethttp.StatusOK {
				falhas[resp.Status]++
			}
		}()
	}
	wg.Wait()
	if len(falhas) > 0 {
		t.Fatalf("requisições falharam: %v", falhas)
	}
	contagem, _ := mod.Escopo.ObterValor("contagem")
	if v := contagem.(hrp.Mapa)["x"]; v != hrp.Inteiro(n) {
		t.Fatalf("esperava %d incrementos, veio %v", n, v)
	}
}

// Um handler que chama o próprio servidor não pode travar: com handlers serializados, a chamada
// interna recebe 503 em poucos segundos em vez de esperar o timeout de 30s do cliente.
func TestHandlerChamandoProprioServidorNaoTrava(t *testing.T) {
	porta := portaLivre(t)
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	_, err := hrp.ExecutarString(ctx, fmt.Sprintf(`de "http" importe Servidor, requisitar
var app = Servidor()
app.obter("/interno", funcao(req, res) { res.corpo = "interno" })
app.obter("/externo", funcao(req, res) {
    var r = requisitar("GET", "http://127.0.0.1:%d/interno")
    res.corpo = texto(r.status)
})
app.escutar("%d")
`, porta, porta))
	if err != nil {
		t.Fatal(err)
	}
	inicio := time.Now()
	resp, err := nethttp.Get(fmt.Sprintf("http://127.0.0.1:%d/externo", porta))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	corpo := make([]byte, 64)
	n, _ := resp.Body.Read(corpo)
	if d := time.Since(inicio); d > 9*time.Second {
		t.Fatalf("auto-chamada levou %v (travou até o timeout)", d)
	}
	if !strings.Contains(string(corpo[:n]), "503") {
		t.Fatalf("esperava a chamada interna receber 503, veio %q", corpo[:n])
	}
}
