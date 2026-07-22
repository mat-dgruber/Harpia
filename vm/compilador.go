package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/mat-dgruber/Harpia/compartilhado"
	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
)

// ProgramaCompilado agrupa o pool de constantes extraídas e o bytecode plano gerado para execução na VM.
type ProgramaCompilado struct {
	Constantes []hrp.Objeto
	Bytecode   []byte
}

// Compilador realiza a tradução de passagem única (single-pass) da AST do Harpia para bytecode.
type Compilador struct {
	Constantes []hrp.Objeto
	Bytecode   []byte
}

// NewCompilador inicializa uma nova instância ativa do compilador de bytecode.
func NewCompilador() *Compilador {
	return &Compilador{
		Constantes: make([]hrp.Objeto, 0),
		Bytecode:   make([]byte, 0),
	}
}

// Compilar realiza a varredura recursiva da AST e monta o bytecode e o pool de constantes.
func (c *Compilador) Compilar(node parser.BaseNode) (*ProgramaCompilado, error) {
	if err := c.visite(node); err != nil {
		return nil, err
	}

	return &ProgramaCompilado{
		Constantes: c.Constantes,
		Bytecode:   c.Bytecode,
	}, nil
}

// visite percorre a árvore de sintaxe abstrata (AST) recursivamente em uma única passagem (single-pass).
// Conforme visita cada tipo de nó específico (declarações, literais, chamadas, condicionais, etc.),
// ele gera a respectiva sequência de bytes (bytecodes/opcodes) no buffer e adiciona as literais
// extraídas ao pool global de constantes do programa compilado.
func (c *Compilador) visite(node parser.BaseNode) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *parser.Programa:
		for _, decl := range n.Declaracoes {
			if err := c.visite(decl); err != nil {
				return err
			}
		}

	case *parser.Bloco:
		for _, decl := range n.Declaracoes {
			if err := c.visite(decl); err != nil {
				return err
			}
		}

	case *parser.InteiroLiteral:
		val, err := compartilhado.StringParaInt(n.Valor)
		if err != nil {
			return err
		}
		obj := hrp.Inteiro(val)
		idx := c.internarConstante(obj)
		c.emitir(OP_PUSH_CONST, idx)

	case *parser.DecimalLiteral:
		val, err := compartilhado.StringParaDec(n.Valor)
		if err != nil {
			return err
		}
		obj := hrp.Decimal(val)
		idx := c.internarConstante(obj)
		c.emitir(OP_PUSH_CONST, idx)

	case *parser.TextoLiteral:
		// Strip aspas limitadoras do lexema do texto (exatamente como o interpretador faz na linha 256)
		textoLimpo := n.Valor[1 : len(n.Valor)-1]
		obj := hrp.Texto(textoLimpo)
		idx := c.internarConstante(obj)
		c.emitir(OP_PUSH_CONST, idx)

	case *parser.ConstanteLiteral:
		switch n.Valor {
		case "Verdadeiro":
			idx := c.internarConstante(hrp.Verdadeiro)
			c.emitir(OP_PUSH_CONST, idx)
		case "Falso":
			idx := c.internarConstante(hrp.Falso)
			c.emitir(OP_PUSH_CONST, idx)
		case "Nulo":
			idx := c.internarConstante(hrp.Nulo)
			c.emitir(OP_PUSH_CONST, idx)
		}

	case *parser.Identificador:
		idx := c.internarConstante(hrp.Texto(n.Nome))
		c.emitir(OP_CARREGAR_VAR, idx)

	case *parser.DeclVar:
		// Se houver um inicializador, compila-o; senão, empilha Nulo por default
		if n.Inicializador != nil {
			if err := c.visite(n.Inicializador); err != nil {
				return err
			}
		} else {
			idx := c.internarConstante(hrp.Nulo)
			c.emitir(OP_PUSH_CONST, idx)
		}

		idx := c.internarConstante(hrp.Texto(n.Nome))
		c.emitir(OP_ARMAZENAR_VAR, idx)

	case *parser.Reatribuicao:
		if err := c.visite(n.Expressao); err != nil {
			return err
		}

		if id, ok := n.Objeto.(*parser.Identificador); ok {
			idx := c.internarConstante(hrp.Texto(id.Nome))
			c.emitir(OP_ARMAZENAR_VAR, idx)
		} else {
			return fmt.Errorf("compilação de reatribuição para objetos complexos ainda não suportada na VM")
		}

	case *parser.OpBinaria:
		// Compila esquerda e direita recursivamente
		if err := c.visite(n.Esq); err != nil {
			return err
		}
		if err := c.visite(n.Dir); err != nil {
			return err
		}

		// Emite opcode aritmético correspondente
		switch n.Operador {
		case "+":
			c.emitir(OP_ADD)
		case "-":
			c.emitir(OP_SUB)
		case "*":
			c.emitir(OP_MUL)
		case "/":
			c.emitir(OP_DIV)
		case "//":
			c.emitir(OP_DIV_INT)
		case "%":
			c.emitir(OP_MOD)
		case "==":
			c.emitir(OP_EQ)
		case "!=":
			c.emitir(OP_NEQ)
		case "<":
			c.emitir(OP_LT)
		case "<=":
			c.emitir(OP_LTE)
		case ">":
			c.emitir(OP_GT)
		case ">=":
			c.emitir(OP_GTE)
		default:
			return fmt.Errorf("operador binário '%s' não suportado na VM de bytecode", n.Operador)
		}

	case *parser.ExpressaoSe:
		if err := c.visite(n.Condicao); err != nil {
			return err
		}

		// Insere slot temporário para JMP_FALSO (2 bytes para offset de pulo)
		c.emitir(OP_JMP_FALSO, 0, 0)
		idxPuloFalso := len(c.Bytecode) - 2

		// Compila bloco do corpo
		if err := c.visite(n.Corpo); err != nil {
			return err
		}

		// Se houver senao (Alternativa), precisamos de JMP incondicional no fim do bloco se
		if n.Alternativa != nil {
			c.emitir(OP_JMP, 0, 0)
			idxPuloSe := len(c.Bytecode) - 2

			// Remenda o salto do JMP_FALSO para o início do bloco senao
			offsetSenao := uint16(len(c.Bytecode))
			binary.BigEndian.PutUint16(c.Bytecode[idxPuloFalso:], offsetSenao)

			// Compila bloco do senao
			if err := c.visite(n.Alternativa); err != nil {
				return err
			}

			// Remenda o salto do fim do bloco se para o final total da condicional
			offsetFim := uint16(len(c.Bytecode))
			binary.BigEndian.PutUint16(c.Bytecode[idxPuloSe:], offsetFim)
		} else {
			// Remenda o salto do JMP_FALSO direto para o final total da condicional
			offsetFim := uint16(len(c.Bytecode))
			binary.BigEndian.PutUint16(c.Bytecode[idxPuloFalso:], offsetFim)
		}

	case *parser.Enquanto:
		enderecoInicio := uint16(len(c.Bytecode))

		if err := c.visite(n.Condicao); err != nil {
			return err
		}

		c.emitir(OP_JMP_FALSO, 0, 0)
		idxPuloFalso := len(c.Bytecode) - 2

		if err := c.visite(n.Corpo); err != nil {
			return err
		}

		// Salta de volta para reavaliar a condição do loop
		c.emitir(OP_JMP, byte(enderecoInicio>>8), byte(enderecoInicio))

		// Remenda o salto falso para o final do loop
		offsetFim := uint16(len(c.Bytecode))
		binary.BigEndian.PutUint16(c.Bytecode[idxPuloFalso:], offsetFim)

	case *parser.RetorneNode:
		if n.Expressao != nil {
			// ponytail: Otimização de Fusão de Bytecodes (Super-Instruções - Fase D)
			switch exp := n.Expressao.(type) {
			case *parser.Identificador:
				idx := c.internarConstante(hrp.Texto(exp.Nome))
				c.emitir(OP_RETORNE_VAR, idx)
				return nil
			case *parser.InteiroLiteral:
				val, err := compartilhado.StringParaInt(exp.Valor)
				if err == nil {
					idx := c.internarConstante(hrp.Inteiro(val))
					c.emitir(OP_RETORNE_CONST, idx)
					return nil
				}
			case *parser.DecimalLiteral:
				val, err := compartilhado.StringParaDec(exp.Valor)
				if err == nil {
					idx := c.internarConstante(hrp.Decimal(val))
					c.emitir(OP_RETORNE_CONST, idx)
					return nil
				}
			case *parser.TextoLiteral:
				textoLimpo := exp.Valor[1 : len(exp.Valor)-1]
				idx := c.internarConstante(hrp.Texto(textoLimpo))
				c.emitir(OP_RETORNE_CONST, idx)
				return nil
			case *parser.ConstanteLiteral:
				var obj hrp.Objeto
				switch exp.Valor {
				case "Verdadeiro":
					obj = hrp.Verdadeiro
				case "Falso":
					obj = hrp.Falso
				case "Nulo":
					obj = hrp.Nulo
				}
				if obj != nil {
					idx := c.internarConstante(obj)
					c.emitir(OP_RETORNE_CONST, idx)
					return nil
				}
			}

			if err := c.visite(n.Expressao); err != nil {
				return err
			}
		} else {
			idx := c.internarConstante(hrp.Nulo)
			c.emitir(OP_RETORNE_CONST, idx)
			return nil
		}
		c.emitir(OP_RETORNE)

	case *parser.ChamadaFuncao:
		// Compila os argumentos primeiro (empilhando-os da esquerda para a direita)
		for _, arg := range n.Argumentos {
			if err := c.visite(arg); err != nil {
				return err
			}
		}
		// Compila o identificador/alvo chamável
		if err := c.visite(n.Identificador); err != nil {
			return err
		}
		// Emite chamada com o número de argumentos
		c.emitir(OP_CHAMAR, byte(len(n.Argumentos)))

	case *parser.DeclFuncao:
		// Para manter a VM de bytecode simples, as funções são registradas no pool de constantes como objetos Funcao nativos,
		// ou compilados. Armazenamos uma representação do escopo local
		funcao := hrp.NewFuncao(n.Nome, n.Corpo, nil, nil)
		nomes := make([]string, len(n.Parametros))
		for idx, param := range n.Parametros {
			nomes[idx] = param.Nome
		}
		funcao.DefinirArgs(nomes)
		funcao.Assincrono = n.Assincrono

		idxFuncao := c.internarConstante(funcao)
		c.emitir(OP_PUSH_CONST, idxFuncao)

		if n.Nome != "" {
			idxNome := c.internarConstante(hrp.Texto(n.Nome))
			c.emitir(OP_ARMAZENAR_VAR, idxNome)
		}

	case *parser.AguardeNode:
		// Compila a expressão que gera a promessa (que será empilhada no topo)
		if err := c.visite(n.Expressao); err != nil {
			return err
		}
		// Emite o opcode de await
		c.emitir(OP_AWAIT)

	case *parser.DeclVarDestructuring:
		return fmt.Errorf("desestruturação var [a, b] = ... ainda não é suportada na VM de Bytecode (use transpilação ou interpretador)")

	case *parser.OpCoalescenciaNula, *parser.AcessoMembroOpcional:
		return fmt.Errorf("operadores nulo-seguros (?? e ?.) ainda não são suportados na VM de Bytecode (use transpilação ou interpretador)")

	case *parser.DeclEnum, *parser.DeclInterface:
		return fmt.Errorf("enumerações e interfaces ainda não são suportadas na VM de Bytecode (use transpilação ou interpretador)")

	default:
		return fmt.Errorf("compilação do nó tipo %T ainda não implementada na VM de bytecode", n)
	}

	return nil
}

