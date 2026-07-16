package tests

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestSSRServidorCompleto(t *testing.T) {
	// Cria uma pasta dist temporária simulada
	tempDir, err := os.MkdirTemp("", "dist_simulada")
	if err != nil {
		t.Fatalf("Erro ao criar dir temp: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Portuscript App</title>
</head>
<body>
    <div id="app"></div>
</body>
</html>`
	err = os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(indexHTML), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar index.html: %v", err)
	}

	// Força o caminho com barras normais para o import
	caminhoFormatado := filepath.ToSlash(tempDir)

	// Define o código em formato de string tradicional com \n explícitos de forma nativa e sem ponto e vírgula
	codigo := "de \"http\" importe Servidor\n\n" +
		"var meuApp = funcao() {\n" +
		"	retorne <div classe=\"p-8\"><h1>Olá do SSR!</h1></div>\n" +
		"}\n\n" +
		"var s = Servidor()\n" +
		"s.servir_app(\"" + caminhoFormatado + "\", meuApp, Nulo)\n" +
		"s.escutar(\"8084\")"

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao compilar: %v", err)
	}

	interpretador := &ptst.Interpretador{
		Ast:      ast,
		Contexto: ctx,
		Escopo:   ptst.NewEscopo(),
	}
	_, err = interpretador.Inicializa()
	if err != nil {
		t.Fatalf("Erro ao iniciar servidor: %v", err)
	}

	// Aguarda estabilização do servidor
	time.Sleep(200 * time.Millisecond)

	// Faz requisição de teste
	resp, err := http.Get("http://localhost:8084/")
	if err != nil {
		t.Fatalf("Erro ao fazer GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Esperado status 200, obtido: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Erro ao ler resposta: %v", err)
	}
	html := string(bodyBytes)

	// Validações de SSR e Injeções
	if !strings.Contains(html, `Olá do SSR`) || !strings.Contains(html, `class="p-8"`) {
		t.Errorf("HTML de SSR esperado no corpo, obtido: %s", html)
	}
}
