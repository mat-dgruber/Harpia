package parser

import "github.com/mat-dgruber/Harpia/lexer"

// BaseNode define a interface de especificação de base para qualquer nó sintático da AST (Árvore de Sintaxe Abstrata).
type BaseNode interface {
	isExpr()
	Pos() (lexer.PosicaoToken, lexer.PosicaoToken)
}

// NoBase é a estrutura embutida base que provê coordenadas espaciais para nós da AST sem alocação no heap.
type NoBase struct {
	Inicio lexer.PosicaoToken
	Fim    lexer.PosicaoToken
}

// Pos retorna o par de coordenadas (início, fim) do nó sem alocação no heap.
func (n NoBase) Pos() (lexer.PosicaoToken, lexer.PosicaoToken) {
	return n.Inicio, n.Fim
}

// DefinirPos atribui as coordenadas de início e fim do nó.
func (n *NoBase) DefinirPos(inicio, fim lexer.PosicaoToken) {
	n.Inicio = inicio
	n.Fim = fim
}

// Programa representa o nó raiz unificado da AST de um script Harpia completo.
type Programa struct {
	NoBase
	Declaracoes []BaseNode                // Lista sequencial de instruções contidas no arquivo.
	Codigo      string                    // Cópia textual integral do código fonte para fins de diagnósticos.
	Arquivo     string                    // Caminho físico ou identificação lógica do arquivo fonte (ex: "<playground>").
	Posicoes    map[BaseNode]*lexer.Token // Tabela de símbolos associando cada nó ao seu respectivo token para rastrear erros físicos de runtime.
}

// DeclVar representa a declaração de variáveis ('var') ou constantes imutáveis ('const').
type DeclVar struct {
	NoBase
	Constante     bool     // Verdadeiro indica que o nó foi gerado por 'const', bloqueando futuras escritas.
	Nome          string   // Nome (identificador) do símbolo.
	Tipo          string   // Identificador de tipo opcional para tipagem estática (ex: "Inteiro").
	Inicializador BaseNode // Expressão inicial opcional para atribuição (ex: x = 10).
}

// DeclVarDestructuring representa a desestruturação de variáveis (ex: var [a, b] = sinal(0)).
type DeclVarDestructuring struct {
	NoBase
	Constante     bool     // Verdadeiro se for const
	Nomes         []string // Lista de variáveis a serem extraídas
	Inicializador BaseNode // O valor/sinal à direita
}

// ElementoDesestruturacao representa uma propriedade ou item extraído na desestruturação.
type ElementoDesestruturacao struct {
	Nome        string
	ValorPadrao BaseNode
}

// DeclDesestruturacao representa desestruturação de listas ou mapas (ex: var [a, b, ...resto] = lista, const { x, y } = mapa).
type DeclDesestruturacao struct {
	NoBase
	Constante  bool
	TipoEstrut string // "lista" ou "mapa"
	Campos     []ElementoDesestruturacao
	Resto      string
	Valor      BaseNode
}

// Reatribuicao representa comandos de sobrescrita de variáveis comuns ou reatribuições compostas (ex: x += 1).
type Reatribuicao struct {
	NoBase
	Objeto    BaseNode // O destino da escrita (geralmente um nó Identificador ou AcessoMembro).
	Operador  string   // O operador de atribuição ("=", "+=", "-=", "*=", "/=", "//=").
	Expressao BaseNode // A expressão à direita cujo valor será avaliado e atribuído.
}

// OpBinaria representa cálculos aritméticos, bitwise ou lógicos que exigem dois operandos.
type OpBinaria struct {
	NoBase
	Esq      BaseNode // Expressão correspondente ao membro esquerdo.
	Operador string   // O caractere operador correspondente (ex: "+", "-", "==", "e", "ou").
	Dir      BaseNode // Expressão correspondente ao membro direito.
}

// OpUnaria representa operações que agem sobre um único operando (ex: sinal negativo '-x' ou negação 'nao x').
type OpUnaria struct {
	NoBase
	Operador  string   // O caractere operador correspondente (ex: "-", "nao", "~").
	Expressao BaseNode // O operando a ser avaliado.
}

