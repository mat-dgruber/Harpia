package ptst

// ============================================================================
// CONVENÇÃO DE NOMENCLATURA:
// De acordo com as diretrizes de design do Harpia, todas as interfaces Go
// destinadas a fins de modelagem de protocolos dinâmicos devem obrigatoriamente
// iniciar com o prefixo "I" e os seus respectivos métodos internos com o prefixo "M" (de "Método").
// ============================================================================

// I__chame__ define o protocolo de chamada de objetos (semelhante ao __call__ do Python).
// Qualquer objeto que possa ser "chamado" como função deve implementar esta interface.
type I__chame__ interface {
	M__chame__(args Tupla) (Objeto, error)
}

// I__texto__ define o protocolo de conversão explícita para o tipo Texto (__texto__).
type I__texto__ interface {
	M__texto__() (Objeto, error)
}

// I__bytes__ define o protocolo de coerção e conversão para o tipo de dados Bytes (__bytes__).
type I__bytes__ interface {
	M__bytes__() (Objeto, error)
}

// I__inteiro__ define o protocolo de coerção e conversão para o tipo de dados Inteiro (__inteiro__).
type I__inteiro__ interface {
	M__inteiro__() (Objeto, error)
}

// I__decimal__ define o protocolo de coerção e conversão para o tipo de dados Decimal (__decimal__).
type I__decimal__ interface {
	M__decimal__() (Objeto, error)
}

// I__booleano__ define o protocolo de coerção e representação lógica booleana (__booleano__).
type I__booleano__ interface {
	M__booleano__() (Objeto, error)
}

// I_conversaoEntreTipos agrupa de forma simplificada todas as interfaces de coerção primitivas de tipos.
type I_conversaoEntreTipos interface {
	I__texto__
	I__inteiro__
	I__decimal__
	I__booleano__
}

// I__obtem_attributo__ define o protocolo de acesso dinâmico de atributos em instâncias (__obtem_atributo__).
type I__obtem_attributo__ interface {
	M__obtem_attributo__(nome string) (Objeto, error)
}

// I__define_atributo__ define o protocolo de escrita/atribuição dinâmica de propriedades (__define_atributo__).
type I__define_atributo__ interface {
	M__define_atributo__(nome string, valor Objeto) error
}

// I__obtem__ define o protocolo para descritores (Descriptors, semelhante ao __get__ do Python).
type I__obtem__ interface {
	M__obtem__(inst Objeto, dono *Tipo) (Objeto, error)
}

// I_ObtemMapa define interfaces internas para recuperação das tabelas hashes de atributos dos objetos.
type I_ObtemMapa interface {
	ObtemMapa() Mapa
}

// I_ObtemDoc define interfaces internas de recuperação rápida de blocos de documentação.
type I_ObtemDoc interface {
	ObtemDoc() string
}

// I_Chamar define rotinas de chamabilidade abstrata de alto nível da VM.
type I_Chamar interface {
	Chamar(inst Objeto, args Tupla) (Objeto, error)
}

// ============================================================================
// PROTOCOLOS DE ARITMÉTICA MATEMÁTICA
// ============================================================================

// I__adiciona__ define a operação de soma aritmética ou concatenação textual (__adiciona__).
type I__adiciona__ interface {
	M__adiciona__(outro Objeto) (Objeto, error)
}

// I__adiciona_e_atribui__ define a operação de soma auto-acumulativa (__adiciona_e_atribui__).
type I__adiciona_e_atribui__ interface {
	M__adiciona_e_atribui__(outro Objeto) (Objeto, error)
}

// I__multiplica__ define a operação aritmética de multiplicação (__multiplica__).
type I__multiplica__ interface {
	M__multiplica__(outro Objeto) (Objeto, error)
}

// I__multiplica_e_atribui__ define a operação de multiplicação auto-acumulativa (__multiplica_e_atribui__).
type I__multiplica_e_atribui__ interface {
	M__multiplica_e_atribui__(outro Objeto) (Objeto, error)
}

// I__subtrai__ define a operação aritmética de subtração (__subtrai__).
type I__subtrai__ interface {
	M__subtrai__(outro Objeto) (Objeto, error)
}

// I__subtrai_e_atribui__ define a operação de subtração auto-acumulativa (__subtrai_e_atribui__).
type I__subtrai_e_atribui__ interface {
	M__subtrai_e_atribui__(outro Objeto) (Objeto, error)
}

// I__divide__ define a operação aritmética de divisão real (__divide__).
type I__divide__ interface {
	M__divide__(outro Objeto) (Objeto, error)
}

// I__divide_e_atribui__ define a operação de divisão real auto-acumulativa (__divide_e_atribui__).
type I__divide_e_atribui__ interface {
	M__divide_e_atribui__(outro Objeto) (Objeto, error)
}

// I__divide_inteiro__ define a operação aritmética de divisão por piso (__divide_inteiro__).
type I__divide_inteiro__ interface {
	M__divide_inteiro__(outro Objeto) (Objeto, error)
}

