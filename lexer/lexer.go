package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Bit flags para classificação O(1) de caracteres na tabela de consulta ASCII.
//
// Estas flags são empacotadas em bits dentro de uma única variável byte para que um
// caractere possa pertencer a múltiplas categorias simultaneamente (por exemplo,
// 'a' é simultaneamente flagLetra e flagHex), evitando múltiplas tabelas paralelas
// e mantendo a consulta de classificação em uma única operação de máscara de bits.
const (
	flagEspaco byte = 1 << iota // Espaço em branco: ' ', '\t', '\r', '\n'
	flagDigito                  // Dígito decimal: '0'-'9'
	flagLetra                   // Letra de identificador ASCII: 'a'-'z', 'A'-'Z', '_'
	flagHex                     // Dígito hexadecimal: '0'-'9', 'a'-'f', 'A'-'F'
)

// tabelaASCII permite classificação O(1) de caracteres ASCII (0..255).
//
// Cada posição indexada pelo código numérico do caractere armazena um OR bitwise
// das flags acima. Esta abordagem substitui uma cadeia de `if/else` ou chamadas a
// `unicode.Is*` da stdlib para o caso comum (ASCII), que representa >95% do
// código-fonte típico e é a parte quente do loop principal de tokenização.
var tabelaASCII [256]byte

// init pré-popula tabelaASCII executando uma única vez no carregamento do pacote.
//
// O custo desta inicialização é amortizado: ela roda apenas quando o binário é
// carregado, e o resultado é compartilhado por todas as instâncias de Lexer
// durante o ciclo de vida do processo. Estratégia deliberada para evitar
// reinicializações repetidas a cada chamada de NewLexer.
func init() {
	tabelaASCII[' '] |= flagEspaco
	tabelaASCII['\t'] |= flagEspaco
	tabelaASCII['\r'] |= flagEspaco
	tabelaASCII['\n'] |= flagEspaco

	for c := '0'; c <= '9'; c++ {
		tabelaASCII[byte(c)] |= flagDigito | flagHex
	}
	for c := 'a'; c <= 'f'; c++ {
		tabelaASCII[byte(c)] |= flagHex | flagLetra
		tabelaASCII[byte(c-'a'+'A')] |= flagHex | flagLetra
	}
	for c := 'g'; c <= 'z'; c++ {
		tabelaASCII[byte(c)] |= flagLetra
		tabelaASCII[byte(c-'g'+'G')] |= flagLetra
	}
	tabelaASCII['_'] |= flagLetra
}

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
//
// Pré-condições:
//   - O parâmetro `entrada` deve conter o código-fonte completo (idealmente já
//     lido do disco pelo chamador). Strings vazias são aceitas e resultam em um
//     lexer cujo primeiro PróximoToken() retorna imediatamente TokenFimDeArquivo.
//
// Comportamento inicial:
//  1. Converte a string para []byte (evita realocações em chamadas sucessivas
//     a SubString e peekN).
//  2. Posiciona linha=1, coluna=0 e dispara um avançar() para carregar o
//     primeiro caractere no campo ch.
//  3. Se o primeiro caractere for '#' e o segundo (peek) for '!', trata-se
//     de um shebang (ex: #!/usr/bin/env harpia) — o conteúdo é descartado
//     para que scripts executáveis em Linux/macOS possam ser tokenizados
//     sem erros léxicos.
//
// Retorna um *Lexer pronto para alimentar o Parser via PróximoToken().
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

// NovoLexer aloca e prepara uma nova instância operacional de Lexer para o código fonte e nome de arquivo informados.
//
// Atualmente o parâmetro variádico `arquivo` é ignorado — mantido apenas como
// ponto de extensão futuro para associar o lexer a um caminho de arquivo
// (útil quando mensagens de erro precisarem reportar nome do arquivo de origem).
//
// Use NewLexer em código novo; NovoLexer existe por compatibilidade com
// integrações legadas.
func NovoLexer(entrada string, arquivo ...string) *Lexer {
	return NewLexer(entrada)
}

