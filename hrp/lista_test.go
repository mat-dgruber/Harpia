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

// extende/pop/junta com argumentos comuns não podem derrubar o processo nem ignorar a lista.
func TestListaExtendePopEJuntaRobustos(t *testing.T) {
	metodo := func(o hrp.Objeto, nome string) hrp.Objeto {
		m, err := hrp.ObtemAtributoS(o, nome)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	l := &hrp.Lista{Itens: hrp.Tupla{hrp.Inteiro(1)}}
	if _, err := hrp.Chamar(metodo(l, "extende"), hrp.Tupla{&hrp.Lista{Itens: hrp.Tupla{hrp.Inteiro(2), hrp.Inteiro(3)}}}); err != nil {
		t.Fatalf("extende(Lista): %v", err)
	}
	if _, err := hrp.Chamar(metodo(l, "extende"), hrp.Tupla{hrp.Tupla{hrp.Inteiro(4)}}); err != nil {
		t.Fatalf("extende(Tupla): %v", err)
	}
	if len(l.Itens) != 4 {
		t.Fatalf("esperava 4 itens, veio %v", l.Itens)
	}
	if _, err := hrp.Chamar(metodo(l, "extende"), hrp.Tupla{hrp.Inteiro(5)}); err == nil {
		t.Fatal("extende(5) deveria ser TipagemErro")
	}
	if _, err := hrp.Chamar(metodo(l, "pop"), hrp.Tupla{hrp.Texto("x")}); err == nil {
		t.Fatal("pop(\"x\") deveria ser TipagemErro")
	}

	sep := hrp.Texto(", ")
	for _, arg := range []hrp.Tupla{
		{&hrp.Lista{Itens: hrp.Tupla{hrp.Texto("a"), hrp.Texto("b"), hrp.Texto("c")}}},
		{hrp.Tupla{hrp.Texto("a"), hrp.Texto("b"), hrp.Texto("c")}},
		{hrp.Texto("a"), hrp.Texto("b"), hrp.Texto("c")},
	} {
		got, err := hrp.Chamar(metodo(sep, "junta"), arg)
		if err != nil || got != hrp.Texto("a, b, c") {
			t.Errorf("junta(%v) = %v (%v)", arg, got, err)
		}
	}
}

// O repr de coleções aninhadas não pode ser afetado pelo espalhamento do método `junta`.
func TestReprDeColecoesAninhadas(t *testing.T) {
	for _, c := range []struct {
		obj  hrp.Objeto
		quer string
	}{
		{&hrp.Lista{Itens: hrp.Tupla{&hrp.Lista{Itens: hrp.Tupla{hrp.Inteiro(1), hrp.Inteiro(2)}}}}, "[[1,2]]"},
		{hrp.Tupla{hrp.Tupla{hrp.Inteiro(1)}}, "((1))"},
	} {
		got, err := hrp.NewTexto(c.obj)
		if err != nil || string(got.(hrp.Texto)) != c.quer {
			t.Errorf("repr = %v (%v), queria %s", got, err, c.quer)
		}
	}
}
