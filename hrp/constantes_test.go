package hrp

import (
	"testing"
)

// TestConstanteImutavel valida a imutabilidade estrita das constantes declaradas,
// assegurando que qualquer tentativa de reatribuição de valor em runtime lance um erro amigável.
func TestConstanteImutavel(t *testing.T) {
	codigoTentativaReatribuicao := `
	const a = 10
	a = 20
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigoTentativaReatribuicao)
	if err == nil {
		t.Fatal("Esperava erro ao tentar reatribuir valor a uma constante")
	}

	erroStr := err.Error()
	if !contains(erroStr, "pois é uma constante") {
		t.Errorf("Mensagem de erro inadequada: %s", erroStr)
	}
}

// TestConstanteSemInicializador verifica se a declaração de uma constante sem seu respectivo
// valor inicializador dispara um erro sintático/semântico esperado no parser/interpretador.
func TestConstanteSemInicializador(t *testing.T) {
	codigoSemInicializador := `
	const a
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigoSemInicializador)
	if err == nil {
		t.Fatal("Esperava erro de compilação/sintaxe ao declarar constante sem inicializador")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr))
}
