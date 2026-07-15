package parser

import "github.com/natanfeitosa/portuscript/lexer"

// BaseNode define a interface de especificação de base para qualquer nó sintático da AST (Árvore de Sintaxe Abstrata).
//
// O método isExpr() serve apenas como uma assinatura de marcação para que o compilador Go verifique
// o polimorfismo estrutural, limitando que apenas structs autorizados implementem a interface.
type BaseNode interface{ isExpr() }

// Programa representa o nó raiz unificado da AST de um script Portuscript completo.
type Programa struct {
	Declaracoes []BaseNode                // Lista sequencial de instruções contidas no arquivo.
	Codigo      string                    // Cópia textual integral do código fonte para fins de diagnósticos.
	Arquivo     string                    // Caminho físico ou identificação lógica do arquivo fonte (ex: "<playground>").
	Posicoes    map[BaseNode]*lexer.Token // Tabela de símbolos associando cada nó ao seu respectivo token para rastrear erros físicos de runtime.
}

// DeclVar representa a declaração de variáveis ('var') ou constantes imutáveis ('const').
type DeclVar struct {
	Constante     bool     // Verdadeiro indica que o nó foi gerado por 'const', bloqueando futuras escritas.
	Nome          string   // Nome (identificador) do símbolo.
	Tipo          string   // Identificador de tipo opcional para tipagem estática (ex: "Inteiro").
	Inicializador BaseNode // Expressão inicial opcional para atribuição (ex: x = 10).
}

// Reatribuicao representa comandos de sobrescrita de variáveis comuns ou reatribuições compostas (ex: x += 1).
type Reatribuicao struct {
	Objeto    BaseNode // O destino da escrita (geralmente um nó Identificador ou AcessoMembro).
	Operador  string   // O operador de atribuição ("=", "+=", "-=", "*=", "/=", "//=").
	Expressao BaseNode // A expressão à direita cujo valor será avaliado e atribuído.
}

// OpBinaria representa cálculos aritméticos, bitwise ou lógicos que exigem dois operandos.
type OpBinaria struct {
	Esq      BaseNode // Expressão correspondente ao membro esquerdo.
	Operador string   // O caractere operador correspondente (ex: "+", "-", "==", "e", "ou").
	Dir      BaseNode // Expressão correspondente ao membro direito.
}

// OpUnaria representa operações que agem sobre um único operando (ex: sinal negativo '-x' ou negação 'nao x').
type OpUnaria struct {
	Operador  string   // O caractere operador correspondente (ex: "-", "nao", "~").
	Expressao BaseNode // O operando a ser avaliado.
}

// DeclFuncao representa a declaração formal de funções ('func') ou métodos de classe.
type DeclFuncao struct {
	Nome        string                 // Nome do método/função.
	Parametros  []*DeclFuncaoParametro // Lista ordenada de parâmetros formais aceitos.
	Corpo       *Bloco                 // Escopo de bloco correspondente às instruções da função.
	Estatico    bool                   // Verdadeiro se for um método estático de classe (ex: 'estatico func').
	TipoRetorno string                 // Anotação opcional de tipo de retorno.
	Assincrono  bool                   // Verdadeiro se a função for assíncrona ('assincrono funcao')
}

// AguardeNode representa a expressão 'aguarde <expr>'
type AguardeNode struct {
	Expressao BaseNode
}

// DeclFuncaoParametro representa a especificação individual de um parâmetro formal de função.
type DeclFuncaoParametro struct {
	Nome   string   // Nome do parâmetro.
	Tipo   string   // Anotação opcional de tipo estático.
	Padrao BaseNode // Valor padrão opcional para o parâmetro (se omitido na chamada).
}

// Bloco representa um escopo físico e lógico encapsulado entre chaves '{}'.
type Bloco struct {
	Declaracoes []BaseNode // Lista sequencial de instruções pertencentes ao escopo interno do bloco.
}

// RetorneNode representa a instrução de saída de controle de fluxo de funções ('retorne').
type RetorneNode struct {
	Expressao BaseNode // Expressão opcional de retorno correspondente ao valor de saída.
}

