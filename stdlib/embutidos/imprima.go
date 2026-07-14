package embutidos

import (
	"fmt"

	"github.com/natanfeitosa/portuscript/ptst"
)

// met_emb_imprima implementa a lógica nativa para a função global 'imprima()'.
//
// Esta função recebe um número dinâmico (variádico) de argumentos de qualquer tipo,
// concatena-os separando por espaço de forma a criar uma representação visual única de string,
// e imprime na saída padrão do terminal (Stdout) acompanhada de uma quebra de linha final ("\n").
//
// Mecânica de Concatenação:
// Ela obtém o método embutido de junta ("junta") de um objeto Texto (" ") e chama-o passando
// a tupla de argumentos recebidos. Isso delega a lógica de concatenação para as regras otimizadas de strings.
func met_emb_imprima(mod ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	const (
		final     = ptst.Texto("\n")
		separador = ptst.Texto(" ")
	)

	junta, err := ptst.ObtemAtributoS(separador, "junta")

	if err != nil {
		return nil, err
	}

	resultado, err := ptst.Chamar(
		junta,
		args,
	)

	if err != nil {
		return nil, err
	}

	fmt.Printf("%s%s", resultado, final)
	return nil, nil
}

// _emb_imprima cria e define a assinatura do método 'imprima' exposto globalmente.
var _emb_imprima = ptst.NewMetodoOuPanic(
	"imprima",
	met_emb_imprima,
	"imprima(...objetos) -> imprime a representação ou a conversão em string dos objetos separados por espaço",
)
