package ptst

// Simbolo representa uma entrada individual registrada no escopo de variáveis da VM.
type Simbolo struct {
	Nome      string // O nome textual identificador da variável ou constante.
	Valor     Objeto // O ponteiro para a instância de Objeto associada.
	Constante bool   // Verdadeiro indica que o símbolo é imutável (constante), bloqueando redefinições.
}

// Escopo gerencia a tabela hash de símbolos e variáveis em tempo de execução.
//
// Os escopos são encadeados de forma léxica (lexical scoping) através do ponteiro opcional 'Pai'.
// Isso permite que estruturas internas (como corpos de funções ou laços) acessem variáveis globais,
// mantendo suas próprias variáveis locais isoladas de colisões externas.
type Escopo struct {
	Simbolos map[string]*Simbolo // Mapa unificador associando nomes textuais aos símbolos.
	Pai      *Escopo             // Ponteiro para o escopo hierárquico imediatamente superior (pai).
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
	return len(e.Simbolos)
}

// DefinirSimbolo registra ou substitui um símbolo na tabela hash local.
//
// Regra de Negócio:
// Se um símbolo idêntico já existir e for uma constante imutável, o registro é abortado
// e um erro de tipagem estruturada (TipagemErro) é lançado.
func (e *Escopo) DefinirSimbolo(simbolo *Simbolo) error {
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

// RedefinirValor executa a reatribuição de valores de variáveis em tempo de execução (ex: x = 10).
//
// Regras e Resolução Recursiva:
//   - Tenta localizar o símbolo na tabela local. Se o encontrar e este for mutável, atualiza o seu valor.
//   - Se não encontrar localmente, mas o escopo possuir um escopo 'Pai', delega recursivamente a redefinição
//     subindo na hierarquia até localizar o escopo onde o símbolo foi originalmente definido.
//   - Lança erro se tentar reatribuir constantes imutáveis ou se o símbolo não existir na hierarquia.
func (e *Escopo) RedefinirValor(nome string, valor Objeto) error {
	simbolo, ok := e.Simbolos[nome]

	if !ok {
		if e.Pai != nil {
			return e.Pai.RedefinirValor(nome, valor)
		}

		return NewErroF(TipagemErro, "Você não pode reatribuir valor a '%s', pois a variável não existe", nome)
	}

	if simbolo.Constante {
		return NewErroF(TipagemErro, "Você não pode reatribuir valor a '%s', pois é uma constante", nome)
	}

	simbolo.Valor = valor
	return nil
}

// ObterValor resolve e retorna o objeto associado ao nome informado de forma léxica.
//
// Realiza a busca incremental: se o símbolo não constar localmente, sobe recursivamente as referências de
// escopos pais. Se atingir o escopo raiz primordial sem sucesso, retorna um erro estruturado NomeErro.
func (e *Escopo) ObterValor(nome string) (Objeto, error) {
	simbolo, ok := e.Simbolos[nome]

	if !ok {
		if e.Pai != nil {
			return e.Pai.ObterValor(nome)
		}

		return nil, NewErroF(NomeErro, "'%s' não foi encontrado/a, talvez você não tenha definido ainda", nome)
	}

	return simbolo.Valor, nil
}

// ExcluirSimbolo remove fisicamente um símbolo da tabela local caso exista. Lança erro se o nome não for achado.
func (e *Escopo) ExcluirSimbolo(nome string) error {
	if _, ok := e.Simbolos[nome]; !ok {
		return NewErroF(NomeErro, "'%s' não foi encontrado/a, talvez você não tenha definido ainda", nome)
	}

	delete(e.Simbolos, nome)
	return nil
}

// NewVarSimbolo é o construtor padrão para alocar variáveis mutáveis.
func NewVarSimbolo(nome string, valor Objeto) *Simbolo {
	return &Simbolo{nome, valor, false}
}

// NewConstSimbolo é o construtor padrão para alocar constantes imutáveis.
func NewConstSimbolo(nome string, valor Objeto) *Simbolo {
	return &Simbolo{nome, valor, true}
}