// ObterTokens executa o lexer até o fim do arquivo e retorna todos os tokens acumulados.
//
// Atalho de conveniência sobre PróximoToken() para consumidores que preferem
// materializar a sequência inteira de uma vez (ex: REPL, ferramentas de
// análise estática, testes).
//
// Importante: a função só termina após receber um TokenFimDeArquivo. Se o
// código-fonte contiver tokens malformados (ex: string não terminada), o
// último token retornado será um TokenErro e o loop se encerra normalmente.
//
// Retorna um slice []*Token contendo todos os tokens, incluindo o
// TokenFimDeArquivo final como sentinela.
func (l *Lexer) ObterTokens() []*Token {
	var tokens []*Token
	for {
		tok := l.PróximoToken()
		tokens = append(tokens, tok)
		if tok.Tipo == TokenFimDeArquivo {
			break
		}
	}
	return tokens
}

// fimDeArquivo verifica se o cursor de leitura do interpretador atingiu ou passou o limite físico
// de bytes presentes no código fonte.
func (l *Lexer) fimDeArquivo() bool {
	return l.offset >= len(l.fonte)
}

// peek faz uma espreitadela (lookahead) de 1 caractere à frente do cursor de leitura.
//
// Não consome o caractere — o cursor permanece na mesma posição. Usada para
// desambiguação de operadores compostos (ex: distinguir `=` de `==` ou `=>`).
//
// Particularidade importante: converte silenciosamente '\r' em '\n' para que
// o restante do lexer não precise tratar CRLF como caso especial. Isso
// garante comportamento consistente entre arquivos gerados no Windows
// (CRLF) e Unix (LF apenas).
//
// Retorna 0 (rune zero) quando o cursor está no fim do arquivo, permitindo
// que o chamador use comparações simples sem precisar verificar EOF
// separadamente.
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
//
// Quando n <= 0, devolve o caractere corrente (l.ch) como sentinela de
// "sem avanço", evitando que chamadores precisem ramificar antes de invocar.
//
// Percorre exatamente n runas UTF-8 adiante, decodificando uma a uma sem
// alocar buffers intermediários. Retorna 0 se o limite do arquivo for
// atingido antes de completar os n saltos.
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
//
// Suporta o formato tradicional <!-- ... --> encontrado em arquivos .html ou em
// fragmentos Harpia embutidos em JSX. Não é recursivo nem reconhece blocos
// aninhados (semântica padrão do HTML).
//
// Caso a sequência de fechamento nunca apareça, o lexer consome até o fim do
// arquivo e devolve o controle — comportamento deliberado para permitir
// recuperação graciosa em arquivos malformados.
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
//
// Comportamento detalhado:
//  1. Avança o cursor físico para proximoOffset e decodifica a runa UTF-8
//     correspondente.
//  2. Normaliza CRLF em fluxo: quando a runa decodificada é '\r', inspeciona
//     o próximo byte. Se for '\n', consome ambos como uma única quebra de
//     linha (padrão Windows). Caso contrário, converte o '\r' solitário
//     em '\n' (padrão Mac clássico).
//  3. Atualiza os contadores espaciais: linha++ e coluna=0 ao detectar '\n';
//     caso contrário, incrementa coluna em uma unidade lógica (não conta
//     bytes adicionais de uma runa multibyte como colunas separadas).
//
// No fim do arquivo, define ch=0 e largura=0 — sentinela que peek() e
// PróximoToken() usam para detectar EOF sem precisar comparar offsets.
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
//
// Retornar por valor (não por ponteiro) é proposital: PosicaoToken é uma struct
// pequena (3 ints = 24 bytes em arquiteturas 64 bits), e cópias por valor
// evitam fugas para a heap e mantêm o padrão zero-alloc do lexer.
func (l *Lexer) posicaoAtual() PosicaoToken {
	return PosicaoToken{
		Coluna: l.coluna,
		Linha:  l.linha,
		Indice: l.offset,
	}
}

// ignorarEspacos avança o cursor descartando e pulando caracteres inofensivos de espaço em branco e tabulação.
//
// Diferente de outras linguagens, Harpia trata espaços e tabulações como
// insignificantes entre tokens. As quebras de linha (\n), por outro lado, são
// significativas e disparam TokenNovaLinha — por isso não são consumidas aqui.
//
// Para garantir comportamento consistente, o método também respeita o limite
// do arquivo para que arquivos terminados abruptamente em whitespace não
// causem loop infinito.
func (l *Lexer) ignorarEspacos() {
	for (l.ch == ' ' || l.ch == '\t') && !l.fimDeArquivo() {
		l.avancar()
	}
}

