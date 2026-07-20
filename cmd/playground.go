package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/spf13/cobra"
)

type ExecucaoPlaygroundRequest struct {
	Codigo string `json:"codigo"`
}

type VariavelEscopo struct {
	Nome  string `json:"nome"`
	Valor string `json:"valor"`
	Tipo  string `json:"tipo"`
}

type ExecucaoPlaygroundResponse struct {
	Sucesso   bool             `json:"sucesso"`
	Saida     string           `json:"saida"`
	Variaveis []VariavelEscopo `json:"variaveis"`
	ErroHtml  string           `json:"erroHtml,omitempty"`
}

// ponytail: garante segurança concorrente contra corridas de dados no Stdout compartilhado
var mutexExecucaoPlayground sync.Mutex

// comandoPlayground inicia o servidor web do playground interativo local
func comandoPlayground() *cobra.Command {
	var porta int
	cmdPlay := &cobra.Command{
		Use:   "playground",
		Short: "Inicia o servidor web local do Playground Interativo do Harpia",
		Run: func(cmd *cobra.Command, args []string) {
			iniciarServidorPlayground(porta)
		},
	}
	cmdPlay.Flags().IntVarP(&porta, "porta", "p", 8090, "Porta de escuta do servidor do playground")
	return cmdPlay
}

func iniciarServidorPlayground(porta int) {
	http.HandleFunc("/", serveInterfacePlayground)
	http.HandleFunc("/runtime-web.js", serveRuntimeWeb)
	http.HandleFunc("/playground.js", servePlaygroundJS)
	http.HandleFunc("/playground.css", servePlaygroundCSS)
	http.HandleFunc("/api/executar", apiExecutarCodigoPlayground)
	// ponytail: endpoints estáticos que alimentam o editor Monaco no cliente (highlight + hover)
	http.HandleFunc("/api/editor-config", apiEditorConfig)
	http.HandleFunc("/api/docs", apiDocsPlayground)
	http.HandleFunc("/editor-monaco.js", serveEditorMonacoJS)

	fmt.Printf("🚀 Playground Interativo do Harpia rodando em: http://localhost:%d\n", porta)
	fmt.Println("Pressione Ctrl+C para encerrar o servidor.")

	// ponytail: servidor HTTP resiliente com timeouts para mitigar DoS/estouro de conexões
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", porta),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao iniciar o servidor do playground: %v\n", err)
		os.Exit(1)
	}
}

func serveInterfacePlayground(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlInterfacePlayground))
}

func serveRuntimeWeb(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	content, err := os.ReadFile("stdlib/web/runtime-web.js")
	if err != nil {
		content, err = os.ReadFile("../stdlib/web/runtime-web.js")
		if err != nil {
			http.Error(w, "Runtime não encontrado", http.StatusInternalServerError)
			return
		}
	}
	w.Write(content)
}

func servePlaygroundJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	conteudo, err := os.ReadFile("playground/interface.hrp")
	if err != nil {
		conteudo, err = os.ReadFile("Harpia/playground/interface.hrp")
		if err != nil {
			http.Error(w, "Interface não encontrada", http.StatusInternalServerError)
			return
		}
	}

	// ponytail: blindagem contra Carriage Returns (\r) do Windows que quebram o lexer
	conteudoStr := strings.ReplaceAll(string(conteudo), "\r", "")

	ast, err := ctx.StringParaAst(conteudoStr, "playground/interface.hrp")
	if err != nil {
		fmt.Printf("❌ [Harpia Server] Erro de compilação do JS do Playground: %v\n", err)
		http.Error(w, fmt.Sprintf("Erro de compilação: %v", err), http.StatusInternalServerError)
		return
	}

	transpiler := &TranspilerWeb{
		Estrito:       false,
		DiretorioBase: "playground",
	}
	jsOutput := transpiler.Transpile(ast)
	finalJS := "import { h, efeito, derivado, armazem, montar } from './runtime-web.js';\n\n" + jsOutput
	w.Write([]byte(finalJS))
}

func servePlaygroundCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	conteudo, err := os.ReadFile("playground/interface.hrp")
	if err != nil {
		conteudo, err = os.ReadFile("Harpia/playground/interface.hrp")
		if err != nil {
			w.Write([]byte("/* Estilo não encontrado */"))
			return
		}
	}

	// ponytail: blindagem contra Carriage Returns (\r) do Windows que quebram o lexer
	conteudoStr := strings.ReplaceAll(string(conteudo), "\r", "")

	ast, err := ctx.StringParaAst(conteudoStr, "playground/interface.hrp")
	if err != nil {
		fmt.Printf("❌ [Harpia Server] Erro de compilação do CSS do Playground: %v\n", err)
		w.Write([]byte("/* Erro de compilação do estilo */"))
		return
	}

	transpiler := &TranspilerWeb{
		Estrito:       false,
		DiretorioBase: "playground",
	}
	transpiler.Transpile(ast)

	cssOutput := strings.Join(transpiler.Styles, "\n")

	// ponytail: CSS dark profissional de alta performance injetado diretamente para dar um visual VS Code impecável
	cssLayoutDark := `
body {
    margin: 0;
    font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background-color: #020617; /* bg-slate-950 */
    color: #f1f5f9; /* text-slate-100 */
    height: 100vh;
    overflow: hidden;
}
.flex { display: flex; }
.flex-col { flex-direction: column; }
.flex-1 { flex: 1 1 0%; }
.grid { display: grid; }
.grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.grid-rows-2 { grid-template-rows: repeat(2, minmax(0, 1fr)); }
.h-full { height: 100%; }
.w-full { width: 100%; }
.overflow-hidden { overflow: hidden; }
.overflow-y-auto { overflow-y: auto; }
.gap-px { gap: 1px; }

header {
    background-color: #0f172a; /* bg-slate-900 */
    border-bottom: 1px solid #1e293b;
    padding: 12px 24px;
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.text-2xl { font-size: 1.5rem; line-height: 2rem; }
.font-black { font-weight: 900; }
.text-blue-500 { color: #3b82f6; }
.bg-slate-900 { background-color: #0f172a; }
.border-slate-800 { border-color: #1e293b; }

button {
    cursor: pointer;
    border: none;
    outline: none;
    display: inline-flex;
    align-items: center;
}
.bg-blue-600 { background-color: #2563eb; color: white; }
.bg-blue-600:hover { background-color: #1d4ed8; }
.font-bold { font-weight: 700; }
.px-6 { padding-left: 24px; padding-right: 24px; }
.py-2 { padding-top: 8px; padding-bottom: 8px; }
.rounded-lg { border-radius: 8px; }
.transition { transition-property: all; transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1); transition-duration: 150ms; }

section {
    background-color: #0f172a;
}
.border-b { border-bottom: 1px solid #1e293b; }
.px-4 { padding-left: 16px; padding-right: 16px; }
.py-2 { padding-top: 8px; padding-bottom: 8px; }
.text-xs { font-size: 0.75rem; line-height: 1rem; }
.text-slate-400 { color: #94a3b8; }
.tracking-wider { letter-spacing: 0.05em; }

pre {
    margin: 0;
    padding: 24px;
    background-color: #090d16;
    font-family: Fira Code, Courier New, monospace;
    font-size: 0.875rem;
    color: #e2e8f0;
    line-height: 1.5;
    white-space: pre-wrap;
    overflow-y: auto;
}
.text-red-300 { color: #fca5a5; }
.bg-red-950\/40 { background-color: rgba(69, 10, 10, 0.4); }
.border-red-900 { border-color: #7f1d1d; }
.p-4 { padding: 16px; }

/* Grade de Dados do Linter */
.grade-dados-container {
    padding: 16px;
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
}
.grade-dados-pesquisa {
    background-color: #1e293b;
    border: 1px solid #334155;
    color: #f1f5f9;
    padding: 8px 12px;
    border-radius: 6px;
    margin-bottom: 12px;
    font-size: 0.875rem;
    outline: none;
}
.grade-dados-tabela {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
}
.grade-dados-tabela th, .grade-dados-tabela td {
    padding: 8px 12px;
    text-align: left;
    border-bottom: 1px solid #1e293b;
}
.grade-dados-tabela th {
    font-weight: bold;
    color: #94a3b8;
    background-color: #0f172a;
}
.grade-dados-tabela td {
    color: #10b981; /* emerald-400 */
}
.grade-dados-paginacao {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 12px;
    font-size: 0.75rem;
}
.grade-dados-paginacao button {
    background-color: #1e293b;
    color: #f1f5f9;
    padding: 6px 12px;
    border-radius: 4px;
}
.grade-dados-paginacao button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}
`
	w.Write([]byte(cssLayoutDark + "\n" + cssOutput))
}

