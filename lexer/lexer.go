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
	fonte := []byte(entrada)
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
	if r == '\r' {
		return '\n'
	}
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
	if r == '\r' {
		if l.proximoOffset+w < len(l.fonte) && l.fonte[l.proximoOffset+w] == '\n' {
			l.proximoOffset += w
			r, w = utf8.DecodeRune(l.fonte[l.proximoOffset:])
		} else {
			r = '\n'
		}
	}
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
func (l *Lexer) posicaoAtual() PosicaoToken {
	return PosicaoToken{
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
	profundidade := 1

	for !l.fimDeArquivo() && profundidade > 0 {
		if l.ch == '/' && l.peek() == '*' {
			profundidade++
			l.avancar() // consome '/'
			l.avancar() // consome '*'
		} else if l.ch == '*' && l.peek() == '/' {
			profundidade--
			l.avancar() // consome '*'
			l.avancar() // consome '/'
		} else {
			l.avancar()
		}
	}
}

// lerComentarioDocLinha consome um comentário de documentação de linha única iniciado por '///'
func (l *Lexer) lerComentarioDocLinha() *Token {
	inicio := l.posicaoAtual()
	for l.ch != '\n' && !l.fimDeArquivo() {
		l.avancar()
	}
	fim := l.posicaoAtual()
	return newToken(TokenComentarioDoc, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
}

// lerComentarioDocBloco consome um comentário de documentação de bloco iniciado por '/**'
func (l *Lexer) lerComentarioDocBloco() *Token {
	inicio := l.posicaoAtual()
	l.avancar() // consome '/'
	l.avancar() // consome '*'
	profundidade := 1

	for !l.fimDeArquivo() && profundidade > 0 {
		if l.ch == '/' && l.peek() == '*' {
			profundidade++
			l.avancar() // consome '/'
			l.avancar() // consome '*'
		} else if l.ch == '*' && l.peek() == '/' {
			profundidade--
			l.avancar() // consome '*'
			l.avancar() // consome '/'
		} else {
			l.avancar()
		}
	}

	fim := l.posicaoAtual()
	return newToken(TokenComentarioDoc, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
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

// isHexDigit verifica se a runa é um dígito hexadecimal válido (0-9, a-f, A-F).
func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// isBinDigit verifica se a runa é um dígito binário válido (0 ou 1).
func isBinDigit(r rune) bool {
	return r == '0' || r == '1'
}

// isOctDigit verifica se a runa é um dígito octal válido (0-7).
func isOctDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

// lerNumero consome sequencialmente dígitos, prefixos de base (0x, 0b, 0o),
// separadores de dígitos '_' e o caractere separador decimal '.' para fatiar literais numéricos.
func (l *Lexer) lerNumero() *Token {
	inicio := l.posicaoAtual()

	if l.ch == '0' {
		p := l.peek()
		if p == 'x' || p == 'X' {
			l.avancar() // consome '0'
			l.avancar() // consome 'x'/'X'
			digitos := 0
			for isHexDigit(l.ch) || l.ch == '_' {
				if isHexDigit(l.ch) {
					digitos++
				}
				l.avancar()
			}
			fim := l.posicaoAtual()
			if digitos == 0 {
				return newToken(TokenErro, "literal numérico malformado", inicio, fim)
			}
			return newToken(TokenInteiro, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
		} else if p == 'b' || p == 'B' {
			l.avancar() // consome '0'
			l.avancar() // consome 'b'/'B'
			digitos := 0
			for isBinDigit(l.ch) || l.ch == '_' {
				if isBinDigit(l.ch) {
					digitos++
				}
				l.avancar()
			}
			fim := l.posicaoAtual()
			if digitos == 0 {
				return newToken(TokenErro, "literal numérico malformado", inicio, fim)
			}
			return newToken(TokenInteiro, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
		} else if p == 'o' || p == 'O' {
			l.avancar() // consome '0'
			l.avancar() // consome 'o'/'O'
			digitos := 0
			for isOctDigit(l.ch) || l.ch == '_' {
				if isOctDigit(l.ch) {
					digitos++
				}
				l.avancar()
			}
			fim := l.posicaoAtual()
			if digitos == 0 {
				return newToken(TokenErro, "literal numérico malformado", inicio, fim)
			}
			return newToken(TokenInteiro, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
		}
	}

	for {
		l.avancar()

		if l.ch == 'e' || l.ch == 'E' {
			p := l.peek()
			if unicode.IsDigit(p) || p == '_' {
				l.avancar() // consome 'e'/'E'
			} else if (p == '+' || p == '-') && (unicode.IsDigit(l.peekN(2)) || l.peekN(2) == '_') {
				l.avancar() // consome 'e'/'E'
				l.avancar() // consome '+' ou '-'
			}
		}

		if !(unicode.IsDigit(l.ch) || l.ch == '.' || l.ch == '_') {
			break
		}
	}

	fim := l.posicaoAtual()
	valor := l.SubString(inicio.Indice, fim.Indice)

	tipo := TokenInteiro
	if strings.Contains(valor, ".") || strings.Contains(valor, "e") || strings.Contains(valor, "E") {
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
		l.avancar()

		if l.fimDeArquivo() || l.ch == '\n' {
			fim := l.posicaoAtual()
			return newToken(TokenErro, "string não terminada", inicio, fim)
		}

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

// lerTextoRaw consome uma cadeia literal multilinha delimitada por crases (`...`),
// sem tratar caracteres de escape.
func (l *Lexer) lerTextoRaw() *Token {
	inicio := l.posicaoAtual()
	l.avancar() // consome '`'

	for !l.fimDeArquivo() {
		if l.ch == '`' {
			l.avancar() // consome '`'
			fim := l.posicaoAtual()
			return newToken(TokenTexto, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
		}
		l.avancar()
	}

	fim := l.posicaoAtual()
	return newToken(TokenErro, "string raw não terminada", inicio, fim)
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
		pos := l.posicaoAtual()
		return &Token{
			Tipo:   TokenFimDeArquivo,
			Valor:  "",
			Inicio: pos,
			Fim:    pos,
		}
	}

	ch := l.ch
	inicio := l.posicaoAtual()

	switch ch {
	case '\n':
		l.avancar()
		return newToken(TokenNovaLinha, "\n", inicio, l.posicaoAtual())

	case '=':
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenIgualIgual, "==", inicio, l.posicaoAtual())
		}
		if l.peek() == '>' {
			l.avancar()
			l.avancar()
			return newToken(TokenSeta, "=>", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenIgual, "=", inicio, l.posicaoAtual())

	case '!':
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenDiferente, "!=", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenBoolNao, "!", inicio, l.posicaoAtual())

	case '+':
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenMaisIgual, "+=", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenMais, "+", inicio, l.posicaoAtual())

	case '-':
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenMenosIgual, "-=", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenMenos, "-", inicio, l.posicaoAtual())

	case '*':
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenAsteriscoIgual, "*=", inicio, l.posicaoAtual())
		}
		if l.peek() == '*' {
			l.avancar()
			l.avancar()
			return newToken(TokenPotencia, "**", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenAsterisco, "*", inicio, l.posicaoAtual())

	case '/':
		if l.peek() == '/' {
			if l.peekN(2) == '/' {
				return l.lerComentarioDocLinha()
			}
			l.ignorarComentario()
			return l.PróximoToken()
		}
		if l.peek() == '*' {
			if l.peekN(2) == '*' {
				return l.lerComentarioDocBloco()
			}
			l.ignorarComentarioBloco()
			return l.PróximoToken()
		}
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenBarraIgual, "/=", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenDivisao, "/", inicio, l.posicaoAtual())

	case '%':
		l.avancar()
		return newToken(TokenModulo, "%", inicio, l.posicaoAtual())

	case '<':
		if l.peek() == '!' && l.peekN(2) == '-' && l.peekN(3) == '-' {
			l.ignorarComentarioHTML()
			return l.PróximoToken()
		}
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenMenorOuIgual, "<=", inicio, l.posicaoAtual())
		}
		if l.peek() == '<' {
			l.avancar()
			l.avancar()
			return newToken(TokenDeslocEsquerda, "<<", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenMenorQue, "<", inicio, l.posicaoAtual())

	case '>':
		if l.peek() == '=' {
			l.avancar()
			l.avancar()
			return newToken(TokenMaiorOuIgual, ">=", inicio, l.posicaoAtual())
		}
		if l.peek() == '>' {
			l.avancar()
			l.avancar()
			return newToken(TokenDeslocDireita, ">>", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenMaiorQue, ">", inicio, l.posicaoAtual())

	case '(':
		l.avancar()
		return newToken(TokenAbreParenteses, "(", inicio, l.posicaoAtual())

	case ')':
		l.avancar()
		return newToken(TokenFechaParenteses, ")", inicio, l.posicaoAtual())

	case ';':
		l.avancar()
		return newToken(TokenPontoEVirgula, ";", inicio, l.posicaoAtual())

	case ',':
		l.avancar()
		return newToken(TokenVirgula, ",", inicio, l.posicaoAtual())

	case '{':
		l.avancar()
		return newToken(TokenAbreChaves, "{", inicio, l.posicaoAtual())

	case '}':
		l.avancar()
		return newToken(TokenFechaChaves, "}", inicio, l.posicaoAtual())

	case '[':
		l.avancar()
		return newToken(TokenAbreColchetes, "[", inicio, l.posicaoAtual())

	case ']':
		l.avancar()
		return newToken(TokenFechaColchetes, "]", inicio, l.posicaoAtual())

	case ':':
		l.avancar()
		return newToken(TokenDoisPontos, ":", inicio, l.posicaoAtual())

	case '@':
		l.avancar()
		return newToken(TokenArroba, "@", inicio, l.posicaoAtual())

	case '|':
		if l.peek() == '|' {
			if l.peekN(2) == '=' {
				l.avancar()
				l.avancar()
				l.avancar()
				return newToken(TokenBoolOuIgual, "||=", inicio, l.posicaoAtual())
			}
			l.avancar()
			l.avancar()
			return newToken(TokenBoolOu, "||", inicio, l.posicaoAtual())
		}
		if l.peek() == '>' {
			l.avancar()
			l.avancar()
			return newToken(TokenPipe, "|>", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenBitABitOu, "|", inicio, l.posicaoAtual())

	case '^':
		l.avancar()
		return newToken(TokenBitABitExOu, "^", inicio, l.posicaoAtual())

	case '&':
		if l.peek() == '#' || unicode.IsLetter(l.peek()) {
			l.lerEntidadeHTML()
			fim := l.posicaoAtual()
			return newToken(TokenEntidadeHTML, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
		}
		if l.peek() == '&' {
			if l.peekN(2) == '=' {
				l.avancar()
				l.avancar()
				l.avancar()
				return newToken(TokenBoolEIgual, "&&=", inicio, l.posicaoAtual())
			}
			l.avancar()
			l.avancar()
			return newToken(TokenBoolE, "&&", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenBitABitE, "&", inicio, l.posicaoAtual())

	case '~':
		l.avancar()
		return newToken(TokenBitABitNao, "~", inicio, l.posicaoAtual())

	case '.':
		if l.peek() == '.' && l.peekN(2) == '.' {
			l.avancar()
			l.avancar()
			l.avancar()
			return newToken(TokenTresPontos, "...", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenPonto, ".", inicio, l.posicaoAtual())

	case '?':
		if l.peek() == '?' {
			if l.peekN(2) == '=' {
				l.avancar()
				l.avancar()
				l.avancar()
				return newToken(TokenInterrogacaoInterrogacaoIgual, "??=", inicio, l.posicaoAtual())
			}
			l.avancar()
			l.avancar()
			return newToken(TokenInterrogacaoInterrogacao, "??", inicio, l.posicaoAtual())
		}
		if l.peek() == '.' {
			l.avancar()
			l.avancar()
			return newToken(TokenPontoInterrogacao, "?.", inicio, l.posicaoAtual())
		}
		l.avancar()
		return newToken(TokenInterrogacao, "?", inicio, l.posicaoAtual())

	case '#':
		l.ignorarComentario()
		return l.PróximoToken()

	case '"', '\'':
		return l.lerTexto()

	case '`':
		return l.lerTextoRaw()

	default:
		if unicode.IsLetter(ch) || ch == '_' {
			return l.lerIdentificador()
		}

		if unicode.IsDigit(ch) {
			return l.lerNumero()
		}
	}

	val := string(ch)
	l.avancar()
	return newToken(TokenErro, val, inicio, l.posicaoAtual())
}