// DeclFuncao representa a declaração formal de funções ('func') ou métodos de classe.
type DeclFuncao struct {
	NoBase
	Nome        string                 // Nome do método/função.
	Parametros  []*DeclFuncaoParametro // Lista ordenada de parâmetros formais aceitos.
	Corpo       *Bloco                 // Escopo de bloco correspondente às instruções da função.
	Estatico    bool                   // Verdadeiro se for um método estático de classe (ex: 'estatico func').
	TipoRetorno string                 // Anotação opcional de tipo de retorno.
	Assincrono  bool                   // Verdadeiro se a função for assíncrona ('assincrono funcao')
}

// AguardeNode representa a expressão 'aguarde <expr>'
type AguardeNode struct {
	NoBase
	Expressao BaseNode
}

// DeclFuncaoParametro representa a especificação individual de um parâmetro formal de função.
type DeclFuncaoParametro struct {
	NoBase
	Nome      string   // Nome do parâmetro.
	Tipo      string   // Anotação opcional de tipo estático.
	Padrao    BaseNode // Valor padrão opcional para o parâmetro (se omitido na chamada).
	Variadico bool     // Verdadeiro se for parâmetro variádico (...)
}

// Bloco representa um escopo físico e lógico encapsulado entre chaves '{}'.
type Bloco struct {
	NoBase
	Declaracoes []BaseNode // Lista sequencial de instruções pertencentes ao escopo interno do bloco.
}

// RetorneNode representa a instrução de saída de controle de fluxo de funções ('retorne').
type RetorneNode struct {
	NoBase
	Expressao BaseNode // Expressão opcional de retorno correspondente ao valor de saída.
}

// ChamadaFuncao representa invocações explícitas de funções, classes ou ponteiros chamáveis.
type ChamadaFuncao struct {
	NoBase
	Identificador BaseNode   // O alvo chamável (geralmente um Identificador ou AcessoMembro).
	Argumentos    []BaseNode // Parâmetros reais fornecidos na chamada.
}

// ExpressaoSe representa ramificações lógicas condicionais baseadas na diretiva 'se/senao'.
type ExpressaoSe struct {
	NoBase
	Condicao    BaseNode // Expressão lógica que determina o caminho condicional.
	Corpo       *Bloco   // Bloco de código principal se a condição for Verdadeira.
	Alternativa BaseNode // Escopo opcional se a condição falhar (outro ExpressaoSe ou Bloco).
}

// Enquanto representa laços condicionais de repetição 'enquanto'.
type Enquanto struct {
	NoBase
	Condicao BaseNode // Expressão booleana de validação a cada loop.
	Corpo    *Bloco   // Bloco físico de instruções executado enquanto a condição for positiva.
}

// TextoLiteral representa strings literais (caracteres cercados por aspas).
type TextoLiteral struct {
	NoBase
	Valor string // A cadeia física textual limpa de aspas limitadoras.
}

// InteiroLiteral representa literais inteiros puramente numéricos (ex: 42).
type InteiroLiteral struct {
	NoBase
	Valor string // Representação textual do valor inteiro.
}

// DecimalLiteral representa números de dízimas ou ponto flutuante (ex: 3.14).
type DecimalLiteral struct {
	NoBase
	Valor string // Representação textual do valor decimal.
}

// ConstanteLiteral representa literais de constantes lógicas globais (como Verdadeiro, Falso, Nulo).
type ConstanteLiteral struct {
	NoBase
	Valor string
}

// Identificador representa variáveis, chaves ou símbolos referenciados nominalmente.
type Identificador struct {
	NoBase
	Nome string // O nome literal do identificador (ex: x).
}

// Anotacao representa blocos de comentários ou comentários documentais específicos no fluxo.
type Anotacao struct {
	NoBase
	Corpo string
}

// AcessoMembro representa expressões de resolução de propriedades ou métodos de objetos (ex: carro.cor).
type AcessoMembro struct {
	NoBase
	Dono   BaseNode // O objeto proprietário do membro.
	Membro BaseNode // A propriedade ou método sendo acessado.
}

