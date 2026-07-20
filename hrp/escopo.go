package hrp

import (
	"sync"
)

// Simbolo representa uma entrada individual registrada no escopo de variáveis da VM.
type Simbolo struct {
	Nome      string       // O nome textual identificador da variável ou constante.
	Valor     Objeto       // O ponteiro para a instância de Objeto associada.
	Constante bool         // Verdadeiro indica que o símbolo é imutável (constante), bloqueando redefinições.
	Tipo      string       // Tipo opcional para verificação de tipagem dinâmica estrita.
	mu        sync.RWMutex // Mutex de grão fino para proteger o valor individual do símbolo em runtime concorrente.
}

func (s *Simbolo) ObterValor() Objeto {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Valor
}

func (s *Simbolo) DefinirValor(v Objeto) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Valor = v
}

// Escopo gerencia a tabela hash de símbolos e variáveis em tempo de execução.
//
// Os escopos são encadeados de forma léxica (lexical scoping) através do ponteiro opcional 'Pai'.
// Isso permite que estruturas internas (como corpos de funções ou laços) acessem variáveis globais,
// mantendo suas próprias variáveis locais isoladas de colisões externas.
type Escopo struct {
	Simbolos map[string]*Simbolo // Mapa unificador associando nomes textuais aos símbolos.
	Pai      *Escopo             // Ponteiro para o escopo hierárquico imediatamente superior (pai).
	mu       sync.RWMutex        // Mutex para proteção de leitura/escrita concorrente em mapas de símbolos.
}

// NewEscopo aloca e retorna uma nova tabela hash de escopo isolada de nível superior.
func NewEscopo() *Escopo {
	return &Escopo{}
}

// NewEscopo instancia um novo escopo filho enlaçado ao escopo atual (Pai = e).
func (e *Escopo) NewEscopo() *Escopo {
	return &Escopo{Pai: e}
}

// Len devolve o número de símbolos definidos localmente no escopo atual.
func (e *Escopo) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.Simbolos)
}

// DefinirSimbolo registra ou substitui um símbolo na tabela hash local.
//
// Regra de Negócio:
// Se um símbolo idêntico já existir e for uma constante imutável, o registro é abortado
// e um erro de tipagem estruturada (TipagemErro) é lançado.
func (e *Escopo) DefinirSimbolo(simbolo *Simbolo) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Simbolos == nil {
		e.Simbolos = make(map[string]*Simbolo)
	}
	exists, ok := e.Simbolos[simbolo.Nome]
	if ok && exists != nil {
		if exists.Constante {
			return NewErroF(TipagemErro, "A constante '%s' já existe", simbolo.Nome)
		}
	}

	e.Simbolos[simbolo.Nome] = simbolo
	return nil
}

// ObterSimbolosSeguro retorna uma cópia rasa e segura de todos os símbolos definidos no escopo atual.
func (e *Escopo) ObterSimbolosSeguro() []*Simbolo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	simbs := make([]*Simbolo, 0, len(e.Simbolos))
	for _, simb := range e.Simbolos {
		if simb != nil {
			simbs = append(simbs, simb)
		}
	}
	return simbs
}

// RedefinirValor executa a reatribuição de valores de variáveis em tempo de execução (ex: x = 10).
//
// Regras e Resolução Recursiva:
//   - Tenta localizar o símbolo na tabela local. Se o encontrar e este for mutável, atualiza o seu valor.
//   - Se não encontrar localmente, mas o escopo possuir um escopo 'Pai', delega recursivamente a redefinição
//     subindo na hierarquia até localizar o escopo onde o símbolo foi originalmente definido.
//   - Lança erro se tentar reatribuir constantes imutáveis ou se o símbolo não existir na hierarquia.
func (e *Escopo) RedefinirValor(nome string, valor Objeto) error {
	e.mu.Lock()
	simbolo, ok := e.Simbolos[nome]
	e.mu.Unlock()

	if !ok {
		if e.Pai != nil {
			return e.Pai.RedefinirValor(nome, valor)
		}

		return NewErroF(TipagemErro, "Você não pode reatribuir valor a '%s', pois a variável não existe", nome)
	}

	if simbolo.Constante {
		return NewErroF(TipagemErro, "Você não pode reatribuir valor a '%s', pois é uma constante", nome)
	}

	simbolo.DefinirValor(valor)
	return nil
}

// ObterValor resolve e retorna o objeto associado ao nome informado de forma léxica.
//
// Realiza a busca incremental: se o símbolo não constar localmente, sobe recursivamente as referências de
// escopos pais. Se atingir o escopo raiz primordial sem sucesso, retorna um erro estruturado NomeErro.
func (e *Escopo) ObterValor(nome string) (Objeto, error) {
	e.mu.RLock()
	simbolo, ok := e.Simbolos[nome]
	e.mu.RUnlock()

	if !ok {
		if e.Pai != nil {
			return e.Pai.ObterValor(nome)
		}

		return nil, NewErroF(NomeErro, "'%s' não foi encontrado no escopo atual", nome)
	}

	return simbolo.ObterValor(), nil
}

// ObterSimbolo busca e retorna a struct Simbolo completa de forma léxica recursiva.
func (e *Escopo) ObterSimbolo(nome string) (*Simbolo, error) {
	e.mu.RLock()
	simbolo, ok := e.Simbolos[nome]
	e.mu.RUnlock()

	if !ok {
		if e.Pai != nil {
			return e.Pai.ObterSimbolo(nome)
		}

		return nil, NewErroF(NomeErro, "'%s' não foi encontrado no escopo atual", nome)
	}

	return simbolo, nil
}

// ExcluirSimbolo remove fisicamente um símbolo da tabela local caso exista. Lança erro se o nome não for achado.
func (e *Escopo) ExcluirSimbolo(nome string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.Simbolos[nome]; !ok {
		return NewErroF(NomeErro, "'%s' não foi encontrado no escopo atual", nome)
	}

	delete(e.Simbolos, nome)
	return nil
}

// NewVarSimbolo é o construtor padrão para alocar variáveis mutáveis.
func NewVarSimbolo(nome string, valor Objeto) *Simbolo {
	return &Simbolo{Nome: nome, Valor: valor, Constante: false}
}

// NewConstSimbolo é o construtor padrão para alocar constantes imutáveis.
func NewConstSimbolo(nome string, valor Objeto) *Simbolo {
	return &Simbolo{Nome: nome, Valor: valor, Constante: true}
}