// ChamadaFuncao representa invocações explícitas de funções, classes ou ponteiros chamáveis.
type ChamadaFuncao struct {
	Identificador BaseNode   // O alvo chamável (geralmente um Identificador ou AcessoMembro).
	Argumentos    []BaseNode // Parâmetros reais fornecidos na chamada.
}

// ExpressaoSe representa ramificações lógicas condicionais baseadas na diretiva 'se/senao'.
type ExpressaoSe struct {
	Condicao    BaseNode // Expressão lógica que determina o caminho condicional.
	Corpo       *Bloco   // Bloco de código principal se a condição for Verdadeira.
	Alternativa BaseNode // Escopo opcional se a condição falhar (outro ExpressaoSe ou Bloco).
}

// Enquanto representa laços condicionais de repetição 'enquanto'.
type Enquanto struct {
	Condicao BaseNode // Expressão booleana de validação a cada loop.
	Corpo    *Bloco   // Bloco físico de instruções executado enquanto a condição for positiva.
}

// TextoLiteral representa strings literais (caracteres cercados por aspas).
type TextoLiteral struct {
	Valor string // A cadeia física textual limpa de aspas limitadoras.
}

// InteiroLiteral representa literais inteiros puramente numéricos (ex: 42).
type InteiroLiteral struct {
	Valor string // Representação textual do valor inteiro.
}

// DecimalLiteral representa números de dízimas ou ponto flutuante (ex: 3.14).
type DecimalLiteral struct {
	Valor string // Representação textual do valor decimal.
}

// ConstanteLiteral representa literais de constantes lógicas globais (como Verdadeiro, Falso, Nulo).
type ConstanteLiteral struct {
	Valor string
}

// Identificador representa variáveis, chaves ou símbolos referenciados nominalmente.
type Identificador struct {
	Nome string // O nome literal do identificador (ex: x).
}

// Anotacao representa blocos de comentários ou comentários documentais específicos no fluxo.
type Anotacao struct {
	Corpo string
}

// AcessoMembro representa expressões de resolução de propriedades ou métodos de objetos (ex: carro.cor).
type AcessoMembro struct {
	Dono   BaseNode // O objeto proprietário do membro.
	Membro BaseNode // A propriedade ou método sendo acessado.
}

// BlocoPara representa laços de repetição iterativos baseados na diretiva 'para-em'.
type BlocoPara struct {
	Identificador string   // Nome da variável de escopo temporário que assume o valor de cada item (ex: x).
	Iterador      BaseNode // O objeto iterável a ser varrido (lista, string ou sequencia).
	Corpo         *Bloco   // Bloco contendo as instruções executadas a cada iteração.
}

// PareNode representa a instrução de quebra imediata de laços estruturados ('pare').
type PareNode struct{}

// ContinueNode representa a instrução de avanço instantâneo de laços estruturados ('continue').
type ContinueNode struct{}

// TuplaLiteral representa estruturas de coleções indexadas imutáveis explícitas na sintaxe (ex: (1, 2, 3)).
type TuplaLiteral struct {
	Elementos []BaseNode // Lista ordenada de itens pertencentes à tupla.
}

// ListaLiteral representa coleções indexadas mutáveis ordenadas de dados (ex: [1, 2, 3]).
type ListaLiteral struct {
	Elementos []BaseNode // Lista ordenada de itens pertencentes à lista.
}

// ImporteDe representa importações parciais qualificadas (ex: de "matematica" importe PI, raiz;).
type ImporteDe struct {
	Caminho *TextoLiteral // Identificador textual do nome ou caminho físico do módulo.
	Nomes   []string      // Relação de símbolos de variáveis/funções a serem importados para o escopo.
}

// Indexacao representa operações de acesso indexado a coleções (ex: lista[0]).
type Indexacao struct {
	Objeto    BaseNode // O objeto sendo indexado (lista, mapa ou string).
	Argumento BaseNode // A expressão que resolve no índice numérico ou chave a ser acessada.
}