// BlocoPara representa laços de repetição iterativos baseados na diretiva 'para-em'.
type BlocoPara struct {
	NoBase
	Identificador string   // Nome da variável de escopo temporário que assume o valor de cada item (ex: x).
	Iterador      BaseNode // O objeto iterável a ser varrido (lista, string ou sequencia).
	Corpo         *Bloco   // Bloco contendo as instruções executadas a cada iteração.
}

// PareNode representa a instrução de quebra imediata de laços estruturados ('pare').
type PareNode struct {
	NoBase
}

// ContinueNode representa a instrução de avanço instantâneo de laços estruturados ('continue').
type ContinueNode struct {
	NoBase
}

// TuplaLiteral representa estruturas de coleções indexadas imutáveis explícitas na sintaxe (ex: (1, 2, 3)).
type TuplaLiteral struct {
	NoBase
	Elementos []BaseNode // Lista ordenada de itens pertencentes à tupla.
}

// ListaLiteral representa coleções indexadas mutáveis ordenadas de dados (ex: [1, 2, 3]).
type ListaLiteral struct {
	NoBase
	Elementos []BaseNode // Lista ordenada de itens pertencentes à lista.
}

// ImporteDe representa importações parciais qualificadas (ex: de "matematica" importe PI, raiz;).
type ImporteDe struct {
	NoBase
	Caminho *TextoLiteral // Identificador textual do nome ou caminho físico do módulo.
	Nomes   []string      // Relação de símbolos de variáveis/funções a serem importados para o escopo.
}

// Indexacao representa operações de acesso indexado a coleções (ex: lista[0]).
type Indexacao struct {
	NoBase
	Objeto    BaseNode // O objeto sendo indexado (lista, mapa ou string).
	Argumento BaseNode // A expressão que resolve no índice numérico ou chave a ser acessada.
}

// MapaPar mapeia chaves e valores estruturais em MapaLiteral.
type MapaPar struct {
	NoBase
	Chave       BaseNode // A chave identificadora (Identificador, Texto ou Expressão em colchetes).
	Valor       BaseNode // O valor correspondente à chave.
	EhImplicito bool     // Verdadeiro se veio de declaração curta implicada (ex: {x, y} mapeando para x: x, y: y).
}

// MapaLiteral representa estruturas de coleções do tipo dicionário ou mapa chave-valor (ex: { "a": 1 }).
type MapaLiteral struct {
	NoBase
	Entradas []MapaPar // Relação de pares de mapeamento chave-valor.
}

// NovaNode representa comandos de instanciação de novas instâncias de classes (ex: nova Pessoa()).
type NovaNode struct {
	NoBase
	Objeto BaseNode // A chamada de construtor correspondente à classe.
}

// AsseguraNode representa asserções e garantias de qualidade de código (ex: assegura x == 10;).
type AsseguraNode struct {
	NoBase
	Condicao BaseNode // Expressão lógica que deve obrigatoriamente retornar Verdadeiro.
	Mensagem BaseNode // Mensagem opcional de erro caso a asserção resulte em Falso.
}

// DeclClasse representa a declaração formal de classes e tipos orientados a objetos.
type DeclClasse struct {
	NoBase
	Nome    string        // Nome da classe declarada.
	Heranca string        // Nome opcional da classe pai da qual herda atributos e comportamentos.
	Metodos []*DeclFuncao // Relação de funções e métodos pertencentes à classe.
}

// OpPipe representa operações de encadeamento fluído de funções baseadas no operador pipe (ex: x |> dobrar()).
type OpPipe struct {
	NoBase
	Esq BaseNode // Membro esquerdo (o valor a ser canalizado).
	Dir BaseNode // Membro direito (a função que receberá o valor esquerdo como primeiro argumento).
}

// ArgumentoNomeado representa um argumento passado pelo nome numa chamada de função (ex: funcao(nome = valor)).
type ArgumentoNomeado struct {
	NoBase
	Nome  string
	Valor BaseNode
}

