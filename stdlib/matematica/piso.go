package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/ptst"
)

// met_mat_piso implementa a lógica nativa para a função 'piso()'.
//
// Esta função recebe um número real, converte-o para o tipo Decimal, realiza o arredondamento
// para baixo (para o menor inteiro mais próximo) utilizando math.Floor do Go
// e retorna um tipo Inteiro nativo (ptst.Inteiro) da VM.
func met_mat_piso(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("piso", false, args, 1, 1); err != nil {
		return nil, err
	}

	num, err := ptst.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}

	return ptst.Inteiro(math.Floor(float64(num.(ptst.Decimal)))), nil
}

// _mat_piso cria e define a assinatura do método 'piso' exposto na stdlib do Harpia.
var _mat_piso = ptst.NewMetodoOuPanic(
	"piso",
	met_mat_piso,
	"piso(decimal) -> Retorna o numero arredondado para baixo",
)