// ignorarComentario pula todos os caracteres que compõem uma linha de comentário (iniciada pelo caractere '#')
//
// Consumido pelo caminho de shebang em NewLexer() e pela detecção de '#'
// no dispatcher principal de PróximoToken(). Não inclui o caractere '#'
// inicial nem retorna TokenComentarioDoc — comentários de linha são
// puramente descartáveis, sem semântica que precise ser preservada na
// sequência de tokens.
func (l *Lexer) ignorarComentario() {
	for l.ch != '\n' && !l.fimDeArquivo() {
		l.avancar()
	}
	l.avancar()
}

// ignorarComentarioBloco pula todos os caracteres até encontrar a sequência de fechamento de comentário de bloco '*/'
//
// Diferente da maioria das linguagens, Harpia suporta comentários de bloco
// ANINHADOS: o contador `profundidade` rastreia quantos `/*` foram abertos
// e exige o mesmo número de `*/` para fechar. Isso permite comentar trechos
// de código que já contêm outros blocos comentados.
//
// Caso o arquivo termine antes do balanceamento ser zerado, o método sai
// silenciosamente — cabe ao parser diagnosticar blocos não fechados
// posteriormente, se aplicável.
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
//
// Diferente do comentário de linha comum (`//` ou `#`), este preserva o conteúdo
// como um TokenComentarioDoc — destinado a alimentar geradores de documentação
// automática, LSP (Language Server Protocol) e ferramentas de inspeção.
//
// O cursor avança até (mas não consome) o caractere de nova linha, deixando-o
// para o dispatcher principal emitir TokenNovaLinha na próxima iteração.
func (l *Lexer) lerComentarioDocLinha() *Token {
	inicio := l.posicaoAtual()
	for l.ch != '\n' && !l.fimDeArquivo() {
		l.avancar()
	}
	fim := l.posicaoAtual()
	return newToken(TokenComentarioDoc, l.SubString(inicio.Indice, fim.Indice), inicio, fim)
}

// lerComentarioDocBloco consome um comentário de documentação de bloco iniciado por '/**'
//
// Assim como o comentário de bloco comum, suporta aninhamento: blocos `/**`
// podem conter outros blocos `/* */` sem fechar prematuramente. Esta função
// é estruturalmente equivalente a ignorarComentarioBloco, exceto que captura
// o conteúdo textual resultante em vez de descartá-lo.
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
//
// Aplica três clamps defensivos antes de fatiar:
//   - `inicio < 0` → trata como 0 (proteção contra chamadas vindas de
//     lógicas de rollback maliciosas ou com snapshots antigos).
//   - `fim > len(fonte)` → limita ao tamanho real do arquivo.
//   - `inicio >= fim` → retorna string vazia em vez de panic por slice bounds.
//
// Caso todas as pré-condições sejam satisfeitas, retorna uma cópia do
// trecho como string (não como []byte), convertendo sob demanda para que
// o consumidor não precise lidar com bytes brutos.
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

// isLetterOrUnderscore verifica se a runa é uma letra ou '_' usando O(1) na tabela ASCII.
//
// Caracteres fora da faixa ASCII (>=128) caem no fallback unicode.IsLetter,
// preservando suporte a identificadores Unicode (acentos, ideogramas,
// emojis estruturais como 🦚).
func isLetterOrUnderscore(r rune) bool {
	if r >= 0 && r < 128 {
		return tabelaASCII[byte(r)]&flagLetra != 0
	}
	return unicode.IsLetter(r)
}

// isDigit verifica se a runa é um dígito decimal usando O(1) na tabela ASCII.
//
// Aceita exclusivamente dígitos ASCII 0-9 via tabela; dígitos em outros
// alfabetos (ex: árabe, devanagari) não são considerados dígitos pelo lexer,
// forçando o uso de ASCII para literais numéricos — comportamento
// consistente com a maioria das linguagens mainstream.
func isDigit(r rune) bool {
	if r >= 0 && r < 128 {
		return tabelaASCII[byte(r)]&flagDigito != 0
	}
	return unicode.IsDigit(r)
}

