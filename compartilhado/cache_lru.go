package compartilhado

import (
	"sync"
)

// elementoLRU representa um nó interno contendo a chave e o valor do cache.
type elementoLRU struct {
	chave string
	valor []int
}

// noLRU representa um nó na lista duplamente encadeada do shard LRU.
type noLRU struct {
	elem elementoLRU
	prox *noLRU
	ant  *noLRU
}

// shardLRU é uma partição individual do cache LRU protegida por seu próprio Mutex.
type shardLRU struct {
	mu         sync.Mutex
	capacidade int
	tabela     map[string]*noLRU
	inicio     *noLRU
	fim        *noLRU
}

// novoShardLRU inicializa um novo shard LRU com a capacidade máxima configurada.
func novoShardLRU(capacidade int) *shardLRU {
	return &shardLRU{
		capacidade: capacidade,
		tabela:     make(map[string]*noLRU),
	}
}

// obter recupera o valor associado à chave e move o nó para o início da lista LRU.
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

// armazenar insere ou atualiza um valor no shard, promovendo o item e removendo o LRU se exceder capacidade.
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

// moverParaInicio posiciona um nó no topo da lista LRU (mais recentemente usado).
func (s *shardLRU) moverParaInicio(no *noLRU) {
	if s.inicio == no {
		return
	}
	s.removerConexoes(no)
	s.inserirNoInicio(no)
}

// removerNo elimina o nó da tabela hash e desfaz os ponteiros da lista encadeada.
func (s *shardLRU) removerNo(no *noLRU) {
	delete(s.tabela, no.elem.chave)
	s.removerConexoes(no)
}

// removerConexoes ajusta os ponteiros anterior e próximo de um nó a ser removido da lista.
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

// inserirNoInicio insere um novo nó na cabeça da lista encadeada.
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

// limpar esvazia a tabela e reseta os ponteiros da lista encadeada do shard.
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

// hashFNV1a calcula a hash FNV-1a de 32 bits de forma inline sem alocações.
func hashFNV1a(chave string) uint32 {
	var h uint32 = fnvOffset32
	for i := 0; i < len(chave); i++ {
		h ^= uint32(chave[i])
		h *= fnvPrime32
	}
	return h
}

// ShardedCache é uma estrutura de cache LRU particionada em M shards independentes.
// Reduz a contenção de locks em cenários de alta concorrência (ex: servidores web/HTTP).
type ShardedCache struct {
	shards    []*shardLRU
	numShards uint32
}

// NovoShardedCache instancia um novo ShardedCache com o número de shards e capacidade por shard.
func NovoShardedCache(numShards int, capPorShard int) *ShardedCache {
	if numShards <= 0 {
		numShards = 256
	}
	if capPorShard <= 0 {
		capPorShard = 64
	}
	shards := make([]*shardLRU, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = novoShardLRU(capPorShard)
	}
	return &ShardedCache{
		shards:    shards,
		numShards: uint32(numShards),
	}
}

// obterShard retorna a partição correspondente à chave via FNV-1a hash.
func (sc *ShardedCache) obterShard(chave string) *shardLRU {
	idx := hashFNV1a(chave) % sc.numShards
	return sc.shards[idx]
}

// Obter busca uma chave no cache sharded de forma thread-safe.
func (sc *ShardedCache) Obter(chave string) ([]int, bool) {
	return sc.obterShard(chave).obter(chave)
}

// Armazenar insere ou atualiza uma chave no cache sharded de forma thread-safe.
func (sc *ShardedCache) Armazenar(chave string, valor []int) {
	sc.obterShard(chave).armazenar(chave, valor)
}

// Limpar reseta todas as entradas de todos os shards.
func (sc *ShardedCache) Limpar() {
	for _, shard := range sc.shards {
		shard.limpar()
	}
}

// globalCache é a instância padrão do cache LRU sharded (256 shards x 64 itens = 16.384 entradas).
var globalCache = NovoShardedCache(256, 64)

// ObterDoCache consulta o cache global sharded por uma chave.
func ObterDoCache(chave string) ([]int, bool) {
	return globalCache.Obter(chave)
}

// ArmazenarNoCache insere ou atualiza uma chave no cache global sharded.
func ArmazenarNoCache(chave string, valor []int) {
	globalCache.Armazenar(chave, valor)
}

// LimparCache limpa todas as entradas do cache global sharded.
func LimparCache() {
	globalCache.Limpar()
}
