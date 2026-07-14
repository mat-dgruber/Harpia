package lexer

// TokenType representa um identificador numérico único que categoriza cada token léxico.
// Mapeado de forma sequencial utilizando 'iota' em Go.
type TokenType int

// Relação de todos os tipos de tokens suportados e classificados pelo Lexer do Portuscript.
const (
	// TokenErro sinaliza a ocorrência de um caractere inválido ou falha léxica em tempo de análise.
	TokenErro TokenType = iota

	// TokenFimDeArquivo indica o encerramento físico da leitura do script (EOF).
	TokenFimDeArquivo

	// TokenNovaLinha delimita instruções físicas no código (ex: \n ou \r\n).
	TokenNovaLinha

	// ============================================================================
	// IDENTIFICADORES E LITERAIS PRIMITIVOS
	// ============================================================================

	// TokenIdentificador representa nomes de variáveis, classes, funções e chaves (ex: minha_var).
	TokenIdentificador

	// TokenInteiro representa valores numéricos inteiros sem sinal (ex: 123).
	TokenInteiro

	// TokenDecimal representa números de ponto flutuante (ex: 123.45).
	TokenDecimal

	// TokenTexto representa literais textuais cercados por aspas duplas (ex: "olá").
	TokenTexto

	// ============================================================================
	// PALAVRAS-CHAVE RESERVADAS E PALAVRAS DE FLUXO
	// ============================================================================

	TokenSe       // se (if)
	TokenSenao    // senao (else)
	TokenEnquanto // enquanto (while)
	TokenPara     // para (for)
	TokenRetorne  // retorne (return)
	TokenPare     // pare (break)
	TokenContinue // continue (continue)

	TokenDe      // de (from)
	TokenImporte // importe (import)

	TokenVerdadeiro // Verdadeiro (true)
	TokenFalso      // Falso (false)
	TokenNulo       // Nulo (null/nil)

	TokenVar   // var (declaração de variável)
	TokenConst // const (declaração de constante)
	TokenFunc  // func (declaração de função)

	// ============================================================================
	// OPERADORES ARITMÉTICOS, RELACIONAIS E DELIMITADORES
	// ============================================================================

	TokenIgual           // = (atribuição)
	TokenMais            // + (adição ou concatenação)
	TokenMenos           // - (subtração ou negativo)
	TokenAsterisco       // * (multiplicação)
	TokenPotencia        // ** (exponenciação)
	TokenDivisao         // / (divisão)
	TokenDivisaoInteira  // // (divisão de piso)
	TokenModulo          // % (resto da divisão)
	TokenMenorQue        // <
	TokenMenorOuIgual    // <=
	TokenIgualIgual      // ==
	TokenDiferente       // !=
	TokenMaiorQue        // >
	TokenMaiorOuIgual    // >=
	TokenAbreParenteses  // (
	TokenFechaParenteses // )
	TokenPontoEVirgula   // ;
	TokenVirgula         // ,
	TokenAbreChaves      // {
	TokenFechaChaves     // }
	TokenAbreColchetes   // [
	TokenFechaColchetes  // ]
	TokenDoisPontos      // :

	// ============================================================================
	// REATRIBUIÇÕES COM ACUMULADORES (OPERADORES COMPOSTOS)
	// ============================================================================

	TokenMaisIgual       // +=
	TokenMenosIgual      // -=
	TokenAsteriscoIgual  // *=
	TokenBarraIgual      // /=
	TokenBarraBarraIgual // //=

	// ============================================================================
	// OPERADORES LÓGICOS BOOLEANOS
	// ============================================================================

	TokenBoolOu  // ou (OR lógico)
	TokenBoolE   // e (AND lógico)
	TokenBoolNao // nao (NOT lógico)

	// ============================================================================
	// OPERADORES DE BITS (BITWISE)
	// ============================================================================

	TokenBitABitOu   // |
	TokenBitABitExOu // ^ (XOR)
	TokenBitABitE    // &
	TokenBitABitNao  // ~

	TokenDeslocEsquerda // << (deslocamento bitwise esquerda)
	TokenDeslocDireita  // >> (deslocamento bitwise direita)

	TokenPonto // . (acesso de atributo/método)

	// ============================================================================
	// PROGRAMAÇÃO ORIENTADA A OBJETOS (POO) E RECURSOS ESPECIAIS
	// ============================================================================

	TokenNova     // nova (instanciação de classe)
	TokenClasse   // classe (definição de classe)
	TokenEstende  // estende (herança)
	TokenSelf     // self (referência à própria instância - similar a 'this')
	TokenEstatico // estatico (atributos/métodos de classe estáticos)

	TokenAssegura // assegura (asserção lógica - assert)

	TokenEm // em (pertencimento/iteração - for item em colecao)

	TokenPipe   // |> (operador pipe para encadeamento de chamadas)
	TokenTestar // testar
)

// PosicaoToken armazena as coordenadas espaciais precisas de um caractere no arquivo fonte.
// Essencial para gerar traçados de traceback e diagnósticos amigáveis ao usuário.
type PosicaoToken struct {
	Coluna int // O offset do caractere a partir do início da linha atual (base 1).
	Linha  int // O número da linha física no arquivo fonte (base 1).
	Indice int // O offset absoluto em bytes desde o início do arquivo (base 0).
}

// Token representa o elemento atômico resultante do processamento do analisador léxico.
type Token struct {
	Tipo  TokenType // A categoria estrutural do token.
	Valor string    // A string de texto exata extraída do código-fonte.

	Inicio *PosicaoToken // Ponteiro marcando a coordenada de início do token.
	Fim    *PosicaoToken // Ponteiro marcando a coordenada de encerramento do token.
}

// newToken é o construtor padrão interno para alocação estruturada de novos Tokens.
func newToken(tipo TokenType, valor string, inicio, fim *PosicaoToken) *Token {
	return &Token{
		tipo,
		valor,
		inicio,
		fim,
	}
}
