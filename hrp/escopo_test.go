package hrp

import (
	"os"
	"path/filepath"
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

// NewContexto em paralelo (handlers HTTP, playground, testes) não pode ter corrida de dados.
func TestNewContextoParaleloSemCorrida(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("HARPIA_TESTE_PARALELO=1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := NewContexto(OpcsContexto{DiretorioTrabalho: dir})
			ctx.Terminar()
		}()
	}
	wg.Wait()
}

// M15: importar "mod" dentro de uma pasta cujo nome contém ".so" (ex.: proj.solar) funciona.
func TestResolveModuloEmPastaComPontoSo(t *testing.T) {
	base := filepath.Join(t.TempDir(), "proj.solar")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	alvo := filepath.Join(base, "mod.hrp")
	if err := os.WriteFile(alvo, []byte("var x = 1"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveArquivohrp("mod", nil, base)
	if err != nil || got != alvo {
		t.Fatalf("esperava %q, veio %q (%v)", alvo, got, err)
	}
}

// M32: imprimir(tipo(x)) de um tipo nativo mostrava a struct do Go.
func TestTipoViraTextoPeloNome(t *testing.T) {
	for _, tp := range []*Tipo{NewTipo("DataHora", "doc"), TipoInteiro, TipoTexto} {
		got, err := NewTexto(tp)
		if err != nil || got != Texto(tp.Nome) {
			t.Errorf("NewTexto(tipo %s) = %v (%v)", tp.Nome, got, err)
		}
	}
}
