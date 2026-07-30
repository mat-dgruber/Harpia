package compartilhado

import (
	"hash/fnv"
	"sync"
)

type elementoLRU struct {
	chave string
	valor []int
}

type noLRU struct {
	elem elementoLRU
	prox *noLRU
	ant  *noLRU
}

type shardLRU struct {
	mu         sync.Mutex
	capacidade int
	tabela     map[string]*noLRU
	inicio     *noLRU
	fim        *noLRU
}

func novoShardLRU(capacidade int) *shardLRU {
	return &shardLRU{
		capacidade: capacidade,
		tabela:     make(map[string]*noLRU),
	}
}

func (s *shardLRU) obter(chave string) ([]int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	no, ok := s.tabela[chave]
	if !ok {
		return nil, false
	}
	s.moverParaInicio(no)
	return no.elem.valor, true
}

func (s *shardLRU) armazenar(chave string, valor []int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if no, ok := s.tabela[chave]; ok {
		no.elem.valor = valor
		s.moverParaInicio(no)
		return
	}

	novoNo := &noLRU{
		elem: elementoLRU{chave: chave, valor: valor},
	}

	if len(s.tabela) >= s.capacidade && s.fim != nil {
		s.removerNo(s.fim)
	}

	s.inserirNoInicio(novoNo)
	s.tabela[chave] = novoNo
}

func (s *shardLRU) moverParaInicio(no *noLRU) {
	if s.inicio == no {
		return
	}
	s.removerConexoes(no)
	s.inserirNoInicio(no)
}

func (s *shardLRU) removerNo(no *noLRU) {
	delete(s.tabela, no.elem.chave)
	s.removerConexoes(no)
}

func (s *shardLRU) removerConexoes(no *noLRU) {
	if no.ant != nil {
		no.ant.prox = no.prox
	} else {
		s.inicio = no.prox
	}
	if no.prox != nil {
		no.prox.ant = no.ant
	} else {
		s.fim = no.ant
	}
	no.ant = nil
	no.prox = nil
}

func (s *shardLRU) inserirNoInicio(no *noLRU) {
	no.prox = s.inicio
	no.ant = nil
	if s.inicio != nil {
		s.inicio.ant = no
	}
	s.inicio = no
	if s.fim == nil {
		s.fim = no
	}
}

func (s *shardLRU) limpar() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tabela = make(map[string]*noLRU)
	s.inicio = nil
	s.fim = nil
}

type shardedCache struct {
	shards    []*shardLRU
	numShards uint32
}

func novoShardedCache(numShards int, capPorShard int) *shardedCache {
	shards := make([]*shardLRU, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = novoShardLRU(capPorShard)
	}
	return &shardedCache{
		shards:    shards,
		numShards: uint32(numShards),
	}
}

func (sc *shardedCache) obterShard(chave string) *shardLRU {
	h := fnv.New32a()
	_, _ = h.Write([]byte(chave))
	idx := h.Sum32() % sc.numShards
	return sc.shards[idx]
}

func (sc *shardedCache) Obter(chave string) ([]int, bool) {
	return sc.obterShard(chave).obter(chave)
}

func (sc *shardedCache) Armazenar(chave string, valor []int) {
	sc.obterShard(chave).armazenar(chave, valor)
}

func (sc *shardedCache) Limpar() {
	for _, shard := range sc.shards {
		shard.limpar()
	}
}

var globalCache = novoShardedCache(256, 64)
