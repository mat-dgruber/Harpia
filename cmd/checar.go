package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
	_ "github.com/mat-dgruber/Harpia/stdlib"
	"github.com/spf13/cobra"
)

// globalsLinter define o conjunto de nomes globais pré-declarados pela VM e pela stdlib
// do Harpia. Cada vez que novos built-ins forem adicionados, sincronize este mapa
// para evitar falsos positivos no linter ("Identificador X não encontrado no escopo").
//
// Manter esta lista sincronizada manualmente é o ponto principal — alternativa seria
// injetar via o Contexto da VM, mas isso acopla o linter ao runtime e quebra a natureza
// "estática" dessa checagem.
var globalsLinter = map[string]bool{
	// I/O básico
	"imprimir": true,
	"imprima":  true,
	"escreva":  true,

	// Asserts e helpers de debug
	"assegura": true,

	// Constantes literais
	"Verdadeiro": true,
	"Falso":      true,
	"Nulo":       true,

	// Classes/tipos globais
	"Texto":   true,
	"Inteiro": true,
	"Decimal": true,
	"Logico":  true,
	"Lista":   true,
	"Tupla":   true,
	"Mapa":    true,
	"Objeto":   true,
	"Erro":     true,
	"Servidor": true,
	"int":      true,
	"string":   true,
	"bool":     true,
	"float":    true,

	// Primitivas Reativas da Web
	"sinal":            true,
	"efeito":           true,
	"derivado":         true,
	"armazem":          true,
	"renderizar":       true,
	"montar":           true,

	// Globais adicionais da Web/Browser para DX profissional (Fase 5-B)
	"fetch":            true,
	"JSON":             true,
	"window":           true,
	"document":         true,
	"console":          true,
	"localStorage":     true,
	"importarHtml":     true,
	"GradeDeDados":     true,
	"FronteiraDeErro":  true,
	"ListaVirtual":     true,
	"Provedor":         true,
	"injetar":          true,
	"sinalPersistente": true,
	"recurso":          true,
	"tamanho":          true,
	"roteador":         true,
}

func init() {
	for k, v := range hrp.ObtemGlobalsDoLinter() {
		globalsLinter[k] = v
	}
}

type EscopoLinter struct {
	Pai        *EscopoLinter
	Variaveis  map[string]bool
	Consts     map[string]bool
	Assincrono bool
	Usadas     map[string]bool
	Declaracoes map[string]parser.BaseNode // guarda nó da declaração para reporte
}

func (e *EscopoLinter) Declarada(nome string) bool {
	if e.Variaveis[nome] || e.Consts[nome] {
		return true
	}
	if e.Pai != nil {
		return e.Pai.Declarada(nome)
	}
	return false
}

func (e *EscopoLinter) DeclaradaLocal(nome string) bool {
	return e.Variaveis[nome] || e.Consts[nome]
}

func (e *EscopoLinter) MarcarComoUsada(nome string) {
	if e.Variaveis[nome] || e.Consts[nome] {
		e.Usadas[nome] = true
		return
	}
	if e.Pai != nil {
		e.Pai.MarcarComoUsada(nome)
	}
}

func (e *EscopoLinter) IsConst(nome string) bool {
	if e.Consts[nome] {
		return true
	}
	if e.Pai != nil {
		return e.Pai.IsConst(nome)
	}
	return false
}

func (e *EscopoLinter) NoContextoAssincrono() bool {
	if e.Assincrono {
		return true
	}
	if e.Pai != nil {
		return e.Pai.NoContextoAssincrono()
	}
	return false
}

type DiagnosticRange struct {
	Start DiagnosticPosition `json:"start"`
	End   DiagnosticPosition `json:"end"`
}

type DiagnosticPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type LSPDiagnostic struct {
	Range    DiagnosticRange `json:"range"`
	Severity int             `json:"severity"`
	Code     string          `json:"code"`
	Source   string          `json:"source"`
	Message  string          `json:"message"`
}

type LinterError struct {
	Message  string
	Node     parser.BaseNode
	Code     string
	Severity int // 1 = Erro, 2 = Aviso
}

type Linter struct {
	Erros             []LinterError
	Escopo            *EscopoLinter
	Posicoes          map[parser.BaseNode]*lexer.Token
	DiretorioAtual    string
	ArquivosVisitados map[string]bool
}

