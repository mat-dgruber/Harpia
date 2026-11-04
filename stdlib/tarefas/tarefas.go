package tarefas

import (
	"fmt"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

type FilaItem struct {
	Funcao hrp.Objeto
	Dados  hrp.Objeto
}

var (
	fila            = make(chan FilaItem, 1000)
	tarefasSemCtxMu sync.Mutex
	tarefasSemCtx   []*TarefaAgendada
)

func pararTarefasSemCtx() {
	tarefasSemCtxMu.Lock()
	defer tarefasSemCtxMu.Unlock()
	for _, t := range tarefasSemCtx {
		t.Parar()
	}
	tarefasSemCtx = nil
}

// TarefaAgendada representa um processo periódico controlado que pode ser cancelado via canal de parada.
type TarefaAgendada struct {
	hrp.GCMixin
	parar chan struct{}
	once  sync.Once
}

var TipoTarefaAgendada = hrp.TipoObjeto.NewTipo("TarefaAgendada", "Instância de tarefa periódica agendada")

func (t *TarefaAgendada) Tipo() *hrp.Tipo {
	return TipoTarefaAgendada
}

func (t *TarefaAgendada) ObterFilhos() []hrp.Objeto {
	return nil
}

func (t *TarefaAgendada) Parar() {
	t.once.Do(func() {
		close(t.parar)
	})
}

var _ hrp.I__obtem_attributo__ = (*TarefaAgendada)(nil)

func (t *TarefaAgendada) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	if nome == "parar" || nome == "cancelar" {
		return hrp.NewMetodoOuPanic(nome, func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			t.Parar()
			return hrp.Nulo, nil
		}, "Encerra a tarefa periódica."), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe na TarefaAgendada", nome)
}

func met_tarefas_agendar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("agendar", false, args, 2, 2); err != nil {
		return nil, err
	}

	intervaloStr, _ := hrp.NewTexto(args[0])
	intervaloSecs := 1
	switch string(intervaloStr.(hrp.Texto)) {
	case "segundo", "* * * * * *":
		intervaloSecs = 1
	case "minuto", "0 * * * * *":
		intervaloSecs = 60
	default:
		// Se for um número (como string), usa como segundos
		var secs int
		_, err := fmt.Sscanf(string(intervaloStr.(hrp.Texto)), "%d", &secs)
		if err == nil && secs > 0 {
			intervaloSecs = secs
		}
	}

	funcao := args[1]
	tarefa := &TarefaAgendada{parar: make(chan struct{})}

	ctx := hrp.ObterContextoAtivo()
	var ctxCancel <-chan struct{}
	if ctx != nil {
		ctxCancel = ctx.CanalCancelamento()
	} else {
		tarefasSemCtxMu.Lock()
		tarefasSemCtx = append(tarefasSemCtx, tarefa)
		tarefasSemCtxMu.Unlock()
	}

	go func() {
		ticker := time.NewTicker(time.Duration(intervaloSecs) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-tarefa.parar:
				return
			case <-ctxCancel:
				return
			case <-ticker.C:
				if ctx != nil && ctx.Fechado() {
					return
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							tarefa.Parar()
						}
					}()
					_, err := hrp.Chamar(funcao, hrp.Tupla{})
					if err != nil {
						// Silencia erro para não derrubar o processo de background
						fmt.Printf("[Tarefas Background] Erro ao executar tarefa agendada: %v\n", err)
					}
				}()
			}
		}
	}()

	return tarefa, nil
}

func met_tarefas_parar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("parar", false, args, 1, 1); err != nil {
		return nil, err
	}
	if t, ok := args[0].(*TarefaAgendada); ok {
		t.Parar()
		return hrp.Verdadeiro, nil
	}
	return nil, hrp.NewErroF(hrp.TipagemErro, "esperava uma TarefaAgendada para parar")
}

func met_tarefas_enfileirar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if hrp.ObterContextoAtivo() == nil {
		pararTarefasSemCtx()
	}
	if err := hrp.VerificaNumeroArgumentos("enfileirar", false, args, 2, 2); err != nil {
		return nil, err
	}

	funcao := args[0]
	dados := args[1]

	select {
	case fila <- FilaItem{Funcao: funcao, Dados: dados}:
		return hrp.Verdadeiro, nil
	default:
		return hrp.Falso, nil
	}
}

func met_tarefas_processarFila(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if hrp.ObterContextoAtivo() == nil {
		pararTarefasSemCtx()
	}
	if err := hrp.VerificaNumeroArgumentos("processarFila", false, args, 1, 1); err != nil {
		return nil, err
	}
	workerFunc := args[0]

	ctx := hrp.ObterContextoAtivo()
	var ctxCancel <-chan struct{}
	if ctx != nil {
		ctxCancel = ctx.CanalCancelamento()
	}

	go func() {
		for {
			select {
			case <-ctxCancel:
				return
			case item, ok := <-fila:
				if !ok {
					return
				}
				if ctx != nil && ctx.Fechado() {
					return
				}
				_, err := hrp.Chamar(workerFunc, hrp.Tupla{item.Funcao, item.Dados})
				if err != nil {
					fmt.Printf("[Tarefas Fila] Erro no worker da fila: %v\n", err)
				}
			}
		}
	}()

	return hrp.Nulo, nil
}

func init() {
	TipoTarefaAgendada.Mapa["parar"] = hrp.NewMetodoOuPanic("parar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if t, ok := inst.(*TarefaAgendada); ok {
			t.Parar()
			return hrp.Nulo, nil
		}
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava TarefaAgendada")
	}, "Encerra a execução periódica da tarefa.")

	TipoTarefaAgendada.Mapa["cancelar"] = hrp.NewMetodoOuPanic("cancelar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if t, ok := inst.(*TarefaAgendada); ok {
			t.Parar()
			return hrp.Nulo, nil
		}
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava TarefaAgendada")
	}, "Encerra a execução periódica da tarefa.")

	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "tarefas",
			Arquivo: "stdlib/tarefas",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("agendar", met_tarefas_agendar, "Agenda a execução periódica de uma função."),
			hrp.NewMetodoOuPanic("parar", met_tarefas_parar, "Interrompe uma tarefa agendada."),
			hrp.NewMetodoOuPanic("enfileirar", met_tarefas_enfileirar, "Adiciona uma tarefa à fila em memória."),
			hrp.NewMetodoOuPanic("processarFila", met_tarefas_processarFila, "Inicia um worker em background para consumir a fila."),
		},
	})
}
