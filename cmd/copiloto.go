package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var modeloCopiloto string
var ollamaURL string

type OllamaGenerateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaGenerateRes struct {
	Response string `json:"response"`
}

func SugerirCopiloto(contexto string) (string, error) {
	prompt := fmt.Sprintf(`Você é o Copiloto da linguagem Harpia (linguagem Full Stack brasileira com sintaxe em português, componentes reativos estilo JSX, OOP simples, tipagem opcional e APIs integradas).
Dado o trecho de código abaixo, complete-o de forma natural e idiomática usando a sintaxe oficial da Harpia.
IMPORTANTE: Retorne APENAS o código de complementação, sem comentários explicativos, sem tags de markdown, sem preâmbulo. Apenas a continuação direta do código.

Contexto de código:
%s`, contexto)

	reqBody := OllamaGenerateReq{
		Model:  modeloCopiloto,
		Prompt: prompt,
		Stream: false,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("Ollama offline em %s: %v", ollamaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro do Ollama (status %d)", resp.StatusCode)
	}

	var resObj OllamaGenerateRes
	if err := json.NewDecoder(resp.Body).Decode(&resObj); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta do Ollama: %v", err)
	}

	return resObj.Response, nil
}

func comandoCopiloto() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copiloto [arquivo]",
		Short: "Sugere complementação de código inteligente via modelo de IA local (Ollama)",
		Long:  "Lê o arquivo ou contexto do código fornecido e aciona o Ollama para autocompletar o código Harpia.",
		Run: func(cmd *cobra.Command, args []string) {
			var contexto string

			if len(args) > 0 {
				bytes, err := os.ReadFile(args[0])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao ler arquivo: %v\n", err)
					os.Exit(1)
				}
				contexto = string(bytes)
			} else {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao ler stdin: %v\n", err)
					os.Exit(1)
				}
				contexto = string(bytes)
			}

			sugestao, err := SugerirCopiloto(contexto)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Copiloto indisponível: %v\n", err)
				os.Exit(0)
			}

			fmt.Print(sugestao)
		},
	}

	cmd.Flags().StringVarP(&modeloCopiloto, "modelo", "m", "llama3", "Modelo do Ollama a ser utilizado")
	cmd.Flags().StringVarP(&ollamaURL, "url", "u", "http://localhost:11434/api/generate", "URL da API do Ollama local")

	// Subcomandos 'revisar' e 'refatorar' ficam ligados ao próprio 'copiloto'
	// para não precisar editar o registro central em cmd.go.
	cmd.AddCommand(comandoCopilotoRevisar())
	cmd.AddCommand(comandoCopilotoRefatorar())

	return cmd
}

// =============================================================================
// Subcomandos: copiloto revisar  /  copiloto refatorar
//
// Análise estática textual sobre arquivos Harpia (.ptst / .pt / .hrp).
// Implementação propositadamente simples: usa apenas os pacotes já importados
// em copiloto.go (bytes, fmt, os). Sem dependências externas para manter a
// ferramenta portátil e despachar rápido dentro da CLI local.
// =============================================================================

const (
	limiteLinhasFuncao      = 80
	limiteParametrosFuncao  = 5
	limiteProfundidadeAni   = 4
	maxLinhasAmostraRefator = 3
)

// achadoRevisao representa um item individual encontrado pela revisão estática.
type achadoRevisao struct {
	Linha    int
	Tipo     string // "função-longa", "muitos-parâmetros", "aninhamento", "variável-não-usada", "TODO/FIXME"
	Mensagem string
}

// funcaoDetectada representa um bloco de função (texto livre) localizado pela
// heurística de chaves. Guarda posições humanas (1-based) e o conteúdo bruto.
type funcaoDetectada struct {
	Nome      string
	LinhaIni  int // 1-based
	LinhaFim  int // 1-based
	Conteudo  []byte
	Parametro string // assinatura do parâmetro extraída (texto da linha de abertura).
}

