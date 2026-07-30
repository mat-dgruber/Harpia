package compartilhado

import (
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
	return append([]int(nil), no.elem.valor...), true
}

func (s *shardLRU) armazenar(chave string, valor []int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	valorCopia := append([]int(nil), valor...)

	if no, ok := s.tabela[chave]; ok {
		no.elem.valor = valorCopia
		s.moverParaInicio(no)
		return
	}

	novoNo := &noLRU{
		elem: elementoLRU{chave: chave, valor: valorCopia},
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

const (
	fnvOffset32 = 2166136261
	fnvPrime32  = 16777619
)

func hashFNV1a(chave string) uint32 {
	var h uint32 = fnvOffset32
	for i := 0; i < len(chave); i++ {
		h ^= uint32(chave[i])
		h *= fnvPrime32
	}
	return h
}

type ShardedCache struct {
	shards    []*shardLRU
	numShards uint32
}

func NovoShardedCache(numShards int, capPorShard int) *ShardedCache {
	shards := make([]*shardLRU, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = novoShardLRU(capPorShard)
	}
	return &ShardedCache{
		shards:    shards,
		numShards: uint32(numShards),
	}
}

func (sc *ShardedCache) obterShard(chave string) *shardLRU {
	idx := hashFNV1a(chave) % sc.numShards
	return sc.shards[idx]
}

func (sc *ShardedCache) Obter(chave string) ([]int, bool) {
	return sc.obterShard(chave).obter(chave)
}

func (sc *ShardedCache) Armazenar(chave string, valor []int) {
	sc.obterShard(chave).armazenar(chave, valor)
}

func (sc *ShardedCache) Limpar() {
	for _, shard := range sc.shards {
		shard.limpar()
	}
}

var globalCache = NovoShardedCache(256, 64)

func ObterDoCache(chave string) ([]int, bool) {
	return globalCache.Obter(chave)
}

func ArmazenarNoCache(chave string, valor []int) {
	globalCache.Armazenar(chave, valor)
}

func LimparCache() {
	globalCache.Limpar()
}
