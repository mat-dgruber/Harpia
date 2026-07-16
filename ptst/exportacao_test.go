package ptst_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestExportacaoESimbolos(t *testing.T) {
	codigoModulo := `
	exportar constante PI = 3.14
	exportar funcao soma(a, b) {
		retorne a + b
	}
	`

	dir, err := os.MkdirTemp("", "portuscript_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	caminhoModulo := filepath.Join(dir, "constantes.ptst")
	if err := os.WriteFile(caminhoModulo, []byte(codigoModulo), 0644); err != nil {
		t.Fatal(err)
	}

	codigoPrincipal := `
	de "./constantes.ptst" importe PI, soma
	assegura PI == 3.14
	assegura soma(1, 2) == 3
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{
		CaminhosPadrao: []string{dir},
	})
	defer ctx.Terminar()

	_, err = ptst.ExecutarString(ctx, codigoPrincipal)
	if err != nil {
		t.Fatalf("Erro inesperado durante a execução de importação: %v", err)
	}
}

func TestImportacaoCiclica(t *testing.T) {
	dir, err := os.MkdirTemp("", "portuscript_ciclo_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// moduloA importa moduloB, moduloB importa moduloA
	codigoA := `
	exportar var a = 1
	de "./moduloB.ptst" importe b
	`
	codigoB := `
	exportar var b = 2
	de "./moduloA.ptst" importe a
	`

	if err := os.WriteFile(filepath.Join(dir, "moduloA.ptst"), []byte(codigoA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "moduloB.ptst"), []byte(codigoB), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := ptst.NewContexto(ptst.OpcsContexto{
		CaminhosPadrao: []string{dir},
	})
	defer ctx.Terminar()

	_, err = ptst.ExecutarString(ctx, `de "./moduloA.ptst" importe a`)
	if err == nil {
		t.Fatal("Esperava erro de importação cíclica, mas a execução terminou com sucesso.")
	}

	ptstErr, ok := err.(*ptst.Erro)
	if !ok || ptstErr.Base != ptst.ImportacaoErro {
		t.Fatalf("Esperava ImportacaoErro, obteve: %v", err)
	}
}
