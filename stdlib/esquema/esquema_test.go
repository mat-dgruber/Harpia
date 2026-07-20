package esquema

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestEsquemaValidacaoSucesso(t *testing.T) {
	regras := hrp.Mapa{
		"nome":  hrp.TipoTexto,
		"idade": hrp.TipoInteiro,
	}
	objEsquema, err := met_esquema_criar(nil, hrp.Tupla{regras})
	if err != nil {
		t.Fatalf("Erro ao criar esquema: %v", err)
	}

	esq := objEsquema.(*Esquema)
	analisarMetodo, errAtt := esq.M__obtem_attributo__("analisar")
	if errAtt != nil {
		t.Fatalf("Erro ao obter método analisar: %v", errAtt)
	}

	dadosValidos := hrp.Mapa{
		"nome":  hrp.Texto("Carlos"),
		"idade": hrp.Inteiro(30),
	}

	res, errCall := hrp.Chamar(analisarMetodo, hrp.Tupla{dadosValidos})
	if errCall != nil {
		t.Fatalf("Erro ao chamar analisar: %v", errCall)
	}

	lista := res.(*hrp.Lista)
	if lista.Itens[0] == hrp.Nulo {
		t.Errorf("Esperava sucesso na validação, obteve erro: %v", lista.Itens[1])
	}
}

func TestEsquemaValidacaoFalhaTipo(t *testing.T) {
	regras := hrp.Mapa{
		"nome": hrp.TipoTexto,
	}
	objEsquema, _ := met_esquema_criar(nil, hrp.Tupla{regras})
	esq := objEsquema.(*Esquema)
	analisarMetodo, _ := esq.M__obtem_attributo__("analisar")

	dadosInvalidos := hrp.Mapa{
		"nome": hrp.Inteiro(123), // Tipo incorreto
	}

	res, _ := hrp.Chamar(analisarMetodo, hrp.Tupla{dadosInvalidos})
	lista := res.(*hrp.Lista)

	if lista.Itens[0] != hrp.Nulo {
		t.Fatalf("Esperava erro na validação devido a tipo incompatível")
	}

	erroStr := string(lista.Itens[1].(hrp.Texto))
	if !stringsContains(erroStr, "deve ser do tipo Texto, obteve Inteiro") {
		t.Errorf("Mensagem de erro inadequada: %s", erroStr)
	}
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
