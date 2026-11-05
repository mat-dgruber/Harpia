package stdlib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Todo pacote da stdlib que registra um módulo precisa ser importado pelo agregador
// (stdlib.go ou stdlib_native.go); senão o módulo existe, tem testes, mas o CLI não o encontra.
func TestAgregadorImportaTodosOsModulos(t *testing.T) {
	agregador := ""
	for _, f := range []string{"stdlib.go", "stdlib_native.go"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		agregador += string(b)
	}
	filepath.WalkDir(".", func(caminho string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(caminho, ".go") || strings.HasSuffix(caminho, "_test.go") {
			return nil
		}
		dir := filepath.Dir(caminho)
		if dir == "." {
			return nil
		}
		b, _ := os.ReadFile(caminho)
		if strings.Contains(string(b), "RegistraModuloImpl(") && !strings.Contains(agregador, `"github.com/mat-dgruber/Harpia/stdlib/`+filepath.ToSlash(dir)+`"`) {
			t.Errorf("stdlib/%s registra um módulo, mas não está importado no agregador", dir)
		}
		return nil
	})
}

func TestModulosAntesAusentesSaoImportaveis(t *testing.T) {
	for _, src := range []string{
		`de "autenticacao" importe criarJwt`,
		`de "fila" importe enfileirar, tamanho`,
		`de "seguranca" importe sanitizarHtml`,
	} {
		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		if _, err := hrp.ExecutarString(ctx, src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
		ctx.Terminar()
	}
}
