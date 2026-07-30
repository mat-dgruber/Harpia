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

programa: declaracoes? EOF;

declaracoes: (declaracao fim_instrucao?)+;

fim_instrucao: PONTO_E_VIRGULA | NOVA_LINHA;

declaracao: declaracao_composta | declaracao_simples;

// ============================================================================
// DECLARAÇÕES SIMPLES (UNILINHA OU INSTRUÇÃO DE CONTROLE)
// ============================================================================

declaracao_simples:
	atribuicao
	| expressao OPERADOR_REATRIBUICAO expressao
	| expressao
	| declaracao_retorne
	| declaracao_importacao
	| declaracao_exportar
	| ASSEGURA expressao (VIRGULA expressao)?
	| PARE
	| CONTINUE;

// ============================================================================
// DECLARAÇÕES COMPOSTAS (BLOCOS E FLUXOS COMPLEXOS)
// ============================================================================

declaracao_composta:
	declaracao_funcao
	| declaracao_se
	| declaracao_enquanto
	| declaracao_para
	| declaracao_tente
	| declaracao_escolha
	| declaracao_classe
	| declaracao_enum
	| declaracao_interface
	| declaracao_estilo
	| declaracao_teste;

// ============================================================================
// ATRIBUIÇÃO, VARIÁVEIS E DESESTRUTURAÇÃO
// ============================================================================

atribuicao:
	atribuicao_constante
	| atribuicao_variavel
	| atribuicao_desestruturacao;

atribuicao_variavel: VAR ID (DOIS_PONTOS ID)? (IGUAL expressao)?;

atribuicao_constante: CONST ID (DOIS_PONTOS ID)? IGUAL expressao;

atribuicao_desestruturacao:
	(VAR | CONST) (padrao_lista | padrao_mapa) IGUAL expressao;

padrao_lista:
	ABRE_COLCHETES ID (VIRGULA ID)* FECHA_COLCHETES;

padrao_mapa: ABRE_CHAVES ID (VIRGULA ID)* FECHA_CHAVES;

// ============================================================================
// IMPORTAÇÃO E EXPORTAÇÃO DE MÓDULOS
// ============================================================================

declaracao_importacao:
	declaracao_importacao_simples
	| declaracao_importacao_de;

declaracao_importacao_simples: IMPORTE TEXTO (VIRGULA TEXTO)*;

declaracao_importacao_de:
	DE TEXTO IMPORTE (ASTERISCO | ID (VIRGULA ID)*);

declaracao_exportar:
	EXPORTAR (
		atribuicao
		| declaracao_funcao
		| declaracao_classe
		| declaracao_enum
		| declaracao_interface
	);

// ============================================================================
// CONTROLE DE FUNÇÕES E RETORNO
// ============================================================================

declaracao_retorne: RETORNE expressao?;

declaracao_funcao:
	ASSINCRONO? FUNC ID ABRE_PARENTESES parametros? FECHA_PARENTESES (
		DOIS_PONTOS ID
	)? bloco;

parametros: parametro (VIRGULA parametro)*;

parametro: TRES_PONTOS? ID (DOIS_PONTOS ID)? (IGUAL expressao)?;

// ============================================================================
// ESTRUTURAS DE DECISÃO CONDICIONAL (SE / SENAO / ENQUANTO / PARA)
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

declaracao_enquanto:
	ENQUANTO ABRE_PARENTESES expressao FECHA_PARENTESES bloco;

declaracao_para: PARA ID EM expressao bloco;

// ============================================================================
// TRATAMENTO DE EXCEÇÕES (TENTE / CAPTURE / FINALMENTE)
// ============================================================================

declaracao_tente:
	TENTE bloco (
		CAPTURE ABRE_PARENTESES ID FECHA_PARENTESES bloco
	)? (FINALMENTE bloco)?;

// ============================================================================
// PATTERN MATCHING / SWITCH (ESCOLHA / CASO / QUANDO)
// ============================================================================

declaracao_escolha:
	ESCOLHA ABRE_PARENTESES expressao FECHA_PARENTESES ABRE_CHAVES caso_item* FECHA_CHAVES;

caso_item:
	CASO expressao (QUANDO expressao)? DOIS_PONTOS (
		bloco
		| declaracoes
	)
	| SENAO DOIS_PONTOS (bloco | declaracoes);

// ============================================================================
// PROGRAMAÇÃO ORIENTADA A OBJETOS (CLASSE / ENUM / INTERFACE)
// ============================================================================

declaracao_classe:
	CLASSE ID (ESTENDE ID)? ABRE_CHAVES membro_classe* FECHA_CHAVES;

membro_classe: ESTATICO? (declaracao_funcao | atribuicao);

declaracao_enum:
	ENUM ID ABRE_CHAVES (
		ID (IGUAL expressao)? (VIRGULA | fim_instrucao)?
	)* FECHA_CHAVES;

declaracao_interface:
	INTERFACE ID ABRE_CHAVES membro_interface* FECHA_CHAVES;

