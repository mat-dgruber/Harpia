package hrp

import (
	"os"
	"path/filepath"
	"strings"
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

// Um pânico dentro de função nativa vira RuntimeErro (antes derrubava o processo).
func TestPanicoEmMetodoNativoViraErro(t *testing.T) {
	m := NewMetodoOuPanic("quebra", func(_ Objeto, _ Tupla) (Objeto, error) {
		var x Objeto = Inteiro(1)
		return x.(Tupla), nil // conversão inválida: pânico
	}, "")
	_, err := m.Chamar(nil, Tupla{})
	if err == nil || !strings.Contains(err.Error(), "interface conversion") {
		t.Fatalf("esperava RuntimeErro com a causa do pânico, veio %v", err)
	}
}

// mapa.chaves()/valores()/atualizar() precisam receber o próprio mapa como self.
func TestMetodosDeMapaVinculamInstancia(t *testing.T) {
	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()
	mod, err := ExecutarString(ctx, `var dados = {"a": 1}
dados.atualizar({"b": 2})
var n = tamanho(dados.chaves()) + tamanho(dados.valores())`)
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := mod.Escopo.ObterValor("n"); n != Inteiro(4) {
		t.Fatalf("esperava 4, veio %v", n)
	}
}

func TestOperadorEmComMapaUsaChaves(t *testing.T) {
	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()
	mod, err := ExecutarString(ctx, `var m = {"a": 1}
var sim = "a" em m
var nao = "b" em m`)
	if err != nil {
		t.Fatal(err)
	}
	sim, _ := mod.Escopo.ObterValor("sim")
	nao, _ := mod.Escopo.ObterValor("nao")
	if sim != Verdadeiro || nao != Falso {
		t.Fatalf("esperava Verdadeiro/Falso, veio %v/%v", sim, nao)
	}
}

// `retorne` sem valor dentro de `se` precisa sair da função.
func TestRetorneSemValorSaiDaFuncao(t *testing.T) {
	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()
	mod, err := ExecutarString(ctx, `var r = 0
funcao f() {
    se Verdadeiro {
        retorne
    }
    r = 1
}
f()`)
	if err != nil {
		t.Fatal(err)
	}
	if r, _ := mod.Escopo.ObterValor("r"); r != Inteiro(0) {
		t.Fatalf("o código depois do retorne rodou (r=%v)", r)
	}
}
