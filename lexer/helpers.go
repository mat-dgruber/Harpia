package lexer

// tokensSimples é uma tabela de consulta rápida (hashmap) que associa caracteres individuais
// ou pequenas sequências consecutivas de símbolos específicos aos seus respectivos tipos de tokens.
//
// Decisão de Projeto / Como funciona:
// Durante a varredura manual de caracteres no Lexer, símbolos especiais como operadores e pontuações
// são resolvidos comparando-os diretamente com as chaves deste mapa em tempo constante O(1).
// Isso elimina a necessidade de árvores de decisão gigantescas ou complexas expressões regulares.
var tokensSimples = map[string]TokenType{
	"\n": TokenNovaLinha,

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

	// Reatribuições (atribuições com atalhos aritméticos)
	"+=":  TokenMaisIgual,
	"-=":  TokenMenosIgual,
	"*=":  TokenAsteriscoIgual,
	"/=":  TokenBarraIgual,
	"//=": TokenBarraBarraIgual,

	// Operadores Bitwise
	"|":  TokenBitABitOu,
	"^":  TokenBitABitExOu,
	"&":  TokenBitABitE,
	"~":  TokenBitABitNao,
	"<<": TokenDeslocEsquerda,
	">>": TokenDeslocDireita,

	// Recursos de acesso e encadeamento
	".":  TokenPonto,
	"|>": TokenPipe,
}

// tokensIdentificadores atua como a tabela de símbolos de palavras-chave reservadas do Portuscript.
//
// Decisão de Projeto / Como funciona:
// No fluxo de análise, assim que o Lexer identifica uma cadeia pura de letras (um identificador válido),
// ele realiza uma consulta rápida nesta tabela. Se o identificador coincidir com alguma palavra reservada
// (ex: "importe", "se", "func"), ele é promovido e classificado sob o token sintático dedicado correspondente.
// Caso contrário, permanece categorizado genericamente como 'TokenIdentificador' (uma variável comum).
var tokensIdentificadores = map[string]TokenType{
	"se":       TokenSe,
	"senao":    TokenSenao,
	"enquanto": TokenEnquanto,
	"para":     TokenPara,
	"retorne":  TokenRetorne,
	"pare":     TokenPare,
	"continue": TokenContinue,

	"de":      TokenDe,
	"importe": TokenImporte,

	"Verdadeiro": TokenVerdadeiro,
	"Falso":      TokenFalso,
	"Nulo":       TokenNulo,

	"var":    TokenVar,
	"const":  TokenConst,
	"func":   TokenFunc,
	"funcao": TokenFunc,

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

	"em": TokenEm,
}
