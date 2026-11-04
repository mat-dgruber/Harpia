package hrp_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestListaPopForaDoIntervaloEInPlace(t *testing.T) {
	// ponytail: teste cirúrgico de regressão para off-by-one e remoção in-place no Pop
	l := &hrp.Lista{Itens: hrp.Tupla{hrp.Inteiro(10), hrp.Inteiro(20), hrp.Inteiro(30)}}

	// Pop(3) deve retornar erro IndiceErro sem pânico (índice == tamanho)
	_, err := l.Pop(hrp.Inteiro(3))
	if err == nil {
		t.Fatal("esperava erro ao fazer Pop no índice igual ao tamanho da lista")
	}
	erroObj, ok := err.(*hrp.Erro)
	if !ok || erroObj.Base != hrp.IndiceErro {
		t.Fatalf("esperava IndiceErro ao fazer Pop(3), obteve %v", err)
	}

	// Pop(-1) deve retornar erro IndiceErro
	_, err = l.Pop(hrp.Inteiro(-1))
	if err == nil {
		t.Fatal("esperava erro ao fazer Pop no índice negativo")
	}
	erroObj, ok = err.(*hrp.Erro)
	if !ok || erroObj.Base != hrp.IndiceErro {
		t.Fatalf("esperava IndiceErro ao fazer Pop(-1), obteve %v", err)
	}

	// Pop(1) deve extrair o elemento 20 e reorganizar a fatia in-place para [10, 30]
	removido, err := l.Pop(hrp.Inteiro(1))
	if err != nil {
		t.Fatalf("erro inesperado no Pop(1): %v", err)
	}
	if removido != hrp.Inteiro(20) {
		t.Fatalf("esperava elemento removido 20, obteve %v", removido)
	}
	if len(l.Itens) != 2 || l.Itens[0] != hrp.Inteiro(10) || l.Itens[1] != hrp.Inteiro(30) {
		t.Fatalf("esperava lista resultante [10, 30], obteve %v", l.Itens)
	}
}
