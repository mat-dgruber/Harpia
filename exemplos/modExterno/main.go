package main

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/hrp"
)

// InicializaModulo é a porta de entrada obrigatória e o símbolo público exportado
// que a Máquina Virtual do Harpia resolve e executa através de reflexão dinâmica de plugins (.so).
//
// Esta função deve retornar o ponteiro para a especificação estática do módulo (*hrp.ModuloImpl),
// declarando o nome do módulo, documentações explicativas de auxílio (Doc) e as assinaturas de seus métodos.
func InicializaModulo() *hrp.ModuloImpl {
	return &hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome: "externos",
			Doc:  "Um módulo de extensão nativa externo compilado em Go para teste",
		},
		Metodos: []*hrp.Metodo{
			// Define a função chamável 'exiba' no escopo do módulo
			hrp.NewMetodoOuPanic("exiba", func(_ hrp.Objeto, args hrp.Tupla) (obj hrp.Objeto, err error) {
				junta, err := hrp.ObtemAtributoS(hrp.Texto(", "), "junta")
				if err != nil {
					return
				}

				juntos, err := hrp.Chamar(junta, args)
				if err != nil {
					return
				}

				fmt.Printf("externos: %s", juntos.(hrp.Texto))
				return
			}, "Exibe algo no terminal com prefixo personalizado, ok?"),
		},
	}
}