func (l *Linter) pushEscopo() {
	l.Escopo = &EscopoLinter{
		Pai:         l.Escopo,
		Variaveis:   make(map[string]bool),
		Consts:      make(map[string]bool),
		Usadas:      make(map[string]bool),
		Declaracoes: make(map[string]parser.BaseNode),
	}
}

func (l *Linter) popEscopo() {
	if l.Escopo != nil {
		// Checa variáveis declaradas localmente mas não usadas antes de sair do escopo
		for varNome, node := range l.Escopo.Declaracoes {
			// Ignora variáveis que começam com "_" (padrão de descarte de DX) e os globais
			// Também ignora componentes (funções/variáveis começando com letra maiúscula) para evitar falsos positivos
			primeiraLetra := ""
			if len(varNome) > 0 {
				primeiraLetra = string(varNome[0])
			}
			ehComponente := primeiraLetra != "" && primeiraLetra == strings.ToUpper(primeiraLetra) && primeiraLetra != strings.ToLower(primeiraLetra)

			if !strings.HasPrefix(varNome, "_") && !ehComponente && !l.Escopo.Usadas[varNome] {
				l.registrarErro(fmt.Sprintf("Variável '%s' foi declarada mas nunca utilizada", varNome), "HRP-0006", 2, node)
			}
		}
		// Propaga marcas de uso para o escopo pai
		if l.Escopo.Pai != nil {
			for u := range l.Escopo.Usadas {
				l.Escopo.Pai.MarcarComoUsada(u)
			}
		}
		l.Escopo = l.Escopo.Pai
	}
}


func (l *Linter) registrarErro(msg string, code string, severity int, node parser.BaseNode) {
	l.Erros = append(l.Erros, LinterError{Message: msg, Code: code, Severity: severity, Node: node})
}

