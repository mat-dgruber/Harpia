package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer é a máquina de estados que executa o processo de varredura (scanning) de caracteres
// e os transforma em uma sequência de tokens lógicos.
//
// O Lexer lê o slice de bytes de entrada incrementalmente, gerenciando o cursor físico, a linha
// e a coluna atuais para alimentar diagnósticos precisos em caso de falhas de análise.
type Lexer struct {
	fonte         []byte // O código fonte completo em formato de slice de bytes UTF-8.
	ch            rune   // O caractere Unicode (runa) sob análise corrente no cursor.
	largura       int    // Largura em bytes do caractere corrente.
	offset        int    // Posição de byte corrente do cursor no slice fonte.
	proximoOffset int    // Próxima posição de byte a ser lida.
	linha         int    // O número da linha física correspondente no código fonte (base 1).
	coluna        int    // O índice de coluna física do caractere sob análise (base 1).
}

// NewLexer aloca e prepara uma nova instância operacional de Lexer para o código fonte informado.
func NewLexer(entrada string) *Lexer {
	fonte := []byte(strings.ReplaceAll(entrada, "\r", ""))
	l := &Lexer{
		fonte:  fonte,
		linha:  1,
		coluna: 0,
	}

	l.avancar()

	if l.ch == '#' && l.peek() == '!' {
		l.ignorarComentario()
	}

	return l
}

// fimDeArquivo verifica se o cursor de leitura do interpretador atingiu ou passou o limite físico
// de bytes presentes no código fonte.
func (l *Lexer) fimDeArquivo() bool {
	return l.offset >= len(l.fonte)
}

// peek faz uma espreitadela (lookahead) de 1 caractere à frente do cursor de leitura.
func (l *Lexer) peek() rune {
	if l.proximoOffset >= len(l.fonte) {
		return 0
	}
	r, _ := utf8.DecodeRune(l.fonte[l.proximoOffset:])
	return r
}

// peekN faz uma espreitadela de n caracteres à frente do cursor de leitura.
func (l *Lexer) peekN(n int) rune {
	if n <= 0 {
		return l.ch
	}
	pos := l.proximoOffset
	var r rune
	for i := 1; i <= n; i++ {
		if pos >= len(l.fonte) {
			return 0
		}
		var w int
		r, w = utf8.DecodeRune(l.fonte[pos:])
		pos += w
	}
	return r
}

// ignorarComentarioHTML pula todos os caracteres até encontrar a sequência de fechamento de comentário HTML '-->'
func (l *Lexer) ignorarComentarioHTML() {
	l.avancar() // consome '<'
	l.avancar() // consome '!'
	l.avancar() // consome '-'
	l.avancar() // consome '-'

	for !l.fimDeArquivo() {
		if l.ch == '-' && l.peek() == '-' {
			l.avancar() // consome o primeiro '-'
			l.avancar() // consome o segundo '-'
			if l.ch == '>' {
				l.avancar() // consome o '>'
				break
			}
			continue
		}
		l.avancar()
	}
}

// avancar incrementa a posição do cursor em um caractere e carrega a nova runa correspondente.
func (l *Lexer) avancar() {
	l.offset = l.proximoOffset
	if l.proximoOffset >= len(l.fonte) {
		l.ch = 0
		l.largura = 0
		return
	}
	r, w := utf8.DecodeRune(l.fonte[l.proximoOffset:])
	l.ch = r
	l.largura = w
	l.proximoOffset += w

	if l.ch == '\n' {
		l.linha++
		l.coluna = 0
	} else {
		l.coluna++
	}
}

// posicaoAtual captura as coordenadas geográficas (coluna, linha, offset absoluto de byte)
// do caractere corrente no código, retornando um objeto *PosicaoToken alocado dinamicamente.
func (l *Lexer) posicaoAtual() *PosicaoToken {
	return &PosicaoToken{
		Coluna: l.coluna,
		Linha:  l.linha,
		Indice: l.offset,
	}
}

