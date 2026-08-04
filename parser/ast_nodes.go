package parser

import "github.com/mat-dgruber/Harpia/lexer"

// # Árvore de Sintaxe Abstrata (AST) do Harpia
//
// Este arquivo define todos os nós da AST gerados pelo Parser. Cada nó é uma
// struct Go que embute NoBase (para rastreamento de coordenadas) e implementa
// o método marcador BaseNode.isExpr().
//
// A AST é usada por três consumidores principais:
//
//  1. A máquina virtual (VM) — para interpretar ou compilar bytecode.
//  2. O resolvedor de módulos — para processar exportar/importar.
//  3. Ferramentas de diagnóstico — Ast2string serializa a árvore em JSON.
//
// # Nota de Segurança
//
// Este arquivo é parte do compilador Harpia. Ele apenas define estruturas de
// dados; não executa lógica de runtime. Não é código malicioso.

// BaseNode é a interface marcadora de toda a Árvore de Sintaxe Abstrata (AST)
//
// Qualquer struct que represente um nó sintático válido deve implementar a
// função interna vazia `isExpr()`. Isso permite ao compilador Go validar em
// tempo de compilação que apenas nós válidos componham a árvore hierárquica,
// mantendo o Parser fortemente tipado e seguro.
//
// O método Pos retorna o par de coordenadas físicas (início, fim) do nó
// dentro do arquivo fonte, fundamental para reportar erros de runtime com
// precisão de linha/coluna pela máquina virtual.
type BaseNode interface {
	isExpr()
	Pos() (lexer.PosicaoToken, lexer.PosicaoToken)
}

// NoBase é o _embedding_ compartilhado que provê coordenadas espaciais para
// nós da AST sem alocação extra no heap.
//
// Cada nó sintático embute NoBase como primeiro campo (em geral), garantindo
// acesso direto às posições de início e fim sem necessidade de polimorfismo
// dinâmico na recuperação de posições.
//
// # Performance
//
// Usar um _embedding_ em vez de uma composição via interface evita uma
// indireção de ponteiro por chamada a Pos(), o que é relevante em benchmarks
// onde o parser gera milhares de nós por segundo.
type NoBase struct {
	Inicio lexer.PosicaoToken // Coordenada do primeiro token que originou este nó.
	Fim    lexer.PosicaoToken // Coordenada do último token consumido pelo nó.
}

// Pos retorna o par de coordenadas (início, fim) do nó sem alocação no heap.
//
// Implementação obrigatória da interface BaseNode. É invocada pela VM e pelo
// sistema de diagnósticos sempre que um erro precisa ser reportado com
// precisão gráfica.
func (n NoBase) Pos() (lexer.PosicaoToken, lexer.PosicaoToken) {
	return n.Inicio, n.Fim
}

// DefinirPos atribui as coordenadas de início e fim do nó de forma explícita.
//
// Usado por Parser.registrar quando o token inicial já é conhecido mas a
// posição final ainda depende de sub-árvores. Permite reter a imutabilidade
// de Pos no escopo externo enquanto a AST é construída.
func (n *NoBase) DefinirPos(inicio, fim lexer.PosicaoToken) {
	n.Inicio = inicio
	n.Fim = fim
}

// Programa representa o nó raiz unificado da AST de um script Harpia completo.
//
// É a estrutura retornada por Parser.Parse e carrega tudo o que a VM precisa
// para executar o script: as declarações, o código fonte (para mensagens de
// erro), o caminho lógico do arquivo e o mapa de posições físicas.
type Programa struct {
	NoBase
	Declaracoes []BaseNode                // Lista sequencial de instruções contidas no arquivo, em ordem de aparição.
	Codigo      string                    // Cópia textual integral do código fonte para fins de diagnósticos.
	Arquivo     string                    // Caminho físico ou identificação lógica do arquivo fonte (ex: "<playground>").
	Posicoes    map[BaseNode]*lexer.Token // Tabela de símbolos associando cada nó ao seu respectivo token para rastrear erros físicos de runtime.
}

