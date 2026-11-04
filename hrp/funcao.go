package hrp

import (
	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

// Funcao representa uma sub-rotina ou função declarada no próprio código fonte pelo programador do Harpia (ex: func somar(a, b)).
//
// Ela armazena o corpo de instruções (AST Bloco), a relação ordenada de parâmetros formais esperados,
// o supervisor (Contexto) de execução da VM, e o escopo léxico estático no qual ela foi originalmente declarada.
type Funcao struct {
	Nome         string                           // Nome identificador da função (disponível sob a constante '__nome__').
	Doc          Texto                            // Bloco de documentação explicativo (disponível sob a constante '__doc__').
	args         []string                         // Relação ordenada de nomes textuais correspondentes aos argumentos.
	defaults     map[string]parser.BaseNode       // Valores padrões para os argumentos (AST)
	corpo        *parser.Bloco                    // O bloco de nós da AST correspondente ao corpo físico da função.
	contexto     *Contexto                        // Ponteiro de referência ao supervisor global da VM.
	escopo       *Escopo                          // O escopo de enquadramento original de declaração (Lexical Scope).
	declArquivo  string                           // Snapshot do arquivo onde a função foi declarada (origem do Posicoes).
	declCodigo   string                           // Snapshot do código-fonte do arquivo de declaração.
	declPosicoes map[parser.BaseNode]*lexer.Token // Snapshot do mapa de posições do parser do arquivo de declaração.
	Estatico     bool                             // Flag opcional que sinaliza se é um método estático de classe (ex: 'estatico func').
	tiposParams  map[string]string                // Mapeamento dos tipos estritos exigidos pelos parâmetros formais.
	tipoRetorno  string                           // O tipo de retorno esperado de forma estrita.
	Assincrono   bool                             // Flag que sinaliza se a função é assíncrona.
}

// TipoFuncao especifica as assinaturas e os metadados de classe do tipo Funcao na VM.
var TipoFuncao = NewTipo("Funcao", "Uma funcao Harpia")

// Tipo retorna a representação de classe (Tipo de Funcao).
func (f *Funcao) Tipo() *Tipo {
	return TipoFuncao
}

// NewFuncao é o construtor padrão em Go para alocar novas Funções na VM.
func NewFuncao(nome string, corpo *parser.Bloco, contexto *Contexto, escopo *Escopo) *Funcao {
	f := &Funcao{
		Nome:     nome,
		corpo:    corpo,
		contexto: contexto,
		escopo:   escopo,
		defaults: make(map[string]parser.BaseNode),
	}
	// ponytail: snapshot do arquivo/codigo/posicoes no momento da declaração para que
	// tracebacks de execução apontem para o arquivo de origem da função, não para o
	// módulo que a chamou (corrige #21: false-positive main.hrp:N ao chamar @backend).
	if contexto != nil {
		arq, cod := contexto.ObterArquivoCodigo()
		f.declArquivo = arq
		f.declCodigo = cod
		if prog, ok := contextoUltimoPrograma(contexto); ok {
			f.declPosicoes = prog.Posicoes
		}
	}
	return f
}

// contextoUltimoPrograma busca o *parser.Programa correspondente ao arquivo atual do contexto
// percorrendo o cache de módulos carregados. Usado em NewFuncao para snapshot do Posicoes.
func contextoUltimoPrograma(c *Contexto) (*parser.Programa, bool) {
	if c == nil || c.Modulos == nil {
		return nil, false
	}
	arq, _ := c.ObterArquivoCodigo()
	if arq == "" {
		return nil, false
	}
	mod, err := c.Modulos.ObterModulo(arq)
	if err != nil || mod == nil || mod.Impl == nil || mod.Impl.Ast == nil {
		return nil, false
	}
	prog, ok := mod.Impl.Ast.(*parser.Programa)
	return prog, ok
}

// DefinirEscopo atualiza o escopo de enquadramento original de declaração (Lexical Scope) da função.
func (f *Funcao) DefinirEscopo(e *Escopo) {
	if f == nil {
		return
	}
	f.escopo = e
}

// ObterEscopo retorna o escopo de enquadramento da função.
func (f *Funcao) ObterEscopo() *Escopo {
	if f == nil {
		return nil
	}
	return f.escopo
}

// DefinirContexto atualiza a referência de supervisor global da VM na função.
func (f *Funcao) DefinirContexto(ctx *Contexto) {
	if f == nil {
		return
	}
	f.contexto = ctx
}

// ObterContexto retorna o ponteiro de referência ao supervisor global da VM da função.
func (f *Funcao) ObterContexto() *Contexto {
	if f == nil {
		return nil
	}
	return f.contexto
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

func (f *Funcao) definirTipoParam(nome string, tipo string) {
	if f.tiposParams == nil {
		f.tiposParams = make(map[string]string)
	}
	f.tiposParams[nome] = tipo
}

func (f *Funcao) definirTipoRetorno(tipo string) {
	f.tipoRetorno = tipo
}

func (f *Funcao) DefinirArgs(nomes []string) {
	f.args = nomes
}

func (f *Funcao) SetContexto(ctx *Contexto) {
	f.contexto = ctx
}

func (f *Funcao) SetEscopo(esc *Escopo) {
	f.escopo = esc
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
		if f.tiposParams != nil && f.contexto.Opcs.Estrito {
			if tipoEsperado, existe := f.tiposParams[nome]; existe && tipoEsperado != "" {
				if !ValidarTipo(tipoEsperado, val) {
					return nil, NewErroF(TipagemErro, "O valor passado para o parâmetro '%s' de %s() não coincide com o tipo '%s' (tipo obtido: '%s')", nome, f.Nome, tipoEsperado, val.Tipo().Nome)
				}
			}
		}

		simbolo := NewVarSimbolo(nome, val)
		if f.tiposParams != nil {
			simbolo.Tipo = f.tiposParams[nome]
		}
		escopo.DefinirSimbolo(simbolo)
	}

	newInterpret := &Interpretador{Ast: f.corpo, Contexto: f.contexto, Escopo: escopo}
	// ponytail: usa o snapshot capturado no momento da declaração para ancorar
	// tracebacks no arquivo onde a função foi escrita, não no módulo chamador.
	if f.declArquivo != "" {
		newInterpret.Arquivo = f.declArquivo
		newInterpret.Codigo = f.declCodigo
		if f.declPosicoes != nil {
			newInterpret.Posicoes = f.declPosicoes
		}
	} else if f.contexto != nil {
		ctxArq, ctxCod := f.contexto.ObterArquivoCodigo()
		if ctxArq != "" {
			newInterpret.Arquivo = ctxArq
			newInterpret.Codigo = ctxCod
			if f.contexto.Modulos != nil {
				if mod, err := f.contexto.Modulos.ObterModulo(ctxArq); err == nil && mod != nil && mod.Impl.Ast != nil {
					if prog, ok := mod.Impl.Ast.(*parser.Programa); ok {
						newInterpret.Posicoes = prog.Posicoes
					}
				}
			}
		}
	}
	retorno, err := newInterpret.Inicializa()
	if err == nil && f.tipoRetorno != "" && f.contexto.Opcs.Estrito {
		valRetorno := retorno
		if valRetorno == nil {
			valRetorno = Nulo
		}
		if !ValidarTipo(f.tipoRetorno, valRetorno) {
			return nil, NewErroF(TipagemErro, "O valor retornado por %s() não coincide com o tipo esperado '%s' (tipo retornado: '%s')", f.Nome, f.tipoRetorno, valRetorno.Tipo().Nome)
		}
	}
	return retorno, err
}

// Garante conformidade de satisfação do protocolo chamável.
var _ I__chame__ = (*Funcao)(nil)
