package vm

// Opcode define a assinatura numérica de 1 byte para cada instrução da VM.
type Opcode = byte

const (
	OP_PUSH_CONST Opcode = 0x01 // Empilha uma constante do pool pelo índice [1 byte operand].
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

	// Controle de Fluxo e Saltos
	OP_JMP       Opcode = 0x10 // Salto absoluto incondicional para IP [2 bytes operand].
	OP_JMP_FALSO Opcode = 0x11 // Salto absoluto se topo for falso ou nulo [2 bytes operand].

	// Escopos e Variáveis
	OP_CARREGAR_VAR  Opcode = 0x12 // Carrega variável local/global por nome [1 byte operand indicando índice no pool].
	OP_ARMAZENAR_VAR Opcode = 0x13 // Armazena/reatribui valor no topo à variável [1 byte operand index pool].

	// Funções, Chamadas e Objetos
	OP_CHAMAR       Opcode = 0x14 // Invoca chamável com N argumentos no topo [1 byte indicando aridade].
	OP_RETORNE      Opcode = 0x15 // Sair do frame corrente e retornar o topo da pilha.
	OP_AWAIT        Opcode = 0x16 // Aguarda a resolução de uma promessa.
	OP_CRIAR_FUNCAO Opcode = 0x17 // Cria uma função em runtime [1 byte index do nome, 1 byte index dos argumentos separados, 1 byte indicando flags (ex: assincrono)].

	// Super-Instruções e Otimizações de Fusão (Fase D)
	OP_RETORNE_CONST Opcode = 0x18 // Carrega constante do pool e retorna de forma atômica [1 byte operand].
	OP_RETORNE_VAR   Opcode = 0x19 // Carrega variável por nome do pool e retorna de forma atômica [1 byte operand].
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
	case OP_CARREGAR_VAR:
		return "OP_CARREGAR_VAR"
	case OP_ARMAZENAR_VAR:
		return "OP_ARMAZENAR_VAR"
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
	case OP_RETORNE_VAR:
		return "OP_RETORNE_VAR"
	}
	return "OP_DESCONHECIDO"
}
