parser grammar HarpiaParser;

options {
	// Importa e vincula o dicionário de tokens declarados em HarpiaLexer.g4.
	tokenVocab = HarpiaLexer;
}

// ============================================================================
// ESPECIFICAÇÃO SINTÁTICA DO HARPIA (ANTLR4 REFERENCE)
// ============================================================================
//
// Este arquivo descreve as regras gramaticais e a estrutura sintática formal da
// linguagem Harpia, detalhando a precedência de operadores e as regras de AST.
//
// ponytail: Assim como o Lexer, o analisador sintático físico do Harpia foi
// escrito inteiramente à mão em Go (/parser/parser.go) para garantir máxima
// velocidade e relatórios de diagnóstico pedagógicos em português com indicação
// precisa de linha, coluna e token sob falha sintática.

// ============================================================================
// PONTO DE ENTRADA GRAMATICAL (RAIZ DO COMPILADOR)
// ============================================================================

// programa representa a árvore raiz de qualquer arquivo ou script Harpia válido.
// Finaliza obrigatoriamente com o sinal de Fim de Arquivo (EOF).
programa: declaracoes? EOF;

// declaracoes agrupa sequencialmente uma ou mais instruções do desenvolvedor.
declaracoes: declaracao+;

// declaracao bifurca as ramificações em estruturas compostas (com bloco) ou simples.
declaracao: declaracao_composta | declaracao_simples;

// ============================================================================
// DECLARAÇÕES SIMPLES (UNILINHA OU INSTRUÇÃO DE CONTROLE)
// ============================================================================

declaracao_simples:
	atribuicao                                                   // Declaração de variável (var) ou constante (const).
	| expressao OPERADOR_REATRIBUICAO expressao                  // Reatribuições aritméticas acumuladas (ex: x += 5).
	| expressao                                                  // Avaliação de expressões isoladas (ex: chamadas de método).
	| declaracao_retorne                                         // Instrução de saída e devolução de valor em funções.
	| declaracao_importacao                                      // Diretiva de importação de bibliotecas ou módulos nativos.
	| ASSEGURA expressao (VIRGULA expressao)?                    // Asserção lógica com mensagem opcional de traceback.
	| PARE                                                       // Parada de loops (break).
	| CONTINUE;                                                  // Avanço para a próxima iteração do loop (continue).

// ============================================================================
// DECLARAÇÕES COMPOSTAS (BLOCOS E FLUXOS COMPLEXOS)
// ============================================================================

declaracao_composta:
	declaracao_funcao                                            // Definição estruturada de funções (func).
	| declaracao_se                                              // Decisões condicionais (se/senao).
	| declaracao_para;                                           // Laços de repetição iterativos (para-em).

// atribuicao unifica a inicialização de variáveis e constantes imutáveis do usuário.
atribuicao: atribuicao_constante | atribuicao_variavel;

// atribuicao_variavel aceita: 'var x;', 'var x = 10;' ou 'var x: Tipo;'
atribuicao_variavel:
	VAR ID (IGUAL expressao | DOIS_PONTOS ID)? PONTO_E_VIRGULA;

// atribuicao_constante exige atribuição e inicialização imediata: 'const PI = 3.14;'
atribuicao_constante: CONST ID IGUAL expressao PONTO_E_VIRGULA;

// ============================================================================
// IMPORTAÇÃO DE MÓDULOS E DEPENDÊNCIAS
// ============================================================================

declaracao_importacao:
	declaracao_importacao_simples
	| declaracao_importacao_de;

// Importação simples direta: 'importe "matematica";' ou 'importe "sistema", "web";'
declaracao_importacao_simples:
	IMPORTE TEXTO (VIRGULA TEXTO)* PONTO_E_VIRGULA;

// Importações qualificadas e parciais: 'de "matematica" importe PI, raiz;' ou 'de "web" importe *;'
declaracao_importacao_de:
	DE TEXTO IMPORTE (ASTERISCO | ID (VIRGULA ID)*) PONTO_E_VIRGULA;

// ============================================================================
// CONTROLE DE FUNÇÕES E RETORNO
// ============================================================================

declaracao_retorne: RETORNE expressao? PONTO_E_VIRGULA;

