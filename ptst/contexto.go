package ptst

import (
	"os"
	"strings"
	"sync"

	"github.com/natanfeitosa/portuscript/lexer"
	"github.com/natanfeitosa/portuscript/parser"
)

// OpcsContexto agrupa as propriedades opcionais de inicialização do contexto da VM.
type OpcsContexto struct {
	Args           []string // Argumentos CLI de terminal repassados ao script.
	CaminhosPadrao []string // Caminhos de varredura do disco para resolução de diretivas de importação.
}

// Contexto representa o orquestrador global e supervisor de estado da VM do Portuscript.
//
// O Contexto armazena o cache de módulos importados, rastreia as coordenadas geográficas físicas da
// instrução em execução (útil para tracebacks precisos de erros) e lida com a sincronização de tarefas concorrentes.
type Contexto struct {
	Modulos      *TabelaModulos // Catálogo e cache de módulos ativos carregados na VM.
	Opcs         OpcsContexto   // Opções de configuração da VM.
	fechado      bool           // Flag reativa que impede operações após o encerramento do contexto.
	waitgroup    sync.WaitGroup // Mecanismo de controle de concorrência para processos ativos em segundo plano.
	once         sync.Once      // Garante encerramento seguro de ciclo único da VM.
	ArquivoAtual string         // O nome do arquivo físico sob execução corrente.
	CodigoAtual  string         // O código fonte textual sob execução.
	LinhaAtual   int            // Linha física corrente sob interpretação (base 1, -1 = indefinida).
	ColunaAtual  int            // Coluna física corrente (base 1).
	TokenAtual   *lexer.Token   // Ponteiro para o token do lexer sob análise operacional.
}

// NewContexto aloca um novo contexto orquestrador da VM, inicializando o cache de módulos
// e pré-carregando de forma global e compulsória o módulo nativo de embutidos.
func NewContexto(opcs OpcsContexto) *Contexto {
	context := &Contexto{
		Modulos:    NewTabelaModulos(),
		Opcs:       opcs,
		fechado:    false,
		LinhaAtual: -1,
	}

	Importe = func(nome string, escopo *Escopo) (Objeto, error) {
		return MaquinarioImporteModulo(context, nome, escopo)
	}

	MultiImporteModulo(context, "embutidos")
	return context
}

// TransformarEmAst localiza um arquivo físico no disco, decodifica sua codificação,
// lê o seu conteúdo textual e dispara o parser para convertê-lo em Árvore de Sintaxe Abstrata (AST).
func (c *Contexto) TransformarEmAst(caminhoInicial string, useSysPaths bool, curDir string) (caminho string, ast parser.BaseNode, err error) {
	if err = c.adicionarTrabalho(); err != nil {
		return
	}
	defer c.encerrarTrabalho()

	caminhos := []string{}
	if useSysPaths {
		caminhos = c.Opcs.CaminhosPadrao
	}

	caminho, err = ResolveArquivoPtst(caminhoInicial, caminhos, curDir)
	if err != nil || strings.HasSuffix(caminho, "so") {
		return
	}

	var codigo []byte
	codigo, err = os.ReadFile(caminho)
	if err != nil {
		err = NewErroF(ErroDeSistema, "Erro ao acessar '%s': %s", caminho, err)
		return
	}

	ast, err = c.StringParaAst(string(codigo), caminho)
	return
}

// StringParaAst é a ponte direta que invoca e aciona o Parser para traduzir uma string de código em nós de AST.
func (c *Contexto) StringParaAst(codigo string, caminho string) (parser.BaseNode, error) {
	ast, err := parser.NewParserFromString(string(codigo), caminho).Parse()
	if err != nil {
		return nil, NewErroF(SintaxeErro, "%s", err)
	}

	return ast, nil
}

// AvaliarAst cria uma instância de Interpretador e processa sequencialmente as instruções da AST sob o escopo informado.
func (c *Contexto) AvaliarAst(ast parser.BaseNode, escopo *Escopo) (Objeto, error) {
	if err := c.adicionarTrabalho(); err != nil {
		return nil, err
	}
	defer c.encerrarTrabalho()

	interpret := &Interpretador{Ast: ast, Contexto: c, Escopo: escopo}
	if prog, ok := ast.(*parser.Programa); ok {
		interpret.Arquivo = prog.Arquivo
		interpret.Codigo = prog.Codigo
		interpret.Posicoes = prog.Posicoes
		c.ArquivoAtual = prog.Arquivo
		c.CodigoAtual = prog.Codigo
	}

	MultiImporteModulo(interpret.Contexto, "embutidos")

	return interpret.Inicializa()
}

// ObterModulo resolve e recupera um módulo ativo a partir da tabela hash de cache da VM.
func (c *Contexto) ObterModulo(nome string) (*Modulo, error) {
	return c.Modulos.ObterModulo(nome)
}

// InicializarModulo cria a representação do módulo dinâmico na tabela de cache e avalia suas respectivas ASTs de código.
func (c *Contexto) InicializarModulo(implementacao *ModuloImpl) (*Modulo, error) {
	if err := c.adicionarTrabalho(); err != nil {
		return nil, err
	}
	defer c.encerrarTrabalho()

	modulo, err := c.Modulos.NewModulo(c, implementacao)
	if err != nil {
		return nil, err
	}

	if implementacao.Ast != nil {
		_, err := c.AvaliarAst(implementacao.Ast, modulo.Escopo)
		if err != nil {
			return nil, err
		}
	}

	return modulo, nil
}

// adicionarTrabalho incrementa de forma concorrente a contagem do WaitGroup, controlando que
// o contexto não seja finalizado abruptamente enquanto tarefas ativas estão em processamento.
func (c *Contexto) adicionarTrabalho() error {
	if c.fechado {
		return NewErro(RuntimeErro, Texto("Contexto já fechado"))
	}

	c.waitgroup.Add(1)
	return nil
}

// encerrarTrabalho sinaliza o encerramento operacional de uma tarefa pendente na VM.
func (c *Contexto) encerrarTrabalho() {
	c.waitgroup.Done()
}

// Terminar executa a destruição segura e controlada do contexto e das variáveis associadas à VM.
// Aguarda a liberação de todas as rotinas concorrentes registradas no WaitGroup antes de encerrar as operações.
func (c *Contexto) Terminar() {
	c.once.Do(func() {
		c.waitgroup.Wait()
		c.fechado = true
		Importe = func(s string, e *Escopo) (Objeto, error) {
			panic("Antes de usar a função `Importe` você precisa criar um contexto")
		}
	})
}