// lerIdentificador consome sequencialmente caracteres alfanuméricos ou sublinhas (_) que compõem
// o nome de uma variável, classe, função ou palavra reservada.
//
// Comportamento passo-a-passo:
//  1. Captura a posição inicial antes de consumir qualquer caractere.
//  2. Avança repetidamente enquanto o caractere corrente for letra, '_' ou dígito.
//     (O dígito é aceito APÓS o primeiro caractere — identificadores em Harpia
//     não podem começar com número, alinhado com convenções C/JS/Go).
//  3. Extrai o lexema via SubString.
//  4. Consulta o mapa `tokensIdentificadores` para promoção a palavra-chave
//     se o lexema coincidir com uma chave. Caso contrário, devolve
//     TokenIdentificador (variável genérica).
//
// Retorna um *Token com Tipo apropriado, Valor sendo o lexema bruto, e
// Inicio/Fim demarcando o intervalo exato ocupado no código-fonte.
func (l *Lexer) lerIdentificador() *Token {
	inicio := l.posicaoAtual()

	for {
		l.avancar()

		if !(isLetterOrUnderscore(l.ch) || isDigit(l.ch)) {
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

// isHexDigit verifica se a runa é um dígito hexadecimal válido (0-9, a-f, A-F) usando O(1) ASCII lookup table.
//
// Como o conjunto hexadecimal está contido no ASCII (128 caracteres), não há
// fallback para unicode — qualquer runa fora desta faixa é considerada não-hex.
func isHexDigit(r rune) bool {
	return r >= 0 && r < 128 && (tabelaASCII[byte(r)]&flagHex != 0)
}

// isBinDigit verifica se a runa é um dígito binário válido (0 ou 1).
//
// Implementação inline trivial (em vez de consultar tabelaASCII) porque
// existem apenas dois valores válidos — a comparação direta é mais rápida
// e igualmente legível.
func isBinDigit(r rune) bool {
	return r == '0' || r == '1'
}

// isOctDigit verifica se a runa é um dígito octal válido (0-7).
//
// Idem isBinDigit: range ASCII inline é mais barato que indireção em tabela.
func isOctDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

// lerNumero consome sequencialmente dígitos, prefixos de base (0x, 0b, 0o),
// separadores de dígitos '_' e o caractere separador decimal '.' para fatiar literais numéricos.
//
// Suporta 4 bases numéricas distintas:
//   - Hexadecimal: prefixo 0x ou 0X, dígitos 0-9 a-f A-F, separadores '_'.
//   - Binário: prefixo 0b ou 0B, dígitos 0 e 1, separadores '_'.
//   - Octal: prefixo 0o ou 0O, dígitos 0-7, separadores '_'.
//   - Decimal (default): dígitos 0-9 com possível separador decimal '.' e
//     expoente 'e'/'E' (com sinal opcional).
//
// Detecção de tipo: ao final do consumo, decide entre TokenInteiro e
// TokenDecimal baseado na presença de '.', 'e' ou 'E' no lexema bruto.
// A validação real do valor (overflow, dígitos duplicados, separador
// mal posicionado) fica a cargo do parser/compilador.
//
// Caso a base seja explícita e nenhum dígito válido siga o prefixo
// (ex: "0x"), emite TokenErro com a mensagem "literal numérico malformado"
// para alimentar diagnósticos amigáveis.
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
			if isDigit(p) || p == '_' {
				l.avancar() // consome 'e'/'E'
			} else if (p == '+' || p == '-') && (isDigit(l.peekN(2)) || l.peekN(2) == '_') {
				l.avancar() // consome 'e'/'E'
				l.avancar() // consome '+' ou '-'
			}
		}

		if !(isDigit(l.ch) || l.ch == '.' || l.ch == '_') {
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
//
// Não produz um token por si só — quem chama (dispatcher principal de '&')
// é responsável por emitir o TokenEntidadeHTML englobando o intervalo
// consumido. Esta função apenas move o cursor adiante até o terminador
// correto.
//
// Termina o consumo ao encontrar ';', espaço em branco ou fim de linha.
// Espaços/brancos servem como terminadores tolerantes para garantir
// recuperação graciosa em documentos HTML malformados (entidade sem
// ponto-e-vírgula final).
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
//
// O delimitador de abertura é capturado em `delimitador` antes de iniciar a
// iteração, permitindo que a string use o mesmo caractere para abrir e
// fechar sem ambiguidade (semelhante a Go).
//
// Comportamentos importantes:
//   - Sequências de escape '\\' fazem o lexer pular o próximo caractere,
//     permitindo que o delimitador apareça literal dentro da string sem
//     encerrá-la (ex: "ela disse \"olá\"").
//   - Quebra de linha não escapada encerra prematuramente como TokenErro
//     "string não terminada" — strings convencionais não podem cruzar
//     linhas físicas; para isso, use lerStringBruta (crase).
//   - EOF antes do delimitador de fechamento também resulta em TokenErro.
//
// Retorna um *TokenTexto com o conteúdo bruto (incluindo delimitadores),
// ficando a cargo do parser remover as aspas e processar escapes
// semânticos (ex: \n → quebra de linha real).
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

// lerStringBruta consome uma cadeia literal multilinha delimitada por crases (`...`),
// sem tratar caracteres de escape.
//
// Inspirada nas raw strings do Rust/C#: tudo entre as crases é preservado
// verbatim, incluindo barras invertidas, aspas e quebras de linha. Útil
// para armazenar HTML, JSON, expressões regulares ou trechos de código
// sem precisar duplicar cada caractere de escape.
//
// Emite TokenErro "string raw não terminada" se o arquivo acabar antes
// da crase de fechamento.
func (l *Lexer) lerStringBruta() *Token {
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
//
// Estrutura por valor (sem ponteiros) usada em conjunto com Estado() e
// Restaurar() para implementar backtracking no lexer — recurso essencial
// quando o parser precisa "experimentar" uma interpretação e desfazer
// caso a hipótese se mostre falsa.
//
// Todos os campos são cópias privadas do estado interno do Lexer, evitando
// que código externo manipule o cursor sem passar pelos métodos oficiais.
type SnapshotLex struct {
	offset        int
	proximoOffset int
	ch            rune
	largura       int
	linha         int
	coluna        int
}

// Estado retorna um snapshot imutável do estado atual do lexer.
//
// O consumidor típico é o parser, quando decide tentar uma nova regra
// sem perder a posição original. Use o padrão:
//
//	estado := lexer.Estado()
//	// ... tenta nova interpretação, possivelmente avançando o cursor ...
//	if err != nil {
//	    lexer.Restaurar(estado)
//	}
//
// O retorno por valor evita alocação na heap, mantendo o padrão zero-alloc.
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
//
// Restaura TODOS os campos do cursor (offset, proximoOffset, ch, largura,
// linha, coluna) garantindo que o lexer volte exatamente ao mesmo ponto
// de quando Estado() foi chamado.
//
// Cuidado: passar um SnapshotLex não proveniente de Estado() neste mesmo
// lexer pode produzir estado inconsistente — os campos não são validados.
func (l *Lexer) Restaurar(s SnapshotLex) {
	l.offset = s.offset
	l.proximoOffset = s.proximoOffset
	l.ch = s.ch
	l.largura = s.largura
	l.linha = s.linha
	l.coluna = s.coluna
}

// PróximoToken é o ponto de entrada principal do analisador léxico.
//
// É a única função que o Parser chama em loop até receber TokenFimDeArquivo.
// Internamente delega para funções especializadas conforme a categoria do
// caractere sob análise:
//   - Operadores/pontuação: dispatch DFA por switch sobre ch + peek/peekN.
//   - Strings/identificadores/números: funções ler* correspondentes.
//   - Comentários: ignorados (comuns) ou emitidos como TokenComentarioDoc.
//
// Cada chamada consome pelo menos um caractere (exceto quando o cursor já
// está em EOF, caso em que devolve TokenFimDeArquivo imediatamente),
// garantindo progresso garantido no loop de tokenização.
//
// Comportamento especial para EOF:
//   - A primeira chamada após o fim do arquivo devolve TokenFimDeArquivo.
//   - Chamadas subsequentes também devolvem TokenFimDeArquivo sem efeito
//     colateral — característica segura para parsers que não verificam
//     se a chamada anterior foi a final.
//
// Retorna TokenErro com mensagem descritiva caso encontre um caractere
// que não pode ser classificado em nenhuma categoria válida.
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
		return l.lerStringBruta()

	default:
		if isLetterOrUnderscore(ch) {
			return l.lerIdentificador()
		}

		if isDigit(ch) {
			return l.lerNumero()
		}
	}

	val := string(ch)
	l.avancar()
	return newToken(TokenErro, val, inicio, l.posicaoAtual())
}
