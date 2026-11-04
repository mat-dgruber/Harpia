package embutidos

import (
	"sync"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestIntSemArgumentos(t *testing.T) {
	res, err := met_emb_int(nil, hrp.Tupla{})
	if err != nil {
		t.Fatalf("erro inesperado em int(): %v", err)
	}
	val, ok := res.(hrp.Inteiro)
	if !ok || val != 0 {
		t.Fatalf("esperava Inteiro(0), obteve: %v", res)
	}
}

func TestTextoSemArgumentos(t *testing.T) {
	res, err := met_emb_texto(nil, hrp.Tupla{})
	if err != nil {
		t.Fatalf("erro inesperado em texto(): %v", err)
	}
	val, ok := res.(hrp.Texto)
	if !ok || val != "" {
		t.Fatalf("esperava Texto(\"\"), obteve: %v", res)
	}
}

func TestImprimaRetornaNulo(t *testing.T) {
	res, err := met_emb_imprima(nil, hrp.Tupla{hrp.Texto("teste")})
	if err != nil {
		t.Fatalf("erro inesperado em imprima: %v", err)
	}
	if res != hrp.Nulo {
		t.Fatalf("esperava hrp.Nulo, obteve: %v", res)
	}
}

func TestObjetoRoteadorConcorrente(t *testing.T) {
	roteador := &ObjetoRoteador{
		caminhoAtual:     hrp.Texto("/"),
		rotasCadastradas: hrp.NewMapaVazio(),
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(id int) {
			defer wg.Done()
			met, _ := roteador.M__obtem_attributo__("definir")
			_, _ = hrp.Chamar(met, hrp.Tupla{hrp.Texto("/rota"), hrp.Inteiro(id)})
		}(i)
		go func(id int) {
			defer wg.Done()
			met, _ := roteador.M__obtem_attributo__("navegar")
			_, _ = hrp.Chamar(met, hrp.Tupla{hrp.Texto("/outra")})
		}(i)
		go func() {
			defer wg.Done()
			met, _ := roteador.M__obtem_attributo__("obterRota")
			_, _ = hrp.Chamar(met, hrp.Tupla{})
		}()
	}
	wg.Wait()
}
