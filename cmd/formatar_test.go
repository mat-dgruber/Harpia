//go:build !js && !wasm

package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestFormatarCodigoHarpia assevera a higienização de identações de blocos e linhas vazias
func TestFormatarCodigoHarpia(t *testing.T) {
	codigoDesorganizado := `
funcao MeuApp() {
var a = 10;

if a > 5 {
imprimir("Alto");
}
}
`

	esperado := strings.TrimSpace(`
funcao MeuApp() {
    var a = 10;

    if a > 5 {
        imprimir("Alto");
    }
}
`)

	formatado := FormatarCodigoHarpia(codigoDesorganizado)
	formatadoLimp := strings.TrimSpace(formatado)

	if formatadoLimp != esperado {
		t.Errorf("Formatação incorreta.\nEsperado:\n%s\n\nObtido:\n%s", esperado, formatadoLimp)
	}
}

func TestGerarUnifiedDiff(t *testing.T) {
	ui := &UI{}
	orig := "se x {\n  imprimir(1)\n}"
	form := "se x {\n    imprimir(1)\n}\n"

	diff := GerarUnifiedDiff("teste.hrp", orig, form, ui)
	if !strings.Contains(diff, "--- a/teste.hrp") || !strings.Contains(diff, "+++ b/teste.hrp") {
		t.Errorf("diff ausente de cabecalho esperado: %s", diff)
	}
}

func TestFormatarStdinEChavesEmStrings(t *testing.T) {
	codigo := "funcao x(){\nimprimir(\"}\")\n}"
	cmd := comandoFormatar()
	cmd.SetIn(strings.NewReader(codigo))
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Comando falhou ao ler de stdin sem argumentos: %v", err)
	}

	res := out.String()
	esperado := "funcao x() {\n    imprimir(\"}\")\n}\n"
	if res != esperado {
		t.Errorf("Formatação inesperada via stdin.\nEsperado:\n%q\nObtido:\n%q", esperado, res)
	}
}