// comandoCopilotoRevisar implementa 'copiloto revisar <arquivo.ptst>'.
func comandoCopilotoRevisar() *cobra.Command {
	return &cobra.Command{
		Use:   "revisar <arquivo.ptst>",
		Short: "Análise estática: detecta funções longas, muitos parâmetros, aninhamento, vars não usadas e TODO/FIXME",
		Long:  "Lê um arquivo Harpia (.ptst/.pt/.hrp) e emite uma lista de achados em PT-BR no formato [ARQUIVO:linha] tipo → mensagem, seguido de um sumário.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			caminho := args[0]

			conteudo, err := os.ReadFile(caminho)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler arquivo %s: %v\n", caminho, err)
				os.Exit(1)
			}

			linhas := bytes.Split(conteudo, []byte{'\n'})

			var achados []achadoRevisao

			// 1) Funções > 80 linhas + parâmetros > 5 + aninhamento > 4
			for _, f := range detectarFuncoesTexto(linhas) {
				totalLinhas := f.LinhaFim - f.LinhaIni + 1
				if totalLinhas > limiteLinhasFuncao {
					achados = append(achados, achadoRevisao{
						Linha:    f.LinhaIni,
						Tipo:     "função-longa",
						Mensagem: fmt.Sprintf("função '%s' possui %d linhas (limite %d)", f.Nome, totalLinhas, limiteLinhasFuncao),
					})
				}
				qntdParams := contarParametros(f.Parametro)
				if qntdParams > limiteParametrosFuncao {
					achados = append(achados, achadoRevisao{
						Linha:    f.LinhaIni,
						Tipo:     "muitos-parâmetros",
						Mensagem: fmt.Sprintf("função '%s' aceita %d parâmetros (limite %d)", f.Nome, qntdParams, limiteParametrosFuncao),
					})
				}
				prof := profundidadeMax(f.Conteudo)
				if prof > limiteProfundidadeAni {
					achados = append(achados, achadoRevisao{
						Linha:    f.LinhaIni,
						Tipo:     "aninhamento",
						Mensagem: fmt.Sprintf("função '%s' atinge profundidade %d (limite %d)", f.Nome, prof, limiteProfundidadeAni),
					})
				}
			}

			// 2) Variáveis começando com '_' nunca referenciadas no escopo abaixo
			achados = append(achados, detectarVarsSublinhadoNaoUsadas(linhas)...)

			// 3) Comentários TODO/FIXME
			achados = append(achados, detectarTodosFixmes(linhas, caminho)...)

			// Saída dos achados + sumário
			for _, a := range achados {
				fmt.Printf("[%s:%d] %s → %s\n", caminho, a.Linha, a.Tipo, a.Mensagem)
			}
			if len(achados) == 0 {
				fmt.Printf("[%s] nenhum problema encontrado.\n", caminho)
				return
			}
			fmt.Printf("\nSumário: %d achado(s) em %s.\n", len(achados), caminho)
		},
	}
}

// comandoCopilotoRefatorar implementa 'copiloto refatorar <arquivo.ptst>'.
func comandoCopilotoRefatorar() *cobra.Command {
	return &cobra.Command{
		Use:   "refatorar <arquivo.ptst>",
		Short: "Sugere extração de helpers a partir de funções/métodos com mais de 80 linhas",
		Long:  "Para cada bloco com mais de 80 linhas, mostra primeira/última linha e sugere um nome de helper baseado em verbos/substantivos do conteúdo (ou 'helper_<N>' como fallback).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			caminho := args[0]
			conteudo, err := os.ReadFile(caminho)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler arquivo %s: %v\n", caminho, err)
				os.Exit(1)
			}
			linhas := bytes.Split(conteudo, []byte{'\n'})

			funcoes := detectarFuncoesTexto(linhas)
			encontrou := false
			for i, f := range funcoes {
				totalLinhas := f.LineaFim(f)
				if totalLinhas <= limiteLinhasFuncao {
					continue
				}
				encontrou = true
				nomeSug := sugerirNomeHelper(f, i+1)
				fmt.Printf("[linha %d – linha %d] → %s  (função '%s', %d linhas)\n",
					f.LinhaIni, f.LinhaFim, nomeSug, f.Nome, totalLinhas)
				fmt.Printf("  primeira: %s\n", bytes.TrimSpace(linhas[f.LinhaIni-1]))
				ult := f.LinhaFim - 1
				if ult < 0 || ult >= len(linhas) {
					ult = len(linhas) - 1
				}
				fmt.Printf("  última:   %s\n", bytes.TrimSpace(linhas[ult]))
			}
			if !encontrou {
				fmt.Printf("[%s] nenhuma função com mais de %d linhas encontrada.\n", caminho, limiteLinhasFuncao)
			}
		},
	}
}

