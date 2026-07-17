package vm

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
	"github.com/mat-dgruber/Harpia/ptst"
)

var poolPilha = sync.Pool{
	New: func() interface{} {
		return make([]ptst.Objeto, 0, 128)
	},
}

type InstrucaoThreaded func(v *VM, frame *Frame) (ptst.Objeto, error)

// Frame representa um contexto isolado de execução de função ou módulo na pilha de chamadas da VM.
type Frame struct {
	Pai          *Frame
	Bytecode     []byte
	IP           int                 // Instrução corrente (Instruction Pointer)
	Consts       []ptst.Objeto       // Referência ao pool de constantes do programa compilado
	Pilha        []ptst.Objeto       // Pilha de operandos local
	Escopo       *ptst.Escopo        // Tabela local de símbolos e fechamento léxico
	ThreadedCode []InstrucaoThreaded // JIT de Traço (Fase F)
}

// NewFrame cria um frame isolado apontando para as constantes e escopo fornecidos.
func NewFrame(bytecode []byte, consts []ptst.Objeto, escopo *ptst.Escopo, pai *Frame) *Frame {
	pilha := poolPilha.Get().([]ptst.Objeto)
	return &Frame{
		Pai:      pai,
		Bytecode: bytecode,
		IP:       0,
		Consts:   consts,
		Pilha:    pilha[:0],
		Escopo:   escopo,
	}
}

// push adiciona o elemento e incrementa suas referências.
func (f *Frame) push(obj ptst.Objeto) {
	f.Pilha = append(f.Pilha, obj)
	ptst.ReterObjeto(obj)
}

// pop remove o elemento sem decrementar (transfere a posse do objeto para o receptor).
func (f *Frame) pop() ptst.Objeto {
	if len(f.Pilha) == 0 {
		return ptst.Nulo
	}
	topoIdx := len(f.Pilha) - 1
	val := f.Pilha[topoIdx]
	f.Pilha = f.Pilha[:topoIdx]
	return val
}

// ObterProfundidade calcula recursivamente a profundidade do frame corrente.
func (f *Frame) ObterProfundidade() int {
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
	Contexto *ptst.Contexto
	Perfil   bool
	Metricas map[Opcode]*PerfilInfo
}

