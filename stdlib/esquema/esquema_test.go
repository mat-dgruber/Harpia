package esquema

import (
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestEsquemaValidacaoSucesso(t *testing.T) {
	regras := ptst.Mapa{
		"nome":  ptst.TipoTexto,
		"idade": ptst.TipoInteiro,
	}
	objEsquema, err := met_esquema_criar(nil, ptst.Tupla{regras})
	if err != nil {
		t.Fatalf("Erro ao criar esquema: %v", err)
	}

	esq := objEsquema.(*Esquema)
	analisarMetodo, errAtt := esq.M__obtem_attributo__("analisar")
	if errAtt != nil {
		t.Fatalf("Erro ao obter método analisar: %v", errAtt)
	}

	dadosValidos := ptst.Mapa{
		"nome":  ptst.Texto("Carlos"),
		"idade": ptst.Inteiro(30),
	}

	res, errCall := ptst.Chamar(analisarMetodo, ptst.Tupla{dadosValidos})
	if errCall != nil {
		t.Fatalf("Erro ao chamar analisar: %v", errCall)
	}

	lista := res.(*ptst.Lista)
	if lista.Itens[0] == ptst.Nulo {
		t.Errorf("Esperava sucesso na validação, obteve erro: %v", lista.Itens[1])
	}
}

func TestEsquemaValidacaoFalhaTipo(t *testing.T) {
	regras := ptst.Mapa{
		"nome": ptst.TipoTexto,
	}
	objEsquema, _ := met_esquema_criar(nil, ptst.Tupla{regras})
	esq := objEsquema.(*Esquema)
	analisarMetodo, _ := esq.M__obtem_attributo__("analisar")

	dadosInvalidos := ptst.Mapa{
		"nome": ptst.Inteiro(123), // Tipo incorreto
	}

	res, _ := ptst.Chamar(analisarMetodo, ptst.Tupla{dadosInvalidos})
	lista := res.(*ptst.Lista)

	if lista.Itens[0] != ptst.Nulo {
		t.Fatalf("Esperava erro na validação devido a tipo incompatível")
	}

	erroStr := string(lista.Itens[1].(ptst.Texto))
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
