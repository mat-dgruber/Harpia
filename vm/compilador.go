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
	Constantes  []hrp.Objeto
	Bytecode    []byte
	NomesLocais map[uint16]string
}

// SimboloLocal mapeia o identificador local para seu slot numérico no frame da VM.
type SimboloLocal struct {
	Nome string
	Slot uint16
}

// EscopoCompilador armazena os símbolos locais resolvidos estaticamente em tempo de compilação.
type EscopoCompilador struct {
	Pai    *EscopoCompilador
	Locais []SimboloLocal
}

func (e *EscopoCompilador) ResolverLocal(nome string) (uint16, bool) {
	if e == nil {
		return 0, false
	}
	for i := len(e.Locais) - 1; i >= 0; i-- {
		if e.Locais[i].Nome == nome {
			return e.Locais[i].Slot, true
		}
	}
	if e.Pai != nil {
		return e.Pai.ResolverLocal(nome)
	}
	return 0, false
}

func (e *EscopoCompilador) AdicionarLocal(nome string) uint16 {
	if e == nil {
		return 0
	}
	slot := uint16(len(e.Locais))
	if e.Pai != nil {
		slot += uint16(e.Pai.TotalLocais())
	}
	e.Locais = append(e.Locais, SimboloLocal{Nome: nome, Slot: slot})
	return slot
}

func (e *EscopoCompilador) TotalLocais() int {
	if e == nil {
		return 0
	}
	tot := len(e.Locais)
	if e.Pai != nil {
		tot += e.Pai.TotalLocais()
	}
	return tot
}

// LoopFrame armazena o contexto de controle de fluxo (pare / continue) de um laço ativo.
type LoopFrame struct {
	EnderecoInicio uint16
	PulosPare      []int
}

// Compilador realiza a tradução de passagem única (single-pass) da AST do Harpia para bytecode.
type Compilador struct {
	Constantes  []hrp.Objeto
	Bytecode    []byte
	Escopo      *EscopoCompilador
	Contexto    *hrp.Contexto
	NomesLocais map[uint16]string
	LoopStack   []LoopFrame
	constMap    map[string]uint16
}

// NewCompilador inicializa uma nova instância ativa do compilador de bytecode.
func NewCompilador(ctx ...*hrp.Contexto) *Compilador {
	var cCtx *hrp.Contexto
	if len(ctx) > 0 {
		cCtx = ctx[0]
	}
	return &Compilador{
		Constantes:  make([]hrp.Objeto, 0),
		Bytecode:    make([]byte, 0),
		Escopo:      &EscopoCompilador{Locais: make([]SimboloLocal, 0)},
		Contexto:    cCtx,
		NomesLocais: make(map[uint16]string),
		LoopStack:   make([]LoopFrame, 0),
		constMap:    make(map[string]uint16),
	}
}

func (c *Compilador) adicionarLocal(nome string) uint16 {
	slot := c.Escopo.AdicionarLocal(nome)
	if c.NomesLocais == nil {
		c.NomesLocais = make(map[uint16]string)
	}
	c.NomesLocais[slot] = nome
	return slot
}