// I__divide_inteiro_e_atribui__ define a operação de divisão por piso auto-acumulativa.
type I__divide_inteiro_e_atribui__ interface {
	M__divide_inteiro_e_atribui__(outro Objeto) (Objeto, error)
}

// I__mod__ define a operação de resto de divisão inteira (__mod__).
type I__mod__ interface {
	M__mod__(outro Objeto) (Objeto, error)
}

// I__neg__ define a operação unária de inversão de sinal aritmético (__neg__).
type I__neg__ interface {
	M__neg__() (Objeto, error)
}

// I__pos__ define o operador unário positivo de identidade de sinal (__pos__).
type I__pos__ interface {
	M__pos__() (Objeto, error)
}

// I_aritmeticaMatematica engloba todas as rotinas básicas de cálculos matemáticos.
type I_aritmeticaMatematica interface {
	I__adiciona__
	I__multiplica__
	I__subtrai__
	I__divide__
	I__divide_inteiro__
	I__mod__
	I__neg__
	I__pos__
}

// ============================================================================
// PROTOCOLOS DE COMPARAÇÕES E COMPARAÇÕES RICAS
// ============================================================================

// I__menor_que__ define o comparador relacional menor que (<) (__menor_que__).
type I__menor_que__ interface {
	M__menor_que__(outro Objeto) (Objeto, error)
}

// I__menor_ou_igual__ define o comparador relacional menor ou igual (<=) (__menor_ou_igual__).
type I__menor_ou_igual__ interface {
	M__menor_ou_igual__(outro Objeto) (Objeto, error)
}

// I__igual__ define o comparador de igualdade lógica de valor (==) (__igual__).
type I__igual__ interface {
	M__igual__(outro Objeto) (Objeto, error)
}

// I__diferente__ define o comparador de diferença lógica de valor (!=) (__diferente__).
type I__diferente__ interface {
	M__diferente__(outro Objeto) (Objeto, error)
}

// I__maior_que__ define o comparador relacional maior que (>) (__maior_que__).
type I__maior_que__ interface {
	M__maior_que__(outro Objeto) (Objeto, error)
}

// I__maior_ou_igual__ define o comparador relacional maior ou igual (>=) (__maior_ou_igual__).
type I__maior_ou_igual__ interface {
	M__maior_ou_igual__(outro Objeto) (Objeto, error)
}

// I_comparacaoRica engloba todas as interfaces de operações lógicas relacionais.
type I_comparacaoRica interface {
	I__menor_que__
	I__menor_ou_igual__
	I__igual__
	I__diferente__
	I__maior_que__
	I__maior_ou_igual__
}

// ============================================================================
// OPERADORES LÓGICOS BOOLEANOS (SHORT-CIRCUIT)
// ============================================================================

// I__ou__ define o operador lógico disjuntivo 'ou' (__ou__).
type I__ou__ interface {
	M__ou__(outro Objeto) (Objeto, error)
}

// I__e__ define o operador lógico conjuntivo 'e' (__e__).
type I__e__ interface {
	M__e__(outro Objeto) (Objeto, error)
}

// I_aritmeticaBooleana engloba os operadores booleanos lógicos.
type I_aritmeticaBooleana interface {
	I__ou__
	I__e__
}

// ============================================================================
// ITERADORES E SEQUÊNCIAS
// ============================================================================

// I__iter__ define o protocolo de aquisição de iterador (__iter__).
type I__iter__ interface {
	M__iter__() (Objeto, error)
}

// I__proximo__ define o protocolo de avanço e devolução de valor de loop (__proximo__).
type I__proximo__ interface {
	M__proximo__() (Objeto, error)
}

// I_iterador encapsula de forma simplificada o protocolo completo de loops da VM.
type I_iterador interface {
	I__iter__
	I__proximo__
}

// ============================================================================
// MÉTODOS DE CONTROLE E PROTOCOLOS DE INFRAESTRUTURA
// ============================================================================

// I__tamanho__ define o protocolo para medição e contagem de itens em coleções (__tamanho__).
type I__tamanho__ interface {
	M__tamanho__() (Objeto, error)
}

// I__obtem_item__ define o protocolo para fatiamento ou indexação por colchetes [] (__obtem_item__).
type I__obtem_item__ interface {
	M__obtem_item__(obj Objeto) (Objeto, error)
}

// I__define_item__ define o protocolo para escrita indexada por colchetes [] (__define_item__).
type I__define_item__ interface {
	M__define_item__(chave, valor Objeto) (Objeto, error)
}

// I__nova_instancia__ define as rotinas de construtor estático alocador (__nova_instancia__).
type I__nova_instancia__ interface {
	M__nova_instancia__(meta *Tipo, args Tupla) (Objeto, error)
}

// I__inicializa__ define as rotinas do inicializador de propriedades (__inicializa__).
type I__inicializa__ interface {
	M__inicializa__(args Tupla) error
}

// I__contem__ define o protocolo de pertinência para verificar se um elemento reside na coleção (__contem__).
type I__contem__ interface {
	M__contem__(args Objeto) (Objeto, error)
}