// LineaFim utilidade (mantido fora de struct para evitar mexer no tipo principal).
func (f funcaoDetectada) LineaFim(f2 funcaoDetectada) int { return f2.LinhaFim - f2.LinhaIni + 1 }

// detectarFuncoesTexto faz varredura textual por blocos 'func nome(...) { ... }'.
// Heurística robusta o suficiente para a CLI local: a assinatura começa com
// 'funcao' (ou 'func'), abre '{' e fechamos balanceando chaves respeitando
// strings de uma linha (sem strings multilinhas na linguagem).
func detectarFuncoesTexto(linhas [][]byte) []funcaoDetectada {
	var resultado []funcaoDetectada
	for i := 0; i < len(linhas); i++ {
		linha := bytes.TrimSpace(linhas[i])
		if !comecaComPalavraChaveFuncao(linha) {
			continue
		}
		// Procura '{' na mesma linha; se não houver, olha as próximas até achar.
		chaveAbre := bytes.IndexByte(linha, '{')
		var cabecalho string
		var parametro string
		idxCabecalho := i
		if chaveAbre < 0 {
			// Função cuja 'assinatura + abre chaves' ocupa várias linhas.
			var agregado []byte
			j := i
			encontrou := false
			for ; j < len(linhas); j++ {
				agregado = append(agregado, linhas[j]...)
				if bytes.IndexByte(linhas[j], '{') >= 0 {
					encontrou = true
					break
				}
				agregado = append(agregado, '\n')
			}
			if !encontrou {
				continue
			}
			cabecalho = string(agregado)
			idxCabecalho = j
			chaveAbre = bytes.IndexByte(agregado, '{')
		} else {
			cabecalho = string(linha)
		}

		// Extrai nome e assinatura de parâmetros do cabeçalho conhecido.
		nome := extrairNomeFuncao(cabecalho)
		parametro = extrairParametros(cabecalho)

		// Conta chaves a partir da linha idxCabecalho (1-based + 1 já que i é 0-based).
		linhaFim, ok := casarChave(linhas, idxCabecalho)
		if !ok || linhaFim <= idxCabecalho+1 {
			// Bloco vazio ou malformado — ignora.
			i = linhaFim
			continue
		}

		// Conteúdo: do início do bloco (linha-i 0-based) até linhaFim (1-based).
		var conteudo []byte
		for k := i; k < linhaFim && k < len(linhas); k++ {
			conteudo = append(conteudo, linhas[k]...)
			conteudo = append(conteudo, '\n')
		}

		resultado = append(resultado, funcaoDetectada{
			Nome:      nome,
			LinhaIni:  i + 1,
			LinhaFim:  linhaFim,
			Conteudo:  conteudo,
			Parametro: parametro,
		})
		i = linhaFim - 1 // avança o for
	}
	return resultado
}

// comecaComPalavraChaveFuncao reconhece 'funcao'/'func'/'função'/'function' no início.
func comecaComPalavraChaveFuncao(linha []byte) bool {
	for _, pref := range []string{"funcao ", "função ", "func ", "function "} {
		if bytes.HasPrefix(linha, []byte(pref)) {
			return true
		}
	}
	return false
}

