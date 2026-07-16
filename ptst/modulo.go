package ptst

import (
	"sync"

	"github.com/natanfeitosa/portuscript/parser"
)

// ModuloInfo armazena os metadados declarativos que descrevem as propriedades básicas de identificação de um módulo.
type ModuloInfo struct {
	Nome    string // Nome identificador legível do módulo (disponível sob a constante '__nome__').
	Doc     string // Bloco de documentação explicativo (disponível sob a constante '__doc__').
	Arquivo string // Caminho físico ou lógico de origem do módulo (disponível sob a constante '__arquivo__').
}

// ModuloImpl define a estrutura de compilação contendo as especificações estáticas do módulo nativo em Go.
type ModuloImpl struct {
	Info       ModuloInfo      // Metadados do módulo.
	Metodos    []*Metodo       // Relação de funções e métodos associados que o módulo expõe.
	Constantes Mapa            // Dicionário de propriedades e chaves estáticas imutáveis.
	Variaveis  Mapa            // Dicionário de chaves mutáveis inicializadas com valores padrão.
	Ast        parser.BaseNode // Árvore Sintática Opcional se o módulo for construído por código do usuário.
}

// GerenciadorModulos é o registro de concorrência que controla as inicializações de módulos nativos (Go).
//
// Utiliza um bloqueador de leitura/escrita RWMutex de forma a garantir segurança concorrente (thread-safe)
// quando múltiplas instâncias da VM tentam importar e carregar pacotes da stdlib simultaneamente.
type GerenciadorModulos struct {
	mu    sync.RWMutex           // Bloqueador concorrente para proteção de escrita.
	Impls map[string]*ModuloImpl // Mapa mapeando nomes às estruturas de compilação estática dos módulos.
}

// RegistraModuloImpl escreve e registra uma especificação de módulo de forma segura na tabela do gerenciador.
func (g *GerenciadorModulos) RegistraModuloImpl(impl *ModuloImpl) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Impls[impl.Info.Nome] = impl
}

// ObtemImplModulo resolve e retorna a especificação Go estática de um módulo a partir de seu nome cadastrado.
func (g *GerenciadorModulos) ObtemImplModulo(nome string) *ModuloImpl {
	g.mu.RLock()
	defer g.mu.RUnlock()
	impl := g.Impls[nome]
	return impl
}

// gerenciador é a instância unificada global do GerenciadorModulos.
var gerenciador = GerenciadorModulos{
	Impls: make(map[string]*ModuloImpl),
}

// RegistraModuloImpl é o atalho público para cadastro de módulos nativos de stdlib.
func RegistraModuloImpl(impl *ModuloImpl) {
	gerenciador.RegistraModuloImpl(impl)
}

// ObtemImplModulo é o atalho público para recuperação rápida de módulos cadastrados.
func ObtemImplModulo(nome string) *ModuloImpl {
	return gerenciador.ObtemImplModulo(nome)
}

// Modulo representa o objeto de instância de escopo dinâmico que a VM utiliza para representar pacotes carregados.
type Modulo struct {
	Impl         *ModuloImpl // Ponteiro para a especificação estática de compilação Go.
	Contexto     *Contexto   // Instância associada do supervisor da VM.
	Escopo       *Escopo     // Tabela local de símbolos e escopo pertencente ao módulo.
	acessoRapido Mapa        // Cache de propriedades resolvidas recentemente para acelerar acessos O(1).
}

// TipoModulo especifica a assinatura e os metadados de Tipo da classe Modulo no Portuscript.
var TipoModulo = NewTipo("Modulo", "Módulos dinâmicos de extensões lógicas e de biblioteca padrão.")

// Tipo retorna a especificação de Tipo da classe Modulo.
func (m *Modulo) Tipo() *Tipo {
	return TipoModulo
}

