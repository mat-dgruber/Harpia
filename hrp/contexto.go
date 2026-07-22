package hrp

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

// OpcsContexto agrupa as propriedades opcionais de inicialização do contexto da VM.
type OpcsContexto struct {
	Args             []string // Argumentos CLI de terminal repassados ao script.
	CaminhosPadrao   []string // Caminhos de varredura do disco para resolução de diretivas de importação.
	Estrito          bool     // Ativa verificação estrita de tipagem dinâmica opcional.
	BloquearArquivos bool     // Ativa o bloqueio de acesso ao sistema de arquivos (Sandbox).
	BloquearRede     bool     // Ativa o bloqueio de conexões e operações de rede (Sandbox).
}

// Contexto representa o orquestrador global e supervisor de estado da VM do Harpia.
//
// O Contexto armazena o cache de módulos importados, rastreia as coordenadas geográficas físicas da
// instrução em execução (útil para tracebacks precisos de erros) e lida com a sincronização de tarefas concorrentes.
type Contexto struct {
	Modulos           *TabelaModulos          // Catálogo e cache de módulos ativos carregados na VM.
	Opcs              OpcsContexto            // Opções de configuração da VM.
	fechado           bool                    // Flag reativa que impede operações após o encerramento do contexto.
	waitgroup         sync.WaitGroup          // Mecanismo de controle de concorrência para processos ativos em segundo plano.
	once              sync.Once               // Garante encerramento seguro de ciclo único da VM.
	ArquivoAtual      string                  // O nome do arquivo físico sob execução corrente.
	CodigoAtual       string                  // O código fonte textual sob execução.
	LinhaAtual        int                     // Linha física corrente sob interpretação (base 1, -1 = indefinida).
	ColunaAtual       int                     // Coluna física corrente (base 1).
	TokenAtual        *lexer.Token            // Ponteiro para o token do lexer sob análise operacional.
	ResolvendoModulos []string                // Pilha de módulos que estão sendo importados/resolvidos para prevenção de ciclos.
	LinhasExecutadas  map[string]map[int]bool // ponytail: mapa para rastreamento de linhas cobertas na execução de testes
	linhasExecMu      sync.Mutex              // mutex para evitar concorrência no mapa de linhas executadas
}

var (
	contextoAtivoMu sync.RWMutex
	contextoAtivo   *Contexto
)

func ObterContextoAtivo() *Contexto {
	contextoAtivoMu.RLock()
	defer contextoAtivoMu.RUnlock()
	return contextoAtivo
}

func DefinirContextoAtivo(ctx *Contexto) {
	contextoAtivoMu.Lock()
	defer contextoAtivoMu.Unlock()
	contextoAtivo = ctx
	ContextoAtivo = ctx
}

var ContextoAtivo *Contexto

// NewContexto aloca um novo contexto orquestrador da VM, inicializando o cache de módulos
// e pré-carregando de forma global e compulsória o módulo nativo de embutidos.
func NewContexto(opcs OpcsContexto) *Contexto {
	context := &Contexto{
		Modulos:          NewTabelaModulos(),
		Opcs:             opcs,
		fechado:          false,
		LinhaAtual:       -1,
		LinhasExecutadas: make(map[string]map[int]bool), // ponytail: inicializa mapa de cobertura
	}

	DefinirContextoAtivo(context)

	// Carrega automaticamente .harpia.env do diretório atual ou dos caminhos padrão
	if dir, err := os.Getwd(); err == nil {
		carregarArquivoEnv(dir)
	}
	for _, dir := range opcs.CaminhosPadrao {
		carregarArquivoEnv(dir)
	}

	Importe = func(nome string, escopo *Escopo) (Objeto, error) {
		return MaquinarioImporteModulo(context, nome, escopo)
	}

	MultiImporteModulo(context, "embutidos")
	return context
}

// carregarArquivoEnv tenta ler o arquivo .harpia.env no diretório dado e exporta suas variáveis para o processo.
// Linhas em branco e comentários (iniciados com #) são ignorados silenciosamente.
func carregarArquivoEnv(dir string) {
	caminho := filepath.Join(dir, ".harpia.env")
	f, err := os.Open(caminho)
	if err != nil {
		return // arquivo não existe ou não é acessível — silencioso
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}
		partes := strings.SplitN(linha, "=", 2)
		if len(partes) == 2 {
			chave := strings.TrimSpace(partes[0])
			valor := strings.TrimSpace(partes[1])
			// Somente define se a variável ainda não estiver no ambiente (não sobrescreve vars de sistema)
			if os.Getenv(chave) == "" {
				os.Setenv(chave, valor)
			}
		}
	}
}

