package vm

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

// poolPilha é um pool de memória sincronizado (sync.Pool) que gerencia fatias de objetos Go reutilizáveis.
// Ele é essencial para mitigar a alocação frequente e desalocação de fatias de memória (heap allocation)
// na pilha de operandos de cada Frame, reduzindo drasticamente a pressão sobre o coletor de lixo (Garbage Collector).
var poolPilha = sync.Pool{
	New: func() interface{} {
		return make([]hrp.Objeto, 0, 128)
	},
}

// poolLocais gerencia a reutilização de fatias de variáveis locais de tamanho padrão (256 slots).
var poolLocais = sync.Pool{
	New: func() interface{} {
		return make([]hrp.Objeto, 256)
	},
}

// TryHandler armazena os contextos de desvio para exceções.
type TryHandler struct {
	CatchIP   int
	FinallyIP int
	PilhaProf int
}

// Frame representa um contexto isolado de execução de função ou módulo na pilha de chamadas da VM.
type Frame struct {
	Pai              *Frame
	Profundidade     int // Profundidade do frame na pilha de chamadas (O(1))
	Bytecode         []byte
	IP               int               // Instrução corrente (Instruction Pointer)
	Consts           []hrp.Objeto      // Referência ao pool de constantes do programa compilado
	Pilha            []hrp.Objeto      // Pilha de operandos local
	Locais           []hrp.Objeto      // Slots de variáveis locais O(1)
	Escopo           *hrp.Escopo       // Tabela local de símbolos e fechamento léxico / globais
	TryStack         []TryHandler      // Pilha de handlers de tente/capture
	NomesLocais      map[uint16]string // Mapeamento estático opcional de slot u16 para nome da variável
	ConstantesLocais map[uint16]bool   // Conjunto de slots que representam constantes locais imutáveis
	MapaTokens       map[int]*lexer.Token
	Arquivo          string
	Codigo           string
	UltimoIP         int
}

// NewFrame cria um frame isolado apontando para as constantes e escopo fornecidos.
func NewFrame(bytecode []byte, consts []hrp.Objeto, escopo *hrp.Escopo, pai *Frame, nomesLocais ...map[uint16]string) *Frame {
	var nl map[uint16]string
	if len(nomesLocais) > 0 {
		nl = nomesLocais[0]
	}
	pilha := poolPilha.Get().([]hrp.Objeto)
	prof := 1
	if pai != nil {
		prof = pai.Profundidade + 1
	}

	var locais []hrp.Objeto
	if raw := poolLocais.Get(); raw != nil {
		locais = raw.([]hrp.Objeto)
		if cap(locais) >= 256 {
			locais = locais[:256]
			for i := range locais {
				locais[i] = nil
			}
		} else {
			locais = make([]hrp.Objeto, 0, 256)
		}
	} else {
		locais = make([]hrp.Objeto, 0, 256)
	}

	return &Frame{
		Pai:          pai,
		Profundidade: prof,
		Bytecode:     bytecode,
		IP:           0,
		Consts:       consts,
		Pilha:        pilha[:0],
		Locais:       locais[:0],
		Escopo:       escopo,
		NomesLocais:  nl,
	}
}

// NovoFrame é um alias conveniente para NewFrame.
func NovoFrame(bytecode []byte, consts []hrp.Objeto, escopo *hrp.Escopo, pai *Frame, nomesLocais ...map[uint16]string) *Frame {
	return NewFrame(bytecode, consts, escopo, pai, nomesLocais...)
}

// push adiciona o elemento e incrementa suas referências.
func (f *Frame) push(obj hrp.Objeto) {
	f.Pilha = append(f.Pilha, obj)
	hrp.ReterObjeto(obj)
}

// pop remove o elemento sem decrementar (transfere a posse do objeto para o receptor).
func (f *Frame) pop() hrp.Objeto {
	if len(f.Pilha) == 0 {
		return hrp.Nulo
	}
	topoIdx := len(f.Pilha) - 1
	val := f.Pilha[topoIdx]
	f.Pilha = f.Pilha[:topoIdx]
	return val
}