// Definição de funções nomeadas: 'func somar(a, b: Inteiro) { retorne a + b; }'
declaracao_funcao:
	FUNC ID ABRE_PARENTESES parametros? FECHA_PARENTESES bloco;

parametros: parametro (VIRGULA parametro)*;

// parametro aceita anotações estáticas opcionais de tipo (ex: 'nome: Texto').
parametro: ID (':' ID)?;

// ============================================================================
// ESTRUTURAS DE DECISÃO CONDICIONAL (SE / SENAO SE / SENAO)
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
// LAÇOS DE REPETIÇÃO COOPERATIVOS (PARA-EM)
// ============================================================================

// para iterador em colecao: 'para item em sequencia(10)'
declaracao_para: PARA ID EM primario;

// bloco delimita escopos de código e instruções aninhadas encapsulados em chaves '{}'.
bloco: ABRE_CHAVES declaracoes? FECHA_CHAVES;

// ============================================================================
// EXPRESSÕES E PRECEDÊNCIA DE OPERADORES (ÁRVORE SINTÁTICA / AST)
// ============================================================================

// expressao é a raiz para construção de operadores e expressões em Harpia.
expressao:
	NOVA primario ABRE_PARENTESES argumentos? FECHA_PARENTESES   // Instanciação de objetos (nova MinhaClasse()).
	| disjuncao;                                                 // Avaliação de expressões lógicas e relacionais.

// Precedência [Nível 11]: Operador Lógico 'ou' (Disjunção).
disjuncao: conjuncao (OU conjuncao)*;

// Precedência [Nível 10]: Operador Lógico 'e' (Conjunção).
conjuncao: inversao (E inversao)*;

// Precedência [Nível 9]: Operador Unário de Negação Lógica 'nao'.
inversao: NAO inversao | comparacao;

// Precedência [Nível 8]: Comparadores lógicos relacionais e associação de inclusão em coleções ('em').
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

// Precedência [Nível 4]: Operadores de Deslocamento de Bits ('<<', '>>').
deslocamento:
	arit_basica ((DESLOC_ESQUERDA | DESLOC_DIREITA) arit_basica)?;

// Precedência [Nível 3]: Soma e Subtração básica ou concatenação de cadeias de caracteres ('+', '-').
arit_basica: termo ((MAIS | MENOS) termo)*;

// Precedência [Nível 2]: Operadores de Multiplicação, Divisão, Divisão Inteira e Resto ('*', '/', '//', '%').
termo:
	fator (
		(ASTERISCO | DIVISAO | DIVISAO_INTEIRA | MODULO) fator
	)*;

// Precedência [Nível 1]: Sinais e Operadores Unários matemáticos e Bitwise ('+', '-', '~').
fator: (MAIS | MENOS | NAO_BIT_A_BIT)* potencia;

// Precedência [Nível 0]: Operador de Exponenciação de Alta Prioridade ('**').
potencia: primario (POTENCIA fator)?;

// ============================================================================
// ACESSO, INDEXAÇÃO, CHAMADAS E TERMINAIS
// ============================================================================

primario:
	primario PONTO primario                                      // Acessos a métodos e atributos (ex: console.log).
	| primario ABRE_PARENTESES argumentos? FECHA_PARENTESES      // Chamadas estruturadas de métodos e funções (ex: calcular()).
	| primario ABRE_COLCHETES expressao FECHA_COLCHETES          // Fatiamento e indexação de coleções ou textos (ex: lista[0]).
	| atomo;

argumentos: argumento (VIRGULA argumento)*;

argumento: expressao;

// atomo agrupa os nós folha literais e de coleções primitivas da árvore sintática.
atomo:
	ID                                                           // Identificadores de variáveis, namespaces ou classes.
	| 'Verdadeiro'                                               // Literal booleano positivo.
	| 'Falso'                                                    // Literal booleano negativo.
	| 'Nulo'                                                     // Constante de nulidade e ausência de valor.
	| TEXTO+                                                     // Strings literais adjacentes (com suporte a fusão implícita).
	| DIGITOS                                                    // Valores numéricos decimais literais.
	| tupla                                                      // Coleção imutável de itens.
	| grupo                                                      // Parentetização matemática para agrupamento de prioridade.
	| lista                                                      // Listas mutáveis ordenadas de elementos.
	| mapa;                                                      // Estruturas de mapas associativos chave-valor (dicionários).

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