// TransformarEmAst localiza um arquivo físico no disco, decodifica sua codificação,
// lê o seu conteúdo textual e dispara o parser para convertê-lo em Árvore de Sintaxe Abstrata (AST).
func (c *Contexto) TransformarEmAst(caminhoInicial string, useSysPaths bool, curDir string) (caminho string, ast parser.BaseNode, err error) {
	if err = c.AdicionarTrabalho(); err != nil {
		return
	}
	defer c.EncerrarTrabalho()

	caminhos := []string{}
	if useSysPaths {
		caminhos = c.Opcs.CaminhosPadrao
	}

	caminho, err = ResolveArquivohrp(caminhoInicial, caminhos, curDir)
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
	c.CodigoAtual = codigo
	c.ArquivoAtual = caminho
	ast, err := parser.NewParserFromString(string(codigo), caminho).Parse()
	if err != nil {
		if errSintatico, ok := err.(*parser.ErroSintatico); ok {
			linha := -1
			coluna := 0
			if errSintatico.Token != nil && errSintatico.Token.Inicio != nil {
				linha = errSintatico.Token.Inicio.Linha - 1
				coluna = errSintatico.Token.Inicio.Coluna
			}
			return nil, &Erro{
				Base:     SintaxeErro,
				Mensagem: Texto(errSintatico.Mensagem),
				Linha:    linha,
				Coluna:   coluna,
				Token:    errSintatico.Token,
				Arquivo:  errSintatico.Arquivo,
				Codigo:   errSintatico.Codigo,
			}
		}
		return nil, NewErroF(SintaxeErro, "%s", err)
	}

	return ast, nil
}

// AvaliarAst cria uma instância de Interpretador e processa sequencialmente as instruções da AST sob o escopo informado.
func (c *Contexto) AvaliarAst(ast parser.BaseNode, escopo *Escopo) (Objeto, error) {
	if err := c.AdicionarTrabalho(); err != nil {
		return nil, err
	}
	defer c.EncerrarTrabalho()

	interpret := &Interpretador{Ast: ast, Contexto: c, Escopo: escopo}
	if prog, ok := ast.(*parser.Programa); ok {
		interpret.Arquivo = prog.Arquivo
		interpret.Codigo = prog.Codigo
		interpret.Posicoes = prog.Posicoes
		c.ArquivoAtual = prog.Arquivo
		c.CodigoAtual = prog.Codigo
	} else {
		interpret.Arquivo = c.ArquivoAtual
		interpret.Codigo = c.CodigoAtual
		// If Contexto already has parsed posicoes map or we can load it:
		if c.ArquivoAtual != "" {
			if mod, err := c.Modulos.ObterModulo(c.ArquivoAtual); err == nil && mod.Impl.Ast != nil {
				if prog, ok := mod.Impl.Ast.(*parser.Programa); ok {
					interpret.Posicoes = prog.Posicoes
				}
			}
		}
	}

	// Fetch Posicoes from Program AST if wrapped, or get them if we can find it
	if interpret.Posicoes == nil {
		if prog, ok := ast.(*parser.Programa); ok {
			interpret.Posicoes = prog.Posicoes
		} else if c.ArquivoAtual != "" {
			// Find the parsed program from cached modules if available
			if mod, err := c.Modulos.ObterModulo(c.ArquivoAtual); err == nil && mod.Impl.Ast != nil {
				if prog, ok := mod.Impl.Ast.(*parser.Programa); ok {
					interpret.Posicoes = prog.Posicoes
				}
			}
		}
	}

	// Save and restore context state because MultiImporteModulo("embutidos") or inner evaluations
	// will mutate c.CodigoAtual and c.ArquivoAtual.
	oldArquivo := c.ArquivoAtual
	oldCodigo := c.CodigoAtual
	defer func() {
		c.ArquivoAtual = oldArquivo
		c.CodigoAtual = oldCodigo
	}()

	res, err := interpret.Inicializa()

	if err != nil {
		if hrpErr, ok := err.(*Erro); ok {
			if hrpErr.Arquivo == "" || hrpErr.Arquivo == "<desconhecido>" {
				hrpErr.Arquivo = interpret.Arquivo
			}
			if hrpErr.Codigo == "" {
				hrpErr.Codigo = interpret.Codigo
			}
			// Enrich token details from visitor state if possible
			if hrpErr.Token == nil && interpret.Posicoes != nil && c.TokenAtual != nil {
				hrpErr.Token = c.TokenAtual
				hrpErr.Linha = c.LinhaAtual
				hrpErr.Coluna = c.ColunaAtual
			}
		}
		return nil, err
	}
	return res, nil
}

