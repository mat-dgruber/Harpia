parser grammar PortuscriptParser;

options {
	// Importa e vincula o dicionário de tokens declarados em PortuscriptLexer.g4.
	tokenVocab = PortuscriptLexer;
}

// ============================================================================
// PONTO DE ENTRADA GRAMATICAL (PÁGINA DO PROGRAMA)
// ============================================================================

// programa define a regra de análise raiz para qualquer script físico em Portuscript.
// Pode conter declarações opcionais seguidas obrigatoriamente pelo token de Fim de Arquivo (EOF).
programa: declaracoes? EOF;

// declaracoes é um agrupamento sequencial de pelo menos uma instrução declarativa.
declaracoes: declaracao+;

// declaracao bifurca as estruturas lógicas em duas classificações sintáticas: compostas e simples.
declaracao: declaracao_composta | declaracao_simples;

// ============================================================================
// DECLARAÇÕES SIMPLES (UNILINHA OU INSTRUÇÃO DE CONTROLE)
// ============================================================================

declaracao_simples:
	atribuicao                                                   // Declaração de variável (var) ou constante (const).
	| expressao OPERADOR_REATRIBUICAO expressao                  // Operação de reatribuição com acumulador (ex: x += 1).
	| expressao                                                  // Avaliação direta de expressão (ex: chamadas de função soltas).
	| declaracao_retorne                                         // Instrução de retorno de valor ('retorne').
	| declaracao_importacao                                      // Diretiva de importação de módulos externos/nativos.
	| ASSEGURA expressao (VIRGULA expressao)?                    // Asserção lógica com mensagem opcional (assegura x == 10, "Erro!").
	| PARE                                                       // Instrução de interrupção de repetição ('pare').
	| CONTINUE;                                                  // Instrução de avanço de repetição ('continue').

// ============================================================================
// DECLARAÇÕES COMPOSTAS (BLOCOS E FLUXOS COMPLEXOS)
// ============================================================================

declaracao_composta:
	declaracao_funcao                                            // Definição estruturada de função (func).
	| declaracao_se                                              // Estrutura de decisão condicional (se/senao).
	| declaracao_para;                                           // Estrutura de repetição iterativa (para-em).

// atribuicao unifica a declaração de variáveis e constantes imutáveis.
atribuicao: atribuicao_constante | atribuicao_variavel;

// atribuicao_variavel aceita: 'var x;', 'var x = 10;' ou 'var x: Tipo;'
atribuicao_variavel:
	VAR ID (IGUAL expressao | DOIS_PONTOS ID)? PONTO_E_VIRGULA;

// atribuicao_constante exige inicialização: 'const x = 10;'
atribuicao_constante: CONST ID IGUAL expressao PONTO_E_VIRGULA;

// ============================================================================
// IMPORTAÇÃO DE MÓDULOS
// ============================================================================

declaracao_importacao:
	declaracao_importacao_simples
	| declaracao_importacao_de;

// Importação direta simples: 'importe "matematica";' ou 'importe "sistema", "colorize";'
declaracao_importacao_simples:
	IMPORTE TEXTO (VIRGULA TEXTO)* PONTO_E_VIRGULA;

// Importação parcial e qualificada: 'de "matematica" importe PI, raiz;' ou 'de "matematica" importe *;'
declaracao_importacao_de:
	DE TEXTO IMPORTE (ASTERISCO | ID (VIRGULA ID)*) PONTO_E_VIRGULA;

// ============================================================================
// CONTROLE DE FUNÇÕES
// ============================================================================

declaracao_retorne: RETORNE expressao? PONTO_E_VIRGULA;

// Definição de funções: 'func somar(a, b: Inteiro) { retorne a + b; }'
declaracao_funcao:
	FUNC ID ABRE_PARENTESES parametros? FECHA_PARENTESES bloco;

parametros: parametro (VIRGULA parametro)*;

parametro: ID (':' ID)?;                                     // Parâmetros opcionais com indicação estática de tipo.

// ============================================================================
// ESTRUTURAS DE DECISÃO CONDICIONAL (SE / SENAO IF / SENAO)
// ============================================================================

declaracao_se:
	SE ABRE_PARENTESES expressao FECHA_PARENTESES bloco (
		declaracao_senao_se
		| SENAO bloco
	)?;

declaracao_senao_se:
	SENAO SE ABRE_PARENTESES expressao FECHA_PARENTESES bloco (
		declaracao_senao_se
		| SENAO bloco
	)?;

declaracao_senao: SENAO bloco;

// ============================================================================
// LAÇO ITERATIVO PARA (FOR-IN)
// ============================================================================

// para id em colecao: 'para item em lista'
declaracao_para: PARA ID EM primario;

// bloco define o corpo encapsulado de instruções delimitado por chaves '{}'.
bloco: ABRE_CHAVES declaracoes? FECHA_CHAVES;

