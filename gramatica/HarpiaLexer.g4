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
// No compilador Go, é mapeado para um objeto de runtime singleton e imutável.
FALSO: 'Falso';

// VERDADEIRO representa o literal booleano positivo 'Verdadeiro'.
// No compilador Go, é mapeado para um objeto de runtime singleton e imutável.
VERDADEIRO: 'Verdadeiro';

// NULO representa a constante nula 'Nulo' (representação de ausência de tipo).
// No runtime do Harpia, é um singleton imutável livre de alocações no Heap.
NULO: 'Nulo';

// ============================================================================
// PALAVRAS-CHAVE RESERVADAS E ESTRUTURAS DE CONTROLE
// ============================================================================

// SE inicia uma estrutura condicional de decisão léxica.
SE: 'se';

// SENAO define desvios alternativos para blocos condicionais de controle de fluxo.
SENAO: 'senao';

// VAR inicializa a declaração de variáveis mutáveis com ou sem amarração estática de tipo.
VAR: 'var';

// CONST inicializa a declaração de constantes imutáveis cujo valor deve ser fixado na criação.
CONST: 'const';

// IMPORTE declara o carregamento e inicialização síncrona de dependências e pacotes.
IMPORTE: 'importe';

// DE qualifica imports de escopo local ou remoto, similar a destruturação (ex: de "x" importe y).
DE: 'de';

// RETORNE sinaliza a devolução opcional de valores e encerramento de escopos de função.
RETORNE: 'retorne';

// FUNC define o delimitador de assinatura para declarações estruturadas de função.
FUNC: 'func';

// OU define o operador de disjunção lógica 'ou' com menor nível de precedência.
OU: 'ou';

// E define o operador de conjunção lógica 'e' com média prioridade.
E: 'e';

// NAO define o operador de negação lógica unária de alta prioridade.
NAO: 'nao';

// PARE interrompe imediatamente laços estruturados de repetição ('para').
PARE: 'pare';

// CONTINUE avança para a próxima iteração lógica útil do laço corrente.
CONTINUE: 'continue';

// PARA define o cabeçalho do laço de repetição iterativo (for-in).
PARA: 'para';

// EM atua como o conector semântico indispensável para laços iterativos (ex: para x em y).
EM: 'em';

// NOVA é a instrução de instanciação de construtores de classe de objetos POO.
NOVA: 'nova';

// ASSEGURA declara asserções de validação para testes lógicos (similar ao 'assert').
ASSEGURA: 'assegura';

// ============================================================================
// TOKENS DE SINTAXE E REGRAS AUXILIARES
// ============================================================================

// NOVA_LINHA delimita instruções físicas no código caso ponto e vírgula não seja usado.
NOVA_LINHA: '\r'? '\n';

// DIGITOS representa o conjunto básico de algarismos numéricos decimais de 0 a 9.
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

// LETRAS define os caracteres alfabéticos padrão ASCII para a composição de identificadores.
LETRAS: 'A' .. 'Z' | 'a' .. 'z';

// ============================================================================
// OPERADORES ARITMÉTICOS, LÓGICOS, BITWISE E ATRIBUIÇÃO
// ============================================================================

// OPERADOR_REATRIBUICAO engloba todas as formas de operadores de atribuição direta e composta.
// Na VM Go, estas são simplificadas e otimizadas para evitar re-avaliações duplicadas de operandos.
OPERADOR_REATRIBUICAO: IGUAL | '+=' | '-=' | '*=' | '@=' | '/=' | '%=' | '&=' | '|=' | '^=' | '<<=' | '>>=' | '**=' | '//=';

// IGUAL define o operador básico de atribuição de variáveis.
IGUAL: '=';

// MAIS define o operador aritmético de adição ou concatenação de cadeias de caracteres.
MAIS: '+';

// MENOS define o operador aritmético de subtração ou sinal unário de inversão matemática.
MENOS: '-';

// ASTERISCO define o operador de multiplicação matemática.
ASTERISCO: '*';

// POTENCIA define o operador aritmético de exponenciação de alta precedência.
POTENCIA: '**';