membro_interface:
	ID ABRE_PARENTESES parametros? FECHA_PARENTESES (
		DOIS_PONTOS ID
	)? fim_instrucao?;

// ============================================================================
// ESTILOS E TESTES NATIVOS
// ============================================================================

declaracao_estilo: ESTILO ID ABRE_CHAVES regra_estilo* FECHA_CHAVES;

regra_estilo:
	ID DOIS_PONTOS expressao fim_instrucao?
	| ID ABRE_CHAVES regra_estilo* FECHA_CHAVES;

declaracao_teste: TESTAR TEXTO bloco;

bloco: ABRE_CHAVES declaracoes? FECHA_CHAVES;

// ============================================================================
// EXPRESSÕES E PRECEDÊNCIA DE OPERADORES (ÁRVORE SINTÁTICA / AST)
// ============================================================================

expressao:
	NOVA primario ABRE_PARENTESES argumentos? FECHA_PARENTESES
	| ternario;

// Precedência [Nível 13]: Operador Ternário Condicional ('? :')
ternario: pipe (INTERROGACAO expressao DOIS_PONTOS expressao)?;

// Precedência [Nível 12]: Operador Pipe ('|>')
pipe: coalescencia (PIPE coalescencia)*;

// Precedência [Nível 11]: Coalescência Nula ('??')
coalescencia: disjuncao (COALESCENCIA_NULA disjuncao)*;

// Precedência [Nível 10]: Disjunção Lógica ('ou')
disjuncao: conjuncao (OU conjuncao)*;

// Precedência [Nível 9]: Conjunção Lógica ('e')
conjuncao: inversao (E inversao)*;

// Precedência [Nível 8]: Negação Lógica ('nao')
inversao: NAO inversao | comparacao;

// Precedência [Nível 7]: Comparações Relacionais e Inclusão ('==', '!=', '<', '<=', '>', '>=', 'em')
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

// Precedência [Nível 6]: Bitwise OR ('|')
ou_bitabit: exou_bitabit (OU_BIT_A_BIT exou_bitabit)*;

// Precedência [Nível 5]: Bitwise XOR ('^')
exou_bitabit: e_bitabit (EX_OU_BIT_A_BIT e_bitabit)*;

// Precedência [Nível 4]: Bitwise AND ('&')
e_bitabit: deslocamento (E_BIT_A_BIT deslocamento)*;

// Precedência [Nível 3]: Deslocamento de Bits ('<<', '>>')
deslocamento:
	arit_basica ((DESLOC_ESQUERDA | DESLOC_DIREITA) arit_basica)?;

// Precedência [Nível 2]: Soma e Subtração ('+', '-')
arit_basica: termo ((MAIS | MENOS) termo)*;

// Precedência [Nível 1]: Multiplicação, Divisão, Piso e Resto ('*', '/', '//', '%')
termo:
	fator (
		(ASTERISCO | DIVISAO | DIVISAO_INTEIRA | MODULO) fator
	)*;

// Precedência [Nível 0 Unário]: Sinais e Operadores Unários ('+', '-', '~')
fator: (MAIS | MENOS | NAO_BIT_A_BIT)* aguarde_expr;

aguarde_expr: AGUARDE potencia | potencia;

// Precedência [Nível Exponenciação]: Exponenciação ('**')
potencia: primario (POTENCIA fator)?;

// ============================================================================
// ACESSO, INDEXAÇÃO, CHAMADAS E TERMINAIS
// ============================================================================

primario:
	primario PONTO ID
	| primario PONTO_INTERROGACAO ID
	| primario ABRE_PARENTESES argumentos? FECHA_PARENTESES
	| primario ABRE_COLCHETES expressao FECHA_COLCHETES
	| atomo;

argumentos: argumento (VIRGULA argumento)*;

argumento: TRES_PONTOS? expressao;

atomo:
	ID
	| VERDADEIRO
	| FALSO
	| NULO
	| SELF
	| TEXTO
	| LITERAL_INTEIRO
	| LITERAL_DECIMAL
	| funcao_anonima
	| tupla
	| grupo
	| lista
	| mapa;

funcao_anonima:
	ASSINCRONO? FUNC ABRE_PARENTESES parametros? FECHA_PARENTESES (
		DOIS_PONTOS ID
	)? (bloco | (SETA expressao));

// ============================================================================
// COLEÇÕES E AGRUPAMENTOS
// ============================================================================

tupla:
	ABRE_PARENTESES expressao VIRGULA (
		expressao VIRGULA?
	)* FECHA_PARENTESES;

grupo: ABRE_PARENTESES expressao FECHA_PARENTESES;

lista:
	ABRE_COLCHETES (expressao (VIRGULA expressao)*)? FECHA_COLCHETES;

mapa:
	ABRE_CHAVES (mapa_item (VIRGULA mapa_item)*)? FECHA_CHAVES;

mapa_item:
	ID
	| (ID | TEXTO | (ABRE_COLCHETES expressao FECHA_COLCHETES)) DOIS_PONTOS expressao;