// ObterProfundidade calcula a profundidade do frame corrente em O(1).
func (f *Frame) ObterProfundidade() int {
	if f == nil {
		return 0
	}
	if f.Profundidade > 0 {
		return f.Profundidade
	}
	profundidade := 0
	curr := f
	for curr != nil {
		profundidade++
		curr = curr.Pai
	}
	return profundidade
}

type PerfilInfo struct {
	Opcode     Opcode
	Vezes      int
	TempoTotal time.Duration
}

// VM representa o mecanismo do motor de execução da máquina virtual de pilha.
type VM struct {
	Contexto *hrp.Contexto
	Perfil   bool
	Metricas map[Opcode]*PerfilInfo
}

// NewVM instancia a máquina virtual associada a um contexto de execução de tipos e módulos.
func NewVM(ctx *hrp.Contexto) *VM {
	return &VM{
		Contexto: ctx,
		Metricas: make(map[Opcode]*PerfilInfo),
	}
}

func init() {
	hrp.DefinirExecutadorVM(func(ctx *hrp.Contexto, ast parser.BaseNode, escopo *hrp.Escopo) (hrp.Objeto, error) {
		comp := NewCompilador(ctx)
		prog, err := comp.Compilar(ast)
		if err != nil {
			return nil, err
		}
		virtualMachine := NewVM(ctx)
		frame := NewFrame(prog.Bytecode, prog.Constantes, escopo, nil, prog.NomesLocais)
		frame.ConstantesLocais = prog.ConstantesLocais
		frame.MapaTokens = prog.MapaTokens
		frame.Arquivo = prog.Arquivo
		frame.Codigo = prog.Codigo
		return virtualMachine.Executar(frame)
	})
}

func (v *VM) registrarMetrica(op Opcode, duracao time.Duration) {
	info, ok := v.Metricas[op]
	if !ok {
		info = &PerfilInfo{Opcode: op}
		v.Metricas[op] = info
	}
	info.Vezes++
	info.TempoTotal += duracao
}

func (v *VM) ImprimirPerfil() {
	if v.Metricas == nil || len(v.Metricas) == 0 {
		return
	}
	fmt.Println("\n📊 RELATÓRIO DE PERFILAMENTO DA VM (PROFILER)")
	fmt.Println("=========================================================")
	fmt.Printf("%-20s | %-10s | %-15s\n", "Instrução (Opcode)", "Chamadas", "Tempo Total")
	fmt.Println("---------------------------------------------------------")
	for op, info := range v.Metricas {
		fmt.Printf("%-20s | %-10d | %-15v\n", NomeOpcode(op), info.Vezes, info.TempoTotal)
	}
	fmt.Println("=========================================================")
	fmt.Println()
}

