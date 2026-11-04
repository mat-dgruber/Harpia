package hrp

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

// OpcsContexto agrupa as propriedades opcionais de inicialização do contexto da VM.
type OpcsContexto struct {
	Args              []string // Argumentos CLI de terminal repassados ao script.
	CaminhosPadrao    []string // Caminhos de varredura do disco para resolução de diretivas de importação.
	DiretorioTrabalho string   // Diretório base/isolado de trabalho do contexto (evita o uso de os.Chdir global).
	Estrito           bool     // Ativa verificação estrita de tipagem dinâmica opcional.
	BloquearArquivos  bool     // Ativa o bloqueio de acesso ao sistema de arquivos (Sandbox).
	BloquearRede      bool     // Ativa o bloqueio de conexões e operações de rede (Sandbox).
}

// Contexto representa o orquestrador global e supervisor de estado da VM do Harpia.
//
// O Contexto armazena o cache de módulos importados, rastreia as coordenadas geográficas físicas da
// instrução em execução (útil para tracebacks precisos de erros) e lida com a sincronização de tarefas concorrentes.
type Contexto struct {
	Modulos           *TabelaModulos          // Catálogo e cache de módulos ativos carregados na VM.
	Opcs              OpcsContexto            // Opções de configuração da VM.
	fechado           bool                    // Flag reativa que impede operações após o encerramento do contexto.
	mu                sync.Mutex              // Mutex para proteção de fechamento e waitgroup de tarefas concorrentes.
	cancelCh          chan struct{}           // Canal fechado quando o contexto encerra suas operações.
	waitgroup         sync.WaitGroup          // Mecanismo de controle de concorrência para processos ativos em segundo plano.
	once              sync.Once               // Garante encerramento seguro de ciclo único da VM.
	posMu             sync.RWMutex            // Mutex para proteção de campos de arquivo, código e coordenadas geográficas
	ArquivoAtual      string                  // O nome do arquivo físico sob execução corrente.
	CodigoAtual       string                  // O código fonte textual sob execução.
	DiretorioTrabalho string                  // Diretório base isolado de trabalho deste contexto.
	LinhaAtual        int                     // Linha física corrente sob interpretação (base 1, -1 = indefinida).
	ColunaAtual       int                     // Coluna física corrente (base 1).
	TokenAtual        *lexer.Token            // Ponteiro para o token do lexer sob análise operacional.
	ResolvendoModulos []string                // Pilha de módulos que estão sendo importados/resolvidos para prevenção de ciclos.
	LinhasExecutadas  map[string]map[int]bool // ponytail: mapa para rastreamento de linhas cobertas na execução de testes
	linhasExecMu      sync.Mutex              // mutex para evitar concorrência no mapa de linhas executadas
	profundidade      atomic.Int32            // chamadas de função em andamento (limite de recursão)
}

// MaxProfundidadeChamadas limita a recursão: acima disso a chamada lança ErroDePilha em vez de
// estourar a pilha do Go (fatal error: stack overflow, que derruba o processo sem recover).
// ponytail: o contador é por contexto e soma goroutines concorrentes (handlers HTTP, assíncronas);
// o limite é generoso para cobrir isso. Upgrade: profundidade por frame/goroutine se houver falso positivo.
const MaxProfundidadeChamadas = 2000

// EntrarChamada registra uma chamada de função e devolve a função que a encerra.
func (c *Contexto) EntrarChamada() (func(), error) {
	if c == nil {
		return func() {}, nil
	}
	if c.profundidade.Add(1) > MaxProfundidadeChamadas {
		c.profundidade.Add(-1)
		return nil, NewErroF(ErroDePilha, "Limite de profundidade de chamadas excedido (%d): provável recursão infinita", MaxProfundidadeChamadas)
	}
	return func() { c.profundidade.Add(-1) }, nil
}

func (c *Contexto) DefinirArquivoCodigo(arquivo, codigo string) {
	if c == nil {
		return
	}
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.ArquivoAtual = arquivo
	c.CodigoAtual = codigo
}

func (c *Contexto) ObterArquivoCodigo() (string, string) {
	if c == nil {
		return "", ""
	}
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.ArquivoAtual, c.CodigoAtual
}

func (c *Contexto) DefinirPosicaoToken(tok *lexer.Token, linha, coluna int) {
	if c == nil {
		return
	}
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.TokenAtual = tok
	c.LinhaAtual = linha
	c.ColunaAtual = coluna
}

func (c *Contexto) ObterPosicaoToken() (*lexer.Token, int, int) {
	if c == nil {
		return nil, -1, 0
	}
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.TokenAtual, c.LinhaAtual, c.ColunaAtual
}