// M__obtem_attributo__ resolve propriedades, métodos e variáveis exportados pelo módulo em tempo de execução.
//
// Algoritmo de Busca e Cache:
//   - Tenta resolver a chave instantaneamente a partir do mapa cache local 'acessoRapido'. Se encontrar, retorna.
//   - Se for um acesso inédito, busca o valor do símbolo na tabela hash do escopo.
//   - Se localizar, grava no mapa 'acessoRapido' para acelerar futuros acessos a este atributo e o retorna.
func (m *Modulo) M__obtem_attributo__(nome string) (objeto Objeto, err error) {
	ok := false
	if objeto, ok = m.acessoRapido[nome]; ok {
		return
	}

	objeto, err = m.Escopo.ObterValor(nome)
	if err != nil {
		return
	}

	m.acessoRapido[nome] = objeto
	return
}

// Garante conformidade de satisfação do protocolo de acesso dinâmico de atributos.
var _ I__obtem_attributo__ = (*Modulo)(nil)

// TabelaModulos é a estrutura de catálogo do contexto que mantém o cache de instâncias de módulos ativos.
type TabelaModulos struct {
	modulos   map[string]*Modulo // Mapa ligando nomes textuais às instâncias ativas de Modulo.
	Embutidos *Modulo            // Atalho referenciando a instância do módulo especial de funções embutidas globais.
}

// NewTabelaModulos aloca o mapa do cache de registros de módulos.
func NewTabelaModulos() *TabelaModulos {
	return &TabelaModulos{modulos: make(map[string]*Modulo)}
}

// NewModulo instancia um objeto Modulo real em tempo de execução a partir de sua especificação ModuloImpl.
//
// Popula o escopo do módulo com constantes implícitas universais (__nome__, __arquivo__, __doc__),
// injeta as variáveis e constantes declaradas na especificação e clona métodos ligando-os ao escopo local.
func (tabela *TabelaModulos) NewModulo(ctx *Contexto, impl *ModuloImpl) (*Modulo, error) {
	nome := impl.Info.Nome
	modulo := &Modulo{
		Impl:         impl,
		Contexto:     ctx,
		Escopo:       NewEscopo(),
		acessoRapido: NewMapaVazio(),
	}

	if nome == "" {
		nome = "__entrada__"
	}

	modulo.Escopo.DefinirSimbolo(NewConstSimbolo("__nome__", Texto(nome)))
	modulo.Escopo.DefinirSimbolo(NewConstSimbolo("__arquivo__", Texto(impl.Info.Arquivo)))
	modulo.Escopo.DefinirSimbolo(NewConstSimbolo("__doc__", Texto(impl.Info.Doc)))

	// Registra métodos e funções exportados ligando-os dinamicamente ao módulo correspondente
	for _, metodo := range impl.Metodos {
		instMetodo := new(Metodo)
		*instMetodo = *metodo
		instMetodo.Modulo = modulo
		modulo.Escopo.DefinirSimbolo(NewVarSimbolo(metodo.Nome, instMetodo))
	}

	// Popula constantes estáticas
	for nome, valor := range impl.Constantes {
		modulo.Escopo.DefinirSimbolo(NewConstSimbolo(string(nome), valor))
	}

	// Popula variáveis padrão
	for nome, valor := range impl.Variaveis {
		modulo.Escopo.DefinirSimbolo(NewVarSimbolo(string(nome), valor))
	}

	tabela.modulos[nome] = modulo
	if impl.Info.Arquivo != "" {
		tabela.modulos[impl.Info.Arquivo] = modulo
	}
	if nome == "embutidos" {
		tabela.Embutidos = modulo
	}

	return modulo, nil
}

// ObterModulo retorna a instância ativa do módulo cadastrado no cache, disparando erro de importação se não existir.
func (tabela *TabelaModulos) ObterModulo(nome string) (*Modulo, error) {
	m, ok := tabela.modulos[nome]
	if !ok {
		return nil, NewErroF(ImportacaoErro, "O módulo '%v' não pode ser achado", nome)
	}

	return m, nil
}