func (l *Linter) Checar(node parser.BaseNode) {
	if node == nil {
		return
	}
	// ponytail: impede pânico se a interface BaseNode encapsular um ponteiro concreto nulo
	val := reflect.ValueOf(node)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return
	}

	switch n := node.(type) {
	case *parser.Programa:
		l.Posicoes = n.Posicoes
		l.pushEscopo()
		// Hidrata o escopo raiz com built-ins e nomes globais conhecidos
		for nome := range globalsLinter {
			l.Escopo.Variaveis[nome] = true
			l.Escopo.Usadas[nome] = true
		}

		for _, decl := range n.Declaracoes {
			l.Checar(decl)
		}
		l.popEscopo()

	case *parser.DeclVar:
		if l.Escopo.DeclaradaLocal(n.Nome) {
			l.registrarErro(fmt.Sprintf("Variável '%s' já declarada neste escopo", n.Nome), "HRP-0002", 1, n)
		} else if l.Escopo.Pai != nil && l.Escopo.Pai.Declarada(n.Nome) && !globalsLinter[n.Nome] {
			// Só alerta sobre shadowing se o pai declarar e não for uma global ou raíz
			l.registrarErro(fmt.Sprintf("O identificador '%s' está sombreando (shadowing) uma variável externa", n.Nome), "HRP-0002", 2, n)
		}
		if n.Constante {
			l.Escopo.Consts[n.Nome] = true
		} else {
			l.Escopo.Variaveis[n.Nome] = true
		}
		l.Escopo.Declaracoes[n.Nome] = n
		l.Checar(n.Inicializador)

		// 2. Detectar vazamento de credenciais
		nomeLower := strings.ToLower(n.Nome)
		if strings.Contains(nomeLower, "senha") || strings.Contains(nomeLower, "token") || strings.Contains(nomeLower, "key") || strings.Contains(nomeLower, "secret") || strings.Contains(nomeLower, "credencial") {
			if _, ok := n.Inicializador.(*parser.TextoLiteral); ok {
				l.registrarErro(fmt.Sprintf("Potencial vazamento de credencial: Evite expor segredos diretamente no código para a variável '%s'. Utilize variáveis de ambiente.", n.Nome), "HRP-SEC-002", 2, n)
			}
		}

	case *parser.Reatribuicao:
		if id, ok := n.Objeto.(*parser.Identificador); ok {
			if !l.Escopo.Declarada(id.Nome) {
				l.registrarErro(fmt.Sprintf("Variável '%s' não foi declarada", id.Nome), "HRP-0005", 1, n)
			} else if l.Escopo.IsConst(id.Nome) {
				l.registrarErro(fmt.Sprintf("Não é permitido reatribuir valor à constante '%s'", id.Nome), "HRP-0002", 1, n)
			}
		}
		l.Checar(n.Expressao)

	case *parser.DeclFuncao:
		if n.Nome != "" {
			if l.Escopo.DeclaradaLocal(n.Nome) {
				l.registrarErro(fmt.Sprintf("Função '%s' conflita com declaração existente", n.Nome), "HRP-0002", 1, n)
			} else if l.Escopo.Declarada(n.Nome) {
				l.registrarErro(fmt.Sprintf("O identificador da função '%s' está sombreando uma variável externa", n.Nome), "HRP-0002", 2, n)
			}
			l.Escopo.Variaveis[n.Nome] = true
		}

		l.pushEscopo()
		l.Escopo.Assincrono = n.Assincrono
		// Detecta parâmetros com o mesmo nome dentro do mesmo escopo de função
		parametrosVistos := make(map[string]bool, len(n.Parametros))
		for _, param := range n.Parametros {
			if parametrosVistos[param.Nome] {
				l.registrarErro(fmt.Sprintf("Parâmetro '%s' declarado mais de uma vez em '%s'", param.Nome, n.Nome), "HRP-0002", 1, param)
			}
			parametrosVistos[param.Nome] = true

			l.Escopo.Variaveis[param.Nome] = true
			l.Escopo.Usadas[param.Nome] = true // Parâmetros de assinatura de função não devem ser reportados como vars não usadas
			if param.Padrao != nil {
				l.Checar(param.Padrao)
			}
		}
		l.Checar(n.Corpo)
		l.popEscopo()

	case *parser.Bloco:
		l.pushEscopo()
		for _, decl := range n.Declaracoes {
			l.Checar(decl)
		}
		l.popEscopo()

	case *parser.ChamadaFuncao:
		l.Checar(n.Identificador)
		for _, arg := range n.Argumentos {
			l.Checar(arg)
		}

		// 1. Detectar SQL Injection
		var funcNome string
		if id, ok := n.Identificador.(*parser.Identificador); ok {
			funcNome = id.Nome
		} else if acesso, ok := n.Identificador.(*parser.AcessoMembro); ok {
			if idMembro, ok := acesso.Membro.(*parser.Identificador); ok {
				funcNome = idMembro.Nome
			}
		}

		if funcNome == "consultar" || funcNome == "executar" {
			if len(n.Argumentos) > 0 {
				primeiroArg := n.Argumentos[0]
				if l.contemConcatenacaoOuVariavel(primeiroArg) {
					l.registrarErro("Potencial SQL Injection detectado: Evite usar concatenação de strings ou variáveis diretamente na consulta SQL. Use parâmetros preparados ou o Query Builder.", "HRP-SEC-001", 2, primeiroArg)
				}
			}
		}

		// 3. Detectar concorrência insegura em canais fora de funções assíncronas
		if funcNome == "receber" || funcNome == "enviar" {
			if !l.Escopo.NoContextoAssincrono() {
				l.registrarErro("Operação de canal síncrona/bloqueante fora de uma função assíncrona. Isso pode causar o travamento permanente da VM. Use 'assincrono funcao' e 'aguarde'.", "HRP-SEC-003", 2, n)
			}
		}

	case *parser.ArgumentoNomeado:
		l.Checar(n.Valor)

	case *parser.Identificador:
		if !l.Escopo.Declarada(n.Nome) {
			l.registrarErro(fmt.Sprintf("Identificador '%s' não encontrado no escopo", n.Nome), "HRP-0005", 1, n)
		} else {
			l.Escopo.MarcarComoUsada(n.Nome)
		}

	case *parser.OpBinaria:
		l.Checar(n.Esq)
		l.Checar(n.Dir)

	case *parser.OpUnaria:
		l.Checar(n.Expressao)

	case *parser.ImporteDe:
		for _, nome := range n.Nomes {
			l.Escopo.Variaveis[nome] = true
			l.Escopo.Usadas[nome] = true
		}

	case *parser.OpPipe:

		l.Checar(n.Esq)
		l.Checar(n.Dir)

	case *parser.TenteCaptureFinalmente:
		l.Checar(n.TenteBlock)
		if n.CaptureBlock != nil {
			l.pushEscopo()
			l.Escopo.Variaveis[n.NomeErro] = true
			l.Checar(n.CaptureBlock)
			l.popEscopo()
		}
		l.Checar(n.FinalmenteBlock)

	case *parser.DeclTeste:
		l.pushEscopo()
		l.Checar(n.Corpo)
		l.popEscopo()

	case *parser.ExpressaoSe:
		l.Checar(n.Condicao)
		l.Checar(n.Corpo)
		if n.Alternativa != nil {
			l.Checar(n.Alternativa)
		}

	case *parser.Enquanto:
		l.Checar(n.Condicao)
		l.Checar(n.Corpo)

	case *parser.BlocoPara:
		l.pushEscopo()
		l.Escopo.Variaveis[n.Identificador] = true
		l.Checar(n.Iterador)
		l.Checar(n.Corpo)
		l.popEscopo()

	case *parser.NoJSX:
		if n.Tag != "" && len(n.Tag) > 0 && (n.Tag[0] >= 'A' && n.Tag[0] <= 'Z') {
			l.Escopo.MarcarComoUsada(n.Tag)
		}
		for _, attr := range n.Atributos {
			l.Checar(attr.Valor)
		}
		for _, filho := range n.Filhos {
			l.Checar(filho)
		}


	case *parser.NoSeJSX:
		l.Checar(n.Condicao)
		for _, filho := range n.Filhos {
			l.Checar(filho)
		}

	case *parser.NoParaJSX:
		l.pushEscopo()
		l.Escopo.Variaveis[n.Item] = true
		l.Checar(n.Lista)
		for _, filho := range n.Filhos {
			l.Checar(filho)
		}
		l.popEscopo()

	case *parser.DeclEstilo:
		// Não precisa de checagem interna já que as regras são mantidas como texto simples CSS.

	case *parser.DeclEnum:
		l.Escopo.Variaveis[n.Nome] = true
		l.Escopo.Consts[n.Nome] = true

	case *parser.DeclInterface:
		l.Escopo.Variaveis[n.Nome] = true
		l.Escopo.Consts[n.Nome] = true

	case *parser.AcessoMembro:
		l.Checar(n.Dono)

	case *parser.AcessoMembroOpcional:
		l.Checar(n.Objeto)

	case *parser.Indexacao:
		l.Checar(n.Objeto)
		l.Checar(n.Argumento)

	case *parser.NovaNode:
		l.Checar(n.Objeto)

	case *parser.RetorneNode:
		if n.Expressao != nil {
			l.Checar(n.Expressao)
		}

	case *parser.AguardeNode:
		l.Checar(n.Expressao)

	case *parser.DeclVarDestructuring:
		for _, nome := range n.Nomes {
			if l.Escopo.DeclaradaLocal(nome) {
				l.registrarErro(fmt.Sprintf("Variável '%s' já declarada neste escopo", nome), "HRP-0002", 1, n)
			}
			if n.Constante {
				l.Escopo.Consts[nome] = true
			} else {
				l.Escopo.Variaveis[nome] = true
			}
			l.Escopo.Declaracoes[nome] = n
		}
		l.Checar(n.Inicializador)

	case *parser.AsseguraNode:
		l.Checar(n.Condicao)
		if n.Mensagem != nil {
			l.Checar(n.Mensagem)
		}

	case *parser.TuplaLiteral:
		for _, elem := range n.Elementos {
			l.Checar(elem)
		}

	case *parser.ListaLiteral:
		for _, elem := range n.Elementos {
			l.Checar(elem)
		}

	case *parser.MapaLiteral:
		for _, entrada := range n.Entradas {
			if entrada.EhImplicito {
				l.Checar(entrada.Valor)
			} else {
				if _, ok := entrada.Chave.(*parser.Identificador); !ok {
					l.Checar(entrada.Chave)
				}
				l.Checar(entrada.Valor)
			}
		}

	case *parser.TemplateLiteral:
		for _, parte := range n.Partes {
			l.Checar(parte)
		}

	case *parser.TemplateExpr:
		l.Checar(n.Expressao)

	case *parser.OpCoalescenciaNula:
		l.Checar(n.Esq)
		l.Checar(n.Dir)

	case *parser.DeclExportar:
		l.Checar(n.Expressao)

	case *parser.DeclClasse:
		if l.Escopo.DeclaradaLocal(n.Nome) {
			l.registrarErro(fmt.Sprintf("Classe '%s' já declarada neste escopo", n.Nome), "HRP-0002", 1, n)
		}
		l.Escopo.Variaveis[n.Nome] = true
		l.Escopo.Consts[n.Nome] = true
		for _, metodo := range n.Metodos {
			l.Checar(metodo)
		}
	}
}

