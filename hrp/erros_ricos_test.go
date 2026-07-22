package hrp

import (
	"strings"
	"testing"
)

// TestErrosRicos valida o motor de diagnósticos ricos do Harpia, garantindo que
// erros de nomes não encontrados exibam o código PSC-xxxx correto, ofereçam sugestões
// contextuais baseadas na distância de Levenshtein (ex: imrpimir -> imprimir) e
// imprimam o trecho de código correspondente com sublinhado ANSI no terminal.
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

	// Validar que contém o sublinhado circunflexo correspondente a "imrpimir" (tamanho 8 ou tamanho 1 do identificador marcado)
	if !strings.Contains(erroStr, "^") {
		t.Errorf("Esperava acentos circunflexos para sublinhar 'imrpimir', obtive:\n%s", erroStr)
	}
}
