package lexer

// tokensSimples é uma tabela de consulta rápida (hashmap) que associa caracteres individuais
// ou pequenas sequências consecutivas de símbolos específicos aos seus respectivos tipos de tokens.
//
// Decisão de Projeto / Como funciona:
// Durante a varredura manual de caracteres no Lexer, símbolos especiais como operadores e pontuações
// são resolvidos comparando-os diretamente com as chaves deste mapa em tempo constante O(1).
// Isso elimina a necessidade de árvores de decisão gigantescas ou complexas expressões regulares.
//
// Organização interna do mapa (de cima para baixo):
//  1. Quebras de linha e operadores unários (incluindo atribuição simples).
//  2. Operadores aritméticos, de comparação e reatribuições com atalho.
//  3. Operadores bitwise e lógicos (com sinônimos `||`/`&&` para `ou`/`e`).
//  4. Pontuação estrutural (parênteses, chaves, colchetes, vírgula, dois-pontos).
//  5. Acessos e encadeamentos (`.`, `?.`, `...`, `|>`, `<|`).
//
// O mapa é consultado em duas etapas dentro do dispatcher de PróximoToken():
// primeiro tenta-se a sequência de 2 caracteres (maximal munch — prefere
// o token mais longo); em caso de miss, tenta-se a sequência de 1 caractere.
// Esse princípio garante que `==` nunca seja confundido com `=` seguido de `=`.
var tokensSimples = map[string]TokenType{
	"\n": TokenNovaLinha,

	"!":  TokenBoolNao,
	"=":  TokenIgual,
	"+":  TokenMais,
	"-":  TokenMenos,
	"*":  TokenAsterisco,
	"**": TokenPotencia,
	"/":  TokenDivisao,
	"//": TokenDivisaoInteira,
	"%":  TokenModulo,
	"<":  TokenMenorQue,
	"<=": TokenMenorOuIgual,
	"==": TokenIgualIgual,
	"!=": TokenDiferente,
	">":  TokenMaiorQue,
	">=": TokenMaiorOuIgual,
	"(":  TokenAbreParenteses,
	")":  TokenFechaParenteses,
	";":  TokenPontoEVirgula,
	",":  TokenVirgula,
	"{":  TokenAbreChaves,
	"}":  TokenFechaChaves,
	"[":  TokenAbreColchetes,
	"]":  TokenFechaColchetes,
	":":  TokenDoisPontos,
	"@":  TokenArroba,

	// Reatribuições (atribuições com atalhos aritméticos)
	"+=":  TokenMaisIgual,
	"-=":  TokenMenosIgual,
	"*=":  TokenAsteriscoIgual,
	"/=":  TokenBarraIgual,
	"//=": TokenBarraBarraIgual,
	"??=": TokenInterrogacaoInterrogacaoIgual,
	"||=": TokenBoolOuIgual,
	"&&=": TokenBoolEIgual,

	// Operadores Bitwise
	"|":  TokenBitABitOu,
	"^":  TokenBitABitExOu,
	"&":  TokenBitABitE,
	"~":  TokenBitABitNao,
	"<<": TokenDeslocEsquerda,
	">>": TokenDeslocDireita,

	// Operadores Lógicos (sinônimos de 'ou' e 'e')
	"||": TokenBoolOu,
	"&&": TokenBoolE,

	// Recursos de acesso e encadeamento
	".":   TokenPonto,
	"...": TokenTresPontos,
	"|>":  TokenPipe,
	"??":  TokenInterrogacaoInterrogacao,
	"?.":  TokenPontoInterrogacao,
	"?":   TokenInterrogacao,

	// Arrow function (params) => expr ou (params) => { ... }
	"=>": TokenSeta,
}

// tokensIdentificadores atua como a tabela de símbolos de palavras-chave reservadas do Harpia.
//
// Decisão de Projeto / Como funciona:
// No fluxo de análise, assim que o Lexer identifica uma cadeia pura de letras (um identificador válido),
// ele realiza uma consulta rápida nesta tabela. Se o identificador coincidir com alguma palavra reservada
// (ex: "importe", "se", "func"), ele é promovido e classificado sob o token sintático dedicado correspondente.
// Caso contrário, permanece categorizado genericamente como 'TokenIdentificador' (uma variável comum).
//
// O mapa inclui também `Verdadeiro`, `Falso` e `Nulo` (com iniciais
// maiúsculas), que apesar de serem identificadores válidos sintaticamente
// têm semântica fixa de valor booleano/nulo no compilador.
//
// Importante: este mapa deve permanecer em sincronia com a lista de palavras
// globais reconhecidas pelo linter (ver globalsLinter em cmd/checar.go) para
// evitar diagnósticos falsos de "identificador não declarado".
var tokensIdentificadores = map[string]TokenType{
	"se":       TokenSe,
	"senao":    TokenSenao,
	"enquanto": TokenEnquanto,
	"para":     TokenPara,
	"retorne":  TokenRetorne,
	"pare":     TokenPare,
	"continue": TokenContinue,
	"continua": TokenContinue,

	"de":      TokenDe,
	"importe": TokenImporte,

	"Verdadeiro": TokenVerdadeiro,
	"Falso":      TokenFalso,
	"Nulo":       TokenNulo,

	"var":       TokenVar,
	"const":     TokenConst,
	"constante": TokenConst,
	"func":      TokenFunc,
	"funcao":    TokenFunc,

	// Operadores lógicos
	"ou":  TokenBoolOu,
	"e":   TokenBoolE,
	"nao": TokenBoolNao,

	// Estruturas de POO e especiais
	"nova":     TokenNova,
	"classe":   TokenClasse,
	"estende":  TokenEstende,
	"self":     TokenSelf,
	"estatico": TokenEstatico,

	"assegura": TokenAssegura,
	"testar":   TokenTestar,

	"tente":      TokenTente,
	"capture":    TokenCapture,
	"finalmente": TokenFinalmente,
	"exportar":   TokenExportar,

	// Pattern matching
	"escolha": TokenEscolha,
	"caso":    TokenCaso,
	"quando":  TokenQuando,

	"em": TokenEm,

	"assincrono": TokenAssincrono,
	"aguarde":    TokenAguarde,
	"estilo":     TokenEstilo,
	"enum":       TokenEnum,
	"interface":  TokenInterface,
}
