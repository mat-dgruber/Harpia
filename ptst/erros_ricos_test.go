package ptst

import (
	"strings"
	"testing"
)

func TestErrosRicos(t *testing.T) {
	codigo := `var a = 1;
imrpimir(a);`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err == nil {
		t.Fatal("Esperava erro ao tentar executar código com identificador inexistente")
	}

	erroStr := err.Error()

	// Validar que contém o código de erro PSC-0005 (NomeErro)
	if !strings.Contains(erroStr, "PSC-0005") {
		t.Errorf("Esperava código PSC-0005 no erro, obtive:\n%s", erroStr)
	}

	// Validar que contém a sugestão contextual para imrpimir -> imprimir
	if !strings.Contains(erroStr, "Você quis dizer 'imprimir'?") {
		t.Errorf("Esperava sugestão contextual, obtive:\n%s", erroStr)
	}

	// Validar que contém a linha do erro impressa
	if !strings.Contains(erroStr, "imrpimir(a)") {
		t.Errorf("Esperava o trecho de código da linha com erro, obtive:\n%s", erroStr)
	}

	// Validar que contém o sublinhado circunflexo correspondente a "imrpimir" (tamanho 8)
	if !strings.Contains(erroStr, "^^^^^^^^") {
		t.Errorf("Esperava 8 acentos circunflexos para sublinhar 'imrpimir', obtive:\n%s", erroStr)
	}
}