// ============================================================================
// EXPRESSÕES E PRECEDÊNCIA DE OPERADORES (ÁRVORE SINTÁTICA / AST)
// ============================================================================

// expressao é a raiz de resolução para construção e avaliação de expressões aritméticas e lógicas.
expressao:
	NOVA primario ABRE_PARENTESES argumentos? FECHA_PARENTESES   // Instanciação de novas instâncias de classes (nova Classe(arg1)).
	| disjuncao;                                                 // Avaliação lógica.

// Precedência [Nível 11]: Operador Lógico 'ou' (Disjunção).
disjuncao: conjuncao (OU conjuncao)*;

// Precedência [Nível 10]: Operador Lógico 'e' (Conjunção).
conjuncao: inversao (E inversao)*;

// Precedência [Nível 9]: Operador Unário de Negação Lógica 'nao'.
inversao: NAO inversao | comparacao;

// Precedência [Nível 8]: Operadores de Comparação Relacional e Associação ('em').
comparacao:
	ou_bitabit (
		(
			IGUAL_IGUAL
			| DIFERENTE
			| MENOR_OU_IGUAL
			| MENOR_QUE
			| MAIOR_OU_IGUAL
			| MAIOR_QUE
			| EM
		) ou_bitabit
	)?;

// Precedência [Nível 7]: Operador Ou Bitwise (Bitwise OR '|').
ou_bitabit: exou_bitabit (OU_BIT_A_BIT exou_bitabit)*;

// Precedência [Nível 6]: Operador XOR Bitwise (Bitwise Exclusive OR '^').
exou_bitabit: e_bitabit (EX_OU_BIT_A_BIT e_bitabit)*;

// Precedência [Nível 5]: Operador E Bitwise (Bitwise AND '&').
e_bitabit: deslocamento (E_BIT_A_BIT deslocamento)*;

// Precedência [Nível 4]: Operadores de Deslocamento de Bits para esquerda/direita ('<<', '>>').
deslocamento:
	arit_basica ((DESLOC_ESQUERDA | DESLOC_DIREITA) arit_basica)?;

// Precedência [Nível 3]: Soma e Subtração básica ('+', '-').
arit_basica: termo ((MAIS | MENOS) termo)*;

// Precedência [Nível 2]: Multiplicação, Divisão real, Divisão inteira e Resto de Divisão ('*', '/', '//', '%').
termo:
	fator (
		(ASTERISCO | DIVISAO | DIVISAO_INTEIRA | MODULO) fator
	)*;

// Precedência [Nível 1]: Sinais e Operadores Unários aritméticos e Bitwise ('+', '-', '~').
fator: (MAIS | MENOS | NAO_BIT_A_BIT)* potencia;

// Precedência [Nível 0]: Operador de Exponenciação de Alta Prioridade ('**').
potencia: primario (POTENCIA fator)?;

// ============================================================================
// ACESSO, INDEXAÇÃO, CHAMADAS E TERMINAIS
// ============================================================================

primario:
	primario PONTO primario                                      // Acesso a atributos/propriedades (ex: console.log).
	| primario ABRE_PARENTESES argumentos? FECHA_PARENTESES      // Chamadas de função ou métodos (ex: somar(a)).
	| primario ABRE_COLCHETES expressao FECHA_COLCHETES          // Fatiamento ou indexação de vetores (ex: lista[0]).
	| atomo;

argumentos: argumento (VIRGULA argumento)*;

argumento: expressao;

// atomo unifica as folhas estruturais terminais (literais e coleções explícitas).
atomo:
	ID                                                           // Identificadores de variáveis ou classes.
	| 'Verdadeiro'                                               // Literal Booleano Positivo.
	| 'Falso'                                                    // Literal Booleano Negativo.
	| 'Nulo'                                                     // Literal Nulo de representação.
	| TEXTO+                                                     // Cadeia física de caracteres literais.
	| DIGITOS                                                    // Valores numéricos decimais.
	| tupla                                                      // Coleção imutável de itens.
	| grupo                                                      // Parentetização de escopo matemático (ex: (2 + 3) * 5).
	| lista                                                      // Coleção mutável ordenada [a, b].
	| mapa;                                                      // Coleção par chave-valor { a: 1 }.

// ============================================================================
// COLEÇÕES E AGRUPAMENTOS
// ============================================================================

tupla:
	ABRE_PARENTESES expressao VIRGULA (expressao VIRGULA?)* FECHA_PARENTESES;

grupo: ABRE_PARENTESES expressao FECHA_PARENTESES;

lista:
	ABRE_COLCHETES (expressao (VIRGULA expressao)*)? FECHA_COLCHETES;

mapa:
	ABRE_CHAVES (mapa_item (VIRGULA mapa_item)*)? FECHA_CHAVES;

mapa_item:
	ID
	| (ID | TEXTO | (ABRE_COLCHETES expressao FECHA_COLCHETES)) DOIS_PONTOS expressao;