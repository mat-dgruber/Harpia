package lexer

import (
	"strings"
	"unicode/utf8"

	"github.com/mat-dgruber/Harpia/compartilhado"
)

// Lexer é a máquina de estados que executa o processo de varredura (scanning) de caracteres
// e os transforma em uma sequência de tokens lógicos.
//
// O Lexer lê a string de entrada incrementalmente, gerenciando o cursor físico, a linha
// e a coluna atuais para alimentar diagnósticos precisos em caso de falhas de análise.
type Lexer struct {
	entrada string  // O código fonte completo em formato textual.
	carater string  // O caractere Unicode (runa) sob análise corrente no cursor.
	coluna  int     // O índice de coluna física do caractere sob análise (base 1).
	linha   int     // O número da linha física correspondente no código fonte (base 1).

	// Campos de suporte operacional
	tamanho   int   // Quantidade conceitual total de caracteres Unicode (runas) na entrada.
	indice    int   // Posição conceitual corrente do cursor em caracteres Unicode (base 0).
	byteCache []int // Cache pré-calculado de offsets de bytes para acesso aleatório O(1) rápido.
}

// NewLexer aloca e prepara uma nova instância operacional de Lexer para o código fonte informado.
//
// Na inicialização, calcula o número total de runas e gera a tabela de mapeamento de bytes
// (byteCache) para garantir fatiamentos ágeis de strings e, por fim, carrega o primeiro caractere.
func NewLexer(entrada string) *Lexer {
	l := &Lexer{
		entrada:   entrada,
		tamanho:   utf8.RuneCountInString(entrada),
		indice:    -1,
		byteCache: compartilhado.IndiceBytePorCarater(entrada),
	}

	l.avancar()

	return l
}

// fimDeArquivo verifica se o cursor de leitura do interpretador atingiu ou passou o limite físico
// de caracteres Unicode presentes no código fonte.
func (l *Lexer) fimDeArquivo() bool {
	return l.indice >= l.tamanho
}

// proximoCarater faz uma espreitadela (lookahead) de um caractere à frente do cursor de leitura.
//
// Permite que o analisador tome decisões condicionais de desvio de fluxo (como diferenciar
// o operador de atribuição '=' de igualdade '==' ou reatribuição '+='), sem avançar o cursor físico.
func (l *Lexer) proximoCarater() string {
	if l.indice+1 >= l.tamanho {
		return ""
	}

	return compartilhado.ObtemCaraterPorIndice(l.entrada, l.indice+1, l.byteCache)
}

// caraterRelativo espreita um caractere em uma posição relativa ao cursor atual de forma segura
func (l *Lexer) caraterRelativo(offset int) string {
	if l.indice+offset >= l.tamanho || l.indice+offset < 0 {
		return ""
	}
	return compartilhado.ObtemCaraterPorIndice(l.entrada, l.indice+offset, l.byteCache)
}

// ignorarComentarioHTML pula todos os caracteres até encontrar a sequência de fechamento de comentário HTML '-->'
func (l *Lexer) ignorarComentarioHTML() {
	// Avança os 4 caracteres iniciais '<!--'
	l.avancar() // consome '<'
	l.avancar() // consome '!'
	l.avancar() // consome '-'
	l.avancar() // consome '-'

	for !l.fimDeArquivo() {
		if l.carater == "-" && l.proximoCarater() == "-" {
			l.avancar() // consome o primeiro '-'
			l.avancar() // consome o segundo '-'
			if l.carater == ">" {
				l.avancar() // consome o '>'
				break
			}
			continue
		}
		l.avancar()
	}
}

// avancar incrementa a posição do cursor em um caractere e carrega a nova runa correspondente.
//
// Se o caractere recém-carregado for uma quebra de linha ('\n'), o Lexer atualiza as coordenadas
// físicas, incrementando o contador de linhas e zerando o cursor de colunas. Do contrário, avança a coluna.
func (l *Lexer) avancar() {
	if l.fimDeArquivo() {
		return
	}

	l.indice += 1
	l.carater = compartilhado.ObtemCaraterPorIndice(l.entrada, l.indice, l.byteCache)

	if l.carater == "\n" {
		l.linha += 1
		l.coluna = 0
		return
	}

	l.coluna += 1
}

// posicaoAtual captura as coordenadas geográficas (linha, coluna, índice) do caractere corrente no código.
func (l *Lexer) posicaoAtual() *PosicaoToken {
	return &PosicaoToken{l.coluna, l.linha, l.indice}
}

// ignorarEspacos avança o cursor descartando e pulando caracteres inofensivos de espaço em branco e tabulação.
func (l *Lexer) ignorarEspacos() {
	for (l.carater == " " || l.carater == "\t") && !l.fimDeArquivo() {
		l.avancar()
	}
}

// ignorarComentario pula todos os caracteres que compõem uma linha de comentário (iniciada pelo caractere '#')
// até encontrar o delimitador de quebra de linha ('\n') ou o término do arquivo.
func (l *Lexer) ignorarComentario() {
	for (l.carater != "\n") && !l.fimDeArquivo() {
		l.avancar()
	}
	l.avancar()
}

