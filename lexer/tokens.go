// Package lexer implementa o analisador léxico (também conhecido como "scanner"
// ou "tokenizer") da linguagem Harpia.
//
// O pacote é responsável por:
//  1. Receber o código-fonte completo como uma string Unicode UTF-8.
//  2. Percorrer caractere a caractere (runa a runa) mantendo o controle de linha
//     e coluna para que mensagens de erro possam reportar coordenadas precisas.
//  3. Agrupar caracteres em lexemas e classificá-los em tipos de token (palavras
//     reservadas, identificadores, literais numéricos, strings, operadores, etc).
//  4. Entregar ao Parser uma sequência ordenada de *Token pronta para análise
//     sintática.
//
// O lexer é escrito inteiramente à mão em Go (sem expressões regulares ou
// geradores) para garantir performance próxima de O(n) sobre o tamanho em bytes
// da entrada e total previsibilidade do consumo de memória.
package lexer

// TokenType representa um identificador numérico único que categoriza cada
// token léxico. Os valores são atribuídos automaticamente pelo prelúdio `iota`
// de Go, em ordem sequencial de declaração.
//
// Cada TokenType carrega semântica suficiente para que o Parser faça a análise
// sintática sem precisar inspecionar o lexema bruto (string) novamente —
// otimização importante para compiladores reais.
type TokenType int

// Constantes que enumeram todos os tipos de tokens reconhecidos pelo lexer do
// Harpia. O comentário ao lado de cada constante indica o lexema canônico
// correspondente na linguagem.
//
// As constantes estão organizadas em blocos lógicos para facilitar a leitura
// humana e localizar rapidamente categorias (identificadores, operadores,
// palavras-chave etc.).
const (
	// TokenErro sinaliza a ocorrência de um caractere inválido ou falha léxica
	// em tempo de análise. Quando o Parser ou diagnóstico recebe um TokenErro
	// deve preferir disparar uma mensagem amigável ao usuário.
	TokenErro TokenType = iota

	// TokenFimDeArquivo indica o encerramento físico da leitura do script (EOF).
	TokenFimDeArquivo

	// TokenNovaLinha delimita instruções físicas no código (ex: \n ou \r\n).
	TokenNovaLinha

	// ============================================================================
	// IDENTIFICADORES E LITERAIS PRIMITIVOS
	// Identificadores são qualquer sequência alfanumérica que começa com letra
	// ou `_`. Os literais primitivos carregam seu lexema bruto em `Token.Valor`
	// para que o Parser possa interpretá-los posteriormente (ex: fazer parse
	// da string contida em um TokenTexto).
	// ============================================================================

	// TokenIdentificador representa nomes de variáveis, classes, funções e chaves (ex: minha_var).
	TokenIdentificador

	// TokenInteiro representa valores numéricos inteiros sem sinal (ex: 123).
	TokenInteiro

	// TokenDecimal representa números de ponto flutuante (ex: 123.45).
	TokenDecimal

	// TokenTexto representa literais textuais cercados por aspas (ex: "olá" ou 'olá').
	TokenTexto

	// ============================================================================
	// PALAVRAS-CHAVE RESERVADAS E PALAVRAS DE FLUXO
	// Cada constante abaixo corresponde a uma palavra escrita em português
	// brasileiro que não pode ser usada como identificador comum. O lexer
	// detecta-as consultando o mapa `tokensIdentificadores` (ver helpers.go).
	// ============================================================================

	TokenSe       // se (if) — desvio condicional.
	TokenSenao    // senao (else) — ramo alternativo de um TokenSe.
	TokenEnquanto // enquanto (while) — loop baseado em condição.
	TokenPara     // para (for) — iteração controlada.
	TokenRetorne  // retorne (return) — devolução de valor de função.
	TokenPare     // pare (break) — interrupção imediata do loop.
	TokenContinue // continue (continue) — salto para a próxima iteração.

	TokenTente      // tente (início do bloco try) — proteção contra exceções.
	TokenCapture    // capture (captura de exceção) — manipulador do erro.
	TokenFinalmente // finalmente (bloco finally) — sempre executado.

	TokenDe      // de (from) — origem em instruções de import.
	TokenImporte // importe (import) — inclusão de módulos externos.

	TokenVerdadeiro // Verdadeiro (true) — literal booleano verdadeiro.
	TokenFalso      // Falso (false) — literal booleano falso.
	TokenNulo       // Nulo (null/nil) — ausência explícita de valor.

	TokenVar   // var (declaração de variável mutável).
	TokenConst // const (declaração de constante imutável).
	TokenFunc  // func (declaração de função/método).

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

	TokenExportar // exportar

	TokenAssincrono // assincrono
	TokenAguarde    // aguarde
	TokenEstilo     // estilo

	TokenInterrogacaoInterrogacao // ?? (coalescência nula)
	TokenPontoInterrogacao         // ?. (encadeamento opcional)
	TokenInterrogacao              // ? (operador condicional ternário)

	TokenEnum      // enum (enumeração)
	TokenInterface // interface (contrato)
)

// PosicaoToken armazena as coordenadas espaciais precisas de um caractere no
// arquivo fonte. É a estrutura fundamental usada por todo o pipeline de
// tooling do Harpia (linter, formatador, LSP, REPL) para emitir mensagens de
// erro e traceback com referência exata ao ponto em que o problema ocorreu.
//
// Os três campos são mantidos em paralelo porque servem propósitos diferentes:
//   - Coluna: usado para render interativos que mostram setas ("^^^^") abaixo
//     do trecho problemático.
//   - Linha: usado em cabeçalhos de erro (ex: `arquivo.hrp:linha 12, coluna 5`).
//   - Indice: usado internamente para operações O(1) sobre slices de bytes,
//     sem precisar reescanear o arquivo fonte.
type PosicaoToken struct {
	Coluna int // O offset do caractere a partir do início da linha atual (base 1).
	Linha  int // O número da linha física no arquivo fonte (base 1).
	Indice int // O offset absoluto em bytes desde o início do arquivo (base 0).
}

// Token representa o elemento atômico resultante do processamento do
// analisador léxico. É a "menina dos olhos" do pipeline de compilação: tudo
// que o Parser enxerga vem em forma de Token.
//
// Tipicamente o Parser consome tokens de forma sequencial via chamadas
// repetidas a `Lexer.ProximoToken()` e nunca inspeciona o *PosicaoToken*
// diretamente; este fica disponível para diagnóstico tardio quando algum nó
// da AST precisa reportar erros referentes à sua origem no código-fonte.
type Token struct {
	Tipo  TokenType // A categoria estrutural do token.
	Valor string    // A string de texto exata extraída do código-fonte.

	Inicio *PosicaoToken // Coordenada do primeiro caractere que originou o token.
	Fim    *PosicaoToken // Coordenada imediatamente após o último caractere do token.
}

// newToken é o construtor padrão interno para alocação estruturada de novos
// Tokens. Centralizar a criação em uma única função evita que chamadas
// dispersas pelo pacote esqueçam de preencher `Inicio` ou `Fim`.
//
// Parâmetros:
//   - tipo: categoria léxica do token (ex: TokenInteiro, TokenSe).
//   - valor: lexema textual exato extraído da entrada.
//   - inicio: coordenada do primeiro caractere consumido.
//   - fim: coordenada imediatamente após o último caractere consumido.
//
// Retorna um *Token pronto para ser devolvido ao Parser.
func newToken(tipo TokenType, valor string, inicio, fim *PosicaoToken) *Token {
	return &Token{
		tipo,
		valor,
		inicio,
		fim,
	}
}