// DIVISAO define o operador aritmético de divisão de ponto flutuante de dupla precisão.
DIVISAO: '/';

// DIVISAO_INTEIRA define o operador de divisão inteira (descarte de fração).
DIVISAO_INTEIRA: '//';

// MODULO obtém o resto da divisão de dois inteiros de 64 bits.
MODULO: '%';

// MENOR_QUE compara se um operando é estritamente menor que o outro.
MENOR_QUE: '<';

// MENOR_OU_IGUAL compara se um operando é menor ou equivalente ao outro.
MENOR_OU_IGUAL: '<=';

// IGUAL_IGUAL compara de forma lógica a equivalência de valores de operandos.
IGUAL_IGUAL: '==';

// DIFERENTE compara se os operandos possuem valores logicamente distintos.
DIFERENTE: '!=';

// MAIOR_QUE compara se um operando é estritamente maior que o outro.
MAIOR_QUE: '>';

// MAIOR_OU_IGUAL compara se um operando é maior ou equivalente ao outro.
MAIOR_OU_IGUAL: '>=';

// ============================================================================
// SÍMBOLOS DELIMITADORES DE SINTAXE
// ============================================================================

// ABRE_PARENTESES inicia agrupamento matemático ou passagens de assinaturas de função.
ABRE_PARENTESES: '(';

// FECHA_PARENTESES finaliza agrupamentos de prioridade e chamadas de função.
FECHA_PARENTESES: ')';

// PONTO_E_VIRGULA representa o delimitador explícito (opcional) de término de instruções.
PONTO_E_VIRGULA: ';';

// VIRGULA atua como separador em argumentos, arrays, mapas ou atribuições múltiplas.
VIRGULA: ',';

// ABRE_CHAVES inicia escopos de bloco estruturados ou definição de dicionários de mapas.
ABRE_CHAVES: '{';

// FECHA_CHAVES finaliza blocos de escopo de chaves e tabelas chave-valor.
FECHA_CHAVES: '}';

// DOIS_PONTOS separa chaves de valores em mapas ou indica tipos estáticos na declaração.
DOIS_PONTOS: ':';

// PONTO executa acessos a propriedades e métodos de namespaces ou objetos.
PONTO: '.';

// ABRE_COLCHETES inicia fatiamento de strings, definição de listas ou indexação de vetores.
ABRE_COLCHETES: '[';

// FECHA_COLCHETES finaliza coleções indexáveis e limites de fatiamento.
FECHA_COLCHETES: ']';

// ============================================================================
// OPERADORES DE BIT (BITWISE)
// ============================================================================

// OU_BIT_A_BIT realiza operação OR em nível de bits (Bitwise OR).
OU_BIT_A_BIT: '|';

// EX_OU_BIT_A_BIT realiza operação XOR em nível de bits (Bitwise XOR).
EX_OU_BIT_A_BIT: '^';

// E_BIT_A_BIT realiza operação AND em nível de bits (Bitwise AND).
E_BIT_A_BIT: '&';

// NAO_BIT_A_BIT realiza operação NOT unária em nível de bits (Bitwise NOT).
NAO_BIT_A_BIT: '~';

// DESLOC_ESQUERDA desloca bits do operando esquerdo para esquerda (Left Shift).
DESLOC_ESQUERDA: '<<';

// DESLOC_DIREITA desloca bits do operando esquerdo para direita (Right Shift).
DESLOC_DIREITA: '>>';

// ============================================================================
// CADEIAS DINÂMICAS E PADRÕES
// ============================================================================

// ID define regras estruturais de composição de variáveis e classes do usuário.
// Devem iniciar por letra ou sublinhado (_), seguido de letras ou dígitos decimais.
ID: (LETRAS | '_') (LETRAS | DIGITOS)*;

// TEXTO representa literais textuais cercados por aspas duplas com suporte a escapes de barra.
TEXTO: '"' ( ~[\\\r\n"] | ('\\' NOVA_LINHA | '\\' .) )* '"';

// WS define a regra de descarte léxico de espaços, tabs e quebras de linha padrão.
WS: [ \t\r\n]+ -> skip;
