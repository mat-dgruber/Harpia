package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestGerarDeHtml(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "teste.pdf")

	args := hrp.Tupla{
		hrp.Texto("<h1>Titulo</h1>"),
		hrp.Texto(outPath),
	}

	res, err := met_pdf_gerarDeHtml(nil, args)
	if err != nil {
		t.Fatalf("Erro ao gerar PDF: %v", err)
	}

	if res != hrp.Verdadeiro {
		t.Errorf("Esperava retorno Verdadeiro, obteve: %v", res)
	}

	conteudo, errRead := os.ReadFile(outPath)
	if errRead != nil {
		t.Fatalf("Erro ao ler arquivo PDF gerado: %v", errRead)
	}

	if !strings.HasPrefix(string(conteudo), "%PDF-1.4") {
		t.Errorf("Cabecalho PDF invalido no arquivo gerado")
	}
}
