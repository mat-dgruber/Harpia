package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/natanfeitosa/portuscript/ptst"
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

// comandoPlayground inicia o servidor web do playground interativo local
func comandoPlayground() *cobra.Command {
	var porta int
	cmdPlay := &cobra.Command{
		Use:   "playground",
		Short: "Inicia o servidor web local do Playground Interativo do Portuscript",
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

	fmt.Printf("🚀 Playground Interativo do Portuscript rodando em: http://localhost:%d\n", porta)
	fmt.Println("Pressione Ctrl+C para encerrar o servidor.")
	err := http.ListenAndServe(fmt.Sprintf(":%d", porta), nil)
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
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	conteudo, err := os.ReadFile("playground/interface.ptst")
	if err != nil {
		http.Error(w, "Interface não encontrada", http.StatusInternalServerError)
		return
	}

	ast, err := ctx.StringParaAst(string(conteudo), "playground/interface.ptst")
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro de compilação: %v", err), http.StatusInternalServerError)
		return
	}

	transpiler := &TranspilerWeb{
		Estrito:       false,
		DiretorioBase: "playground",
	}
	jsOutput := transpiler.Transpile(ast)
	finalJS := "import { h, sinal, GradeDeDados, efeito, derivado, armazem, montar } from './runtime-web.js';\n\n" + jsOutput
	w.Write([]byte(finalJS))
}

func servePlaygroundCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	conteudo, err := os.ReadFile("playground/interface.ptst")
	if err != nil {
		w.Write([]byte("/* Estilo não encontrado */"))
		return
	}

	ast, err := ctx.StringParaAst(string(conteudo), "playground/interface.ptst")
	if err != nil {
		w.Write([]byte("/* Erro de compilação do estilo */"))
		return
	}

	transpiler := &TranspilerWeb{
		Estrito:       false,
		DiretorioBase: "playground",
	}
	transpiler.Transpile(ast)

	cssOutput := strings.Join(transpiler.Styles, "\n")
	w.Write([]byte(cssOutput))
}

func apiExecutarCodigoPlayground(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		return
	}

	var req ExecucaoPlaygroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Captura stdout para coletar impressões de logs do console
	oldStdout := os.Stdout
	readP, writeP, _ := os.Pipe()
	os.Stdout = writeP

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	escopo := ptst.NewEscopo()

	// Compila a string de código para AST
	ast, err := ctx.StringParaAst(req.Codigo, "playground.ptst")

	var response ExecucaoPlaygroundResponse
	if err != nil {
		// Fecha os pipes e restaura a saída padrão
		writeP.Close()
		os.Stdout = oldStdout

		response.Sucesso = false
		response.ErroHtml = ansiParaHtml(err.Error())
	} else {
		// Executa o script sob o escopo isolado
		_, errExec := ctx.AvaliarAst(ast, escopo)

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
	res = strings.ReplaceAll(res, "\u001b[31m", "<span class='text-red-400 font-bold'>") // vermelho
	res = strings.ReplaceAll(res, "\u001b[33m", "<span class='text-yellow-400 font-bold'>") // amarelo
	res = strings.ReplaceAll(res, "\u001b[32m", "<span class='text-green-400'>") // verde
	res = strings.ReplaceAll(res, "\u001b[36m", "<span class='text-cyan-400'>") // ciano
	res = strings.ReplaceAll(res, "\u001b[35m", "<span class='text-pink-400'>") // magenta
	res = strings.ReplaceAll(res, "\u001b[1m", "<span class='font-bold'>") // negrito
	res = strings.ReplaceAll(res, "\u001b[0m", "</span>") // reset
	return res
}

// htmlInterfacePlayground define a interface física do editor web em Dark Mode corporativo
const htmlInterfacePlayground = `<!DOCTYPE html>
<html lang="pt-BR" class="h-full bg-slate-950 text-slate-100">
<head>
    <meta charset="UTF-8">
    <title>Portuscript — Playground Interativo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/playground.css">
    <style>
        .editor-textarea {
            font-family: 'Fira Code', 'Courier New', monospace;
            tab-size: 4;
        }
    </style>
</head>
<body class="h-full flex flex-col overflow-hidden">
    <div id="app" class="h-full flex flex-col"></div>
    <script type="module">
        import { montar } from './runtime-web.js';
        import { PlaygroundApp } from './playground.js';
        montar(PlaygroundApp, document.getElementById('app'));
    </script>
</body>
</html>`
