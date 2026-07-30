package compartilhado

import (
	"fmt"
	"sync"
	"testing"
)

func TestShardedCache_GetPutClear(t *testing.T) {
	cache := novoShardedCache(256, 16)

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
	val, ok = cache.Obter("teste")
	if ok {
		t.Fatalf("esperava miss apos Limpar()")
	}
}

func TestShardedCache_Concorrencia(t *testing.T) {
	cache := novoShardedCache(16, 32)
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