// DeclVar representa a declaração de variáveis ('var') ou constantes imutáveis ('const').
//
// Sintaxe suportada:
//
//	var nome        = expressao
//	const NOME: Tipo = expressao
//
// # Validação Semântica
//
// A aplicação de tipos é opt-in (Sprint 6): o campo Tipo é apenas capturado na
// AST, mas não é validado em runtime. Constantes (Constante=true) bloqueiam
// reatribuição — semântica essa já aplicada pela VM.
type DeclVar struct {
	NoBase
	Constante     bool     // Verdadeiro indica que o nó foi gerado por 'const', bloqueando futuras escritas.
	Nome          string   // Nome (identificador) do símbolo.
	Tipo          string   // Identificador de tipo opcional para tipagem estática (ex: "Inteiro"). Vazio se não anotado.
	Inicializador BaseNode // Expressão inicial opcional para atribuição (ex: x = 10). Constantes exigem inicializador.
	DocComments   string   // Comentários de documentação associados (/// ou /** */).
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

// DeclFuncao representa a declaração formal de funções ('funcao') ou métodos
// de classe.
//
// Sintaxe suportada:
//
//	funcao nome(p1, p2: Tipo = padrao): TipoRetorno { corpo }
//
//	estatico funcao metodo() { ... }           // Método estático de classe
//	assincrono funcao buscar() { ... }          // Função assíncrona (Promise)
//
// O corpo é sempre não-nulo: funções anônimas atribuídas a variáveis ou
// retornadas são representadas como DeclFuncao com Nome vazio.
type DeclFuncao struct {
	NoBase
	Nome        string                 // Nome do método/função. Vazio para funções anônimas.
	Parametros  []*DeclFuncaoParametro // Lista ordenada de parâmetros formais aceitos.
	Corpo       *Bloco                 // Escopo de bloco correspondente às instruções da função.
	Estatico    bool                   // Verdadeiro se for um método estático de classe (ex: 'estatico funcao').
	TipoRetorno string                 // Anotação opcional de tipo de retorno (captura literal, sem validação em runtime).
	Assincrono  bool                   // Verdadeiro se a função for assíncrona ('assincrono funcao').
	DocComments string                 // Comentários de documentação associados (/// ou /** */).
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

// DeclClasse representa a declaração formal de classes e tipos OOP.
//
// Sintaxe:
//
//	classe MinhaClasse estende Pai {
//	    funcao metodo() { ... }
//	    estatico funcao util() { ... }
//	}
//
// # Restrições
//
//   - Herança múltipla não é suportada (apenas uma classe pai).
//   - Métodos são sempre DeclFuncao com Estatico marcado conforme a presença
//     do modificador 'estatico'.
type DeclClasse struct {
	NoBase
	Nome        string        // Nome da classe declarada.
	Heranca     string        // Nome opcional da classe pai da qual herda atributos e comportamentos.
	Metodos     []*DeclFuncao // Relação de funções e métodos pertencentes à classe.
	DocComments string        // Comentários de documentação associados (/// ou /** */).
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

// TenteCaptureFinalmente representa a estrutura de tratamento de exceções.
//
// Sintaxe aceita:
//
//	tente { ... } capture (erro) { ... }
//	tente { ... } finalmente { ... }
//	tente { ... } capture (erro) { ... } finalmente { ... }
//
// `capture` e `finalmente` são opcionais e independentes. Pelo menos um deles
// deve estar presente após o bloco `tente` — a validação é feita em
// Parser.parseTenteCapture.
//
// # Semântica em runtime
//
//   - Se uma exceção Harpia (ou erro Go empacotado em RuntimeErro) for lançada
//     dentro de TenteBlock e CaptureBlock != nil, o controle salta para ele
//     com NomeErro vinculado ao valor lançado.
//   - FinalmenteBlock, se presente, sempre executa via defer (semântica
//     Python/Java) — mesmo se TenteBlock terminar em exceção.
//   - Erros lançados dentro de FinalmenteBlock substituem o erro original
//     propagado.
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

func (*DeclInterface) isExpr() {}

// DeclEnum representa uma enumeração nativa.
//
// Sintaxe:
//
//	enum Status { Pendente, Concluido, Erro }
//
// Os valores são armazenados apenas como strings (nomes simbólicos). A VM
// atribui internamente índices sequenciais para comparação.
type DeclEnum struct {
	NoBase
	Nome        string   // Identificador da enumeração.
	Valores     []string // Lista ordenada de identificadores de valores.
	DocComments string   // Comentários de documentação associados.
}

// MetodoAssinatura representa a assinatura de um método exigido em uma Interface.
//
// Apenas os nomes dos parâmetros são preservados (não há anotação de tipo nem
// valores padrão em interfaces) — interfaces Harpia são puramente estruturais.
type MetodoAssinatura struct {
	NoBase
	Nome        string   // Nome do método exigido.
	Parametros  []string // Nomes dos parâmetros formais (apenas identificadores).
	TipoRetorno string   // Anotação opcional de tipo de retorno.
}

// DeclInterface representa um contrato de interface.
//
// Sintaxe:
//
//	interface Repositorio {
//	    funcao salvar(item): Logico
//	    funcao remover(id): Logico
//	}
//
// Classes que implementem todos os métodos exigidos satisfazem a interface
// (duck typing estrutural) — não há palavra-chave `implementa`.
type DeclInterface struct {
	NoBase
	Nome        string             // Nome da interface.
	Metodos     []MetodoAssinatura // Lista de assinaturas exigidas.
	DocComments string             // Comentários de documentação associados.
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

// DeclEscolha representa a instrução de pattern matching estrutural.
//
// Sintaxe aceita:
//
//	escolha expressao {
//	    caso padrao quando guarda { ... }
//	    caso _ { ... }                          // wildcard
//	    caso [a, b, ...resto] { ... }            // desestruturação de lista
//	    caso { chave: valor } { ... }            // desestruturação de mapa
//	}
//
// # Algoritmo
//
// A VM percorre os casos em ordem; o primeiro cujo padrão casar (e guarda
// opcional retornar verdadeiro) executa seu Corpo. Não há fall-through.
type DeclEscolha struct {
	NoBase
	Expressao BaseNode      // Expressão discriminada avaliada uma única vez antes dos casos.
	Casos     []CasoEscolha // Lista ordenada de alternativas avaliadas sequencialmente.
}

// AtributoJSX representa um atributo em uma tag JSX.
//
// Sintaxe suportada:
//
//	classe="p-4"               // string literal
//	aoClicar={handleClick}     // expressão entre chaves
//	desabilitado               // atributo booleano implícito (Valor = Identificador "Verdadeiro")
type AtributoJSX struct {
	NoBase
	Nome  string   // Nome do atributo (preserva kebab-case original).
	Valor BaseNode // Pode ser TextoLiteral, Identificador ou qualquer expressão embrulhada.
}

// NoJSX representa uma tag JSX genérica (frontend).
//
// Sintaxe:
//
//	<div classe="app">conteudo</div>
//	<Botao desabilitado />
//	<Componente prop1={x} prop2="texto">
//	    <Filho />
//	</Componente>
//
// # Compilação
//
// Quando compilado para web (alvo=web), NoJSX é transpilado para chamadas
// h(tag, props, ...children) do runtime web nativo do Harpia.
type NoJSX struct {
	NoBase
	Tag         string         // Nome da tag (preserva o caso original).
	Atributos   []*AtributoJSX // Lista de atributos/props da tag.
	Filhos      []BaseNode     // Pode ser NoJSX, NoSeJSX, NoParaJSX, TextoLiteral ou Expressao.
	AutoFechado bool           // Verdadeiro para tags <X /> sem conteúdo.
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

// Decorador representa a sintaxe de decoradores aplicados sobre funções ou classes.
//
// Sintaxe:
//
//	@depreciado
//	@rota("/api", metodo = "GET")
//	funcao endpoint() { ... }
//
// # Aplicação
//
// Decoradores são processados pelo resolvedor de módulos; cada decorador é
// uma função Harpia regular que recebe a declaração decorada como primeiro
// argumento e retorna a declaração (possivelmente modificada).
type Decorador struct {
	NoBase
	Nome       string     // Nome do decorador (ex: "depreciado", "rota").
	Argumentos []BaseNode // Argumentos opcionais entre parênteses (ex: "/api", expressoes).
	Alvo       BaseNode   // A declaração decorada (ex: *DeclFuncao ou *DeclClasse).
}

func (*Decorador) isExpr() {}