// DeclTeste representa o bloco de teste unitário integrado ('testar "nome" { ... }').
type DeclTeste struct {
	NoBase
	Nome  string // Descrição ou nome do caso de teste unitário.
	Corpo *Bloco // Escopo contendo as asserções e instruções do teste.
}

// TenteCaptureFinalmente representa a estrutura de tratamento de exceções ('tente ... capture (erro) ... finalmente ...').
type TenteCaptureFinalmente struct {
	NoBase
	TenteBlock      *Bloco // Bloco protegido de execução primária.
	NomeErro        string // Identificador da variável que recebe a exceção capturada no bloco 'capture'.
	CaptureBlock    *Bloco // Bloco de tratamento de erros disparado caso ocorra uma exceção.
	FinalmenteBlock *Bloco // Bloco opcional de limpeza final executado incondicionalmente ao término.
}

// DeclExportar representa a instrução de exportação de símbolos ('exportar <expr>').
type DeclExportar struct {
	NoBase
	Expressao BaseNode // Nó ou declaração a ser exportada para outros módulos.
}

func (*Programa) isExpr()            {}
func (*DeclVar) isExpr()             {}
func (*Reatribuicao) isExpr()        {}
func (*OpBinaria) isExpr()           {}
func (*OpUnaria) isExpr()            {}
func (*DeclFuncao) isExpr()          {}
func (*DeclFuncaoParametro) isExpr() {}
func (*Bloco) isExpr()               {}
func (*RetorneNode) isExpr()         {}
func (*ChamadaFuncao) isExpr()       {}
func (*ExpressaoSe) isExpr()         {}
func (*Enquanto) isExpr()            {}
func (*TextoLiteral) isExpr()        {}
func (*InteiroLiteral) isExpr()      {}
func (*DecimalLiteral) isExpr()      {}
func (*ConstanteLiteral) isExpr()    {}
func (*Identificador) isExpr()       {}
func (*Anotacao) isExpr()            {}
func (*AcessoMembro) isExpr()        {}

type TemplateLiteral struct {
	NoBase
	Partes []BaseNode // Pode ser TextoLiteral ou TemplateExpr
}

type TemplateExpr struct {
	NoBase
	Expressao BaseNode
}

// NoSpread representa o operador spread/rest (...) em listas, mapas ou argumentos.
type NoSpread struct {
	NoBase
	Expressao BaseNode
}

func (*NoSpread) isExpr()               {}
func (*BlocoPara) isExpr()              {}
func (*PareNode) isExpr()               {}
func (*ContinueNode) isExpr()           {}
func (*TuplaLiteral) isExpr()           {}
func (*ListaLiteral) isExpr()           {}
func (*ImporteDe) isExpr()              {}
func (*Indexacao) isExpr()              {}
func (*MapaLiteral) isExpr()            {}
func (*NovaNode) isExpr()               {}
func (*AsseguraNode) isExpr()           {}
func (*DeclClasse) isExpr()             {}
func (*OpPipe) isExpr()                 {}
func (*ArgumentoNomeado) isExpr()       {}
func (*DeclTeste) isExpr()              {}
func (*TenteCaptureFinalmente) isExpr() {}
func (*DeclExportar) isExpr()           {}
func (*TemplateLiteral) isExpr()        {}
func (*TemplateExpr) isExpr()           {}
func (*AguardeNode) isExpr()            {}
func (*DeclVarDestructuring) isExpr()   {}
func (*DeclDesestruturacao) isExpr()    {}
func (*OpCoalescenciaNula) isExpr()     {}
func (*AcessoMembroOpcional) isExpr()   {}
func (*ChamadaOpcional) isExpr()        {}
func (*IndexacaoOpcional) isExpr()      {}
func (*CasoEscolha) isExpr()            {}
func (*DeclEscolha) isExpr()            {}
func (*DeclEnum) isExpr()               {}


func (*DeclInterface) isExpr()          {}

// DeclEnum representa uma enumeração nativa (ex: enum Status { Pendente, Concluido })
type DeclEnum struct {
	NoBase
	Nome    string
	Valores []string
}