// subString é um fatiador seguro que extrai uma partição da string original delimitada
// por posições em caracteres Unicode, convertendo-as de forma segura e rápida em offsets de bytes.
func (l *Lexer) subString(inicio, fim int) string {
	inicioByte := compartilhado.IndiceCaraterParaByte(l.entrada, inicio, l.byteCache)
	fimByte := compartilhado.IndiceCaraterParaByte(l.entrada, fim, l.byteCache)
	return l.entrada[inicioByte:fimByte]
}

// lerIdentificador consome sequencialmente caracteres alfanuméricos ou sublinhas (_) que compõem
// o nome de uma variável ou instrução.
//
// Ao final da leitura, faz a verificação na tabela de palavras reservadas (tokensIdentificadores).
// Se coincidir, o token genérico é promovido à palavra-chave correspondente (ex: TokenSe).
func (l *Lexer) lerIdentificador() *Token {
	inicio := l.posicaoAtual()

	for {
		l.avancar()

		if !(compartilhado.ContemApenasAlfaNum(l.carater) || l.carater == "_") {
			break
		}
	}

	fim := l.posicaoAtual()
	valor := l.subString(inicio.Indice, fim.Indice)
	tipo := TokenIdentificador

	if t, ok := tokensIdentificadores[valor]; ok {
		tipo = t
	}

	return newToken(tipo, valor, inicio, fim)
}

// lerNumero consome sequencialmente dígitos e o caractere separador decimal '.' para fatiar literais numéricos.
//
// Diferencia de forma automatizada números inteiros (TokenInteiro) de dízimas ou reais (TokenDecimal).
func (l *Lexer) lerNumero() *Token {
	inicio := l.posicaoAtual()

	for {
		l.avancar()

		if !(compartilhado.ContemApenasDigitos(l.carater) || l.carater == ".") {
			break
		}
	}

	fim := l.posicaoAtual()
	valor := l.subString(inicio.Indice, fim.Indice)

	tipo := TokenInteiro
	if strings.Contains(valor, ".") {
		tipo = TokenDecimal
	}

	return newToken(tipo, valor, inicio, fim)
}

// lerTexto consome uma cadeia literal delimitada por aspas simples (') ou aspas duplas ("),
// tratando de forma transparente sequências de escape do delimitador (ex: \" ou \').
func (l *Lexer) lerTexto() *Token {
	inicio := l.posicaoAtual()
	delimitador := l.carater

	for {
		l.avancar()

		if l.carater == delimitador {
			l.avancar()
			break
		}

		if l.carater == "\\" && l.proximoCarater() == delimitador {
			l.avancar()
		}
	}

	fim := l.posicaoAtual()
	return newToken(TokenTexto, l.subString(inicio.Indice, fim.Indice), inicio, fim)
}

// ProximoToken é o ponto de entrada principal do analisador léxico.
//
// Executa a máquina de transições lógica: ignora espaços em branco, lida com comentários,
// avalia caracteres únicos nos dicionários estáticos de operadores, consome números, strings,
// identificadores de palavras-chave e retorna a próxima estrutura Token de forma sequencial.
func (l *Lexer) ProximoToken() *Token {
	l.ignorarEspacos()

	if l.fimDeArquivo() {
		return &Token{
			Tipo:   TokenFimDeArquivo,
			Valor:  "",
			Inicio: l.posicaoAtual(),
			Fim:    l.posicaoAtual(),
		}
	}

	carater := l.carater
	inicio := l.posicaoAtual()

	// Trata comentários de linha única '#' ou '//' (deve rodar antes de operadores simples para evitar colisão com '/')
	if carater == "#" || (carater == "/" && l.proximoCarater() == "/") {
		l.ignorarComentario()
		return l.ProximoToken()
	}

	// ponytail: trata comentários HTML '<!-- ... -->' de forma nativa para excelente DX no JSX
	if carater == "<" && l.proximoCarater() == "!" {
		if l.caraterRelativo(2) == "-" && l.caraterRelativo(3) == "-" {
			l.ignorarComentarioHTML()
			return l.ProximoToken()
		}
	}

	// Se for um caractere operador catalogado ou operador de negação '!'
	if tipo, ok := tokensSimples[carater]; ok || carater == "!" {
		for {
			if l.fimDeArquivo() {
				break
			}
			l.avancar()
			// Guloso: tenta aglutinar caracteres subsequentes para formar operadores compostos (ex: '=' + '=' = '==')
			if t, ok := tokensSimples[carater+l.carater]; ok {
				carater += l.carater
				tipo = t
				continue
			}

			break
		}

		return newToken(tipo, carater, inicio, l.posicaoAtual())
	}

	switch carater {
	case "\"", "'":
		return l.lerTexto()
	default:
		// Se iniciar com letra ou sublinha, lê como identificador
		if compartilhado.ContemApenasLetras(carater) || carater == "_" {
			return l.lerIdentificador()
		}

		// Se iniciar com dígito, lê como literal numérico
		if compartilhado.ContemApenasDigitos(carater) {
			return l.lerNumero()
		}
	}

	// Caso não coincida com nenhuma regra, sinaliza um erro léxico
	return &Token{Tipo: TokenErro, Valor: l.carater}
}
