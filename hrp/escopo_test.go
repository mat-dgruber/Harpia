package hrp

import (
	"sync"
	"testing"
)

func TestEscopoConcorrenteRace(t *testing.T) {
	escopoPai := NewEscopo(nil)
	if err := escopoPai.DefinirSimbolo(NewVarSimbolo("contador", Inteiro(0))); err != nil {
		t.Fatalf("falha ao definir símbolo inicial: %v", err)
	}
	escopoFilho := NewEscopo(escopoPai)

	const goroutines = 16
	const iteracoes = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(v int) {
			defer wg.Done()
			for j := 0; j < iteracoes; j++ {
				_ = escopoFilho.RedefinirValor("contador", Inteiro(v+j))
				_, _ = escopoFilho.ObterValor("contador")
				_ = escopoPai.ObterSimbolosSeguro()
			}
		}(i)
	}

	wg.Wait()
}
