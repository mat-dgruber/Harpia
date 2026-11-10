package http

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// app.obter(rota, mw, handler): o middleware recebe proximo e decide se o handler roda.
func TestEncadearRotaPassaProximo(t *testing.T) {
	var ordem []string
	mw := func(nome string, segue bool) hrp.Objeto {
		return hrp.NewMetodoOuPanic(nome, func(_ hrp.Objeto, a hrp.Tupla) (hrp.Objeto, error) {
			ordem = append(ordem, nome)
			if len(a) != 3 {
				t.Fatalf("%s deveria receber (req, res, proximo), veio %d args", nome, len(a))
			}
			if segue {
				return hrp.Chamar(a[2], nil)
			}
			return hrp.Nulo, nil
		}, "")
	}
	final := hrp.NewMetodoOuPanic("final", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
		ordem = append(ordem, "final")
		return hrp.Nulo, nil
	}, "")
	req, res := hrp.Texto("req"), hrp.Texto("res")

	hrp.Chamar(encadearRota(hrp.Tupla{mw("a", true), mw("b", true)}, final), hrp.Tupla{req, res})
	if got := len(ordem); got != 3 || ordem[2] != "final" {
		t.Fatalf("esperava a, b, final; veio %v", ordem)
	}

	ordem = nil
	hrp.Chamar(encadearRota(hrp.Tupla{mw("bloqueia", false)}, final), hrp.Tupla{req, res})
	if len(ordem) != 1 {
		t.Fatalf("middleware sem proximo() deveria barrar o handler; veio %v", ordem)
	}
}

func TestRespostaEnviarTextoPuro(t *testing.T) {
	r := &Resposta{Cabecalho: hrp.NewMapaVazio()}
	enviar, err := r.M__obtem_attributo__("enviar")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hrp.Chamar(enviar, hrp.Tupla{hrp.Texto("ok 1")}); err != nil {
		t.Fatal(err)
	}
	tipo, _ := r.Cabecalho.M__obtem_item__(hrp.Texto("Content-Type"))
	if r.Corpo != hrp.Texto("ok 1") || tipo != hrp.Texto("text/plain; charset=utf-8") {
		t.Fatalf("corpo=%v tipo=%v", r.Corpo, tipo)
	}
}

// Middleware anexa dados à requisição; campos nativos continuam somente leitura.
func TestRequisicaoAceitaAtributosDeMiddleware(t *testing.T) {
	r := &Requisicao{Metodo: hrp.Texto("GET")}
	if err := r.M__define_atributo__("usuario", hrp.Texto("ada")); err != nil {
		t.Fatal(err)
	}
	if v, err := r.M__obtem_attributo__("usuario"); err != nil || v != hrp.Texto("ada") {
		t.Fatalf("usuario = %v (%v)", v, err)
	}
	if err := r.M__define_atributo__("metodo", hrp.Texto("POST")); err == nil {
		t.Fatal("metodo deveria ser somente leitura")
	}
}