var (
	contextoAtivoMu sync.RWMutex
	contextoAtivo   *Contexto
	OnEnvLoaded     []func()
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
}

var ContextoAtivo *Contexto

// NewContexto aloca um novo contexto orquestrador da VM, inicializando o cache de módulos
// e pré-carregando de forma global e compulsória o módulo nativo de embutidos.
func NewContexto(opcs OpcsContexto) *Contexto {
	dirTrabalho := opcs.DiretorioTrabalho
	if dirTrabalho == "" {
		if d, err := os.Getwd(); err == nil {
			dirTrabalho = d
		} else {
			dirTrabalho = "."
		}
	}

	context := &Contexto{
		Modulos:           NewTabelaModulos(),
		Opcs:              opcs,
		fechado:           false,
		cancelCh:          make(chan struct{}),
		DiretorioTrabalho: dirTrabalho,
		LinhaAtual:        -1,
		LinhasExecutadas:  make(map[string]map[int]bool), // ponytail: inicializa mapa de cobertura
	}

	DefinirContextoAtivo(context)

	// Carrega automaticamente .harpia.env do diretório atual de trabalho do contexto
	carregarArquivoEnv(dirTrabalho)
	for _, dir := range opcs.CaminhosPadrao {
		carregarArquivoEnv(dir)
	}

	MultiImporteModulo(context, "embutidos")
	return context
}

// ObterDiretorioTrabalho retorna o diretório isolado de trabalho do contexto.
func (c *Contexto) ObterDiretorioTrabalho() string {
	if c != nil && c.DiretorioTrabalho != "" {
		return c.DiretorioTrabalho
	}
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "."
}