// Executar inicia o loop switch zero-alloc de despacho sobre o bytecode plano do frame.
func (v *VM) Executar(frame *Frame) (hrp.Objeto, error) {
	if frame.ObterProfundidade() > 1000 {
		return nil, hrp.NewErroF(hrp.ErroDePilha, "Limite máximo de recursão excedido (1000 frames)")
	}

	defer func() {
		// Limpeza de fim de frame: libera todos os operandos remanescentes na pilha
		for len(frame.Pilha) > 0 {
			hrp.LiberarObjeto(frame.pop())
		}

		// Limpa a fatia para evitar vazamentos de referências a objetos Go antes de devolver ao pool
		fatiaLimpa := frame.Pilha[:cap(frame.Pilha)]
		for i := range fatiaLimpa {
			fatiaLimpa[i] = nil
		}
		poolPilha.Put(frame.Pilha[:0])

		if cap(frame.Locais) >= 256 {
			fatiaLocais := frame.Locais[:cap(frame.Locais)]
			for i := range fatiaLocais {
				fatiaLocais[i] = nil
			}
			poolLocais.Put(fatiaLocais[:256])
		}

		// Limpeza de escopo: realiza a varredura cíclica e libera as referências retidas pelo escopo local
		// PERF-01: só executa ColetarCiclos quando frame.Pai == nil (retorno do frame raiz)
		if frame.Escopo != nil {
			if frame.Pai == nil {
				hrp.ColetarCiclos(frame.Escopo)
			}

			for _, simb := range frame.Escopo.ObterSimbolosSeguro() {
				if simb != nil {
					hrp.LiberarObjeto(simb.Valor)
				}
			}
		}
	}()

	code := frame.Bytecode
	codeLen := len(code)

	for frame.IP < codeLen {
		frame.UltimoIP = frame.IP
		op := code[frame.IP]
		var inicio time.Time
		if v.Perfil {
			inicio = time.Now()
		}
		frame.IP++

		switch op {
		case OP_PUSH_CONST:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			frame.push(frame.Consts[idx])

		case OP_POP:
			val := frame.pop()
			hrp.LiberarObjeto(val)

		case OP_DUP:
			val := frame.pop()
			frame.push(val)
			frame.push(val)
			hrp.LiberarObjeto(val)

		case OP_ADD:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.Adiciona(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_SUB:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.Subtrai(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_MUL:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.Multiplica(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_DIV:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.Divide(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_DIV_INT:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.DivideInteiro(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_MOD:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.Mod(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_EQ:
			b := frame.pop()
			a := frame.pop()
			igual, err := hrp.Igual(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(igual)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_NEQ:
			b := frame.pop()
			a := frame.pop()
			igual, err := hrp.Igual(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			if igual == hrp.Verdadeiro {
				frame.push(hrp.Falso)
			} else {
				frame.push(hrp.Verdadeiro)
			}
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_LT:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.MenorQue(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_LTE:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.MenorOuIgual(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_GT:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.MaiorQue(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_GTE:
			b := frame.pop()
			a := frame.pop()
			res, err := hrp.MaiorOuIgual(a, b)
			if err != nil {
				hrp.LiberarObjeto(a)
				hrp.LiberarObjeto(b)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(a)
			hrp.LiberarObjeto(b)

		case OP_NEG:
			val := frame.pop()
			res, err := hrp.Neg(val)
			if err != nil {
				hrp.LiberarObjeto(val)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(val)

		case OP_NOT:
			val := frame.pop()
			res, err := hrp.Nao(val)
			if err != nil {
				hrp.LiberarObjeto(val)
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)
			hrp.LiberarObjeto(val)

		case OP_JMP:
			destino := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP = int(destino)

		case OP_JMP_FALSO:
			destino := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			cond := frame.pop()
			if cond == hrp.Falso || cond == hrp.Nulo {
				frame.IP = int(destino)
			}
			hrp.LiberarObjeto(cond)

		case OP_CARREGAR_LOCAL:
			slot := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			if int(slot) < len(frame.Locais) && frame.Locais[slot] != nil {
				frame.push(frame.Locais[slot])
			} else if frame.Escopo != nil && frame.NomesLocais != nil {
				if nome, ok := frame.NomesLocais[slot]; ok {
					if val, err := frame.Escopo.ObterValor(nome); err == nil {
						frame.push(val)
						break
					}
				}
				frame.push(hrp.Nulo)
			} else {
				frame.push(hrp.Nulo)
			}

		case OP_ARMAZENAR_LOCAL:
			slot := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			val := frame.pop()
			if int(slot) >= len(frame.Locais) {
				if int(slot) < cap(frame.Locais) {
					antigoLen := len(frame.Locais)
					frame.Locais = frame.Locais[:slot+1]
					for i := antigoLen; i < int(slot); i++ {
						frame.Locais[i] = nil
					}
				} else {
					novosLocais := make([]hrp.Objeto, slot+1, (slot+1)*2)
					copy(novosLocais, frame.Locais)
					frame.Locais = novosLocais
				}
			}
			frame.Locais[slot] = val
			if frame.Escopo != nil && frame.NomesLocais != nil {
				if nome, ok := frame.NomesLocais[slot]; ok {
					if frame.ConstantesLocais != nil && frame.ConstantesLocais[slot] {
						frame.Escopo.DefinirSimbolo(hrp.NewConstSimbolo(nome, val))
					} else {
						frame.Escopo.DefinirSimbolo(hrp.NewVarSimbolo(nome, val))
					}
				}
			}
			frame.push(val)

		case OP_CARREGAR_GLOBAL:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			nomeStr := string(frame.Consts[idx].(hrp.Texto))
			if frame.Escopo != nil {
				val, err := frame.Escopo.ObterValor(nomeStr)
				if err == nil {
					frame.push(val)
					break
				}
			}
			if v.Contexto != nil && v.Contexto.Modulos != nil && v.Contexto.Modulos.Embutidos != nil {
				if embObj, errEmb := v.Contexto.Modulos.Embutidos.M__obtem_attributo__(nomeStr); errEmb == nil {
					frame.push(embObj)
					break
				}
			}
			resExc, errExc := v.tratarExcecao(frame, hrp.NewErroF(hrp.NomeErro, "'%s' não foi encontrado no escopo atual", nomeStr))
			if errExc != nil {
				return nil, errExc
			}
			if resExc != nil {
				return resExc, nil
			}

		case OP_ARMAZENAR_GLOBAL:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			nomeStr := string(frame.Consts[idx].(hrp.Texto))
			val := frame.pop()
			if frame.Escopo != nil {
				err := frame.Escopo.RedefinirValor(nomeStr, val)
				if err != nil {
					if simb, errSimb := frame.Escopo.ObterSimbolo(nomeStr); errSimb == nil && simb.Constante {
						resExc, errExc := v.tratarExcecao(frame, err)
						if errExc != nil {
							return nil, errExc
						}
						if resExc != nil {
							return resExc, nil
						}
						continue
					}
					simbolo := hrp.NewVarSimbolo(nomeStr, val)
					if errDef := frame.Escopo.DefinirSimbolo(simbolo); errDef != nil {
						resExc, errExc := v.tratarExcecao(frame, errDef)
						if errExc != nil {
							return nil, errExc
						}
						if resExc != nil {
							return resExc, nil
						}
						continue
					}
				}
			}
			frame.push(val)

		case OP_CRIAR_FUNCAO:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			fnObj := frame.Consts[idx]
			frame.push(fnObj)

		case OP_CHAMAR:
			aridade := int(code[frame.IP])
			frame.IP++
			var args []hrp.Objeto
			if aridade > 0 {
				args = make([]hrp.Objeto, aridade)
				for i := aridade - 1; i >= 0; i-- {
					args[i] = frame.pop()
				}
			}
			fnObj := frame.pop()
			if fn, ok := fnObj.(*hrp.Funcao); ok {
				if fn.ObterContexto() == nil {
					fn.DefinirContexto(v.Contexto)
				}
				if fn.ObterEscopo() == nil {
					fn.DefinirEscopo(frame.Escopo)
				}
			}
			var res hrp.Objeto
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(*hrp.Erro); ok {
							err = e
						} else if e, ok := r.(error); ok {
							err = hrp.NewErro(hrp.RuntimeErro, hrp.Texto(e.Error()))
						} else {
							err = hrp.NewErroF(hrp.RuntimeErro, "%v", r)
						}
					}
				}()
				res, err = hrp.Chamar(fnObj, hrp.Tupla(args))
			}()
			for _, arg := range args {
				hrp.LiberarObjeto(arg)
			}
			hrp.LiberarObjeto(fnObj)
			if err != nil {
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)

		case OP_AWAIT:
			val := frame.pop()
			if prom, ok := val.(*hrp.Promessa); ok {
				ch := make(chan struct{})
				var res hrp.Objeto
				var resErr error
				prom.Registre(func(o hrp.Objeto, err error) {
					res = o
					resErr = err
					close(ch)
				})
				<-ch
				hrp.LiberarObjeto(val)
				if resErr != nil {
					resExc, errExc := v.tratarExcecao(frame, resErr)
					if errExc != nil {
						return nil, errExc
					}
					if resExc != nil {
						return resExc, nil
					}
					continue
				}
				frame.push(res)
			} else {
				frame.push(val)
			}

		case OP_RETORNE:
			if len(frame.Pilha) > 0 {
				return frame.pop(), nil
			}
			return hrp.Nulo, nil

		case OP_RETORNE_CONST:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			return frame.Consts[idx], nil

		case OP_RETORNE_LOCAL:
			slot := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			if int(slot) < len(frame.Locais) {
				return frame.Locais[slot], nil
			}
			return hrp.Nulo, nil

		case OP_CRIAR_LISTA:
			qtd := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			itens := make([]hrp.Objeto, qtd)
			for i := int(qtd) - 1; i >= 0; i-- {
				itens[i] = frame.pop()
			}
			lista := &hrp.Lista{Itens: itens}
			frame.push(lista)

		case OP_CRIAR_MAPA:
			qtdPares := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			mapa := make(hrp.Mapa, qtdPares)
			pares := make([][2]hrp.Objeto, qtdPares)
			for i := int(qtdPares) - 1; i >= 0; i-- {
				val := frame.pop()
				chave := frame.pop()
				pares[i] = [2]hrp.Objeto{chave, val}
			}
			for _, par := range pares {
				var chaveStr string
				if txt, ok := par[0].(hrp.Texto); ok {
					chaveStr = string(txt)
				} else {
					chaveStr = fmt.Sprintf("%v", par[0])
				}
				mapa[chaveStr] = par[1]
			}
			frame.push(mapa)

		case OP_OBTER_INDICE:
			idxObj := frame.pop()
			alvo := frame.pop()
			res, err := hrp.ObtemItem(alvo, idxObj)
			hrp.LiberarObjeto(alvo)
			hrp.LiberarObjeto(idxObj)
			if err != nil {
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)

		case OP_DEFINIR_INDICE:
			val := frame.pop()
			idxObj := frame.pop()
			alvo := frame.pop()
			res, err := hrp.DefineItem(alvo, idxObj, val)
			hrp.LiberarObjeto(alvo)
			hrp.LiberarObjeto(idxObj)
			if err != nil {
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)

		case OP_OBTER_ATRIBUTO:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			attrNome := string(frame.Consts[idx].(hrp.Texto))
			alvo := frame.pop()
			res, err := hrp.ObtemAtributoS(alvo, attrNome)
			hrp.LiberarObjeto(alvo)
			if err != nil {
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(res)

		case OP_DEFINIR_ATRIBUTO:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			attrNome := string(frame.Consts[idx].(hrp.Texto))
			val := frame.pop()
			alvo := frame.pop()
			err := hrp.DefineAtributo(alvo, attrNome, val)
			hrp.LiberarObjeto(alvo)
			if err != nil {
				resExc, errExc := v.tratarExcecao(frame, err)
				if errExc != nil {
					return nil, errExc
				}
				if resExc != nil {
					return resExc, nil
				}
				continue
			}
			frame.push(val)

		case OP_COALESCE:
			offset := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			topo := frame.Pilha[len(frame.Pilha)-1]
			if topo != hrp.Nulo {
				frame.IP += int(offset)
			} else {
				frame.pop()
			}

		case OP_ACESSAR_OPCIONAL:
			idx := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			attrNome := string(frame.Consts[idx].(hrp.Texto))
			alvo := frame.pop()
			if alvo == hrp.Nulo {
				frame.push(hrp.Nulo)
			} else {
				res, err := hrp.ObtemAtributoS(alvo, attrNome)
				hrp.LiberarObjeto(alvo)
				if err != nil {
					resExc, errExc := v.tratarExcecao(frame, err)
					if errExc != nil {
						return nil, errExc
					}
					if resExc != nil {
						return resExc, nil
					}
					continue
				}
				frame.push(res)
			}

		case OP_SETUP_TRY:
			catchIP := binary.BigEndian.Uint16(code[frame.IP:])
			frame.IP += 2
			frame.TryStack = append(frame.TryStack, TryHandler{
				CatchIP:   int(catchIP),
				PilhaProf: len(frame.Pilha),
			})

		case OP_POP_TRY:
			if len(frame.TryStack) > 0 {
				frame.TryStack = frame.TryStack[:len(frame.TryStack)-1]
			}

		case OP_LANCAR:
			errObj := frame.pop()
			resExc, errExc := v.tratarExcecao(frame, hrp.NewErroF(hrp.RuntimeErro, "%v", errObj))
			if errExc != nil {
				return nil, errExc
			}
			if resExc != nil {
				return resExc, nil
			}

		default:
			return nil, fmt.Errorf("instrução opcode '0x%X' desconhecida ou não suportada no runtime da VM", op)
		}

		if v.Perfil {
			v.registrarMetrica(op, time.Since(inicio))
		}
	}

	if len(frame.Pilha) > 0 {
		return frame.pop(), nil
	}
	return hrp.Nulo, nil
}

func (v *VM) enriquecerErro(frame *Frame, err error) error {
	if err == nil {
		return nil
	}
	hrpErr, ok := err.(*hrp.Erro)
	if !ok {
		return err
	}
	if hrpErr.Linha == -1 && frame != nil && frame.MapaTokens != nil {
		tok := frame.MapaTokens[frame.UltimoIP]
		if tok == nil {
			maiorOffset := -1
			for offset, t := range frame.MapaTokens {
				if offset <= frame.UltimoIP && offset > maiorOffset {
					maiorOffset = offset
					tok = t
				}
			}
		}
		if tok != nil {
			hrpErr.Token = tok
			if tok.Inicio.Linha > 0 {
				hrpErr.Linha = tok.Inicio.Linha - 1
				hrpErr.Coluna = tok.Inicio.Coluna
			}
		}
	}
	if frame != nil {
		if hrpErr.Arquivo == "" || hrpErr.Arquivo == "<desconhecido>" {
			if frame.Arquivo != "" {
				hrpErr.Arquivo = frame.Arquivo
			} else if v.Contexto != nil {
				arq, _ := v.Contexto.ObterArquivoCodigo()
				hrpErr.Arquivo = arq
			}
		}
		if hrpErr.Codigo == "" {
			if frame.Codigo != "" {
				hrpErr.Codigo = frame.Codigo
			} else if v.Contexto != nil {
				_, cod := v.Contexto.ObterArquivoCodigo()
				hrpErr.Codigo = cod
			}
		}
	} else if v.Contexto != nil {
		arq, cod := v.Contexto.ObterArquivoCodigo()
		if hrpErr.Arquivo == "" || hrpErr.Arquivo == "<desconhecido>" {
			hrpErr.Arquivo = arq
		}
		if hrpErr.Codigo == "" {
			hrpErr.Codigo = cod
		}
	}
	if v.Contexto != nil {
		hrpErr.AdicionarContexto(v.Contexto)
	}
	return hrpErr
}

func (v *VM) tratarExcecao(frame *Frame, err error) (hrp.Objeto, error) {
	errEnriquecido := v.enriquecerErro(frame, err)
	if len(frame.TryStack) > 0 {
		handler := frame.TryStack[len(frame.TryStack)-1]
		frame.TryStack = frame.TryStack[:len(frame.TryStack)-1]

		// Restaura pilha de operandos
		for len(frame.Pilha) > handler.PilhaProf {
			hrp.LiberarObjeto(frame.pop())
		}

		// Empilha o erro capturado e pula para o catch IP
		// CORE-10: Se o erro já for *hrp.Erro, preserva-o integralmente
		var erroObjeto hrp.Objeto
		if hrpErr, ok := errEnriquecido.(*hrp.Erro); ok {
			erroObjeto = hrpErr
		} else {
			erroObjeto = hrp.NewErroF(hrp.RuntimeErro, "%v", errEnriquecido)
		}
		if hrpErr, ok := erroObjeto.(*hrp.Erro); ok && v.Contexto != nil {
			hrpErr.AdicionarContexto(v.Contexto)
		}
		frame.push(erroObjeto)
		frame.IP = handler.CatchIP
		return nil, nil
	}
	return nil, errEnriquecido
}
