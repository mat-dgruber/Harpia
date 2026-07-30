lexer grammar HarpiaLexer;

// ============================================================================
// ESPECIFICAÇÃO LÉXICA DO HARPIA (ANTLR4 REFERENCE)
// ============================================================================
//
// Este arquivo descreve a gramática léxica formal da linguagem Harpia.
//
// ponytail: Embora seja a especificação formal em ANTLR4, o compilador físico do
// Harpia utiliza um Lexer escrito à mão em Go (/lexer/lexer.go) por motivos
// de performance extrema, suporte completo e resiliente a caracteres Unicode
// multibyte (UTF-8) e para emissão de tracebacks e relatórios de diagnóstico ricos.
//
// Esta especificação serve como contrato de fidelidade sintática para o analisador.

// ============================================================================
// TOKENS PRIMORDIAIS E VALORES LITERAIS
// ============================================================================

// FALSO representa o literal booleano negativo 'Falso'.
FALSO: 'Falso';

// VERDADEIRO representa o literal booleano positivo 'Verdadeiro'.
VERDADEIRO: 'Verdadeiro';

// NULO representa a constante nula 'Nulo' (representação de ausência de tipo).
NULO: 'Nulo';

// ============================================================================
// PALAVRAS-CHAVE RESERVADAS E ESTRUTURAS DE CONTROLE
// ============================================================================

SE: 'se';
SENAO: 'senao';
ENQUANTO: 'enquanto';
PARA: 'para';
EM: 'em';

VAR: 'var';
CONST: 'const';
FUNC: 'func';
RETORNE: 'retorne';
PARE: 'pare';
CONTINUE: 'continue';

TENTE: 'tente';
CAPTURE: 'capture';
FINALMENTE: 'finalmente';

ESCOLHA: 'escolha';
CASO: 'caso';
QUANDO: 'quando';

IMPORTE: 'importe';
DE: 'de';
EXPORTAR: 'exportar';

NOVA: 'nova';
CLASSE: 'classe';
ESTENDE: 'estende';
SELF: 'self';
ESTATICO: 'estatico';

ENUM: 'enum';
INTERFACE: 'interface';

ASSINCRONO: 'assincrono';
AGUARDE: 'aguarde';

ESTILO: 'estilo';
TESTAR: 'testar';
ASSEGURA: 'assegura';

// ============================================================================
// OPERADORES LÓGICOS TEXTUAIS (EM PORTUGUÊS)
// ============================================================================

OU: 'ou';
E: 'e';
NAO: 'nao';

// ============================================================================
// OPERADORES ARITMÉTICOS, RELACIONAIS, BITWISE E DE ATRIBUIÇÃO
// ============================================================================

OPERADOR_REATRIBUICAO:
	'='
	| '+='
	| '-='
	| '*='
	| '/='
	| '//='
	| '%='
	| '&='
	| '|='
	| '^='
	| '<<='
	| '>>='
	| '**='
	| '??='
	| '||='
	| '&&=';

IGUAL: '=';
MAIS: '+';
MENOS: '-';
ASTERISCO: '*';
POTENCIA: '**';
DIVISAO: '/';
DIVISAO_INTEIRA: '//';
MODULO: '%';

MENOR_QUE: '<';
MENOR_OU_IGUAL: '<=';
IGUAL_IGUAL: '==';
DIFERENTE: '!=';
MAIOR_QUE: '>';
MAIOR_OU_IGUAL: '>=';

// Bitwise
OU_BIT_A_BIT: '|';
EX_OU_BIT_A_BIT: '^';
E_BIT_A_BIT: '&';
NAO_BIT_A_BIT: '~';
DESLOC_ESQUERDA: '<<';
DESLOC_DIREITA: '>>';

// Operadores de Encadeamento, Pipe e Nulidade
PIPE: '|>';
COALESCENCIA_NULA: '??';
PONTO_INTERROGACAO: '?.';
INTERROGACAO: '?';
ARROBA: '@';
TRES_PONTOS: '...';
SETA: '->';

// ============================================================================
// SÍMBOLOS DELIMITADORES DE SINTAXE
// ============================================================================

ABRE_PARENTESES: '(';
FECHA_PARENTESES: ')';
PONTO_E_VIRGULA: ';';
VIRGULA: ',';
ABRE_CHAVES: '{';
FECHA_CHAVES: '}';
DOIS_PONTOS: ':';
PONTO: '.';
ABRE_COLCHETES: '[';
FECHA_COLCHETES: ']';

// ============================================================================
// LITERAIS NUMÉRICOS, TEXTUAIS E IDENTIFICADORES
// ============================================================================

// LITERAL_INTEIRO representa inteiros decimais, hexadecimais, binários e octais.
LITERAL_INTEIRO:
	[0-9]+
	| '0' [xX] [0-9a-fA-F]+
	| '0' [bB] [01]+
	| '0' [oO] [0-7]+;

// LITERAL_DECIMAL representa números de ponto flutuante.
LITERAL_DECIMAL: [0-9]+ '.' [0-9]+;

// TEXTO representa literais textuais cercados por aspas duplas, simples ou crases (template string).
TEXTO:
	'"' ( ~[\\\r\n"] | ('\\' .) )* '"'
	| '\'' ( ~[\\\r\n'] | ('\\' .) )* '\''
	| '`' ( ~[`\\] | ('\\' .) )* '`';

// ID define identificadores de variáveis, classes e funções com suporte a Unicode (letras acentuadas PT-BR).
ID: (LETRAS_UNICODE | '_') (LETRAS_UNICODE | DIGITOS | '_')*;

fragment LETRAS_UNICODE:
	[a-zA-Z]
	| [\u00C0-\u00D6]
	| [\u00D8-\u00F6]
	| [\u00F8-\u024F];

fragment DIGITOS: [0-9];

// ============================================================================
// ESPAÇOS EM BRANCO E COMENTÁRIOS
// ============================================================================

NOVA_LINHA: '\r'? '\n';

COMENTARIO_LINHA: '//' ~[\r\n]* -> skip;

COMENTARIO_BLOCO: '/*' .*? '*/' -> skip;

WS: [ \t\r\n]+ -> skip;
