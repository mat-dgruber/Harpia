lexer grammar HarpiaLexer;

// ============================================================================
// TOKENS PRIMORDIAIS E VALORES LITERAIS
// ============================================================================

// Representações diretas de constantes primitivas da linguagem.
FALSO: 'Falso';
VERDADEIRO: 'Verdadeiro';
NULO: 'Nulo';

// ============================================================================
// PALAVRAS-CHAVE RESERVADAS E ESTRUTURAS DE CONTROLE
// ============================================================================

SE: 'se';               // Início de fluxo condicional.
SENAO: 'senao';         // Ramificação alternativa para fluxo condicional 'se'.
VAR: 'var';             // Declaração de variável mutável.
CONST: 'const';         // Declaração de constante imutável.
IMPORTE: 'importe';     // Declaração de carregamento de dependências e módulos.
DE: 'de';               // Usado em conjunto com 'importe' para importações qualificadas.
RETORNE: 'retorne';     // Declaração de devolução de valor em funções.
FUNC: 'func';           // Delimitador construtor de definição de funções.
OU: 'ou';               // Operador lógico de disjunção (OR).
E: 'e';                 // Operador lógico de conjunção (AND).
NAO: 'nao';             // Operador lógico de negação (NOT).
PARE: 'pare';           // Interrupção imediata de laços estruturados de repetição ('para').
CONTINUE: 'continue';   // Avanço imediato para a próxima iteração de laço.
PARA: 'para';           // Estrutura de laço de repetição iterativo (for-in).
EM: 'em';               // Conector semântico para laços de repetição (ex: 'para x em sequencia').
NOVA: 'nova';           // Usado para instanciar novas classes ou objetos.
ASSEGURA: 'assegura';   // Declaração de asserção lógica (teste/garantia, similar a 'assert').

// ============================================================================
// TOKENS DE SINTAXE E REGRAS AUXILIARES
// ============================================================================

// NOVA_LINHA é o delimitador físico de instruções na gramática.
NOVA_LINHA: '\r'? '\n';

// DIGITOS representa o conjunto numérico básico decimal.
DIGITOS:
	'0'
	| '1'
	| '2'
	| '3'
	| '4'
	| '5'
	| '6'
	| '7'
	| '8'
	| '9';

// LETRAS define os caracteres alfabéticos padrão ASCII suportados na gramática do compilador.
LETRAS: 'A' .. 'Z' | 'a' .. 'z';

// ============================================================================
// OPERADORES ARITMÉTICOS, LÓGICOS, BITWISE E ATRIBUIÇÃO
// ============================================================================

// OPERADOR_REATRIBUICAO engloba operadores de atribuição compostos (com atalho aritmético).
OPERADOR_REATRIBUICAO: IGUAL | '+=' | '-=' | '*=' | '@=' | '/=' | '%=' | '&=' | '|=' | '^=' | '<<=' | '>>=' | '**=' | '//=';

IGUAL: '=';                  // Atribuição de valor básica.
MAIS: '+';                   // Adição aritmética ou concatenação de strings.
MENOS: '-';                  // Subtração aritmética ou sinal negativo unário.
ASTERISCO: '*';              // Multiplicação aritmética.
POTENCIA: '**';              // Exponenciação real.
DIVISAO: '/';                // Divisão real de ponto flutuante.
DIVISAO_INTEIRA: '//';       // Divisão inteira (piso da divisão).
MODULO: '%';                 // Resto da divisão inteira.

MENOR_QUE: '<';              // Comparador relacional menor que.
MENOR_OU_IGUAL: '<=';        // Comparador relacional menor ou igual.
IGUAL_IGUAL: '==';           // Comparador lógico de igualdade de valor.
DIFERENTE: '!=';             // Comparador lógico de diferença de valor.
MAIOR_QUE: '>';              // Comparador relacional maior que.
MAIOR_OU_IGUAL: '>=';        // Comparador relacional maior ou igual.

// ============================================================================
// SÍMBOLOS DELIMITADORES DE SINTAXE
// ============================================================================

ABRE_PARENTESES: '(';        // Agrupamento de expressões ou passagem de argumentos de chamada.
FECHA_PARENTESES: ')';
PONTO_E_VIRGULA: ';';        // Finalizador de instruções explícito (opcional).
VIRGULA: ',';                // Separador de listas, parâmetros ou coleções.
ABRE_CHAVES: '{';            // Delimitador de início de escopo de bloco físico ou de declaração de mapas.
FECHA_CHAVES: '}';
DOIS_PONTOS: ':';            // Separador chave-valor em dicionários ou herança.
PONTO: '.';                  // Acesso a métodos e propriedades de instâncias.
ABRE_COLCHETES: '[';         // Delimitador de início de listas ou expressões de indexação.
FECHA_COLCHETES: ']';

// ============================================================================
// OPERADORES DE BIT (BITWISE)
// ============================================================================

OU_BIT_A_BIT: '|';           // Operação lógica OR em nível de bits.
EX_OU_BIT_A_BIT: '^';        // Operação lógica XOR (OU Exclusivo) em nível de bits.
E_BIT_A_BIT: '&';            // Operação lógica AND em nível de bits.
NAO_BIT_A_BIT: '~';          // Operação lógica NOT (Inversão) em nível de bits.
DESLOC_ESQUERDA: '<<';       // Deslocamento de bits para a esquerda (Bitwise Left Shift).
DESLOC_DIREITA: '>>';        // Deslocamento de bits para a direita (Bitwise Right Shift).

// ============================================================================
// CADEIAS DINÂMICAS E PADRÕES
// ============================================================================

// ID define um identificador válido (nomes de variáveis, funções, classes, etc.).
// Deve obrigatoriamente iniciar com uma letra ou sublinhado (_), seguido de qualquer combinação de letras ou dígitos.
ID: (LETRAS | '_') (LETRAS | DIGITOS)*;

// TEXTO representa literais textuais cercados por aspas duplas, aceitando sequências de escape variadas.
TEXTO: '"' ( ~[\\\r\n"] | ('\\' NOVA_LINHA | '\\' .) )* '"';

// WS define a regra padrão para ignorar e descartar espaços em branco, tabulações e quebras do analisador.
WS: [ \t\r\n]+ -> skip;