// emitir anexa o opcode e seus operandos adicionais no buffer do bytecode gerado.
func (c *Compilador) emitir(op Opcode, ops ...byte) {
	c.Bytecode = append(c.Bytecode, op)
	c.Bytecode = append(c.Bytecode, ops...)
}

// internarConstante busca se o objeto já existe no pool de constantes.
// Se existir, retorna seu índice; senão, adiciona-o ao pool e retorna o novo índice.
func (c *Compilador) internarConstante(val hrp.Objeto) byte {
	// Procura duplicações de valores de constantes de forma linear (lazy/ponytail)
	for i, constVal := range c.Constantes {
		// Comparações básicas estáveis para deduplicação
		if constVal == val {
			return byte(i)
		}
		// Fallback para tipos de valores literais suportados
		switch v1 := constVal.(type) {
		case hrp.Texto:
			if v2, ok := val.(hrp.Texto); ok && v1 == v2 {
				return byte(i)
			}
		case hrp.Inteiro:
			if v2, ok := val.(hrp.Inteiro); ok && v1 == v2 {
				return byte(i)
			}
		case hrp.Decimal:
			if v2, ok := val.(hrp.Decimal); ok && v1 == v2 {
				return byte(i)
			}
		case hrp.Booleano:
			if v2, ok := val.(hrp.Booleano); ok && v1 == v2 {
				return byte(i)
			}
		}
	}

	c.Constantes = append(c.Constantes, val)
	return byte(len(c.Constantes) - 1)
}