func apiExecutarCodigoPlayground(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		return
	}

	// ponytail: limita payload do editor a no máximo 1MB contra ataques de DoS/estouro de buffer
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	// ponytail: serializa execuções concorrentes do terminal local para evitar Race Conditions em os.Stdout
	mutexExecucaoPlayground.Lock()
	defer mutexExecucaoPlayground.Unlock()

	var req ExecucaoPlaygroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Captura stdout para coletar impressões de logs do console
	oldStdout := os.Stdout
	readP, writeP, _ := os.Pipe()
	os.Stdout = writeP

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	escopo := hrp.NewEscopo()

	// Compila a string de código para AST
	ast, err := ctx.StringParaAst(req.Codigo, "playground.hrp")

	var response ExecucaoPlaygroundResponse
	if err != nil {
		// Fecha os pipes e restaura a saída padrão
		writeP.Close()
		os.Stdout = oldStdout

		fmt.Printf("⚠️ [Harpia Server] Erro sintático no editor: %v\n", err)

		response.Sucesso = false
		response.ErroHtml = ansiParaHtml(err.Error())
	} else {
		// Executa o script sob o escopo isolado
		_, errExec := ctx.AvaliarAst(ast, escopo)

		if errExec != nil {
			fmt.Printf("⚠️ [Harpia Server] Erro de execução na VM: %v\n", errExec)
		}

		// Fecha a gravação e lê o buffer capturado de Stdout
		writeP.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(readP)
		response.Saida = buf.String()

		if errExec != nil {
			response.Sucesso = false
			response.ErroHtml = ansiParaHtml(errExec.Error())
		} else {
			response.Sucesso = true
		}

		// Coleta variáveis do escopo local pós-execução de forma segura
		simbolos := escopo.ObterSimbolosSeguro()
		var vars []VariavelEscopo
		for _, simb := range simbolos {
			if simb != nil {
				// Ignora built-ins internos de inicialização
				if strings.HasPrefix(simb.Nome, "_") {
					continue
				}
				valStr := fmt.Sprintf("%v", simb.ObterValor())
				tipoStr := simb.Tipo
				if tipoStr == "" {
					tipoStr = "Dinâmico"
				}
				vars = append(vars, VariavelEscopo{
					Nome:  simb.Nome,
					Valor: valStr,
					Tipo:  tipoStr,
				})
			}
		}
		response.Variaveis = vars
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ansiParaHtml(ansi string) string {
	res := ansi
	res = strings.ReplaceAll(res, "\n", "<br/>")
	res = strings.ReplaceAll(res, " ", "&nbsp;")
	res = strings.ReplaceAll(res, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;")
	// Substituições simples de cores ANSI comuns para HTML formatado do playground
	res = strings.ReplaceAll(res, "\u001b[31m", "<span class='text-red-400 font-bold'>")    // vermelho
	res = strings.ReplaceAll(res, "\u001b[33m", "<span class='text-yellow-400 font-bold'>") // amarelo
	res = strings.ReplaceAll(res, "\u001b[32m", "<span class='text-green-400'>")            // verde
	res = strings.ReplaceAll(res, "\u001b[36m", "<span class='text-cyan-400'>")             // ciano
	res = strings.ReplaceAll(res, "\u001b[35m", "<span class='text-pink-400'>")             // magenta
	res = strings.ReplaceAll(res, "\u001b[1m", "<span class='font-bold'>")                  // negrito
	res = strings.ReplaceAll(res, "\u001b[0m", "</span>")                                   // reset
	return res
}

// ponytail: implementação completa do suporte ao Monaco Editor para coloração, hover e inicialização de código
func apiEditorConfig(w http.ResponseWriter, r *http.Request) {
	// Tenta ler app_teste.hrp de forma resiliente
	caminhos := []string{"app_teste.hrp", "Harpia/app_teste.hrp", "../app_teste.hrp"}
	var defaultCode []byte
	var err error

	for _, cam := range caminhos {
		defaultCode, err = os.ReadFile(cam)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Fallback robusto se o arquivo não for encontrado
		defaultCode = []byte(`// Exemplo de App Portuscript SPA Reativo
estilo MeuApp {
    corDeFundo: "#f4f4f9";
    padding: "20px";
}

funcao MeuApp() {
    var contadorSinal = sinal(10);
    var contador = contadorSinal[0];
    var definirContador = contadorSinal[1];

    var incrementar = funcao() {
        definirContador(contador() + 1);
    }

    retorne <div classe="MeuApp">
        <h1>Meu App Harpia</h1>
        <p>Contador atual: <strong>{contador()}</strong></p>
        <button aoClicar={incrementar}>Incrementar (+)</button>
    </div>
}`)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"defaultCode": string(defaultCode),
		"contador": `var contadorSinal = sinal(10)
var contador = contadorSinal[0]
var setContador = contadorSinal[1]

funcao incrementar() {
    setContador(contador() + 1)
}

# Incrementando o contador reativo
incrementar()
incrementar()

imprimir("Contador reativo inicializado em: " + contador())`,
		"fibonacci": `funcao fibonacci(n) {
    se (n <= 1) {
        retorne n
    }
    retorne fibonacci(n - 1) + fibonacci(n - 2)
}

imprimir("Fibonacci de 8 termos:")
imprimir("fib(1) = " + fibonacci(1))
imprimir("fib(2) = " + fibonacci(2))
imprimir("fib(3) = " + fibonacci(3))
imprimir("fib(4) = " + fibonacci(4))
imprimir("fib(5) = " + fibonacci(5))
imprimir("fib(6) = " + fibonacci(6))
imprimir("fib(7) = " + fibonacci(7))
imprimir("fib(8) = " + fibonacci(8))`,
		"classes": `classe Usuario {
    var nome
    var email

    funcao inicializar(n, e) {
        self.nome = n
        self.email = e
    }

    funcao apresentar() {
        imprimir("Olá! Meu nome é " + self.nome + " e meu email é " + self.email)
    }
}

var user = Usuario("Guilherme", "guilherme@harpia.dev")
user.apresentar()`,
	})
}

func apiDocsPlayground(w http.ResponseWriter, r *http.Request) {
	palavra := r.URL.Query().Get("palavra")
	docs := map[string]string{
		"sinal":      "### sinal(valor)\n\nCria um **Sinal Reativo** (Fine-Grained). Retorna uma lista de dois elementos:\n- `sinal[0]`: A função de leitura (getter) para obter o valor atual.\n- `sinal[1]`: A função de escrita (setter) para atualizar o valor.\n\n**Exemplo:**\n```portuscript\nvar contSinal = sinal(0);\nvar cont = contSinal[0];\nvar setCont = contSinal[1];\n```",
		"funcao":     "### funcao() / func()\n\nDeclara uma nova função ou bloco executável autônomo. Pode receber parâmetros e retornar valores usando `retorne`.\n\n**Exemplo:**\n```portuscript\nvar dobrar = funcao(n) {\n    retorne n * 2;\n}\n```",
		"func":       "### funcao() / func()\n\nDeclara uma nova função ou bloco executável autônomo. Pode receber parâmetros e retornar valores usando `retorne`.\n\n**Exemplo:**\n```portuscript\nvar dobrar = func(n) {\n    retorne n * 2;\n}\n```",
		"estilo":     "### estilo Nome {\n    propriedade: valor;\n}\n\nDeclara um bloco de estilos CSS reativo associado de forma exclusiva a uma classe de elemento no template.\n\n**Exemplo:**\n```portuscript\nestilo BotaoDestaque {\n    corDeFundo: \"#3b82f6\";\n    padding: \"8px 16px\";\n}\n```",
		"aoClicar":   "### aoClicar={funcao}\n\nAtributo de evento especial para vincular uma função ao clique do mouse em elementos HTML/JSX no Harpia.\n\n**Exemplo:**\n```portuscript\n<button aoClicar={incrementar}>Clique-me</button>\n```",
		"retorne":    "### retorne\n\nEncerra a execução da função atual e retorna o controle e o valor especificado para o chamador.",
		"se":         "### se condicao {\n    ...\n}\n\nEstrutura de controle condicional básica. Executa o bloco associado caso a condição seja avaliada como verdadeira.",
		"senao":      "### senao\n\nEstrutura complementar de decisão condicional. Executa o bloco caso a condição do `se` precedente falhe.",
		"enquanto":   "### enquanto condicao {\n    ...\n}\n\nCria um loop que executa continuamente o bloco de código enquanto a condição lógica especificada for verdadeira.",
		"para":       "### para item em colecao {\n    ...\n}\n\nEstrutura de repetição (iteração) para percorrer cada item em uma coleção ou lista especificada.",
		"var":        "### var\n\nDeclara uma nova variável mutável no escopo atual.",
		"const":      "### const / constante\n\nDeclara uma constante local de somente leitura. O valor atribuído não pode ser alterado posteriormente.",
		"constante":  "### const / constante\n\nDeclara uma constante local de somente leitura. O valor atribuído não pode ser alterado posteriormente.",
		"classe":     "### classe Nome {\n    ...\n}\n\nDefine uma classe de Objeto para programação orientada a objetos (POO), encapsulando propriedades e métodos.",
		"importe":    "### de \"modulo\" importe recurso;\n\nImporta funções ou dados de um módulo específico da biblioteca padrão do Harpia.",
		"de":         "### de \"modulo\" importe recurso;\n\nParte da instrução de importação para especificar o módulo de origem (ex: de \"web\" importe sinal;).",
		"Verdadeiro": "### Verdadeiro\n\nValor lógico booleano verdadeiro.",
		"Falso":      "### Falso\n\nValor lógico booleano falso.",
		"Nulo":       "### Nulo\n\nRepresenta a ausência intencional de qualquer valor de objeto (similar a null/nil).",
	}

	docText, ok := docs[palavra]
	if !ok {
		docText = fmt.Sprintf("### %s\n\nIdentificador ou palavra-chave do ecossistema Portuscript / Harpia.", palavra)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"doc": docText,
	})
}

func serveEditorMonacoJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")

	// Usamos strings.ReplaceAll para injetar as crases (backticks) de JS de forma limpa sem conflitar com Go
	jsCode := `
// ponytail: JS inline resiliente carregando Monaco, Monarch e estabelecendo ponte reativa bidirecional
(function() {
	require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs' } });

	require(['vs/editor/editor.main'], function() {
		// 1. Registra a linguagem Portuscript
		monaco.languages.register({ id: 'portuscript' });

		// 2. Configura o analisador léxico Monarch para coloração sintática perfeita
		monaco.languages.setMonarchTokensProvider('portuscript', {
			keywords: [
				'se', 'senao', 'enquanto', 'para', 'retorne', 'pare', 'continue',
				'de', 'importe', 'Verdadeiro', 'Falso', 'Nulo',
				'var', 'const', 'constante', 'func', 'funcao',
				'ou', 'e', 'nao', 'nova', 'classe', 'estende', 'self', 'estatico',
				'assegura', 'testar', 'tente', 'capture', 'finalmente', 'exportar',
				'em', 'assincrono', 'aguarde', 'estilo'
			],
			operators: [
				'=', '+', '-', '*', '**', '/', '//', '%', '<', '<=', '==', '!=', '>', '>=',
				'+=', '-=', '*=', '/=', '//=', '|', '^', '&', '~', '<<', '>>', '.', '|>'
			],
			symbols: /[=><!~?:&|+\-*\/\^%]+/,
			escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

			tokenizer: {
				root: [
					// Identificadores especiais (sinal e atributos de evento do JSX)
					[/(sinal|aoClicar|classe|condicao)\b/, 'keyword'],

					// Palavras-chave normais e identificadores
					[/[a-zA-ZÀ-ÿ_][a-zA-ZÀ-ÿ0-9_]*/, {
						cases: {
							'@keywords': 'keyword',
							'@default': 'identifier'
						}
					}],

					// Espaços em branco
					{ include: '@whitespace' },

					// Delimitadores e chaves
					[/[{}()\[\]]/, '@brackets'],

					// Operadores e delimitadores simples
					[/@symbols/, {
						cases: {
							'@operators': 'operator',
							'@default': ''
						}
					}],

					// Números (inteiros e decimais)
					[/\d*\.\d+([eE][\-+]?\d+)?/, 'number'],
					[/\d+/, 'number'],

					// Strings
					[/"([^"\\]|\\.)*$/, 'string.invalid'],  // string não terminada
					[/'([^'\\]|\\.)*$/, 'string.invalid'],
					[/"/,  { token: 'string.quote', bracket: '@open', next: '@string_double' }],
					[/'/,  { token: 'string.quote', bracket: '@open', next: '@string_single' }],
				],

				string_double: [
					[/[^\\"]+/,  'string'],
					[/@escapes/, 'string.escape'],
					[/\\./,      'string.escape.invalid'],
					[/"/,        { token: 'string.quote', bracket: '@close', next: '@pop' }]
				],

				string_single: [
					[/[^\\']+/,  'string'],
					[/@escapes/, 'string.escape'],
					[/\\./,      'string.escape.invalid'],
					[/'/,        { token: 'string.quote', bracket: '@close', next: '@pop' }]
				],

				whitespace: [
					[/[ \t\r\n]+/, ''],
					[/#.*$/, 'comment'],
					[/<!--/, 'comment', '@comment_html']
				],

				comment_html: [
					[/[^<\-]+/, 'comment'],
					[/-->/, 'comment', '@pop'],
					[/[<\-]/, 'comment']
				]
			}
		});

		// 3. Define as regras de pareamento de parênteses, colchetes e chaves
		monaco.languages.setLanguageConfiguration('portuscript', {
			brackets: [
				['{', '}'],
				['[', ']'],
				['(', ')']
			],
			autoClosingPairs: [
				{ open: '{', close: '}' },
				{ open: '[', close: ']' },
				{ open: '(', close: ')' },
				{ open: '"', close: '"' },
				{ open: '\'', close: '\'' }
			]
		});

		// 4. Registra o HoverProvider para explicações interativas (DX)
		monaco.languages.registerHoverProvider('portuscript', {
			provideHover: function(model, position) {
				var word = model.getWordAtPosition(position);
				if (!word) return null;

				return fetch('/api/docs?palavra=' + encodeURIComponent(word.word))
					.then(function(res) { return res.json(); })
					.then(function(dados) {
						return {
							range: new monaco.Range(position.lineNumber, word.startColumn, position.lineNumber, word.endColumn),
							contents: [
								{ value: dados.doc }
							]
						};
					})
					.catch(function() { return null; });
			}
		});

		// 5. Configura o Tema customizado de cores (One Dark adaptado para Slate Tailwind)
		monaco.editor.defineTheme('harpia-dark', {
			base: 'vs-dark',
			inherit: true,
			rules: [
				{ token: 'keyword',       foreground: 'c678dd', fontStyle: 'bold' },  // roxo elegante
				{ token: 'string',        foreground: '98c379' },                     // verde esmeralda
				{ token: 'number',        foreground: 'd19a66' },                     // laranja
				{ token: 'comment',       foreground: '7f848e', fontStyle: 'italic' },// cinza
				{ token: 'operator',      foreground: '56b6c2' },                     // ciano
				{ token: 'identifier',    foreground: '61afef' },                     // azul claro de variáveis
				{ token: 'delimiter.bracket', foreground: 'e5c07b' },                // chaves em amarelo suave
				{ token: 'delimiter',     foreground: 'e5c07b' }                      // delimitadores comuns em amarelo
			],
			colors: {
				'editor.background': '#0f172a',                     // bg-slate-950 perfeitamente alinhado
				'editor.foreground': '#f1f5f9',                     // slate-100
				'editor.lineHighlightBackground': '#1e293b',        // slate-800
				'editor.selectionBackground': '#334155',            // slate-700
				'editorCursor.foreground': '#3b82f6',               // azul principal
				'editorLineNumber.foreground': '#475569',           // slate-600
				'editorLineNumber.activeForeground': '#94a3b8'     // slate-400
			}
		});

		// 6. Expõe uma função global para inicializar o editor quando o mount point no VDOM estiver pronto
		window.__initHarpiaEditor = function(mountId, initialCode, onChange, executar) {
			var container = document.getElementById(mountId);
			if (!container) return;

			// Limpa conteúdo pré-existente
			container.innerHTML = '';

			var editor = monaco.editor.create(container, {
				value: initialCode,
				language: 'portuscript',
				theme: 'harpia-dark',
				automaticLayout: true,
				fontSize: 15,
				fontFamily: 'Fira Code, Courier New, monospace',
				minimap: { enabled: false },
				scrollBeyondLastLine: false,
				lineHeight: 24,
				padding: { top: 12, bottom: 12 }
			});

			window.__psiEditor = editor;

			// Lock de segurança contra loops reativos
			var lock = false;

			editor.onDidChangeModelContent(function() {
				if (lock) return;
				lock = true;
				if (typeof onChange === 'function') {
					onChange(editor.getValue());
				}
				lock = false;
			});

			window.__psiSetCodigo = function(newVal) {
				if (lock) return;
				if (editor.getValue() !== newVal) {
					lock = true;
					editor.setValue(newVal);
					lock = false;
				}
			};

			// Atalho físico de execução Ctrl+Enter
			editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, function() {
				if (typeof executar === 'function') {
					executar();
				}
			});
		};

		// ponytail: ponte global de rede para execução do playground (livre de bugs sintáticos do linter Go)
		window.executarCodigoHarpia = function(codigo, callback) {
			fetch("/api/executar", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ codigo: codigo })
			})
			.then(function(res) { return res.json(); })
			.then(function(dados) {
				if (typeof callback === 'function') {
					callback(dados);
				}
			})
			.catch(function(err) {
				if (typeof callback === 'function') {
					callback({ saida: "Erro de conexão: " + err, erroHtml: "", variaveis: [] });
				}
			});
		};

		// Processa fila de pontes registradas pelo Harpia que estavam aguardando o Monaco carregar
		if (window.__psiPontesPending && window.__psiPontesPending.length > 0) {
			window.__psiPontesPending.forEach(function(ponte) {
				fetch("/api/editor-config")
					.then(function(res) { return res.json(); })
					.then(function(dados) {
						window.__psiExemplos = dados;
						ponte.setCodigo(dados.defaultCode);
						window.__initHarpiaEditor("editor-harpia-container", dados.defaultCode, ponte.setCodigo, ponte.executar);
					});
			});
		}
	});
})();
`
	jsCode = strings.ReplaceAll(jsCode, "BACKTICK", "`")
	w.Write([]byte(jsCode))
}