// carregarArquivoEnv tenta ler arquivos de ambiente (.harpia.env e .env) no diretório dado e exporta suas variáveis para o processo.
// Linhas em branco e comentários (iniciados com #) são ignorados silenciosamente.
func carregarArquivoEnv(dir string) {
	for _, nome := range []string{".harpia.env", ".env"} {
		caminho := filepath.Join(dir, nome)
		f, err := os.Open(caminho)
		if err != nil {
			continue // arquivo não existe ou não é acessível — silencioso
		}

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
		if err := scanner.Err(); err != nil {
			// Ignorado de forma silenciosa e segura
		}
		f.Close()
	}

	for _, fn := range OnEnvLoaded {
		fn()
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
	c.DefinirArquivoCodigo(caminho, codigo)
	ast, err := parser.NewParserFromString(string(codigo), caminho).Parse()
	if err != nil {
		if errSintatico, ok := err.(*parser.ErroSintatico); ok {
			linha := -1
			coluna := 0
			if errSintatico.Token != nil && errSintatico.Token.Inicio.Linha > 0 {
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

// ErrNaoSuportadoVM marca construções que o compilador de bytecode ainda não cobre;
// AvaliarAst usa errors.Is para cair no interpretador AST apenas nesses casos.
var ErrNaoSuportadoVM = errors.New("construção ainda não suportada na VM de bytecode")

var executadorVM func(ctx *Contexto, ast parser.BaseNode, escopo *Escopo) (Objeto, error)

func DefinirExecutadorVM(exec func(ctx *Contexto, ast parser.BaseNode, escopo *Escopo) (Objeto, error)) {
	executadorVM = exec
}

// AvaliarAst cria uma instância de Interpretador e processa sequencialmente as instruções da AST sob o escopo informado.
func (c *Contexto) AvaliarAst(ast parser.BaseNode, escopo *Escopo) (Objeto, error) {
	if err := c.AdicionarTrabalho(); err != nil {
		return nil, err
	}
	defer c.EncerrarTrabalho()

	if executadorVM != nil {
		resVM, errVM := executadorVM(c, ast, escopo)
		if errVM == nil {
			return resVM, nil
		}
		// Só erros de COMPILAÇÃO carregam o sentinela: um erro de runtime do script nunca
		// dispara reexecução pelo interpretador (o que duplicaria efeitos colaterais).
		if !errors.Is(errVM, ErrNaoSuportadoVM) {
			return nil, errVM
		}
	}

	interpret := &Interpretador{Ast: ast, Contexto: c, Escopo: escopo}
	oldArquivo, oldCodigo := c.ObterArquivoCodigo()
	defer func() {
		c.DefinirArquivoCodigo(oldArquivo, oldCodigo)
	}()

	if prog, ok := ast.(*parser.Programa); ok {
		interpret.Arquivo = prog.Arquivo
		interpret.Codigo = prog.Codigo
		interpret.Posicoes = prog.Posicoes
		c.DefinirArquivoCodigo(prog.Arquivo, prog.Codigo)
	} else {
		interpret.Arquivo = oldArquivo
		interpret.Codigo = oldCodigo
		// If Contexto already has parsed posicoes map or we can load it:
		if oldArquivo != "" {
			if mod, err := c.Modulos.ObterModulo(oldArquivo); err == nil && mod.Impl.Ast != nil {
				if prog, ok := mod.Impl.Ast.(*parser.Programa); ok {
					interpret.Posicoes = prog.Posicoes
				}
			}
		}
	}

	// Fetch Posicoes from Program AST if wrapped, or get them if we can find it
	if interpret.Posicoes == nil {
		curArq, _ := c.ObterArquivoCodigo()
		if prog, ok := ast.(*parser.Programa); ok {
			interpret.Posicoes = prog.Posicoes
		} else if curArq != "" {
			// Find the parsed program from cached modules if available
			if mod, err := c.Modulos.ObterModulo(curArq); err == nil && mod.Impl.Ast != nil {
				if prog, ok := mod.Impl.Ast.(*parser.Programa); ok {
					interpret.Posicoes = prog.Posicoes
				}
			}
		}
	}

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
			tok, lin, col := c.ObterPosicaoToken()
			if hrpErr.Token == nil && interpret.Posicoes != nil && tok != nil {
				hrpErr.Token = tok
				hrpErr.Linha = lin
				hrpErr.Coluna = col
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

	oldArquivo, oldCodigo := c.ObterArquivoCodigo()
	defer func() {
		c.DefinirArquivoCodigo(oldArquivo, oldCodigo)
	}()

	modulo, err := c.Modulos.NewModulo(c, implementacao)
	if err != nil {
		return nil, err
	}

	if implementacao.Ast != nil {
		if prog, ok := implementacao.Ast.(*parser.Programa); ok {
			c.DefinirArquivoCodigo(prog.Arquivo, prog.Codigo)
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
	c.mu.Lock()
	defer c.mu.Unlock()

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
		c.mu.Lock()
		c.fechado = true
		if c.cancelCh != nil {
			close(c.cancelCh)
		}
		c.mu.Unlock()

		c.waitgroup.Wait()

		contextoAtivoMu.Lock()
		if contextoAtivo == c {
			contextoAtivo = nil
		}
		contextoAtivoMu.Unlock()
	})
}

// CanalCancelamento retorna um canal que é fechado quando o contexto encerra suas operações.
func (c *Contexto) CanalCancelamento() <-chan struct{} {
	if c == nil {
		return nil
	}
	return c.cancelCh
}

// VerificarCancelamento retorna erro se o contexto foi encerrado (ex.: timeout do playground).
// Laços do interpretador e saltos para trás da VM chamam isto a cada iteração para que
// um script em loop infinito pare quando Terminar() é chamado.
func (c *Contexto) VerificarCancelamento() error {
	if c == nil || c.cancelCh == nil {
		return nil
	}
	select {
	case <-c.cancelCh:
		return NewErroF(RuntimeErro, "Execução cancelada: o contexto foi encerrado")
	default:
		return nil
	}
}

// Fechado retorna true se o contexto foi marcado como encerrado.
func (c *Contexto) Fechado() bool {
	if c == nil {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fechado
}

// Recurso identifica a capacidade sensível que uma função nativa quer usar.
type Recurso int

const (
	RecursoArquivos  Recurso = iota // ler/escrever disco (inclui SQLite e FFI)
	RecursoRede                     // conexões e requisições de rede
	RecursoProcessos                // subprocessos e variáveis de ambiente (usa a flag BloquearArquivos)
)

// ContextoDe resolve o contexto dono da chamada: o do módulo receptor (o script que importou
// a função), com fallback para o contexto ativo global.
func ContextoDe(inst Objeto) *Contexto {
	if m, ok := inst.(*Modulo); ok && m != nil && m.Contexto != nil {
		return m.Contexto
	}
	return ObterContextoAtivo()
}

// ExigirPermissao é o guard único do sandbox. Ele resolve o contexto pelo receptor, e não só pelo
// ativo global (que outro NewContexto concorrente pode trocar), e nega quando não há contexto
// nenhum para consultar (fail-closed).
func ExigirPermissao(inst Objeto, r Recurso) error {
	return ExigirPermissaoNoContexto(ContextoDe(inst), r)
}

// ExigirPermissaoNoContexto aplica o guard a um contexto já conhecido (ex.: capturado por um
// objeto nativo na criação). nil = negado.
func ExigirPermissaoNoContexto(ctx *Contexto, r Recurso) error {
	if ctx == nil {
		return NewErroF(ErroDeSistema, "Acesso Negado: nenhum contexto de execução ativo para autorizar a operação.")
	}
	if r == RecursoRede {
		return ctx.VerificarPermissaoRede()
	}
	return ctx.VerificarPermissaoArquivos()
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
