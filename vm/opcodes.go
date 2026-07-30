package vm

// Opcode define a assinatura numérica de 1 byte para cada instrução da VM V2.
type Opcode = byte

const (
	OP_PUSH_CONST Opcode = 0x01 // Empilha uma constante do pool pelo índice [2 bytes operand].
	OP_POP        Opcode = 0x02 // Desempilha o elemento no topo da pilha.
	OP_DUP        Opcode = 0x03 // Duplica o valor no topo da pilha.

	// Aritmética e Operações Binárias
	OP_ADD     Opcode = 0x04 // Polimorfismo hrp.Adiciona (a + b)
	OP_SUB     Opcode = 0x05 // Polimorfismo hrp.Subtrai (a - b)
	OP_MUL     Opcode = 0x06 // Polimorfismo hrp.Multiplica (a * b)
	OP_DIV     Opcode = 0x07 // Polimorfismo hrp.Divide (a / b)
	OP_DIV_INT Opcode = 0x08 // Divisão inteira (a // b)
	OP_MOD     Opcode = 0x09 // Resto da divisão (a % b)

	// Comparações Ricas
	OP_EQ  Opcode = 0x0A // Igualdade (a == b)
	OP_NEQ Opcode = 0x0B // Diferença (a != b)
	OP_LT  Opcode = 0x0C // Menor que (a < b)
	OP_LTE Opcode = 0x0D // Menor ou igual (a <= b)
	OP_GT  Opcode = 0x0E // Maior que (a > b)
	OP_GTE Opcode = 0x0F // Maior ou igual (a >= b)

	// Controle de Fluxo e Saltos (Operandos u16)
	OP_JMP       Opcode = 0x10 // Salto absoluto incondicional para IP [2 bytes operand].
	OP_JMP_FALSO Opcode = 0x11 // Salto absoluto se topo for falso ou nulo [2 bytes operand].

	// Escopos, Variáveis Locais O(1) e Globais
	OP_CARREGAR_LOCAL   Opcode = 0x12 // Carrega variável local por slot O(1) [2 bytes operand index].
	OP_ARMAZENAR_LOCAL  Opcode = 0x13 // Armazena valor no slot local O(1) [2 bytes operand index].
	OP_CARREGAR_GLOBAL  Opcode = 0x14 // Carrega variável global por nome no pool [2 bytes operand index].
	OP_ARMAZENAR_GLOBAL Opcode = 0x15 // Armazena variável global por nome no pool [2 bytes operand index].

	// Aliases legados de transição
	OP_CARREGAR_VAR  = OP_CARREGAR_GLOBAL
	OP_ARMAZENAR_VAR = OP_ARMAZENAR_GLOBAL
	OP_RETORNE_VAR   = OP_RETORNE_LOCAL

	// Funções, Chamadas e Objetos
	OP_CHAMAR       Opcode = 0x16 // Invoca chamável com N argumentos no topo [1 byte indicando aridade].
	OP_RETORNE      Opcode = 0x17 // Sair do frame corrente e retornar o topo da pilha.
	OP_AWAIT        Opcode = 0x18 // Aguarda a resolução de uma promessa.
	OP_CRIAR_FUNCAO Opcode = 0x19 // Cria uma função em runtime [2 bytes const index].

	// Super-Instruções e Otimizações
	OP_RETORNE_CONST Opcode = 0x1A // Carrega constante do pool e retorna de forma atômica [2 bytes operand].
	OP_RETORNE_LOCAL Opcode = 0x1B // Carrega local por slot e retorna de forma atômica [2 bytes operand].

	// Operações Unárias
	OP_NEG Opcode = 0x1C // Operação unária de negação aritmética (-x).
	OP_NOT Opcode = 0x1D // Operação unária de negação lógica (nao x).

	// Coleções Nativa e Indexação
	OP_CRIAR_LISTA   Opcode = 0x1E // Cria lista com N elementos da pilha [2 bytes operand].
	OP_CRIAR_MAPA    Opcode = 0x1F // Cria mapa com N pares (2*N elementos) da pilha [2 bytes operand].
	OP_OBTER_INDICE  Opcode = 0x20 // Acesso por índice: alvo[indice].
	OP_DEFINIR_INDICE Opcode = 0x21 // Atribuição por índice: alvo[indice] = valor.

	// Membros e Operadores Nulo-Seguros
	OP_OBTER_ATRIBUTO   Opcode = 0x22 // Acesso a membro por nome [2 bytes const index].
	OP_DEFINIR_ATRIBUTO Opcode = 0x23 // Atribuição a membro por nome [2 bytes const index].
	OP_COALESCE         Opcode = 0x24 // Operador nulo-seguro ?? [2 bytes offset pulo].
	OP_ACESSAR_OPCIONAL Opcode = 0x25 // Acesso opcional ?. [2 bytes const index].

	// Tratamento de Exceções
	OP_SETUP_TRY Opcode = 0x26 // Registra handler de exceção [2 bytes catch IP].
	OP_POP_TRY   Opcode = 0x27 // Remove handler de exceção ativo.
	OP_LANCAR    Opcode = 0x28 // Lança a exceção presente no topo da pilha.
)

