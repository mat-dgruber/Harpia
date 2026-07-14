package main

import (
	"fmt"

	"github.com/natanfeitosa/portuscript/ptst"
)

// InicializaModulo é a porta de entrada obrigatória e o símbolo público exportado
// que a Máquina Virtual do Portuscript resolve e executa através de reflexão dinâmica de plugins (.so).
//
// Esta função deve retornar o ponteiro para a especificação estática do módulo (*ptst.ModuloImpl),
// declarando o nome do módulo, documentações explicativas de auxílio (Doc) e as assinaturas de seus métodos.
func InicializaModulo() *ptst.ModuloImpl {
	return &ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome: "externos",
			Doc:  "Um módulo de extensão nativa externo compilado em Go para teste",
		},
		Metodos: []*ptst.Metodo{
			// Define a função chamável 'exiba' no escopo do módulo
			ptst.NewMetodoOuPanic("exiba", func(_ ptst.Objeto, args ptst.Tupla) (obj ptst.Objeto, err error) {
				junta, err := ptst.ObtemAtributoS(ptst.Texto(", "), "junta")
				if err != nil {
					return
				}

				juntos, err := ptst.Chamar(junta, args)
				if err != nil {
					return
				}

				fmt.Printf("externos: %s", juntos.(ptst.Texto))
				return
			}, "Exibe algo no terminal com prefixo personalizado, ok?"),
		},
	}
}