// MetodoAssinatura representa a assinatura de um método exigido em uma Interface
type MetodoAssinatura struct {
	NoBase
	Nome        string
	Parametros  []string
	TipoRetorno string
}

// DeclInterface representa um contrato de interface (ex: interface Repositorio { func salvar(item) })
type DeclInterface struct {
	NoBase
	Nome    string
	Metodos []MetodoAssinatura
}

// OpCoalescenciaNula representa o operador de coalescência nula 'a ?? b'.
type OpCoalescenciaNula struct {
	NoBase
	Esq BaseNode
	Dir BaseNode
}

// AcessoMembroOpcional representa o acesso a atributo com navegabilidade segura 'a?.b'.
type AcessoMembroOpcional struct {
	NoBase
	Objeto BaseNode
	Membro BaseNode
}

// ChamadaOpcional representa a chamada de função com navegabilidade segura 'fn?.(a, b)'.
type ChamadaOpcional struct {
	NoBase
	Funcao     BaseNode
	Argumentos []BaseNode
}

// IndexacaoOpcional representa a indexação/acesso por colchete com navegabilidade segura 'lista?.[0]'.
type IndexacaoOpcional struct {
	NoBase
	Objeto    BaseNode
	Argumento BaseNode
}

// CasoEscolha representa uma ramificação individual em uma instrução 'escolha'.
type CasoEscolha struct {
	NoBase
	Padrao BaseNode // Literal, Identificador, DeclDesestruturacao ou wildcard "_"
	Guarda BaseNode // Expressão condicional opcional ("quando <expr>")
	Corpo  BaseNode // Bloco de instruções
}

// DeclEscolha representa a instrução de pattern matching 'escolha <expr> { caso ... }'.
type DeclEscolha struct {
	NoBase
	Expressao BaseNode
	Casos     []CasoEscolha
}



// AtributoJSX representa um atributo em uma tag JSX (ex: classe="p-4" ou aoClicar={clique})
type AtributoJSX struct {
	NoBase
	Nome  string
	Valor BaseNode // Pode ser TextoLiteral, Identificador ou qualquer expressão embrulhada
}

// NoJSX representa uma tag JSX (ex: <div classe="app">...</div> ou <botao />)
type NoJSX struct {
	NoBase
	Tag         string
	Atributos   []*AtributoJSX
	Filhos      []BaseNode // Pode ser NoJSX, NoSeJSX, NoParaJSX, TextoLiteral ou Expressao
	AutoFechado bool
}

// NoSeJSX representa a diretiva condicional no JSX: <se condicao={...}>...</se>
type NoSeJSX struct {
	NoBase
	Condicao BaseNode
	Filhos   []BaseNode
}

// NoParaJSX representa a diretiva de loop no JSX: <para item em lista={...}>...</para>
type NoParaJSX struct {
	NoBase
	Item   string
	Lista  BaseNode
	Filhos []BaseNode
}

// DeclEstilo representa uma declaração de estilo: estilo MeuComponente { cor: "azul"; }
type DeclEstilo struct {
	NoBase
	Nome   string
	Regras string // Guardamos as regras internas como texto simples para compilar para CSS
}

func (*AtributoJSX) isExpr() {}
func (*NoJSX) isExpr()       {}
func (*NoSeJSX) isExpr()     {}
func (*NoParaJSX) isExpr()   {}
func (*DeclEstilo) isExpr()  {}

// OpTernaria representa o operador condicional ternário (condicao ? entao : senao).
type OpTernaria struct {
	NoBase
	Condicao BaseNode
	Entao    BaseNode
	Senao    BaseNode
}

func (*OpTernaria) isExpr() {}

// Decorador representa a sintaxe de decoradores (ex: @depreciado ou @rota("/api")) sobre funções ou classes.
type Decorador struct {
	NoBase
	Nome       string     // Nome do decorador (ex: "depreciado", "rota")
	Argumentos []BaseNode // Argumentos opcionais entre parênteses (ex: "/api")
	Alvo       BaseNode   // A declaração decorada (ex: *DeclFuncao ou *DeclClasse)
}

func (*Decorador) isExpr() {}
