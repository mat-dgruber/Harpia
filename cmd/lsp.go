package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
	"github.com/spf13/cobra"
)

// Estruturas de dados JSON-RPC e LSP oficiais para troca de pacotes
type RequestMessage struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ResponseMessage struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type NotificationMessage struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type PublishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []LSPDiagnostic `json:"diagnostics"`
}

// ponytail: cache global em memória para formatação do documento reativo
var cacheArquivosLSP = make(map[string]string)

// ponytail: cache do último AST compilado por URI, evita refazer parse a cada hover/go-to-def
type entradaAstLSP struct {
	prog   *parser.Programa
	codigo string
}

var cacheAstLSP = make(map[string]entradaAstLSP)

// comandoLsp gerencia conexões de Language Server Protocol direto da IDE
func comandoLsp() *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Inicia o servidor de Language Server Protocol (LSP) para IDEs",
		Run: func(cmd *cobra.Command, args []string) {
			iniciarServidorLSP()
		},
	}
}

func iniciarServidorLSP() {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := lerMensagemLSP(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		var req RequestMessage
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}

		tratarRequisicaoLSP(req)
	}
}

func lerMensagemLSP(reader *bufio.Reader) ([]byte, error) {
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				val, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
				contentLength = val
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("Content-Length inválido")
	}

	buf := make([]byte, contentLength)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func enviarMensagemLSP(msg interface{}) {
	bytes, _ := json.Marshal(msg)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(bytes), string(bytes))
}