// extrairNomeFuncao retorna o identificador logo após a palavra-chave.
func extrairNomeFuncao(cabecalho string) string {
	for _, pref := range []string{"funcao ", "função ", "func ", "function "} {
		idx := -1
		for i := 0; i+len(pref) <= len(cabecalho); i++ {
			if cabecalho[i:i+len(pref)] == pref {
				idx = i + len(pref)
				break
			}
		}
		if idx < 0 {
			continue
		}
		resto := cabecalho[idx:]
		// Pula espaços.
		p := 0
		for p < len(resto) && (resto[p] == ' ' || resto[p] == '\t') {
			p++
		}
		inicio := p
		for p < len(resto) && isIdentChar(resto[p]) {
			p++
		}
		if p > inicio {
			return resto[inicio:p]
		}
	}
	return "<anônima>"
}

// isIdentChar verifica se é válido em identificador Harpia (ASCII simplificado).
func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' || c == ':'
}

// extrairParametros retorna o conteúdo entre '(' e ')' próximos, ou literal vazio.
func extrairParametros(cabecalho string) string {
	abre := -1
	for i := 0; i < len(cabecalho); i++ {
		if cabecalho[i] == '(' {
			abre = i
			break
		}
	}
	if abre < 0 {
		return ""
	}
	for j := abre + 1; j < len(cabecalho); j++ {
		if cabecalho[j] == ')' {
			return cabecalho[abre+1 : j]
		}
	}
	return ""
}

// contarParametros conta vírgulas no nível superior (sem levar '(' em conta).
func contarParametros(params string) int {
	if len(bytes.TrimSpace([]byte(params))) == 0 {
		return 0
	}
	n := 1
	prof := 0
	for i := 0; i < len(params); i++ {
		switch params[i] {
		case '(':
			prof++
		case ')':
			prof--
		case ',':
			if prof == 0 {
				n++
			}
		}
	}
	return n
}

// casarChave fecha a chave '{' a partir do índice (0-based) e retorna a linha (1-based) do '}' correspondente.
func casarChave(linhas [][]byte, inicio int) (int, bool) {
	prof := 0
	for i := inicio; i < len(linhas); i++ {
		linha := linhas[i]
		for j := 0; j < len(linha); j++ {
			c := linha[j]
			// Ignora conteúdo de string entre aspas duplas ou simples na mesma linha.
			if c == '"' || c == '\'' || c == '`' {
				aspas := c
				j++
				for j < len(linha) && linha[j] != aspas {
					if linha[j] == '\\' && j+1 < len(linha) {
						j++
					}
					j++
				}
				continue
			}
			// Ignora // comentário de linha.
			if c == '/' && j+1 < len(linha) && linha[j+1] == '/' {
				break
			}
			switch c {
			case '{':
				prof++
			case '}':
				prof--
				if prof == 0 {
					return i + 1, true
				}
			}
		}
	}
	return 0, false
}

