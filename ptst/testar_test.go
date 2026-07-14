package ptst

import (
	"testing"
)

func TestTestarEDiretivas(t *testing.T) {
	codigo := `
	testar "soma bem sucedida" {
		assegura(1 + 1 == 2)
	}

	testar "soma mal sucedida" {
		assegura(1 + 1 == 3, "um mais um deve ser dois")
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err == nil {
		t.Fatal("Esperava falha no teste mal sucedido, mas rodou sem erro")
	}

	if objErr, ok := err.(*Erro); ok {
		if objErr.Tipo() != ErroDeAsseguracao {
			t.Errorf("Tipo do erro incorreto: %v", objErr.Tipo())
		}
	} else {
		t.Fatalf("Erro não é do tipo Erro: %v", err)
	}
}