func tratarRequisicaoLSP(req RequestMessage) {
	switch req.Method {
	case "initialize":
		res := ResponseMessage{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"capabilities": map[string]interface{}{
					"textDocumentSync":           1,    // Full sync (didOpen, didChange, didClose, didSave)
					"documentFormattingProvider": true, // ponytail: ativa suporte síncrono para 'On-Save' na IDE
					"hoverProvider":              true, // ponytail: hover de palavras-chave e símbolos usando o Lexer
					"definitionProvider":         true, // ponytail: F12 navega para a declaração via walk do AST
					"completionProvider": map[string]interface{}{
						"resolveProvider":   false,
						"triggerCharacters": []string{".", " "},
					},
				},
			},
		}
		enviarMensagemLSP(res)

	case "initialized":
		// Confirmação de conexão vazia

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err == nil {
			cacheArquivosLSP[params.TextDocument.URI] = params.TextDocument.Text
			processarDiagnosticosLSP(params.TextDocument.URI, params.TextDocument.Text)
		}

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err == nil {
			if len(params.ContentChanges) > 0 {
				cacheArquivosLSP[params.TextDocument.URI] = params.ContentChanges[0].Text
				processarDiagnosticosLSP(params.TextDocument.URI, params.ContentChanges[0].Text)
			}
		}

	case "textDocument/formatting":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			codigo := cacheArquivosLSP[params.TextDocument.URI]
			if codigo != "" {
				codigoFormatado := FormatarCodigoHarpia(codigo)
				res := ResponseMessage{
					Jsonrpc: "2.0",
					ID:      req.ID,
					Result: []map[string]interface{}{
						{
							"range": map[string]interface{}{
								"start": map[string]interface{}{"line": 0, "character": 0},
								"end":   map[string]interface{}{"line": 100000, "character": 0},
							},
							"newText": codigoFormatado,
						},
					},
				}
				enviarMensagemLSP(res)
			}
		}

	case "textDocument/completion":
		// ponytail: retorna lista de CompletionItems contendo as palavras-chave e embutidos reativos
		items := []map[string]interface{}{
			{"label": "funcao", "kind": 14, "detail": "Declara uma nova função em português"},
			{"label": "classe", "kind": 14, "detail": "Declara uma nova classe com suporte a herança"},
			{"label": "enum", "kind": 14, "detail": "Declara uma nova enumeração de valores imutáveis"},
			{"label": "interface", "kind": 14, "detail": "Declara um novo contrato de interface"},
			{"label": "retorne", "kind": 14, "detail": "Retorna um valor de dentro de uma função"},
			{"label": "se", "kind": 14, "detail": "Estrutura de decisão condicional se"},
			{"label": "senao", "kind": 14, "detail": "Estrutura de decisão condicional senão"},
			{"label": "para", "kind": 14, "detail": "Estrutura de repetição para"},
			{"label": "enquanto", "kind": 14, "detail": "Estrutura de repetição enquanto"},
			{"label": "tente", "kind": 14, "detail": "Estrutura de tratamento de erros tente"},
			{"label": "capture", "kind": 14, "detail": "Estrutura de tratamento de erros capture"},
			{"label": "importar", "kind": 14, "detail": "Importa um arquivo ou módulo"},
			{"label": "exportar", "kind": 14, "detail": "Exporta uma variável, função ou classe para módulos externos"},
			{"label": "estilo", "kind": 14, "detail": "Declara um bloco de estilo estático em português"},
			{"label": "var", "kind": 14, "detail": "Declara uma variável mutável"},
			{"label": "constante", "kind": 14, "detail": "Declara uma constante imutável"},
			{"label": "assincrono", "kind": 14, "detail": "Declara uma função assíncrona não-bloqueante"},
			{"label": "aguarde", "kind": 14, "detail": "Aguarda a resolução de uma promessa assíncrona"},

			// Built-ins and Frontend/Reactivity (Kind: 3)
			{"label": "imprimir", "kind": 3, "detail": "Imprime texto na saída padrão", "insertText": "imprimir($1)"},
			{"label": "sinal", "kind": 3, "detail": "Cria um sinal reativo contendo um valor", "insertText": "var [$1, set$1] = sinal($2)"},
			{"label": "efeito", "kind": 3, "detail": "Cria um efeito colateral que roda sob mudanças de sinais", "insertText": "efeito(funcao() {\n    $1\n})"},
			{"label": "derivado", "kind": 3, "detail": "Cria um valor reativo derivado e memoizado", "insertText": "derivado(funcao() {\n    retorne $1;\n})"},
			{"label": "armazem", "kind": 3, "detail": "Gerenciador de Estado Global reativo", "insertText": "armazem($1)"},
			{"label": "montar", "kind": 3, "detail": "Inicializa a montagem reativa da aplicação", "insertText": "montar($1, $2)"},
			{"label": "roteador", "kind": 6, "detail": "Roteador SPA nativo com URLs limpas e navegação por History API"},
			{"label": "importarHtml", "kind": 3, "detail": "Inlinou dinamicamente um layout HTML de arquivo externo", "insertText": "importarHtml(\"$1\")"},
			{"label": "sinalPersistente", "kind": 3, "detail": "Cria um sinal reativo que persiste no localStorage", "insertText": "sinalPersistente(\"$1\", $2)"},
			{"label": "recurso", "kind": 3, "detail": "Cria uma primitiva de estado assíncrono para chamadas de rede", "insertText": "recurso(funcao() {\n    $1\n})"},
			{"label": "injetar", "kind": 3, "detail": "Injeta/recupera um serviço provido por Provedor superior", "insertText": "injetar(\"$1\")"},
		}

		res := ResponseMessage{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result:  items,
		}
		enviarMensagemLSP(res)

	case "shutdown":
		res := ResponseMessage{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result:  nil,
		}
		enviarMensagemLSP(res)

	case "exit":
		os.Exit(0)

	case "textDocument/hover":
		var params TextDocumentPositionParams
		if err := json.Unmarshal(req.Params, &params); err == nil {
			responderHoverLSP(req.ID, params)
		}

	case "textDocument/definition":
		var params TextDocumentPositionParams
		if err := json.Unmarshal(req.Params, &params); err == nil {
			responderDefinicaoLSP(req.ID, params)
		}
	}
}

