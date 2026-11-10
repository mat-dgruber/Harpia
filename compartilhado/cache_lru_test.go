package compartilhado

import (
	"fmt"
	"sync"
	"testing"
)

func TestShardedCache_GetPutClear(t *testing.T) {
	cache := NovoShardedCache(256, 16)

	val, ok := cache.Obter("teste")
	if ok || val != nil {
		t.Fatalf("esperava miss para chave inexistente")
	}

	dados := []int{0, 1, 3, 4}
	cache.Armazenar("teste", dados)

	val, ok = cache.Obter("teste")
	if !ok || len(val) != 4 {
		t.Fatalf("esperava hit e slice de 4 elementos, obteve %v (ok=%v)", val, ok)
	}

	cache.Limpar()
	_, ok = cache.Obter("teste")
	if ok {
		t.Fatalf("esperava miss apos Limpar()")
	}
}

func TestShardedCache_FuncoesPacote(t *testing.T) {
	LimparCache()
	defer LimparCache()

	val, ok := ObterDoCache("global_test")
	if ok || val != nil {
		t.Fatalf("esperava miss global inicial")
	}

	ArmazenarNoCache("global_test", []int{10, 20})
	val, ok = ObterDoCache("global_test")
	if !ok || len(val) != 2 || val[0] != 10 {
		t.Fatalf("esperava hit global com [10, 20], obteve %v (ok=%v)", val, ok)
	}

	LimparCache()
	_, ok = ObterDoCache("global_test")
	if ok {
		t.Fatalf("esperava miss global apos LimparCache()")
	}
}

func TestShardedCache_Eviccao(t *testing.T) {
	cache := NovoShardedCache(1, 2)

	cache.Armazenar("chave1", []int{1})
	cache.Armazenar("chave2", []int{2})

	// Acessa chave1 para torna-la recentemente usada (chave2 vira LRU)
	val, ok := cache.Obter("chave1")
	if !ok || len(val) != 1 || val[0] != 1 {
		t.Fatalf("esperava encontrar chave1")
	}

	// Insere chave3. Como capacidade e 2, chave2 (LRU) deve ser descartada
	cache.Armazenar("chave3", []int{3})

	_, ok = cache.Obter("chave2")
	if ok {
		t.Fatalf("esperava que chave2 tivesse sido evictada")
	}

	_, ok = cache.Obter("chave1")
	if !ok {
		t.Fatalf("esperava encontrar chave1 que foi recentemente acessada")
	}

	_, ok = cache.Obter("chave3")
	if !ok {
		t.Fatalf("esperava encontrar chave3")
	}
}

func TestShardedCache_CopiaDefensiva(t *testing.T) {
	cache := NovoShardedCache(1, 10)

	origem := []int{1, 2, 3}
	cache.Armazenar("copia", origem)

	// Mutacao no slice original nao deve afetar o cache
	origem[0] = 999

	val, ok := cache.Obter("copia")
	if !ok || val[0] != 1 {
		t.Fatalf("mutacao no slice original alterou o cache: %v", val)
	}

	// Mutacao no slice retornado nao deve afetar o cache
	val[0] = 888

	val2, ok := cache.Obter("copia")
	if !ok || val2[0] != 1 {
		t.Fatalf("mutacao no slice retornado alterou o cache: %v", val2)
	}
}

func TestShardedCache_Concorrencia(t *testing.T) {
	cache := NovoShardedCache(16, 32)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			chave := fmt.Sprintf("chave_%d", id%20)
			cache.Armazenar(chave, []int{id, id + 1})
			cache.Obter(chave)
		}(i)
	}
	wg.Wait()
}