// Compilar realiza a varredura recursiva da AST e monta o bytecode e o pool de constantes.
func (c *Compilador) Compilar(node parser.BaseNode) (*ProgramaCompilado, error) {
	if err := c.visite(node); err != nil {
		return nil, err
	}

	return &ProgramaCompilado{
		Constantes:  c.Constantes,
		Bytecode:    c.Bytecode,
		NomesLocais: c.NomesLocais,
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
		c.emitirConstante(OP_PUSH_CONST, hrp.Inteiro(val))

	case *parser.DecimalLiteral:
		val, err := compartilhado.StringParaDec(n.Valor)
		if err != nil {
			return err
		}
		c.emitirConstante(OP_PUSH_CONST, hrp.Decimal(val))

	case *parser.TextoLiteral:
		textoLimpo := n.Valor[1 : len(n.Valor)-1]
		c.emitirConstante(OP_PUSH_CONST, hrp.Texto(textoLimpo))

	case *parser.ConstanteLiteral:
		switch n.Valor {
		case "Verdadeiro":
			c.emitirConstante(OP_PUSH_CONST, hrp.Verdadeiro)
		case "Falso":
			c.emitirConstante(OP_PUSH_CONST, hrp.Falso)
		case "Nulo":
			c.emitirConstante(OP_PUSH_CONST, hrp.Nulo)
		}

	case *parser.Identificador:
		if slot, ok := c.Escopo.ResolverLocal(n.Nome); ok {
			c.emitirU16(OP_CARREGAR_LOCAL, slot)
		} else {
			switch n.Nome {
			case "Verdadeiro":
				c.emitirConstante(OP_PUSH_CONST, hrp.Verdadeiro)
			case "Falso":
				c.emitirConstante(OP_PUSH_CONST, hrp.Falso)
			case "Nulo":
				c.emitirConstante(OP_PUSH_CONST, hrp.Nulo)
			default:
				idx := c.internarConstante(hrp.Texto(n.Nome))
				c.emitirU16(OP_CARREGAR_GLOBAL, idx)
			}
		}

	case *parser.DeclVar:
		if n.Inicializador != nil {
			if err := c.visite(n.Inicializador); err != nil {
				return err
			}
		} else {
			c.emitirConstante(OP_PUSH_CONST, hrp.Nulo)
		}

		slot := c.adicionarLocal(n.Nome)
		c.emitirU16(OP_ARMAZENAR_LOCAL, slot)
		c.emitir(OP_POP)

	case *parser.Reatribuicao:
		opBase := n.Operador
		if len(opBase) > 1 && opBase[len(opBase)-1] == '=' {
			opBase = opBase[:len(opBase)-1]
		}

		switch obj := n.Objeto.(type) {
		case *parser.Identificador:
			slot, okLoc := c.Escopo.ResolverLocal(obj.Nome)
			if n.Operador != "=" {
				if okLoc {
					c.emitirU16(OP_CARREGAR_LOCAL, slot)
				} else {
					idx := c.internarConstante(hrp.Texto(obj.Nome))
					c.emitirU16(OP_CARREGAR_GLOBAL, idx)
				}
				if err := c.visite(n.Expressao); err != nil {
					return err
				}
				if err := c.emitirOpBinaria(opBase); err != nil {
					return err
				}
			} else {
				if err := c.visite(n.Expressao); err != nil {
					return err
				}
			}

			if okLoc {
				c.emitirU16(OP_ARMAZENAR_LOCAL, slot)
				c.emitir(OP_POP)
			} else {
				idx := c.internarConstante(hrp.Texto(obj.Nome))
				c.emitirU16(OP_ARMAZENAR_GLOBAL, idx)
				c.emitir(OP_POP)
			}

		case *parser.Indexacao:
			if n.Operador != "=" {
				if err := c.visite(obj.Objeto); err != nil {
					return err
				}
				if err := c.visite(obj.Argumento); err != nil {
					return err
				}
				if err := c.visite(obj.Objeto); err != nil {
					return err
				}
				if err := c.visite(obj.Argumento); err != nil {
					return err
				}
				c.emitir(OP_OBTER_INDICE)
				if err := c.visite(n.Expressao); err != nil {
					return err
				}
				if err := c.emitirOpBinaria(opBase); err != nil {
					return err
				}
			} else {
				if err := c.visite(obj.Objeto); err != nil {
					return err
				}
				if err := c.visite(obj.Argumento); err != nil {
					return err
				}
				if err := c.visite(n.Expressao); err != nil {
					return err
				}
			}
			c.emitir(OP_DEFINIR_INDICE)
			c.emitir(OP_POP)

		case *parser.AcessoMembro:
			nomeMembro := ""
			if id, ok := obj.Membro.(*parser.Identificador); ok {
				nomeMembro = id.Nome
			} else if txt, ok := obj.Membro.(*parser.TextoLiteral); ok {
				nomeMembro = txt.Valor
				if len(txt.Valor) >= 2 && txt.Valor[0] == '"' {
					nomeMembro = txt.Valor[1 : len(txt.Valor)-1]
				}
			} else {
				return fmt.Errorf("nome do membro deve ser identificador ou texto")
			}
			idx := c.internarConstante(hrp.Texto(nomeMembro))

			if n.Operador != "=" {
				if err := c.visite(obj.Dono); err != nil {
					return err
				}
				if err := c.visite(obj.Dono); err != nil {
					return err
				}
				c.emitirU16(OP_OBTER_ATRIBUTO, idx)
				if err := c.visite(n.Expressao); err != nil {
					return err
				}
				if err := c.emitirOpBinaria(opBase); err != nil {
					return err
				}
			} else {
				if err := c.visite(obj.Dono); err != nil {
					return err
				}
				if err := c.visite(n.Expressao); err != nil {
					return err
				}
			}
			c.emitirU16(OP_DEFINIR_ATRIBUTO, idx)
			c.emitir(OP_POP)

		default:
			return fmt.Errorf("compilação de reatribuição para o nó tipo %T ainda não suportada na VM", obj)
		}

	case *parser.OpBinaria:
		if err := c.visite(n.Esq); err != nil {
			return err
		}
		if err := c.visite(n.Dir); err != nil {
			return err
		}
		if err := c.emitirOpBinaria(n.Operador); err != nil {
			return err
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
		c.pushLoopContext(enderecoInicio)

		if err := c.visite(n.Condicao); err != nil {
			c.popLoopContext()
			return err
		}

		c.emitir(OP_JMP_FALSO, 0, 0)
		idxPuloFalso := len(c.Bytecode) - 2

		if err := c.visite(n.Corpo); err != nil {
			c.popLoopContext()
			return err
		}

		// Salta de volta para reavaliar a condição do loop
		c.emitir(OP_JMP, byte(enderecoInicio>>8), byte(enderecoInicio))

		// Remenda o salto falso e quaisquer pulos pare para o final do loop
		offsetFim := uint16(len(c.Bytecode))
		binary.BigEndian.PutUint16(c.Bytecode[idxPuloFalso:], offsetFim)
		c.remendarPulosPare()
		c.popLoopContext()

	case *parser.BlocoPara:
		// Compila o iterável
		if err := c.visite(n.Iterador); err != nil {
			return err
		}
		iterSlot := c.adicionarLocal("__iter_" + n.Identificador)
		c.emitirU16(OP_ARMAZENAR_LOCAL, iterSlot)
		c.emitir(OP_POP)

		idxSlot := c.adicionarLocal("__idx_" + n.Identificador)
		c.emitirConstante(OP_PUSH_CONST, hrp.Inteiro(0))
		c.emitirU16(OP_ARMAZENAR_LOCAL, idxSlot)
		c.emitir(OP_POP)

		itemSlot := c.adicionarLocal(n.Identificador)

		enderecoInicio := uint16(len(c.Bytecode))
		c.pushLoopContext(enderecoInicio)

		// Condição do laço: idx < iteravel.tamanho
		c.emitirU16(OP_CARREGAR_LOCAL, idxSlot)
		c.emitirU16(OP_CARREGAR_LOCAL, iterSlot)
		idxTam := c.internarConstante(hrp.Texto("tamanho"))
		c.emitirU16(OP_OBTER_ATRIBUTO, idxTam)
		c.emitir(OP_LT)

		c.emitir(OP_JMP_FALSO, 0, 0)
		idxPuloFalso := len(c.Bytecode) - 2

		// item = iteravel[idx]
		c.emitirU16(OP_CARREGAR_LOCAL, iterSlot)
		c.emitirU16(OP_CARREGAR_LOCAL, idxSlot)
		c.emitir(OP_OBTER_INDICE)
		c.emitirU16(OP_ARMAZENAR_LOCAL, itemSlot)
		c.emitir(OP_POP)

		// Compila corpo
		if err := c.visite(n.Corpo); err != nil {
			c.popLoopContext()
			return err
		}

		// Incrementa idx: idx = idx + 1
		c.emitirU16(OP_CARREGAR_LOCAL, idxSlot)
		c.emitirConstante(OP_PUSH_CONST, hrp.Inteiro(1))
		c.emitir(OP_ADD)
		c.emitirU16(OP_ARMAZENAR_LOCAL, idxSlot)
		c.emitir(OP_POP)

		c.emitir(OP_JMP, byte(enderecoInicio>>8), byte(enderecoInicio))

		offsetFim := uint16(len(c.Bytecode))
		binary.BigEndian.PutUint16(c.Bytecode[idxPuloFalso:], offsetFim)
		c.remendarPulosPare()
		c.popLoopContext()

	case *parser.PareNode:
		if len(c.LoopStack) == 0 {
			return fmt.Errorf("instrução 'pare' fora de um laço de repetição")
		}
		c.emitir(OP_JMP, 0, 0)
		topoIdx := len(c.LoopStack) - 1
		c.LoopStack[topoIdx].PulosPare = append(c.LoopStack[topoIdx].PulosPare, len(c.Bytecode)-2)

	case *parser.ContinueNode:
		if len(c.LoopStack) == 0 {
			return fmt.Errorf("instrução 'continue' fora de um laço de repetição")
		}
		topoIdx := len(c.LoopStack) - 1
		inicio := c.LoopStack[topoIdx].EnderecoInicio
		c.emitir(OP_JMP, byte(inicio>>8), byte(inicio))

	case *parser.RetorneNode:
		if n.Expressao != nil {
			// ponytail: Otimização de Fusão de Bytecodes (Super-Instruções - Fase D)
			switch exp := n.Expressao.(type) {
			case *parser.Identificador:
				if slot, ok := c.Escopo.ResolverLocal(exp.Nome); ok {
					c.emitirU16(OP_RETORNE_LOCAL, slot)
					return nil
				}
			case *parser.InteiroLiteral:
				val, err := compartilhado.StringParaInt(exp.Valor)
				if err == nil {
					idx := c.internarConstante(hrp.Inteiro(val))
					c.emitirU16(OP_RETORNE_CONST, idx)
					return nil
				}
			case *parser.DecimalLiteral:
				val, err := compartilhado.StringParaDec(exp.Valor)
				if err == nil {
					idx := c.internarConstante(hrp.Decimal(val))
					c.emitirU16(OP_RETORNE_CONST, idx)
					return nil
				}
			case *parser.TextoLiteral:
				textoLimpo := exp.Valor[1 : len(exp.Valor)-1]
				idx := c.internarConstante(hrp.Texto(textoLimpo))
				c.emitirU16(OP_RETORNE_CONST, idx)
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
					c.emitirU16(OP_RETORNE_CONST, idx)
					return nil
				}
			}

			if err := c.visite(n.Expressao); err != nil {
				return err
			}
		} else {
			idx := c.internarConstante(hrp.Nulo)
			c.emitirU16(OP_RETORNE_CONST, idx)
			return nil
		}
		c.emitir(OP_RETORNE)

	case *parser.ChamadaFuncao:
		// Compila o identificador/alvo chamável primeiro
		if err := c.visite(n.Identificador); err != nil {
			return err
		}
		// Compila os argumentos (empilhando-os em seguida)
		for _, arg := range n.Argumentos {
			if err := c.visite(arg); err != nil {
				return err
			}
		}
		// Emite chamada com o número de argumentos
		c.emitir(OP_CHAMAR, byte(len(n.Argumentos)))

	case *parser.DeclFuncao:
		// Para manter a VM de bytecode simples, as funções são registradas no pool de constantes como objetos Funcao nativos,
		// ou compilados. Armazenamos uma representação do escopo local
		funcao := hrp.NewFuncao(n.Nome, n.Corpo, c.Contexto, nil)
		nomes := make([]string, len(n.Parametros))
		for idx, param := range n.Parametros {
			nomes[idx] = param.Nome
		}
		funcao.DefinirArgs(nomes)
		funcao.Assincrono = n.Assincrono

		idxFuncao := c.internarConstante(funcao)
		c.emitirU16(OP_PUSH_CONST, idxFuncao)

		if n.Nome != "" {
			idxNome := c.internarConstante(hrp.Texto(n.Nome))
			c.emitirU16(OP_ARMAZENAR_GLOBAL, idxNome)
		}

	case *parser.AguardeNode:
		// Compila a expressão que gera a promessa (que será empilhada no topo)
		if err := c.visite(n.Expressao); err != nil {
			return err
		}
		// Emite o opcode de await
		c.emitir(OP_AWAIT)

	case *parser.ListaLiteral:
		for _, elem := range n.Elementos {
			if err := c.visite(elem); err != nil {
				return err
			}
		}
		c.emitirU16(OP_CRIAR_LISTA, uint16(len(n.Elementos)))

	case *parser.MapaLiteral:
		for _, par := range n.Entradas {
			if id, ok := par.Chave.(*parser.Identificador); ok {
				c.emitirConstante(OP_PUSH_CONST, hrp.Texto(id.Nome))
			} else if txt, ok := par.Chave.(*parser.TextoLiteral); ok {
				textoLimpo := txt.Valor
				if len(txt.Valor) >= 2 && txt.Valor[0] == '"' {
					textoLimpo = txt.Valor[1 : len(txt.Valor)-1]
				}
				c.emitirConstante(OP_PUSH_CONST, hrp.Texto(textoLimpo))
			} else {
				if err := c.visite(par.Chave); err != nil {
					return err
				}
			}

			if err := c.visite(par.Valor); err != nil {
				return err
			}
		}
		c.emitirU16(OP_CRIAR_MAPA, uint16(len(n.Entradas)))

	case *parser.Indexacao:
		if err := c.visite(n.Objeto); err != nil {
			return err
		}
		if err := c.visite(n.Argumento); err != nil {
			return err
		}
		c.emitir(OP_OBTER_INDICE)

	case *parser.AcessoMembro:
		if err := c.visite(n.Dono); err != nil {
			return err
		}
		nomeMembro := ""
		if id, ok := n.Membro.(*parser.Identificador); ok {
			nomeMembro = id.Nome
		} else if txt, ok := n.Membro.(*parser.TextoLiteral); ok {
			nomeMembro = txt.Valor
			if len(txt.Valor) >= 2 && txt.Valor[0] == '"' {
				nomeMembro = txt.Valor[1 : len(txt.Valor)-1]
			}
		} else {
			return fmt.Errorf("nome do membro deve ser identificador ou texto")
		}
		idx := c.internarConstante(hrp.Texto(nomeMembro))
		c.emitirU16(OP_OBTER_ATRIBUTO, idx)

	case *parser.OpCoalescenciaNula:
		if err := c.visite(n.Esq); err != nil {
			return err
		}
		c.emitir(OP_COALESCE, 0, 0)
		idxPulo := len(c.Bytecode) - 2

		if err := c.visite(n.Dir); err != nil {
			return err
		}

		offset := uint16(len(c.Bytecode) - (idxPulo + 2))
		binary.BigEndian.PutUint16(c.Bytecode[idxPulo:], offset)

	case *parser.AcessoMembroOpcional:
		if err := c.visite(n.Objeto); err != nil {
			return err
		}
		nomeMembro := ""
		if id, ok := n.Membro.(*parser.Identificador); ok {
			nomeMembro = id.Nome
		} else if txt, ok := n.Membro.(*parser.TextoLiteral); ok {
			nomeMembro = txt.Valor
			if len(txt.Valor) >= 2 && txt.Valor[0] == '"' {
				nomeMembro = txt.Valor[1 : len(txt.Valor)-1]
			}
		} else {
			return fmt.Errorf("nome do membro opcional deve ser identificador ou texto")
		}
		idx := c.internarConstante(hrp.Texto(nomeMembro))
		c.emitirU16(OP_ACESSAR_OPCIONAL, idx)

	case *parser.TenteCaptureFinalmente:
		c.emitir(OP_SETUP_TRY, 0, 0)
		idxSetupTry := len(c.Bytecode) - 2

		if err := c.visite(n.TenteBlock); err != nil {
			return err
		}

		c.emitir(OP_POP_TRY)

		c.emitir(OP_JMP, 0, 0)
		idxPuloFim := len(c.Bytecode) - 2

		catchIP := uint16(len(c.Bytecode))
		binary.BigEndian.PutUint16(c.Bytecode[idxSetupTry:], catchIP)

		if n.CaptureBlock != nil {
			if n.NomeErro != "" {
				slot := c.adicionarLocal(n.NomeErro)
				c.emitirU16(OP_ARMAZENAR_LOCAL, slot)
				c.emitir(OP_POP)
			} else {
				c.emitir(OP_POP)
			}

			if err := c.visite(n.CaptureBlock); err != nil {
				return err
			}
		} else {
			c.emitir(OP_POP)
		}

		offsetFim := uint16(len(c.Bytecode))
		binary.BigEndian.PutUint16(c.Bytecode[idxPuloFim:], offsetFim)

		if n.FinalmenteBlock != nil {
			if err := c.visite(n.FinalmenteBlock); err != nil {
				return err
			}
		}

	case *parser.DeclVarDestructuring, *parser.DeclDesestruturacao:
		return fmt.Errorf("desestruturação var [a, b] = ... ainda não é suportada na VM de Bytecode (use transpilação ou interpretador)")

	case *parser.DeclEnum, *parser.DeclInterface:
		return fmt.Errorf("enumerações e interfaces ainda não são suportadas na VM de Bytecode (use transpilação ou interpretador)")

	case *parser.OpUnaria:
		if err := c.visite(n.Expressao); err != nil {
			return err
		}
		switch n.Operador {
		case "-":
			c.emitir(OP_NEG)
		case "+":
			// ponytail: operador unário positivo é um no-op
		case "nao":
			c.emitir(OP_NOT)
		default:
			return fmt.Errorf("operador unário '%s' não suportado na VM de bytecode", n.Operador)
		}

	default:
		return fmt.Errorf("compilação do nó tipo %T ainda não implementada na VM de bytecode", n)
	}

	return nil
}