func (l *Linter) contemConcatenacaoOuVariavel(node parser.BaseNode) bool {
	if node == nil {
		return false
	}
	switch n := node.(type) {
	case *parser.OpBinaria:
		if n.Operador == "+" {
			return true
		}
		return l.contemConcatenacaoOuVariavel(n.Esq) || l.contemConcatenacaoOuVariavel(n.Dir)
	case *parser.Identificador:
		return true
	case *parser.ChamadaFuncao:
		return true
	case *parser.AcessoMembro:
		return true
	}
	return false
}

// ExecutarChecagemSilenciosa executa a análise semântica e linter em um arquivo antes de rodá-lo.
// Retorna a quantidade de erros bloqueantes encontrados (0 = limpo).
func ExecutarChecagemSilenciosa(caminho string) int {
	arquivos, err := encontrarArquivosTeste(caminho)

	if err != nil || len(arquivos) == 0 {
		return 0
	}

	totalErros := 0


	for _, arq := range arquivos {
		conteudo, errRead := os.ReadFile(arq)
		if errRead != nil {
			continue
		}
		ast, err := parser.NewParserFromString(string(conteudo), arq).Parse()
		if err != nil {
			fmt.Printf("\nErro de Sintaxe em %s:\n  ➔ %v\n", arq, err)
			totalErros++
			continue
		}


		dirArq := filepath.Dir(arq)
		linter := &Linter{
			DiretorioAtual:    dirArq,
			ArquivosVisitados: map[string]bool{filepath.Clean(arq): true},
		}
		linter.Checar(ast)

		if len(linter.Erros) > 0 {
			for _, errObj := range linter.Erros {
				if errObj.Severity == 1 { // Erro gravíssimo (impedimento)
					totalErros++
					var line, col int
					if tok, ok := linter.Posicoes[errObj.Node]; ok && tok != nil {
						line = tok.Inicio.Linha
						col = tok.Inicio.Coluna
					}
					fmt.Printf("  ❌ [%s:%d:%d] %s\n", filepath.Base(arq), line, col, errObj.Message)
				}
			}
		}
	}

	return totalErros
}