func processarDiagnosticosLSP(uriStr, codigo string) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return
	}
	caminhoArquivo := u.Path
	if os.PathSeparator == '\\' && strings.HasPrefix(caminhoArquivo, "/") {
		caminhoArquivo = caminhoArquivo[1:]
	}

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	var diagnostics []LSPDiagnostic

	ast, err := ctx.StringParaAst(codigo, caminhoArquivo)
	if err != nil {
		// Adiciona erro de sintaxe do parser como diagnóstico de severity 1
		diagnostics = append(diagnostics, LSPDiagnostic{
			Range: DiagnosticRange{
				Start: DiagnosticPosition{Line: 0, Character: 0},
				End:   DiagnosticPosition{Line: 0, Character: 1},
			},
			Severity: 1,
			Code:     "PSC-0001",
			Source:   "Harpia-parser",
			Message:  err.Error(),
		})
	} else {
		// ponytail: alimenta o cache do AST para uso em hover/go-to-def sem reparsear
		if prog, ok := ast.(*parser.Programa); ok {
			cacheAstLSP[uriStr] = entradaAstLSP{prog: prog, codigo: codigo}
		}

		// Roda o linter no AST
		linter := &Linter{}
		linter.Checar(ast)

		// ponytail: linter reativo de segurança de Clean Architecture inline
		if prog, ok := ast.(*parser.Programa); ok {
			if strings.Contains(caminhoArquivo, "/dominio/") {
				for _, decl := range prog.Declaracoes {
					if imp, ok := decl.(*parser.ImporteDe); ok {
						caminhoImp := imp.Caminho.Valor
						if strings.Contains(caminhoImp, "/infra/") || strings.Contains(caminhoImp, "/web/") {
							var line, col, length int
							if tok, ok := linter.Posicoes[imp]; ok && tok != nil {
								line = tok.Inicio.Linha - 1
								col = tok.Inicio.Coluna - 1
								length = len(tok.Valor)
								if length == 0 {
									length = 1
								}
							}
							diagnostics = append(diagnostics, LSPDiagnostic{
								Range: DiagnosticRange{
									Start: DiagnosticPosition{Line: line, Character: col},
									End:   DiagnosticPosition{Line: line, Character: col + length},
								},
								Severity: 1, // Erro
								Code:     "PSC-ARCH-001",
								Source:   "Harpia-arquitetura",
								Message:  "Violação de Clean Architecture: Arquivos sob '/dominio' não podem importar camadas de infraestrutura ou web!",
							})
						}
					}
				}
			}
		}

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

			diagnostics = append(diagnostics, LSPDiagnostic{
				Range: DiagnosticRange{
					Start: DiagnosticPosition{Line: line, Character: col},
					End:   DiagnosticPosition{Line: line, Character: col + length},
				},
				Severity: errObj.Severity,
				Code:     errObj.Code,
				Source:   "Harpia-linter",
				Message:  errObj.Message,
			})
		}
	}

	// Envia notificação de diagnósticos para a IDE
	notif := NotificationMessage{
		Jsonrpc: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: PublishDiagnosticsParams{
			URI:         uriStr,
			Diagnostics: diagnostics,
		},
	}
	enviarMensagemLSP(notif)
}

// =============================================================================
// PONYTAIL: hover + go-to-definition nativos do LSP do Harpia
// =============================================================================

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     DiagnosticPosition     `json:"position"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type HoverResult struct {
	Contents MarkupContent    `json:"contents"`
	Range    *DiagnosticRange `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Location struct {
	URI   string          `json:"uri"`
	Range DiagnosticRange `json:"range"`
}

func palavraSobCursor(codigo string, linha, char int) string {
	linhas := strings.Split(codigo, "\n")
	if linha < 0 || linha >= len(linhas) {
		return ""
	}
	linhaTexto := linhas[linha]
	if char < 0 || char > len(linhaTexto) {
		return ""
	}

	// Anda para a esquerda
	inicio := char
	for inicio > 0 {
		r := linhaTexto[inicio-1]
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			inicio--
		} else {
			break
		}
	}

	// Anda para a direita
	fim := char
	for fim < len(linhaTexto) {
		r := linhaTexto[fim]
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			fim++
		} else {
			break
		}
	}

	if inicio == fim {
		return ""
	}
	return linhaTexto[inicio:fim]
}

