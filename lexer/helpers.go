package lexer

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