func comandoChecar() *cobra.Command {
	var formato string
	var estrito bool
	var estritoArquitetura bool

	checar := &cobra.Command{
		Use:   "checar [caminho]",
		Short: "Realiza a checagem semântica/linting estático no código",
		Run: func(cmd *cobra.Command, args []string) {
			caminho := "."
			if len(args) > 0 {
				caminho = args[0]
			}


			arquivos, err := encontrarArquivosTeste(caminho)
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao encontrar arquivos:", err)
				os.Exit(1)
			}

			totalErros := 0
			totalAvisos := 0
			var diagnostics []LSPDiagnostic



			for _, arq := range arquivos {
				conteudo, errRead := os.ReadFile(arq)
				if errRead != nil {
					continue
				}
				ast, err := parser.NewParserFromString(string(conteudo), arq).Parse()
				if err != nil {
					if formato == "json" {
						diagnostics = append(diagnostics, LSPDiagnostic{
							Range:    DiagnosticRange{Start: DiagnosticPosition{Line: 0, Character: 0}, End: DiagnosticPosition{Line: 0, Character: 1}},
							Severity: 1,
							Code:     "HRP-0001",
							Source:   "Harpia-parser",
							Message:  err.Error(),
						})
					} else {
						fmt.Printf("Erro de Sintaxe em %s: %v\n", arq, err)
					}
					totalErros++
					continue
				}


				dirArq := filepath.Dir(arq)
				linter := &Linter{
					DiretorioAtual:    dirArq,
					ArquivosVisitados: map[string]bool{arq: true},
				}
				linter.Checar(ast)

				if estritoArquitetura {
					archErros := checarArquitetura(arq, ast)
					for _, e := range archErros {
						linter.Erros = append(linter.Erros, LinterError{Message: e.Message, Code: e.Code, Severity: 1, Node: e.Node})
					}
				}

				if len(linter.Erros) > 0 {
					for _, errObj := range linter.Erros {
						var line, col, length int
						if tok, ok := linter.Posicoes[errObj.Node]; ok && tok != nil {
							line = tok.Inicio.Linha - 1
							col = tok.Inicio.Coluna - 1
							length = len(tok.Valor)
							if length == 0 {
								length = 1
							}
						}

						if errObj.Severity == 2 {
							totalAvisos++
						} else {
							totalErros++
						}

						if formato == "json" {
							diagnostics = append(diagnostics, LSPDiagnostic{
								Range:    DiagnosticRange{Start: DiagnosticPosition{Line: line, Character: col}, End: DiagnosticPosition{Line: line, Character: col + length}},
								Severity: errObj.Severity,
								Code:     errObj.Code,
								Source:   "Harpia-linter",
								Message:  errObj.Message,
							})
						} else {
							icon := "❌"
							if errObj.Severity == 2 {
								icon = "⚠️"
							}
							fmt.Printf("  %s  [%s:%d:%d] %s\n", icon, filepath.Base(arq), line+1, col+1, errObj.Message)
						}
					}
				}
			}

			if formato == "json" {
				bytes, _ := json.MarshalIndent(diagnostics, "", "  ")
				fmt.Println(string(bytes))
				if totalErros > 0 {
					os.Exit(1)
				}
				return
			}

			if totalErros > 0 || totalAvisos > 0 {
				fmt.Printf("\nChecagem concluída com %d erro(s) e %d aviso(s) semântico(s).\n", totalErros, totalAvisos)
				if totalErros > 0 {
					os.Exit(1)
				}
			} else {
				fmt.Println("Nenhum problema semântico encontrado.")
			}
		},
	}

	checar.Flags().StringVar(&formato, "formato", "texto", "Formato de saída dos erros: 'texto' ou 'json'")
	checar.Flags().BoolVar(&estrito, "estrito", false, "Ativa a validação estrita de tipos na análise estática")
	checar.Flags().BoolVar(&estritoArquitetura, "estrito-arquitetura", false, "Ativa validação de Clean Architecture: camadas não podem importar de camadas superiores (dominio<-infra<-web).")
	return checar
}

