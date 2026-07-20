package tarefas

import (
	"fmt"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

type FilaItem struct {
	Funcao hrp.Objeto
	Dados  hrp.Objeto
}

var (
	fila      = make(chan FilaItem, 1000)
	workersMu = 0
)

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

	go func() {
		ticker := time.NewTicker(time.Duration(intervaloSecs) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_, err := hrp.Chamar(funcao, hrp.Tupla{})
			if err != nil {
				// Silencia erro para não derrubar o processo de background
				fmt.Printf("[Tarefas Background] Erro ao executar tarefa agendada: %v\n", err)
			}
		}
	}()

	return hrp.Nulo, nil
}

func met_tarefas_enfileirar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
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
	if err := hrp.VerificaNumeroArgumentos("processarFila", false, args, 1, 1); err != nil {
		return nil, err
	}
	workerFunc := args[0]

	go func() {
		for item := range fila {
			_, err := hrp.Chamar(workerFunc, hrp.Tupla{item.Funcao, item.Dados})
			if err != nil {
				fmt.Printf("[Tarefas Fila] Erro no worker da fila: %v\n", err)
			}
		}
	}()

	return hrp.Nulo, nil
}

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "tarefas",
			Arquivo: "stdlib/tarefas",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("agendar", met_tarefas_agendar, "Agenda a execução periódica de uma função."),
			hrp.NewMetodoOuPanic("enfileirar", met_tarefas_enfileirar, "Adiciona uma tarefa à fila em memória."),
			hrp.NewMetodoOuPanic("processarFila", met_tarefas_processarFila, "Inicia um worker em background para consumir a fila."),
		},
	})
}
