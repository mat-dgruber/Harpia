package tests

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

// TestIteradorSobreTexto verifica que 'para-em' sobre um Texto itera caractere a caractere (Unicode-safe).
func TestIteradorSobreTexto(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
var chars = []
var texto = "Olá"
para c em texto {
    chars.adiciona(c)
}
`
	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao iterar sobre Texto: %v", err)
	}

	val, err := res.Escopo.ObterValor("chars")
	if err != nil {
		t.Fatalf("Variável 'chars' não encontrada: %v", err)
	}

	lista, ok := val.(*hrp.Lista)
	if !ok {
		t.Fatalf("Esperava Lista, obteve %T", val)
	}

	// "Olá" tem 3 caracteres Unicode: O, l, á
	if len(lista.Itens) != 3 {
		t.Errorf("Esperava 3 caracteres, obteve %d: %v", len(lista.Itens), lista.Itens)
	}

	esperados := []string{"O", "l", "á"}
	for i, esperado := range esperados {
		if string(lista.Itens[i].(hrp.Texto)) != esperado {
			t.Errorf("char[%d]: esperava '%s', obteve '%s'", i, esperado, lista.Itens[i])
		}
	}
}

// TestIteradorSobreTextoPalavra verifica iteração básica sobre texto simples ASCII.
func TestIteradorSobreTextoPalavra(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
var total = 0
para c em "abc" {
    total = total + 1
}
`
	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao iterar sobre Texto ASCII: %v", err)
	}

	val, err := res.Escopo.ObterValor("total")
	if err != nil {
		t.Fatalf("Variável 'total' não encontrada: %v", err)
	}

	if int64(val.(hrp.Inteiro)) != 3 {
		t.Errorf("Esperava total=3 para 'abc', obteve %v", val)
	}
}

// TestIteradorNativoTexto verifica a API de baixo nível do Iterador com Texto diretamente via Go.
func TestIteradorNativoTexto(t *testing.T) {
	iter, err := hrp.NewIterador(hrp.Texto("Harpia"))
	if err != nil {
		t.Fatalf("NewIterador(Texto) falhou: %v", err)
	}

	esperados := []string{"H", "a", "r", "p", "i", "a"}
	for i, esp := range esperados {
		item, err := iter.M__proximo__()
		if err != nil {
			t.Fatalf("M__proximo__() falhou no índice %d: %v", i, err)
		}
		if string(item.(hrp.Texto)) != esp {
			t.Errorf("item[%d]: esperava '%s', obteve '%v'", i, esp, item)
		}
	}

	// Próxima chamada deve retornar FimIteracao
	_, err = iter.M__proximo__()
	if err == nil {
		t.Error("Esperava FimIteracao após esgotar o Texto, mas não houve erro")
	}
}

// TestIteradorNativoBytes verifica a API de baixo nível do Iterador com *Bytes diretamente via Go.
func TestIteradorNativoBytes(t *testing.T) {
	bs := &hrp.Bytes{Itens: []byte("ABC")}
	iter, err := hrp.NewIterador(bs)
	if err != nil {
		t.Fatalf("NewIterador(*Bytes) falhou: %v", err)
	}

	esperados := []int64{65, 66, 67}
	for i, esp := range esperados {
		item, err := iter.M__proximo__()
		if err != nil {
			t.Fatalf("M__proximo__() falhou no índice %d: %v", i, err)
		}
		v, ok := item.(hrp.Inteiro)
		if !ok {
			t.Fatalf("item[%d]: esperava Inteiro, obteve %T", i, item)
		}
		if int64(v) != esp {
			t.Errorf("byte[%d]: esperava %d, obteve %d", i, esp, v)
		}
	}

	// Próxima chamada deve retornar FimIteracao
	_, err = iter.M__proximo__()
	if err == nil {
		t.Error("Esperava FimIteracao após esgotar o Bytes, mas não houve erro")
	}
}

// TestBytesConversaoErroAmigavel verifica que converter Bytes -> Inteiro/Decimal retorna erro amigável.
func TestBytesConversaoErroAmigavel(t *testing.T) {
	b := &hrp.Bytes{Itens: []byte("A")}

	_, err := b.M__inteiro__()
	if err == nil {
		t.Error("M__inteiro__() em *Bytes deveria retornar erro NaoImplementadoErro")
	}

	_, err = b.M__decimal__()
	if err == nil {
		t.Error("M__decimal__() em *Bytes deveria retornar erro NaoImplementadoErro")
	}
}