func buscarDocInserida(codigo string, linhaDecl int) []string {
	linhas := strings.Split(codigo, "\n")
	idx := linhaDecl - 2 // Linha imediatamente anterior ao início do nó (já que linhaDecl é 1-based, a linha anterior está em linhaDecl - 2)
	var docs []string

	for idx >= 0 {
		linha := strings.TrimSpace(linhas[idx])
		if strings.HasPrefix(linha, "///") {
			docs = append([]string{strings.TrimSpace(strings.TrimPrefix(linha, "///"))}, docs...)
			idx--
		} else {
			break
		}
	}
	return docs
}

func encontrarDeclNoAST(prog *parser.Programa, nome string) (parser.BaseNode, *lexer.Token) {
	if prog == nil {
		return nil, nil
	}

	for _, decl := range prog.Declaracoes {
		switch d := decl.(type) {
		case *parser.DeclFuncao:
			if d.Nome == nome {
				return d, prog.Posicoes[d]
			}
		case *parser.DeclClasse:
			if d.Nome == nome {
				return d, prog.Posicoes[d]
			}
			for _, m := range d.Metodos {
				if m.Nome == nome {
					return m, prog.Posicoes[m]
				}
			}
		case *parser.DeclVar:
			if d.Nome == nome {
				return d, prog.Posicoes[d]
			}
		case *parser.DeclEnum:
			if d.Nome == nome {
				return d, prog.Posicoes[d]
			}
		case *parser.DeclInterface:
			if d.Nome == nome {
				return d, prog.Posicoes[d]
			}
			for _, m := range d.Metodos {
				if m.Nome == nome {
					return d, prog.Posicoes[d]
				}
			}
		}
	}
	return nil, nil
}

func assinaturaFuncao(d *parser.DeclFuncao) string {
	var sb strings.Builder
	// Nota: Como não temos d.Assincrono no AST de forma evidente,
	// podemos assumir formato padrão
	sb.WriteString("funcao ")
	sb.WriteString(d.Nome)
	sb.WriteString("(")
	var params []string
	for _, p := range d.Parametros {
		paramStr := p.Nome
		if p.Tipo != "" {
			paramStr += ": " + p.Tipo
		}
		params = append(params, paramStr)
	}
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(")")
	return sb.String()
}

func assinaturaClasse(d *parser.DeclClasse) string {
	sig := "classe " + d.Nome
	if d.Heranca != "" {
		sig += " estende " + d.Heranca
	}
	return sig
}

func assinaturaVar(d *parser.DeclVar) string {
	prefix := "var "
	if d.Constante {
		prefix = "constante "
	}
	sig := prefix + d.Nome
	if d.Tipo != "" {
		sig += ": " + d.Tipo
	}
	return sig
}

func obterDescricaoBuiltin(palavra string) string {
	switch palavra {
	case "imprimir":
		return "```Harpia\nimprimir(valor)\n```\n\nImprime um valor na saída padrão do console."
	case "sinal":
		return "```Harpia\nsinal(valorInicial)\n```\n\nCria um sinal reativo contendo um valor mutável."
	case "efeito":
		return "```Harpia\nefeito(funcao)\n```\n\nCria um efeito colateral que roda automaticamente sempre que os sinais dependentes mudam."
	case "derivado":
		return "```Harpia\nderivado(funcao)\n```\n\nCria um valor reativo derivado de outros sinais e memoizado."
	case "armazem":
		return "```Harpia\narmazem(estadoInicial)\n```\n\nCria um armazenamento de estado global reativo para componentes."
	case "montar":
		return "```Harpia\nmontar(componente, elementoAlvo)\n```\n\nInicializa e renderiza a montagem reativa da aplicação em um elemento alvo."
	}
	return ""
}