// MapaPar mapeia chaves e valores estruturais em MapaLiteral.
type MapaPar struct {
	Chave       BaseNode // A chave identificadora (Identificador, Texto ou Expressão em colchetes).
	Valor       BaseNode // O valor correspondente à chave.
	EhImplicito bool     // Verdadeiro se veio de declaração curta implicada (ex: {x, y} mapeando para x: x, y: y).
}

// MapaLiteral representa estruturas de coleções do tipo dicionário ou mapa chave-valor (ex: { "a": 1 }).
type MapaLiteral struct {
	Entradas []MapaPar // Relação de pares de mapeamento chave-valor.
}

// NovaNode representa comandos de instanciação de novas instâncias de classes (ex: nova Pessoa()).
type NovaNode struct {
	Objeto BaseNode // A chamada de construtor correspondente à classe.
}

// AsseguraNode representa asserções e garantias de qualidade de código (ex: assegura x == 10;).
type AsseguraNode struct {
	Condicao BaseNode // Expressão lógica que deve obrigatoriamente retornar Verdadeiro.
	Mensagem BaseNode // Mensagem opcional de erro caso a asserção resulte em Falso.
}

// DeclClasse representa a declaração formal de classes e tipos orientados a objetos.
type DeclClasse struct {
	Nome    string        // Nome da classe declarada.
	Heranca string        // Nome opcional da classe pai da qual herda atributos e comportamentos.
	Metodos []*DeclFuncao // Relação de funções e métodos pertencentes à classe.
}

// OpPipe representa operações de encadeamento fluído de funções baseadas no operador pipe (ex: x |> dobrar()).
type OpPipe struct {
	Esq BaseNode // Membro esquerdo (o valor a ser canalizado).
	Dir BaseNode // Membro direito (a função que receberá o valor esquerdo como primeiro argumento).
}

// ArgumentoNomeado representa um argumento passado pelo nome numa chamada de função (ex: funcao(nome = valor)).
type ArgumentoNomeado struct {
	Nome  string
	Valor BaseNode
}

type DeclTeste struct {
	Nome  string
	Corpo *Bloco
}

type TenteCaptureFinalmente struct {
	TenteBlock      *Bloco
	NomeErro        string
	CaptureBlock    *Bloco
	FinalmenteBlock *Bloco
}

type DeclExportar struct {
	Expressao BaseNode
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
	Partes []BaseNode // Pode ser TextoLiteral ou TemplateExpr
}

type TemplateExpr struct {
	Expressao BaseNode
}

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

// AtributoJSX representa um atributo em uma tag JSX (ex: classe="p-4" ou aoClicar={clique})
type AtributoJSX struct {
	Nome  string
	Valor BaseNode // Pode ser TextoLiteral, Identificador ou qualquer expressão embrulhada
}

// NoJSX representa uma tag JSX (ex: <div classe="app">...</div> ou <botao />)
type NoJSX struct {
	Tag         string
	Atributos   []*AtributoJSX
	Filhos      []BaseNode // Pode ser NoJSX, NoSeJSX, NoParaJSX, TextoLiteral ou Expressao
	AutoFechado bool
}

// NoSeJSX representa a diretiva condicional no JSX: <se condicao={...}>...</se>
type NoSeJSX struct {
	Condicao BaseNode
	Filhos   []BaseNode
}

// NoParaJSX representa a diretiva de loop no JSX: <para item em lista={...}>...</para>
type NoParaJSX struct {
	Item   string
	Lista  BaseNode
	Filhos []BaseNode
}

// DeclEstilo representa uma declaração de estilo: estilo MeuComponente { cor: "azul"; }
type DeclEstilo struct {
	Nome   string
	Regras string // Guardamos as regras internas como texto simples para compilar para CSS
}

func (*AtributoJSX) isExpr() {}
func (*NoJSX) isExpr()       {}
func (*NoSeJSX) isExpr()     {}
func (*NoParaJSX) isExpr()   {}
func (*DeclEstilo) isExpr()  {}

