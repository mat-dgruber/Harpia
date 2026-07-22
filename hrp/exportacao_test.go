package hrp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// TestExportacaoESimbolos valida o mecanismo de exportação e importação de variáveis, constantes
// e funções entre módulos físicos do Harpia, garantindo a resolução correta de dependências.
func TestExportacaoESimbolos(t *testing.T) {
	codigoModulo := `
	exportar constante PI = 3.14
	exportar funcao soma(a, b) {
		retorne a + b
	}
	`

	dir, err := os.MkdirTemp("", "Harpia_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	caminhoModulo := filepath.Join(dir, "constantes.hrp")
	if err := os.WriteFile(caminhoModulo, []byte(codigoModulo), 0644); err != nil {
		t.Fatal(err)
	}

	codigoPrincipal := `
	de "./constantes.hrp" importe PI, soma
	assegura PI == 3.14
	assegura soma(1, 2) == 3
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{
		CaminhosPadrao: []string{dir},
	})
	defer ctx.Terminar()

	_, err = hrp.ExecutarString(ctx, codigoPrincipal)
	if err != nil {
		t.Fatalf("Erro inesperado durante a execução de importação: %v", err)
	}
}

// TestImportacaoCiclica assegura que o mecanismo de resolução de dependências em runtime do Harpia
// detecte adequadamente cenários de importação circular/cíclica e lance um erro apropriado
// de forma controlada, impedindo loops infinitos ou estouro de pilha.
func TestImportacaoCiclica(t *testing.T) {
	dir, err := os.MkdirTemp("", "Harpia_ciclo_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// moduloA importa moduloB, moduloB importa moduloA
	codigoA := `
	exportar var a = 1
	de "./moduloB.hrp" importe b
	`
	codigoB := `
	exportar var b = 2
	de "./moduloA.hrp" importe a
	`

	if err := os.WriteFile(filepath.Join(dir, "moduloA.hrp"), []byte(codigoA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "moduloB.hrp"), []byte(codigoB), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := hrp.NewContexto(hrp.OpcsContexto{
		CaminhosPadrao: []string{dir},
	})
	defer ctx.Terminar()

	_, err = hrp.ExecutarString(ctx, `de "./moduloA.hrp" importe a`)
	if err == nil {
		t.Fatal("Esperava erro de importação cíclica, mas a execução terminou com sucesso.")
	}

	hrpErr, ok := err.(*hrp.Erro)
	if !ok || hrpErr.Base != hrp.ImportacaoErro {
		t.Fatalf("Esperava ImportacaoErro, obteve: %v", err)
	}
}
