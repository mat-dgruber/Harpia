package ptst

import (
	"github.com/natanfeitosa/portuscript/parser"
)

// Funcao representa uma sub-rotina ou função declarada no próprio código fonte pelo programador do Portuscript (ex: func somar(a, b)).
//
// Ela armazena o corpo de instruções (AST Bloco), a relação ordenada de parâmetros formais esperados,
// o supervisor (Contexto) de execução da VM, e o escopo léxico estático no qual ela foi originalmente declarada.
type Funcao struct {
	Nome     string                     // Nome identificador da função (disponível sob a constante '__nome__').
	Doc      Texto                      // Bloco de documentação explicativo (disponível sob a constante '__doc__').
	args     []string                   // Relação ordenada de nomes textuais correspondentes aos argumentos.
	defaults map[string]parser.BaseNode // Valores padrões para os argumentos (AST)
	corpo    *parser.Bloco              // O bloco de nós da AST correspondente ao corpo físico da função.
	contexto *Contexto                  // Ponteiro de referência ao supervisor global da VM.
	escopo   *Escopo                    // O escopo de enquadramento original de declaração (Lexical Scope).
	Estatico bool                       // Flag opcional que sinaliza se é um método estático de classe (ex: 'estatico func').
}

// TipoFuncao especifica as assinaturas e os metadados de classe do tipo Funcao na VM.
var TipoFuncao = NewTipo("Funcao", "Uma funcao Portuscript")

// Tipo retorna a representação de classe (Tipo de Funcao).
func (f *Funcao) Tipo() *Tipo {
	return TipoFuncao
}

// NewFuncao é o construtor padrão em Go para alocar novas Funções na VM.
func NewFuncao(nome string, corpo *parser.Bloco, contexto *Contexto, escopo *Escopo) *Funcao {
	return &Funcao{
		Nome:     nome,
		corpo:    corpo,
		contexto: contexto,
		escopo:   escopo,
		defaults: make(map[string]parser.BaseNode),
	}
}

// Args retorna a relação ordenada de parâmetros formais da função.
func (f *Funcao) Args() []string { return f.args }

// definirArgs configura e registra a lista de parâmetros formais aceitos.
func (f *Funcao) definirArgs(nomes []string) { f.args = nomes }

func (f *Funcao) definirDefault(nome string, node parser.BaseNode) {
	if f.defaults == nil {
		f.defaults = make(map[string]parser.BaseNode)
	}
	f.defaults[nome] = node
}

// M__chame__ satisfaz o protocolo de chamabilidade da VM (I__chame__), realizando a execução física da função.
//
// Algoritmo de Execução:
//  1. Valida se o número de argumentos reais fornecidos coincide com os parâmetros formais esperados;
//  2. Instancia um novo escopo de variáveis enlaçado ao escopo léxico estático da função ('f.escopo.NewEscopo()');
//  3. Associa de forma pareada os nomes dos parâmetros locais aos valores recebidos na chamada;
//  4. Instancia e aciona o Interpretador sobre o bloco da AST correspondente ao corpo da função;
//  5. Limpa os recursos locais e retorna o Objeto resultante da avaliação ou erros de tempo de execução.
func (f *Funcao) M__chame__(args Tupla) (Objeto, error) {
	escopo := f.escopo.NewEscopo()
	resolvidos := make(map[string]Objeto)

	// 1. Processa argumentos posicionais da chamada
	indicePosicional := 0
	for _, arg := range args {
		if nomeado, ok := arg.(*ArgumentoNomeadoObj); ok {
			// Argumento nomeado
			resolvidos[nomeado.Nome] = nomeado.Valor
		} else {
			// Argumento posicional
			if indicePosicional >= len(f.args) {
				return nil, NewErroF(TipagemErro, "%v() esperava no máximo %v argumentos posicionais, mas recebeu a mais", f.Nome, len(f.args))
			}
			nomeParam := f.args[indicePosicional]
			resolvidos[nomeParam] = arg
			indicePosicional++
		}
	}

	// 2. Preenche parâmetros ausentes com valores padrão ou falha se obrigatório
	interpretadorTemp := &Interpretador{
		Contexto: f.contexto,
		Escopo:   f.escopo, // Avalia expressões padrão no escopo de definição
	}

	for _, nomeParam := range f.args {
		if _, existe := resolvidos[nomeParam]; !existe {
			if noPadrao, temDefault := f.defaults[nomeParam]; temDefault && noPadrao != nil {
				valPadrao, err := interpretadorTemp.visite(noPadrao)
				if err != nil {
					return nil, err
				}
				resolvidos[nomeParam] = valPadrao
			} else {
				return nil, NewErroF(TipagemErro, "%v() requer o argumento '%v' que não foi fornecido", f.Nome, nomeParam)
			}
		}
	}

	// 3. Define no escopo local da função todos os parâmetros resolvidos
	for nome, val := range resolvidos {
		escopo.DefinirSimbolo(NewVarSimbolo(nome, val))
	}

	return (&Interpretador{Ast: f.corpo, Contexto: f.contexto, Escopo: escopo}).Inicializa()
}

// Garante conformidade de satisfação do protocolo chamável.
var _ I__chame__ = (*Funcao)(nil)
