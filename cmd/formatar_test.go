package cmd

import (
	"strings"
	"testing"
)

// TestFormatarCodigoPortuscript assevera a higienização de identações de blocos e linhas vazias
func TestFormatarCodigoPortuscript(t *testing.T) {
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

	formatado := FormatarCodigoPortuscript(codigoDesorganizado)
	formatadoLimp := strings.TrimSpace(formatado)

	if formatadoLimp != esperado {
		t.Errorf("Formatação incorreta.\nEsperado:\n%s\n\nObtido:\n%s", esperado, formatadoLimp)
	}
}