func responderHoverLSP(id interface{}, params TextDocumentPositionParams) {
	uri := params.TextDocument.URI
	cache, existe := cacheAstLSP[uri]
	if !existe {
		enviarMensagemLSP(ResponseMessage{Jsonrpc: "2.0", ID: id, Result: nil})
		return
	}

	palavra := palavraSobCursor(cache.codigo, params.Position.Line, params.Position.Character)
	if palavra == "" {
		enviarMensagemLSP(ResponseMessage{Jsonrpc: "2.0", ID: id, Result: nil})
		return
	}

	node, tok := encontrarDeclNoAST(cache.prog, palavra)
	if node == nil {
		descBuiltin := obterDescricaoBuiltin(palavra)
		if descBuiltin != "" {
			enviarMensagemLSP(ResponseMessage{
				Jsonrpc: "2.0",
				ID:      id,
				Result: HoverResult{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: descBuiltin,
					},
				},
			})
			return
		}

		enviarMensagemLSP(ResponseMessage{Jsonrpc: "2.0", ID: id, Result: nil})
		return
	}

	var assinatura string
	switch d := node.(type) {
	case *parser.DeclFuncao:
		assinatura = assinaturaFuncao(d)
	case *parser.DeclClasse:
		assinatura = assinaturaClasse(d)
	case *parser.DeclVar:
		assinatura = assinaturaVar(d)
	case *parser.DeclEnum:
		assinatura = fmt.Sprintf("enum %s { %s }", d.Nome, strings.Join(d.Valores, ", "))
	case *parser.DeclInterface:
		assinatura = fmt.Sprintf("interface %s", d.Nome)
	}

	var markdown strings.Builder
	markdown.WriteString(fmt.Sprintf("```Harpia\n%s\n```\n", assinatura))

	if tok != nil {
		docs := buscarDocInserida(cache.codigo, tok.Inicio.Linha)
		if len(docs) > 0 {
			markdown.WriteString("\n---\n")
			for _, doc := range docs {
				markdown.WriteString(doc + "\n")
			}
		}
	}

	var lspRange *DiagnosticRange
	if tok != nil {
		lspRange = &DiagnosticRange{
			Start: DiagnosticPosition{Line: tok.Inicio.Linha - 1, Character: tok.Inicio.Coluna - 1},
			End:   DiagnosticPosition{Line: tok.Fim.Linha - 1, Character: tok.Fim.Coluna - 1},
		}
	}

	enviarMensagemLSP(ResponseMessage{
		Jsonrpc: "2.0",
		ID:      id,
		Result: HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: markdown.String(),
			},
			Range: lspRange,
		},
	})
}

func responderDefinicaoLSP(id interface{}, params TextDocumentPositionParams) {
	uri := params.TextDocument.URI
	cache, existe := cacheAstLSP[uri]
	if !existe {
		enviarMensagemLSP(ResponseMessage{Jsonrpc: "2.0", ID: id, Result: nil})
		return
	}

	palavra := palavraSobCursor(cache.codigo, params.Position.Line, params.Position.Character)
	if palavra == "" {
		enviarMensagemLSP(ResponseMessage{Jsonrpc: "2.0", ID: id, Result: nil})
		return
	}

	node, tok := encontrarDeclNoAST(cache.prog, palavra)
	if node == nil || tok == nil {
		enviarMensagemLSP(ResponseMessage{Jsonrpc: "2.0", ID: id, Result: nil})
		return
	}

	loc := Location{
		URI: uri,
		Range: DiagnosticRange{
			Start: DiagnosticPosition{Line: tok.Inicio.Linha - 1, Character: tok.Inicio.Coluna - 1},
			End:   DiagnosticPosition{Line: tok.Fim.Linha - 1, Character: tok.Fim.Coluna - 1},
		},
	}

	enviarMensagemLSP(ResponseMessage{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  loc,
	})
}