func (c *Compilador) emitirOpBinaria(operador string) error {
	switch operador {
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
		return fmt.Errorf("operador binário '%s' não suportado na VM de bytecode", operador)
	}
	return nil
}

// emitir anexa o opcode e seus operandos de byte no buffer.
func (c *Compilador) emitir(op Opcode, ops ...byte) {
	c.Bytecode = append(c.Bytecode, op)
	c.Bytecode = append(c.Bytecode, ops...)
}

// emitirU16 anexa o opcode e seu operando u16 BigEndian no buffer.
func (c *Compilador) emitirU16(op Opcode, operand uint16) {
	c.Bytecode = append(c.Bytecode, op, byte(operand>>8), byte(operand))
}

// emitirConstante insere constante no pool e emite opcode com operando uint16.
func (c *Compilador) emitirConstante(op Opcode, val hrp.Objeto) uint16 {
	idx := c.internarConstante(val)
	c.emitirU16(op, idx)
	return idx
}

func (c *Compilador) pushLoopContext(enderecoInicio uint16) {
	c.LoopStack = append(c.LoopStack, LoopFrame{
		EnderecoInicio: enderecoInicio,
		PulosPare:      make([]int, 0),
	})
}

func (c *Compilador) popLoopContext() {
	if len(c.LoopStack) > 0 {
		c.LoopStack = c.LoopStack[:len(c.LoopStack)-1]
	}
}

