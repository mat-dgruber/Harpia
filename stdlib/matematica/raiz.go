package matematica

import "github.com/natanfeitosa/portuscript/ptst"

// met_mat_raiz implementa a lógica nativa para a função 'raiz()'.
//
// Esta função recebe um radicando e opcionalmente um índice de raiz. Se o índice for omitido,
// assume o valor padrão de 2.0 (calculando a raiz quadrada).
//
// Mecânica de Cálculo:
// Ela calcula a raiz de forma puramente aritmética convertendo a operação para potenciação fracionária.
// Ou seja, calcular a raiz N de X é equivalente a calcular X elevado à potência (1 / N).
// Para realizar o cálculo, ela delega o processamento para a função nativa 'met_mat_potencia'
// passando uma tupla contendo o radicando e o expoente fracionário calculado (1.0 / indice).
func met_mat_raiz(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("raiz", false, args, 1, 2); err != nil {
		return nil, err
	}

	indice := ptst.Decimal(2.0)

	if len(args) > 1 {
		dec, err := ptst.NewDecimal(args[1])
		if err != nil {
			return nil, err
		}

		indice = dec.(ptst.Decimal)
	}

	return met_mat_potencia(inst, ptst.Tupla{args[0], 1.0/indice})
}

// _mat_raiz cria e define a assinatura do método 'raiz' exposto na stdlib do Portuscript.
var _mat_raiz = ptst.NewMetodoOuPanic(
	"raiz",
	met_mat_raiz,
	"raiz(radicando, indice?) -> Retorna a raiz de radicando por indice. Se indice não for definido, o padrão é 2 (raiz quadrada do radicando)",
)
