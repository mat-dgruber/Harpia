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