// ignorarEspacos avança o cursor descartando e pulando caracteres inofensivos de espaço em branco e tabulação.
func (l *Lexer) ignorarEspacos() {
	for (l.ch == ' ' || l.ch == '\t') && !l.fimDeArquivo() {
		l.avancar()
	}
}

// ignorarComentario pula todos os caracteres que compõem uma linha de comentário (iniciada pelo caractere '#')
func (l *Lexer) ignorarComentario() {
	for l.ch != '\n' && !l.fimDeArquivo() {
		l.avancar()
	}
	l.avancar()
}

// ignorarComentarioBloco pula todos os caracteres até encontrar a sequência de fechamento de comentário de bloco '*/'
func (l *Lexer) ignorarComentarioBloco() {
	l.avancar() // consome '/'
	l.avancar() // consome '*'

	for !l.fimDeArquivo() {
		if l.ch == '*' && l.peek() == '/' {
			l.avancar() // consome '*'
			l.avancar() // consome '/'
			break
		}
		l.avancar()
	}
}

// SubString é um fatiador seguro que extrai uma partição do slice de bytes original
// delimitado por offsets de bytes.
func (l *Lexer) SubString(inicio, fim int) string {
	if inicio < 0 {
		inicio = 0
	}
	if fim > len(l.fonte) {
		fim = len(l.fonte)
	}
	if inicio >= fim {
		return ""
	}
	return string(l.fonte[inicio:fim])
}

// lerIdentificador consome sequencialmente caracteres alfanuméricos ou sublinhas (_) que compõem
// o nome de uma variável, classe, função ou palavra reservada.
func (l *Lexer) lerIdentificador() *Token {
	inicio := l.posicaoAtual()

	for {
		l.avancar()

		if !(unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_') {
			break
		}
	}

	fim := l.posicaoAtual()
	valor := l.SubString(inicio.Indice, fim.Indice)
	tipo := TokenIdentificador

	if t, ok := tokensIdentificadores[valor]; ok {
		tipo = t
	}

	return newToken(tipo, valor, inicio, fim)
}

// lerNumero consome sequencialmente dígitos e o caractere separador decimal '.' para fatiar literais numéricos.
func (l *Lexer) lerNumero() *Token {
	inicio := l.posicaoAtual()

	for {
		l.avancar()

		if !(unicode.IsDigit(l.ch) || l.ch == '.') {
			break
		}
	}

	fim := l.posicaoAtual()
	valor := l.SubString(inicio.Indice, fim.Indice)

	tipo := TokenInteiro
	if strings.Contains(valor, ".") {
		tipo = TokenDecimal
	}

	return newToken(tipo, valor, inicio, fim)
}

// lerEntidadeHTML consome uma entidade HTML do tipo &#DIGITS; ou &NAME; (como &#8595; ou &nbsp;).
func (l *Lexer) lerEntidadeHTML() {
	l.avancar() // consome '&'
	for !l.fimDeArquivo() && l.ch != ';' && l.ch != ' ' && l.ch != '\n' {
		l.avancar()
	}
	if !l.fimDeArquivo() && l.ch == ';' {
		l.avancar() // consome ';'
	}
}

