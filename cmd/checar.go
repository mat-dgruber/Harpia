package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natanfeitosa/portuscript/lexer"
	"github.com/natanfeitosa/portuscript/parser"
	"github.com/natanfeitosa/portuscript/ptst"
	"github.com/spf13/cobra"
)

// globalsLinter define o conjunto de nomes globais pré-declarados pela VM e pela stdlib
// do Portuscript. Cada vez que novos built-ins forem adicionados, sincronize este mapa
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
	"Objeto":  true,
	"Erro":    true,

	// Primitivas Reativas da Web
	"sinal":    true,
	"efeito":   true,
	"derivado": true,
	"armazem":  true,

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
}

type EscopoLinter struct {
	Pai       *EscopoLinter
	Variaveis map[string]bool
	Consts    map[string]bool
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

func (e *EscopoLinter) IsConst(nome string) bool {
	if e.Consts[nome] {
		return true
	}
	if e.Pai != nil {
		return e.Pai.IsConst(nome)
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
	Erros    []LinterError
	Escopo   *EscopoLinter
	Posicoes map[parser.BaseNode]*lexer.Token
}

func (l *Linter) pushEscopo() {
	l.Escopo = &EscopoLinter{
		Pai:       l.Escopo,
		Variaveis: make(map[string]bool),
		Consts:    make(map[string]bool),
	}
}

func (l *Linter) popEscopo() {
	if l.Escopo != nil {
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

	switch n := node.(type) {
	case *parser.Programa:
		l.Posicoes = n.Posicoes
		l.pushEscopo()
		// Hidrata o escopo raiz com built-ins e nomes globais conhecidos
		for nome := range globalsLinter {
			l.Escopo.Variaveis[nome] = true
		}

		for _, decl := range n.Declaracoes {
			l.Checar(decl)
		}
		l.popEscopo()

	case *parser.DeclVar:
		if l.Escopo.DeclaradaLocal(n.Nome) {
			l.registrarErro(fmt.Sprintf("Variável '%s' já declarada neste escopo", n.Nome), "HRP-0002", 1, n)
		} else if l.Escopo.Declarada(n.Nome) {
			l.registrarErro(fmt.Sprintf("O identificador '%s' está sombreando (shadowing) uma variável externa", n.Nome), "HRP-0002", 2, n)
		}
		if n.Constante {
			l.Escopo.Consts[n.Nome] = true
		} else {
			l.Escopo.Variaveis[n.Nome] = true
		}
		l.Checar(n.Inicializador)

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
		if l.Escopo.DeclaradaLocal(n.Nome) {
			l.registrarErro(fmt.Sprintf("Função '%s' conflita com declaração existente", n.Nome), "HRP-0002", 1, n)
		} else if l.Escopo.Declarada(n.Nome) {
			l.registrarErro(fmt.Sprintf("O identificador da função '%s' está sombreando uma variável externa", n.Nome), "HRP-0002", 2, n)
		}
		l.Escopo.Variaveis[n.Nome] = true

		l.pushEscopo()
		// Detecta parâmetros com o mesmo nome dentro do mesmo escopo de função
		parametrosVistos := make(map[string]bool, len(n.Parametros))
		for _, param := range n.Parametros {
			if parametrosVistos[param.Nome] {
				l.registrarErro(fmt.Sprintf("Parâmetro '%s' declarado mais de uma vez em '%s'", param.Nome, n.Nome), "HRP-0002", 1, param)
			}
			parametrosVistos[param.Nome] = true

			l.Escopo.Variaveis[param.Nome] = true
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

	case *parser.ArgumentoNomeado:
		l.Checar(n.Valor)

	case *parser.Identificador:
		if !l.Escopo.Declarada(n.Nome) {
			l.registrarErro(fmt.Sprintf("Identificador '%s' não encontrado no escopo", n.Nome), "HRP-0005", 1, n)
		}

	case *parser.OpBinaria:
		l.Checar(n.Esq)
		l.Checar(n.Dir)

	case *parser.OpUnaria:
		l.Checar(n.Expressao)

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
	}
}

func comandoChecar() *cobra.Command {
	var formato string
	var estrito bool

	checar := &cobra.Command{
		Use:   "checar [caminho]",
		Short: "Realiza a checagem semântica/linting estático no código",
		Run: func(cmd *cobra.Command, args []string) {
			cur, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "erro ao obter diretório atual:", err)
				os.Exit(1)
			}

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
			ctx := ptst.NewContexto(ptst.OpcsContexto{CaminhosPadrao: []string{cur}})
			defer ctx.Terminar()

			var diagnostics []LSPDiagnostic

			for _, arq := range arquivos {
				_, ast, err := ctx.TransformarEmAst(arq, false, cur)
				if err != nil {
					if formato == "json" {
						diagnostics = append(diagnostics, LSPDiagnostic{
							Range:    DiagnosticRange{Start: DiagnosticPosition{Line: 0, Character: 0}, End: DiagnosticPosition{Line: 0, Character: 1}},
							Severity: 1,
							Code:     "HRP-0001",
							Source:   "portuscript-parser",
							Message:  err.Error(),
						})
					} else {
						fmt.Printf("Erro de Sintaxe em %s: %v\n", arq, err)
					}
					totalErros++
					continue
				}

				linter := &Linter{}
				linter.Checar(ast)

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
								Source:   "portuscript-linter",
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
	return checar
}