// ObterModulo resolve e recupera um módulo ativo a partir da tabela hash de cache da VM.
func (c *Contexto) ObterModulo(nome string) (*Modulo, error) {
	return c.Modulos.ObterModulo(nome)
}

// InicializarModulo cria a representação do módulo dinâmico na tabela de cache e avalia suas respectivas ASTs de código.
func (c *Contexto) InicializarModulo(implementacao *ModuloImpl) (*Modulo, error) {
	if err := c.AdicionarTrabalho(); err != nil {
		return nil, err
	}
	defer c.EncerrarTrabalho()

	modulo, err := c.Modulos.NewModulo(c, implementacao)
	if err != nil {
		return nil, err
	}

	if implementacao.Ast != nil {
		if prog, ok := implementacao.Ast.(*parser.Programa); ok {
			c.ArquivoAtual = prog.Arquivo
			c.CodigoAtual = prog.Codigo
		}
		_, err := c.AvaliarAst(implementacao.Ast, modulo.Escopo)
		if err != nil {
			return nil, err
		}
	}

	return modulo, nil
}

// adicionarTrabalho incrementa de forma concorrente a contagem do WaitGroup, controlando que
// o contexto não seja finalizado abruptamente enquanto tarefas ativas estão em processamento.
func (c *Contexto) AdicionarTrabalho() error {
	if c.fechado {
		return NewErro(RuntimeErro, Texto("Contexto já fechado"))
	}

	c.waitgroup.Add(1)
	return nil
}

// EncerrarTrabalho sinaliza o encerramento operacional de uma tarefa pendente na VM.
func (c *Contexto) EncerrarTrabalho() {
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

// VerificarPermissaoArquivos valida se o contexto possui a flag de permissão para ler ou escrever arquivos.
func (c *Contexto) VerificarPermissaoArquivos() error {
	if c.Opcs.BloquearArquivos {
		return NewErroF(ErroDeSistema, "Acesso Negado: Operação de manipulação de arquivos bloqueada pelo Sandbox.")
	}
	return nil
}

// VerificarPermissaoRede valida se o contexto possui permissão ativa para conexões de rede (abrir portas ou requisições cliente).
func (c *Contexto) VerificarPermissaoRede() error {
	if c.Opcs.BloquearRede {
		return NewErroF(ErroDeSistema, "Acesso Negado: Operação de rede bloqueada pelo Sandbox.")
	}
	return nil
}

// RegistrarLinhaExecutada registra uma linha executada de forma concorrente-segura.
func (c *Contexto) RegistrarLinhaExecutada(arquivo string, linha int) {
	if c == nil || c.LinhasExecutadas == nil || arquivo == "" {
		return
	}
	c.linhasExecMu.Lock()
	defer c.linhasExecMu.Unlock()
	if c.LinhasExecutadas[arquivo] == nil {
		c.LinhasExecutadas[arquivo] = make(map[int]bool)
	}
	c.LinhasExecutadas[arquivo][linha] = true
}

// ObterLinhasExecutadas retorna uma cópia das linhas executadas de forma concorrente-segura.
func (c *Contexto) ObterLinhasExecutadas() map[string]map[int]bool {
	if c == nil {
		return nil
	}
	c.linhasExecMu.Lock()
	defer c.linhasExecMu.Unlock()
	copia := make(map[string]map[int]bool)
	for k, v := range c.LinhasExecutadas {
		subCopia := make(map[int]bool)
		for k2, v2 := range v {
			subCopia[k2] = v2
		}
		copia[k] = subCopia
	}
	return copia
}