// lerTexto consome uma cadeia literal delimitada por aspas simples (') ou aspas duplas ("),
// tratando de forma transparente sequências de escape do delimitador (ex: \" ou \').
func (l *Lexer) lerTexto() *Token {
	inicio := l.posicaoAtual()
	delimitador := l.ch

	for {
		if l.fimDeArquivo() {
			break
		}
		l.avancar()

		if l.ch == '\\' {
			l.avancar() // Pula o caractere de escape '\' e preserva o caractere seguinte
			continue
		}

		if l.ch == delimitador {
			l.avancar()
			break
		}
	}

	fim := l.posicaoAtual()
	return newToken(TokenTexto, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
}

// SnapshotLex captura o estado completo do cursor léxico para permitir rollback.
type SnapshotLex struct {
	offset        int
	proximoOffset int
	ch            rune
	largura       int
	linha         int
	coluna        int
}

// Estado retorna um snapshot imutável do estado atual do lexer.
func (l *Lexer) Estado() SnapshotLex {
	return SnapshotLex{
		offset:        l.offset,
		proximoOffset: l.proximoOffset,
		ch:            l.ch,
		largura:       l.largura,
		linha:         l.linha,
		coluna:        l.coluna,
	}
}

// Restaurar devolve o lexer ao estado capturado em um snapshot anterior.
func (l *Lexer) Restaurar(s SnapshotLex) {
	l.offset = s.offset
	l.proximoOffset = s.proximoOffset
	l.ch = s.ch
	l.largura = s.largura
	l.linha = s.linha
	l.coluna = s.coluna
}

// PróximoToken é o ponto de entrada principal do analisador léxico.
func (l *Lexer) PróximoToken() *Token {
	l.ignorarEspacos()

	if l.fimDeArquivo() {
		return &Token{
			Tipo:   TokenFimDeArquivo,
			Valor:  "",
			Inicio: l.posicaoAtual(),
			Fim:    l.posicaoAtual(),
		}
	}

	ch := l.ch
	inicio := l.posicaoAtual()

	// ponytail: reconhece entidades HTML (&#DIGITS; ou &NAME; ex: &#8595; ou &nbsp;)
	if ch == '&' && (l.peek() == '#' || unicode.IsLetter(l.peek())) {
		l.lerEntidadeHTML()
		fim := l.posicaoAtual()
		return newToken(TokenEntidadeHTML, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
	}

	// Trata comentários de linha única '#' ou '//' (deve rodar antes de operadores simples para evitar colisão com '/')
	if ch == '#' || (ch == '/' && l.peek() == '/') {
		l.ignorarComentario()
		return l.PróximoToken()
	}

	// Trata comentários de bloco '/* ... */'
	if ch == '/' && l.peek() == '*' {
		l.ignorarComentarioBloco()
		return l.PróximoToken()
	}

	// ponytail: trata comentários HTML '<!-- ... -->' de forma nativa para excelente DX no JSX
	if ch == '<' && l.peek() == '!' {
		if l.peekN(2) == '-' && l.peekN(3) == '-' {
			l.ignorarComentarioHTML()
			return l.PróximoToken()
		}
	}

	// ponytail: trata operador spread '...'
	if ch == '.' && l.peek() == '.' && l.peekN(2) == '.' {
		l.avancar()
		l.avancar()
		l.avancar()
		fim := l.posicaoAtual()
		return newToken(TokenTresPontos, "...", inicio, fim)
	}

	// Se for um caractere operador catalogado ou operador de negação '!' ou '?'
	strVal := string(ch)
	if tipo, ok := tokensSimples[strVal]; ok || ch == '!' || ch == '?' {
		for {
			if l.fimDeArquivo() {
				break
			}
			l.avancar()
			// Guloso: tenta aglutinar caracteres subsequentes para formar operadores compostos (ex: '=' + '=' = '==')
			if t, ok := tokensSimples[strVal+string(l.ch)]; ok {
				strVal += string(l.ch)
				tipo = t
				continue
			}

			break
		}

		fim := l.posicaoAtual()
		return newToken(tipo, strVal, inicio, fim)
	}

	switch ch {
	case '"', '\'':
		return l.lerTexto()
	default:
		// Se iniciar com letra ou sublinha, lê como identificador
		if unicode.IsLetter(ch) || ch == '_' {
			return l.lerIdentificador()
		}

		// Se iniciar com dígito, lê como literal numérico
		if unicode.IsDigit(ch) {
			return l.lerNumero()
		}
	}

	// Caso não coincida com nenhuma regra, sinaliza um erro léxico e avança o cursor
	errTok := &Token{Tipo: TokenErro, Valor: string(l.ch), Inicio: inicio, Fim: l.posicaoAtual()}
	l.avancar()
	return errTok
}