// NewVM instancia a máquina virtual associada a um contexto de execução de tipos e módulos.
func NewVM(ctx *ptst.Contexto) *VM {
	return &VM{
		Contexto: ctx,
		Metricas: make(map[Opcode]*PerfilInfo),
	}
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

// Executar inicia o loop de threaded callbacks JIT sobre o frame fornecido.
func (v *VM) Executar(frame *Frame) (ptst.Objeto, error) {
	if frame.ObterProfundidade() > 1000 {
		return nil, ptst.NewErroF(ptst.ErroDePilha, "Limite máximo de recursão excedido (1000 frames)")
	}

	defer func() {
		// Limpeza de fim de frame: libera todos os operandos remanescentes na pilha
		for len(frame.Pilha) > 0 {
			ptst.LiberarObjeto(frame.pop())
		}

		// Limpa a fatia para evitar vazamentos de referências a objetos Go antes de devolver ao pool
		fatiaLimpa := frame.Pilha[:cap(frame.Pilha)]
		for i := range fatiaLimpa {
			fatiaLimpa[i] = nil
		}
		poolPilha.Put(frame.Pilha[:0])

		// Limpeza de escopo: realiza a varredura cíclica e libera as referências retidas pelo escopo local
		if frame.Escopo != nil {
			ptst.ColetarCiclos(frame.Escopo)

			for _, simb := range frame.Escopo.ObterSimbolosSeguro() {
				if simb != nil {
					ptst.LiberarObjeto(simb.ObterValor())
				}
			}
		}
	}()

	// Ativação dinâmica do JIT de Traço (Threaded Code Execution - Fase F)
	if frame.ThreadedCode == nil {
		frame.ThreadedCode = v.compilarThreadedCode(frame)
	}

	for frame.IP < len(frame.ThreadedCode) {
		inst := frame.ThreadedCode[frame.IP]
		if inst == nil {
			frame.IP++
			continue
		}

		opIdx := frame.IP
		frame.IP++

		var res ptst.Objeto
		var err error

		if v.Perfil {
			op := frame.Bytecode[opIdx]
			inicio := time.Now()
			res, err = inst(v, frame)
			v.registrarMetrica(op, time.Since(inicio))
		} else {
			res, err = inst(v, frame)
		}

		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}

	if len(frame.Pilha) > 0 {
		return frame.pop(), nil
	}
	return ptst.Nulo, nil
}

// compilarThreadedCode compila em tempo de execução o bytecode plano de 1 byte em fatias de ponteiros de funções (callbacks JIT).
func (v *VM) compilarThreadedCode(frame *Frame) []InstrucaoThreaded {
	code := frame.Bytecode
	threaded := make([]InstrucaoThreaded, len(code))
	ip := 0

	for ip < len(code) {
		op := code[ip]
		currentIP := ip
		ip++

		switch op {
		case OP_PUSH_CONST:
			idx := code[ip]
			ip++
			val := frame.Consts[idx]
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				frame.push(val)
				return nil, nil
			}

		case OP_POP:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				val := frame.pop()
				ptst.LiberarObjeto(val)
				return nil, nil
			}

		case OP_DUP:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				val := frame.pop()
				frame.push(val)
				frame.push(val)
				ptst.LiberarObjeto(val)
				return nil, nil
			}

		case OP_ADD:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Adiciona(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_SUB:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Subtrai(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_MUL:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Multiplica(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_DIV:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Divide(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_DIV_INT:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.DivideInteiro(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_MOD:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Mod(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_EQ:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Igual(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_NEQ:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.Diferente(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_LT:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.MenorQue(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_LTE:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.MenorOuIgual(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_GT:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.MaiorQue(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_GTE:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				b := frame.pop()
				a := frame.pop()
				res, err := ptst.MaiorOuIgual(a, b)
				if err != nil {
					ptst.LiberarObjeto(a)
					ptst.LiberarObjeto(b)
					return nil, err
				}
				frame.push(res)
				ptst.LiberarObjeto(a)
				ptst.LiberarObjeto(b)
				return nil, nil
			}

		case OP_JMP:
			addr := binary.BigEndian.Uint16(code[ip : ip+2])
			ip += 2
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				frame.IP = int(addr)
				return nil, nil
			}

		case OP_JMP_FALSO:
			addr := binary.BigEndian.Uint16(code[ip : ip+2])
			ip += 2
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				val := frame.pop()
				if val == ptst.Falso || val == ptst.Nulo {
					frame.IP = int(addr)
				}
				ptst.LiberarObjeto(val)
				return nil, nil
			}

		case OP_CARREGAR_VAR:
			idx := code[ip]
			ip++
			nome := string(frame.Consts[idx].(ptst.Texto))

			// Monomorphic Inline Cache (MIC)
			var cacheEscopo *ptst.Escopo = nil
			var cacheSimbolo *ptst.Simbolo = nil

			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				// Cache Hit: se o escopo for o mesmo, retorna o valor direto
				if frame.Escopo == cacheEscopo && cacheSimbolo != nil {
					frame.push(cacheSimbolo.Valor)
					return nil, nil
				}

				// Cache Miss: realiza a busca léxica
				simb, err := frame.Escopo.ObterSimbolo(nome)
				if err == nil && simb != nil {
					cacheEscopo = frame.Escopo
					cacheSimbolo = simb
					frame.push(simb.Valor)
					return nil, nil
				}

				// Fallback de embutidos
				if val, errEmb := v.Contexto.Modulos.Embutidos.M__obtem_attributo__(nome); errEmb == nil {
					frame.push(val)
					return nil, nil
				}

				return nil, err
			}

		case OP_ARMAZENAR_VAR:
			idx := code[ip]
			ip++
			nome := string(frame.Consts[idx].(ptst.Texto))
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				valor := frame.pop()
				simb, errS := frame.Escopo.ObterSimbolo(nome)
				if errS == nil && simb != nil {
					ptst.LiberarObjeto(simb.Valor)
					simb.Valor = valor
					ptst.ReterObjeto(valor)
				} else {
					simbolo := ptst.NewVarSimbolo(nome, valor)
					ptst.ReterObjeto(valor)
					if errDef := frame.Escopo.DefinirSimbolo(simbolo); errDef != nil {
						ptst.LiberarObjeto(valor)
						return nil, errDef
					}
				}
				ptst.LiberarObjeto(valor)
				return nil, nil
			}

		case OP_CHAMAR:
			aridade := code[ip]
			ip++
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				objeto := frame.pop()
				args := make(ptst.Tupla, aridade)
				for i := int(aridade) - 1; i >= 0; i-- {
					args[i] = frame.pop()
				}

				// Se o chamável for uma função assíncrona, interceptamos e retornamos uma Promessa
				if fn, ok := objeto.(*ptst.Funcao); ok {
					fn.SetContexto(v.Contexto)
					fn.SetEscopo(frame.Escopo)
					if fn.Assincrono {
						prom := ptst.NewPromessa()
						if v.Contexto != nil {
							v.Contexto.AdicionarTrabalho()
						}
						go func() {
							defer func() {
								if v.Contexto != nil {
									v.Contexto.EncerrarTrabalho()
								}
							}()
							res, err := fn.M__chame__(args)
							if err != nil {
								prom.Rejeitar(err)
							} else {
								prom.Resolver(res)
							}
							for _, arg := range args {
								ptst.LiberarObjeto(arg)
							}
							ptst.LiberarObjeto(objeto)
						}()
						frame.push(prom)
						return nil, nil
					}
				}

				res, err := ptst.Chamar(objeto, args)
				if err != nil {
					ptst.LiberarObjeto(objeto)
					for _, arg := range args {
						ptst.LiberarObjeto(arg)
					}
					return nil, err
				}

				frame.push(res)
				ptst.LiberarObjeto(objeto)
				for _, arg := range args {
					ptst.LiberarObjeto(arg)
				}
				return nil, nil
			}

		case OP_RETORNE:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				if len(frame.Pilha) > 0 {
					val := frame.pop()
					return val, nil
				}
				return ptst.Nulo, nil
			}

		case OP_RETORNE_CONST:
			idx := code[ip]
			ip++
			val := frame.Consts[idx]
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				return val, nil
			}

		case OP_RETORNE_VAR:
			idx := code[ip]
			ip++
			nome := string(frame.Consts[idx].(ptst.Texto))
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				val, err := frame.Escopo.ObterValor(nome)
				if err != nil {
					return nil, err
				}
				return val, nil
			}

		case OP_AWAIT:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				val := frame.pop()
				prom, ok := val.(*ptst.Promessa)
				if !ok {
					frame.push(val)
					ptst.LiberarObjeto(val)
					return nil, nil
				}

				channel := make(chan ptst.Objeto, 1)
				var errProm error
				prom.Registre(func(res ptst.Objeto, err error) {
					if err != nil {
						errProm = err
						channel <- nil
					} else {
						channel <- res
					}
				})

				res := <-channel
				if errProm != nil {
					ptst.LiberarObjeto(val)
					return nil, errProm
				}

				frame.push(res)
				ptst.LiberarObjeto(val)
				return nil, nil
			}

		case OP_CRIAR_FUNCAO:
			idxNome := code[ip]
			ip++
			nome := string(frame.Consts[idxNome].(ptst.Texto))
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				funcao := ptst.NewFuncao(nome, nil, v.Contexto, frame.Escopo)
				frame.push(funcao)
				return nil, nil
			}

		default:
			threaded[currentIP] = func(v *VM, frame *Frame) (ptst.Objeto, error) {
				return nil, fmt.Errorf("instrução opcode '0x%X' desconhecida ou não suportada no runtime da VM", op)
			}
		}
	}

	return threaded
}