// profundidadeMax calcula a profundidade de aninhamento máxima do bloco (apenas chaves).
func profundidadeMax(conteudo []byte) int {
	max, prof := 0, 0
	for i := 0; i < len(conteudo); i++ {
		c := conteudo[i]
		if c == '"' || c == '\'' || c == '`' {
			aspas := c
			i++
			for i < len(conteudo) && conteudo[i] != aspas {
				if conteudo[i] == '\\' && i+1 < len(conteudo) {
					i++
				}
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(conteudo) && conteudo[i+1] == '/' {
			// comentário de linha: pular até \n
			for i < len(conteudo) && conteudo[i] != '\n' {
				i++
			}
			continue
		}
		if c == '{' {
			prof++
			if prof > max {
				max = prof
			}
		} else if c == '}' {
			prof--
		}
	}
	return max
}

// detectarVarsSublinhadoNaoUsadas varrer linhas procurando 'var _nome' e verificar
// se '_nome' aparece em alguma linha posterior (mesmo escopo textual).
func detectarVarsSublinhadoNaoUsadas(linhas [][]byte) []achadoRevisao {
	var achados []achadoRevisao
	for i, ln := range linhas {
		linha := ln
		// Heurística: a linha inicia (após espaços) com 'var _'.
		inicio := 0
		for inicio < len(linha) && (linha[inicio] == ' ' || linha[inicio] == '\t') {
			inicio++
		}
		resto := linha[inicio:]
		if !bytes.HasPrefix(resto, []byte("var ")) && !bytes.HasPrefix(resto, []byte("const ")) {
			continue
		}
		// Extrai o nome (primeira palavra após 'var ').
		depois := resto[4:]
		k := 0
		for k < len(depois) && depois[k] == ' ' {
			k++
		}
		inicioNome := k
		for k < len(depois) && isIdentCharMedio(depois[k]) {
			k++
		}
		if k == inicioNome {
			continue
		}
		nome := string(depois[inicioNome:k])
		if len(nome) == 0 || nome[0] != '_' {
			continue
		}
		// Procura referências posteriores — olha todo o arquivo abaixo.
		refs := 0
		corpo := bytes.Join(linhas[i+1:], []byte{'\n'})
		refs += bytes.Count(corpo, []byte(nome))
		// A própria linha de declaração contém o nome; subtraímos 1.
		if refs == 0 {
			achados = append(achados, achadoRevisao{
				Linha:    i + 1,
				Tipo:     "variável-não-usada",
				Mensagem: fmt.Sprintf("variável '%s' começa com '_' e não é referenciada em nenhum ponto abaixo", nome),
			})
		}
	}
	return achados
}

func isIdentCharMedio(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// detectarTodosFixmes varre linhas por padrões TODO/FIXME em comentários.
func detectarTodosFixmes(linhas [][]byte, caminho string) []achadoRevisao {
	var achados []achadoRevisao
	for i, ln := range linhas {
		linha := ln
		// aceita // ou # como início de comentário
		if !bytes.Contains(linha, []byte("TODO")) && !bytes.Contains(linha, []byte("FIXME")) {
			continue
		}
		// Localiza cada ocorrência e adiciona um achado.
		for _, padrao := range []string{"TODO", "FIXME"} {
			pos := 0
			for {
				idx := indexAsString(linha, []byte(padrao), pos)
				if idx < 0 {
					break
				}
				msg := fmt.Sprintf("comentário %s encontrado: '%s'", padrao, trimRight(string(linha)))
				achados = append(achados, achadoRevisao{
					Linha:    i + 1,
					Tipo:     "TODO/FIXME",
					Mensagem: msg,
				})
				pos = idx + len(padrao)
			}
		}
	}
	_ = caminho
	return achados
}

// indexAsString faz busca de substring sem usar 'strings' nem 'bytes.Index'
// repetido (que poderia ser usado direto, mas evito imports novos).
func indexAsString(haystack, needle []byte, desde int) int {
	for i := desde; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// trimRight remove espaços à direita.
func trimRight(s string) string {
	i := len(s)
	for i > 0 && (s[i-1] == ' ' || s[i-1] == '\t' || s[i-1] == '\r') {
		i--
	}
	return s[:i]
}

// sugerirNomeHelper tenta extrair um verbo + substantivo a partir das
// primeiras linhas de conteúdo; cai para 'helper_<N>' se nada plausível.
func sugerirNomeHelper(f funcaoDetectada, n int) string {
	linhas := bytes.Split(f.Conteudo, []byte{'\n'})
	amostra := make([][]byte, 0, maxLinhasAmostraRefator)
	for _, l := range linhas {
		tl := bytes.TrimSpace(l)
		if len(tl) == 0 {
			continue
		}
		tl = stripComentarioLinha(tl)
		amostra = append(amostra, tl)
		if len(amostra) == maxLinhasAmostraRefator {
			break
		}
	}

	verbosPadrao := []string{
		"validar", "verificar", "calcular", "montar", "enviar", "processar",
		"renderizar", "render", "parse", "formatar", "gerar", "somar", "carregar",
		"buscar", "filtrar", "ordenar", "construir", "atualizar", "remover",
		"converter", "transformar", "criar", "inicializar", "normalizar", "limpar",
		"executar", "abrir", "fechar", "ler", "escrever", "imprimir", "logar",
		"emitir", "compor", "resolver", "preparar", "configurar",
	}
	substantivosPadrao := []string{
		"usuario", "usuarios", "cliente", "clientes", "pedido", "pedidos",
		"produto", "produtos", "mensagem", "mensagens", "resposta", "respostas",
		"requisicao", "requisicoes", "request", "lista", "listas", "evento",
		"eventos", "config", "configuracao", "configuracoes", "dados", "contexto",
		"estado", "estado_inicial", "relatorio", "relatorios", "token", "tokens",
		"sessao", "sessoes", "erro", "erros", "resultado", "resultados",
		"pagina", "paginas", "rota", "rotas", "componente", "componentes",
		"html", "css", "valor", "valores", "chave", "chaves", "buffer", "buffers",
		"texto", "numeros", "numero", "data", "horario", "horarios",
		"parametros", "parametro", "registro", "registros", "linha", "linhas",
		"coluna", "colunas", "matriz", "tabela", "tabelas", "no", "nos",
		"cabecalho", "corpo", "indice", "indices",
	}

	var verbo, subst string
	for _, l := range amostra {
		// Tokeniza por whitespace e pontuação simples.
		palavras := tokenizarLinha(string(l))
		for _, p := range palavras {
			pl := lowerAscii(p)
			if verbo == "" && containsString(verbosPadrao, pl) {
				verbo = pl
			}
			if subst == "" && containsString(substantivosPadrao, pl) {
				subst = pl
			}
		}
		if verbo != "" && (subst != "" || stringEmTudo(amostra, substantivosPadrao)) {
			break
		}
	}

	if verbo == "" && subst == "" {
		return fmt.Sprintf("helper_%d", n)
	}
	if verbo == "" {
		// pega primeiro substantivo que aparecer em qualquer linha do conteúdo
		for _, l := range f.ConteudoStringLines() {
			for _, p := range tokenizarLinha(string(l)) {
				if containsString(substantivosPadrao, lowerAscii(p)) {
					subst = lowerAscii(p)
					break
				}
			}
			if subst != "" {
				break
			}
		}
		if subst == "" {
			return fmt.Sprintf("helper_%d", n)
		}
		return "executar_" + subst
	}
	if subst == "" {
		return verbo + "_interno"
	}
	return verbo + "_" + subst
}

// ConteudoStringLines parte o conteúdo em linhas para reuso do sugerirNomeHelper.
func (f funcaoDetectada) ConteudoStringLines() [][]byte {
	return bytes.Split(f.Conteudo, []byte{'\n'})
}

// stripComentarioLinha remove eventual // comentário de linha do trecho.
func stripComentarioLinha(linha []byte) []byte {
	for i := 0; i+1 < len(linha); i++ {
		if linha[i] == '/' && linha[i+1] == '/' {
			return bytes.TrimSpace(linha[:i])
		}
	}
	return linha
}

// tokenizarLinha quebra por whitespace e parênteses/colchetes genéricos.
func tokenizarLinha(s string) []string {
	var out []string
	atual := 0
	emit := func() {
		if atual < len(s) {
			out = append(out, s[atual:len(s)])
			atual = len(s)
		}
	}
	inWord := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ', '\t', '\n', '\r', '(', ')', '{', '}', '[', ']', ',', ';', '.', ':':
			if inWord {
				out = append(out, s[atual:i])
				atual = i + 1
				inWord = false
			} else {
				atual = i + 1
			}
		default:
			inWord = true
		}
	}
	emit()
	return out
}

// containsString verifica pertinência em slice sem sort.Search (mantém determinístico).
func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// stringEmTudo retorna true se as amostras contêm pelo menos um dos substantivos.
func stringEmTudo(amostra [][]byte, lista []string) bool {
	for _, l := range amostra {
		for _, p := range tokenizarLinha(string(l)) {
			if containsString(lista, lowerAscii(p)) {
				return true
			}
		}
	}
	return false
}

// lowerAscii converte letras ASCII para minúsculas.
func lowerAscii(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
