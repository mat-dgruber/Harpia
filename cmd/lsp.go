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

	"github.com/natanfeitosa/portuscript/parser"
	"github.com/natanfeitosa/portuscript/ptst"
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
	TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`
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
					"textDocumentSync":           1, // Full sync (didOpen, didChange, didClose, didSave)
					"documentFormattingProvider": true, // ponytail: ativa suporte síncrono para 'On-Save' na IDE
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
				codigoFormatado := FormatarCodigoPortuscript(codigo)
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
			// Keywords (Kind: 14)
			{"label": "funcao", "kind": 14, "detail": "Declara uma nova função em português"},
			{"label": "classe", "kind": 14, "detail": "Declara uma nova classe com suporte a herança"},
			{"label": "retorne", "kind": 14, "detail": "Retorna um valor de dentro de uma função"},
			{"label": "se", "kind": 14, "detail": "Estrutura de decisão condicional se"},
			{"label": "senao", "kind": 14, "detail": "Estrutura de decisão condicional senão"},
			{"label": "para", "kind": 14, "detail": "Estrutura de repetição para"},
			{"label": "enquanto", "kind": 14, "detail": "Estrutura de repetição enquanto"},
			{"label": "tente", "kind": 14, "detail": "Estrutura de tratamento de erros tente"},
			{"label": "capture", "kind": 14, "detail": "Estrutura de tratamento de erros capture"},
			{"label": "importar", "kind": 14, "detail": "Importa um arquivo ou módulo"},
			{"label": "estilo", "kind": 14, "detail": "Declara um bloco de estilo estático em português"},
			{"label": "var", "kind": 14, "detail": "Declara uma variável mutável"},
			{"label": "constante", "kind": 14, "detail": "Declara uma constante imutável"},

			// Built-ins and Frontend/Reactivity (Kind: 3)
			{"label": "imprimir", "kind": 3, "detail": "Imprime texto na saída padrão", "insertText": "imprimir($1)"},
			{"label": "sinal", "kind": 3, "detail": "Cria um sinal reativo contendo um valor", "insertText": "sinal($1)"},
			{"label": "efeito", "kind": 3, "detail": "Cria um efeito colateral que roda sob mudanças de sinais", "insertText": "efeito(funcao() {\n    $1\n})"},
			{"label": "derivado", "kind": 3, "detail": "Cria um valor reativo derivado e memoizado", "insertText": "derivado(funcao() {\n    retorne $1;\n})"},
			{"label": "armazem", "kind": 3, "detail": "Gerenciador de Estado Global reativo", "insertText": "armazem($1)"},
			{"label": "montar", "kind": 3, "detail": "Inicializa a montagem reativa da aplicação", "insertText": "montar($1, $2)"},
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

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
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
			Source:   "portuscript-parser",
			Message:  err.Error(),
		})
	} else {
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
								Source:   "portuscript-arquitetura",
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
				Source:   "portuscript-linter",
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