// checarArquitetura valida regras de Clean Architecture com base nos imports relativos.
// dominio/ e subpastas: pode importar apenas de si mesmo e da stdlib (sem caminho relativo).
// infra/ e subpastas: não pode importar de web/.
// web/ e subpastas: pode importar de qualquer camada.
func checarArquitetura(arq string, node parser.BaseNode) []LinterError {
	var erros []LinterError
	arqNorm := filepath.ToSlash(filepath.Clean(arq))

	camada := ""
	switch {
	case strings.HasPrefix(arqNorm, "dominio/"):
		camada = "dominio"
	case strings.HasPrefix(arqNorm, "infra/"):
		camada = "infra"
	case strings.HasPrefix(arqNorm, "web/"):
		camada = "web"
	default:
		return nil
	}

	prog, ok := node.(*parser.Programa)
	if !ok {
		return nil
	}
	for _, decl := range prog.Declaracoes {
		imp, ok := decl.(*parser.ImporteDe)
		if !ok || imp.Caminho == nil {
			continue
		}
		caminho := ""
		if imp.Caminho != nil {
			caminho = imp.Caminho.Valor
		}
		if !strings.HasPrefix(caminho, "../") && !strings.HasPrefix(caminho, "./") && !strings.HasPrefix(caminho, "/") {
			continue
		}
		switch camada {
		case "dominio":
			erros = append(erros, LinterError{
				Message: fmt.Sprintf("HRP-ARCH-001: camada 'dominio' não pode importar '%s' (use 'infra' para dependências externas)", caminho),
				Code:    "HRP-ARCH-001",
				Node:    imp,
			})
		case "infra":
			if strings.Contains(caminho, "/web/") || strings.Contains(caminho, "../web") || strings.Contains(caminho, "web/") || strings.HasSuffix(strings.TrimPrefix(caminho, "./"), "web") {
				erros = append(erros, LinterError{
					Message: fmt.Sprintf("HRP-ARCH-002: camada 'infra' não pode importar da camada 'web' (%s)", caminho),
					Code:    "HRP-ARCH-002",
					Node:    imp,
				})
			}
		}
	}
	return erros
}