// NomeOpcode traduz o byte de instrução em sua correspondente representação string legível para depuração.
func NomeOpcode(op Opcode) string {
	switch op {
	case OP_PUSH_CONST:
		return "OP_PUSH_CONST"
	case OP_POP:
		return "OP_POP"
	case OP_DUP:
		return "OP_DUP"
	case OP_ADD:
		return "OP_ADD"
	case OP_SUB:
		return "OP_SUB"
	case OP_MUL:
		return "OP_MUL"
	case OP_DIV:
		return "OP_DIV"
	case OP_DIV_INT:
		return "OP_DIV_INT"
	case OP_MOD:
		return "OP_MOD"
	case OP_EQ:
		return "OP_EQ"
	case OP_NEQ:
		return "OP_NEQ"
	case OP_LT:
		return "OP_LT"
	case OP_LTE:
		return "OP_LTE"
	case OP_GT:
		return "OP_GT"
	case OP_GTE:
		return "OP_GTE"
	case OP_JMP:
		return "OP_JMP"
	case OP_JMP_FALSO:
		return "OP_JMP_FALSO"
	case OP_CARREGAR_LOCAL:
		return "OP_CARREGAR_LOCAL"
	case OP_ARMAZENAR_LOCAL:
		return "OP_ARMAZENAR_LOCAL"
	case OP_CARREGAR_GLOBAL:
		return "OP_CARREGAR_GLOBAL"
	case OP_ARMAZENAR_GLOBAL:
		return "OP_ARMAZENAR_GLOBAL"
	case OP_CHAMAR:
		return "OP_CHAMAR"
	case OP_RETORNE:
		return "OP_RETORNE"
	case OP_AWAIT:
		return "OP_AWAIT"
	case OP_CRIAR_FUNCAO:
		return "OP_CRIAR_FUNCAO"
	case OP_RETORNE_CONST:
		return "OP_RETORNE_CONST"
	case OP_RETORNE_LOCAL:
		return "OP_RETORNE_LOCAL"
	case OP_NEG:
		return "OP_NEG"
	case OP_NOT:
		return "OP_NOT"
	case OP_CRIAR_LISTA:
		return "OP_CRIAR_LISTA"
	case OP_CRIAR_MAPA:
		return "OP_CRIAR_MAPA"
	case OP_OBTER_INDICE:
		return "OP_OBTER_INDICE"
	case OP_DEFINIR_INDICE:
		return "OP_DEFINIR_INDICE"
	case OP_OBTER_ATRIBUTO:
		return "OP_OBTER_ATRIBUTO"
	case OP_DEFINIR_ATRIBUTO:
		return "OP_DEFINIR_ATRIBUTO"
	case OP_COALESCE:
		return "OP_COALESCE"
	case OP_ACESSAR_OPCIONAL:
		return "OP_ACESSAR_OPCIONAL"
	case OP_SETUP_TRY:
		return "OP_SETUP_TRY"
	case OP_POP_TRY:
		return "OP_POP_TRY"
	case OP_LANCAR:
		return "OP_LANCAR"
	}
	return "OP_DESCONHECIDO"
}
