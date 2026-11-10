//go:build !js && !wasm

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"
)

func executarNoPlayground(t *testing.T, codigo string, ajustar func(*http.Request)) (*httptest.ResponseRecorder, ExecucaoPlaygroundResponse) {
	t.Helper()
	corpo, _ := json.Marshal(ExecucaoPlaygroundRequest{Codigo: codigo})
	req := httptest.NewRequest(http.MethodPost, "/api/executar", strings.NewReader(string(corpo)))
	req.Host = "localhost:8090"
	req.Header.Set("Content-Type", "application/json")
	if ajustar != nil {
		ajustar(req)
	}
	rec := httptest.NewRecorder()
	apiExecutarCodigoPlayground(rec, req)
	var resp ExecucaoPlaygroundResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return rec, resp
}

func semNbsp(s string) string { return strings.ReplaceAll(s, "&nbsp;", " ") }

func TestPlaygroundNegaProcessosArquivosERede(t *testing.T) {
	casos := map[string]string{
		"processo": "de \"sistema\" importe executar\nimprimir(executar(\"echo\", [\"pwned\"]))",
		"arquivo":  "de \"arquivos\" importe ler\nimprimir(ler(\"/etc/hosts\"))",
		"rede":     "de \"http\" importe requisitar\nimprimir(requisitar(\"GET\", \"http://127.0.0.1:1\"))",
	}
	for nome, codigo := range casos {
		t.Run(nome, func(t *testing.T) {
			_, resp := executarNoPlayground(t, codigo, nil)
			if resp.Sucesso || strings.Contains(resp.Saida, "pwned") || strings.Contains(resp.Saida, "localhost") {
				t.Fatalf("sandbox não bloqueou: sucesso=%v saida=%q", resp.Sucesso, resp.Saida)
			}
			if !strings.Contains(semNbsp(resp.ErroHtml), "Acesso Negado") {
				t.Fatalf("esperava 'Acesso Negado', veio %q", resp.ErroHtml)
			}
		})
	}
}

func TestPlaygroundAbortaLacoInfinito(t *testing.T) {
	antes := runtime.NumGoroutine()
	inicio := time.Now()
	_, resp := executarNoPlayground(t, "enquanto Verdadeiro { }", nil)
	if d := time.Since(inicio); d > 8*time.Second {
		t.Fatalf("demorou %v para abortar", d)
	}
	if resp.Sucesso || !strings.Contains(semNbsp(resp.ErroHtml), "Tempo limite") {
		t.Fatalf("esperava erro de tempo limite, veio sucesso=%v erro=%q", resp.Sucesso, resp.ErroHtml)
	}
	time.Sleep(1 * time.Second)
	if depois := runtime.NumGoroutine(); depois > antes+2 {
		t.Fatalf("script zumbi continua rodando: goroutines %d -> %d", antes, depois)
	}
}

func TestPlaygroundRecusaOrigemExterna(t *testing.T) {
	casos := map[string]func(*http.Request){
		"text/plain sem preflight": func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") },
		"origin externa":           func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") },
		"host da LAN":              func(r *http.Request) { r.Host = "192.168.0.10:8090" },
		"dns rebinding":            func(r *http.Request) { r.Host = "evil.example:8090" },
	}
	for nome, ajustar := range casos {
		t.Run(nome, func(t *testing.T) {
			rec, _ := executarNoPlayground(t, "imprimir(1)", ajustar)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("esperava 403, veio %d", rec.Code)
			}
		})
	}
	rec, _ := executarNoPlayground(t, "imprimir(1)", func(r *http.Request) { r.Header.Set("Origin", "http://localhost:8090") })
	if rec.Code != http.StatusOK {
		t.Fatalf("origem própria deveria passar, veio %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("não deveria emitir CORS, veio %q", got)
	}
}

func TestCaminhoOcultoBloqueado(t *testing.T) {
	for _, p := range []string{"/.env", "/.git/config", "/sub/.segredo", "/a/../.env"} {
		if !caminhoTemSegmentoOculto(p) {
			t.Errorf("%q deveria ser bloqueado", p)
		}
	}
	for _, p := range []string{"/index.html", "/assets/app.js", "/a.b/c"} {
		if caminhoTemSegmentoOculto(p) {
			t.Errorf("%q não deveria ser bloqueado", p)
		}
	}
}