func (c *Compilador) remendarPulosPare() {
	if len(c.LoopStack) == 0 {
		return
	}
	topoIdx := len(c.LoopStack) - 1
	offsetFim := uint16(len(c.Bytecode))
	for _, ip := range c.LoopStack[topoIdx].PulosPare {
		binary.BigEndian.PutUint16(c.Bytecode[ip:], offsetFim)
	}
}

// internarConstante busca se o objeto já existe no pool de constantes.
// Se existir, retorna seu índice u16; senão, adiciona-o ao pool e retorna o novo índice.
func (c *Compilador) internarConstante(val hrp.Objeto) uint16 {
	if c.constMap == nil {
		c.constMap = make(map[string]uint16)
	}

	chave := fmt.Sprintf("%T:%v", val, val)
	if idx, ok := c.constMap[chave]; ok {
		return idx
	}

	for i, constVal := range c.Constantes {
		if constVal == val {
			idx := uint16(i)
			c.constMap[chave] = idx
			return idx
		}
		switch v1 := constVal.(type) {
		case hrp.Texto:
			if v2, ok := val.(hrp.Texto); ok && v1 == v2 {
				idx := uint16(i)
				c.constMap[chave] = idx
				return idx
			}
		case hrp.Inteiro:
			if v2, ok := val.(hrp.Inteiro); ok && v1 == v2 {
				idx := uint16(i)
				c.constMap[chave] = idx
				return idx
			}
		case hrp.Decimal:
			if v2, ok := val.(hrp.Decimal); ok && v1 == v2 {
				idx := uint16(i)
				c.constMap[chave] = idx
				return idx
			}
		case hrp.Booleano:
			if v2, ok := val.(hrp.Booleano); ok && v1 == v2 {
				idx := uint16(i)
				c.constMap[chave] = idx
				return idx
			}
		}
	}

	idx := uint16(len(c.Constantes))
	c.Constantes = append(c.Constantes, val)
	c.constMap[chave] = idx
	return idx
}
