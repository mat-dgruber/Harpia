package fila

import (
	"fmt"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

type GerenciadorFila struct {
	mu     sync.Mutex
	tarefas []map[string]interface{}
}

var globalFila = &GerenciadorFila{
	tarefas: make([]map[string]interface{}, 0),
}

func met_fila_enfileirar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("enfileirar", false, args, 2, 2); err != nil {
		return nil, err
	}

	nomeFila, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	dadosMapa, ok := args[1].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipoErro, "O segundo argumento de 'enfileirar' deve ser um Mapa")
	}

	dados := make(map[string]interface{})
	for k, v := range dadosMapa {
		dados[k] = fmt.Sprintf("%v", v)
	}

	globalFila.mu.Lock()
	globalFila.tarefas = append(globalFila.tarefas, map[string]interface{}{
		"fila":  string(nomeFila.(hrp.Texto)),
		"dados": dados,
		"criadoEm": time.Now(),
	})
	globalFila.mu.Unlock()

	return hrp.Booleano(true), nil
}

func met_fila_tamanho(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	globalFila.mu.Lock()
	t := len(globalFila.tarefas)
	globalFila.mu.Unlock()
	return hrp.Inteiro(t), nil
}

var _enfileirar = hrp.NewMetodoOuPanic("enfileirar", met_fila_enfileirar, "")
var _tamanho = hrp.NewMetodoOuPanic("tamanho", met_fila_tamanho, "")

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "fila",
			Arquivo: "stdlib/fila",
		},
		Metodos: []*hrp.Metodo{
			_enfileirar,
			_tamanho,
		},
	})
}