// htmlInterfacePlayground define a interface física do editor web em Dark Mode corporativo
const htmlInterfacePlayground = `<!DOCTYPE html>
<html lang="pt-BR" class="h-full bg-slate-950 text-slate-100">
<head>
    <meta charset="UTF-8">
    <title>Harpia — Playground Interativo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/playground.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs/editor/editor.main.css">
    <style>
        .editor-container { width: 100%; height: 100%; min-height: 200px; }
        .editor-textarea { font-family: 'Fira Code', 'Courier New', monospace; tab-size: 4; }
        .monaco-hover { padding: 0 12px !important; }
        .monaco-hover code { font-family: 'Fira Code', monospace; }
    </style>
    <script>
        // ponytail: ponte global resiliente contra corridas de carregamento assíncrono entre o SPA e o Monaco CDN
        window.Verdadeiro = true;
        window.Falso = false;
        window.Nulo = null;

        window.__psiExemplos = {};
        window.carregarExemploNoMonaco = function(nome) {
            var codigo = window.__psiExemplos[nome];
            if (codigo && window.__psiSetCodigo) {
                window.__psiSetCodigo(codigo);
            }
        };

        window.__psiPontesPending = [];
        window.registrarPontePlayground = function(setCodigo, executar) {
            if (window.__initHarpiaEditor) {
                fetch("/api/editor-config")
                    .then(res => res.json())
                    .then(dados => {
                        window.__psiExemplos = dados;
                        setCodigo(dados.defaultCode);
                        window.__initHarpiaEditor("editor-harpia-container", dados.defaultCode, setCodigo, executar);
                    });
            } else {
                window.__psiPontesPending.push({ setCodigo: setCodigo, executar: executar });
            }
        };
    </script>
</head>
<body class="h-full flex flex-col overflow-hidden">
    <div id="app" class="h-full flex flex-col"></div>
    <script type="module">
        import { montar } from './runtime-web.js';
        import { PlaygroundApp } from './playground.js';
        montar(PlaygroundApp, document.getElementById('app'));
    </script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs/loader.js" onload="
        var script = document.createElement('script');
        script.src = '/editor-monaco.js';
        document.body.appendChild(script);
    "></script>
</body>
</html>`
