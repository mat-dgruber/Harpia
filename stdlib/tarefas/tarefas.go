package tarefas

import (
	"fmt"
	"time"

	"github.com/natanfeitosa/portuscript/ptst"
)

type FilaItem struct {
	Funcao ptst.Objeto
	Dados  ptst.Objeto
}

var (
	fila      = make(chan FilaItem, 1000)
	workersMu = 0
)

func met_tarefas_agendar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("agendar", false, args, 2, 2); err != nil {
		return nil, err
	}

	intervaloStr, _ := ptst.NewTexto(args[0])
	intervaloSecs := 1
	switch string(intervaloStr.(ptst.Texto)) {
	case "segundo", "* * * * * *":
		intervaloSecs = 1
	case "minuto", "0 * * * * *":
		intervaloSecs = 60
	default:
		// Se for um número (como string), usa como segundos
		var secs int
		_, err := fmt.Sscanf(string(intervaloStr.(ptst.Texto)), "%d", &secs)
		if err == nil && secs > 0 {
			intervaloSecs = secs
		}
	}

	funcao := args[1]

	go func() {
		ticker := time.NewTicker(time.Duration(intervaloSecs) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_, err := ptst.Chamar(funcao, ptst.Tupla{})
			if err != nil {
				// Silencia erro para não derrubar o processo de background
				fmt.Printf("[Tarefas Background] Erro ao executar tarefa agendada: %v\n", err)
			}
		}
	}()

	return ptst.Nulo, nil
}

func met_tarefas_enfileirar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("enfileirar", false, args, 2, 2); err != nil {
		return nil, err
	}

	funcao := args[0]
	dados := args[1]

	select {
	case fila <- FilaItem{Funcao: funcao, Dados: dados}:
		return ptst.Verdadeiro, nil
	default:
		return ptst.Falso, nil
	}
}

func met_tarefas_processarFila(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("processarFila", false, args, 1, 1); err != nil {
		return nil, err
	}
	workerFunc := args[0]

	go func() {
		for item := range fila {
			_, err := ptst.Chamar(workerFunc, ptst.Tupla{item.Funcao, item.Dados})
			if err != nil {
				fmt.Printf("[Tarefas Fila] Erro no worker da fila: %v\n", err)
			}
		}
	}()

	return ptst.Nulo, nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "tarefas",
			Arquivo: "stdlib/tarefas",
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("agendar", met_tarefas_agendar, "Agenda a execução periódica de uma função."),
			ptst.NewMetodoOuPanic("enfileirar", met_tarefas_enfileirar, "Adiciona uma tarefa à fila em memória."),
			ptst.NewMetodoOuPanic("processarFila", met_tarefas_processarFila, "Inicia um worker em background para consumir a fila."),
		},
	})
}
