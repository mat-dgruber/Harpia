package tarefas

import (
	"sync"
	"testing"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestAgendarTarefa(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	executou := false
	funcao := hrp.NewMetodoOuPanic("callback", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		executou = true
		wg.Done()
		return hrp.Nulo, nil
	}, "")

	// Agenda a cada 1 segundo
	_, err := met_tarefas_agendar(nil, hrp.Tupla{hrp.Texto("1"), funcao})
	if err != nil {
		t.Fatalf("Erro ao agendar: %v", err)
	}

	// Aguarda execução
	canalWG := make(chan struct{})
	go func() {
		wg.Wait()
		close(canalWG)
	}()

	select {
	case <-canalWG:
		// Sucesso
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout: tarefa agendada não foi executada")
	}

	if !executou {
		t.Errorf("Esperava que a tarefa tivesse executado")
	}
}

func TestFilaDeTarefas(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	executou := false
	var dadosRecebidos hrp.Objeto

	workerFunc := hrp.NewMetodoOuPanic("worker", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		// Args[0] é a função a executar, Args[1] são os dados
		f := args[0]
		d := args[1]
		_, err := hrp.Chamar(f, hrp.Tupla{})
		if err != nil {
			t.Errorf("Erro ao executar função enfileirada: %v", err)
		}
		dadosRecebidos = d
		executou = true
		wg.Done()
		return hrp.Nulo, nil
	}, "")

	// Inicia processador de fila
	_, errProc := met_tarefas_processarFila(nil, hrp.Tupla{workerFunc})
	if errProc != nil {
		t.Fatalf("Erro ao iniciar processador: %v", errProc)
	}

	funcTarefa := hrp.NewMetodoOuPanic("tarefa", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		return hrp.Nulo, nil
	}, "")

	// Enfileira
	res, errEnf := met_tarefas_enfileirar(nil, hrp.Tupla{funcTarefa, hrp.Texto("dados_da_fila")})
	if errEnf != nil || res != hrp.Verdadeiro {
		t.Fatalf("Erro ao enfileirar: %v, res: %v", errEnf, res)
	}

	// Aguarda processamento
	canalWG := make(chan struct{})
	go func() {
		wg.Wait()
		close(canalWG)
	}()

	select {
	case <-canalWG:
		// Sucesso
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout: tarefa da fila não foi processada")
	}

	if !executou {
		t.Errorf("Esperava que o worker tivesse processado")
	}

	if string(dadosRecebidos.(hrp.Texto)) != "dados_da_fila" {
		t.Errorf("Dados incorretos recebidos na fila: %v", dadosRecebidos)
	}
}
