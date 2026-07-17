package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/ptst"
)

// met_mat_teto implementa a lógica nativa para a função 'teto()'.
//
// Esta função recebe um número real, valida os argumentos, converte-o para o tipo Decimal,
// realiza o arredondamento para cima (para o menor inteiro maior ou igual) utilizando math.Ceil do Go
// e retorna um tipo Inteiro nativo (ptst.Inteiro) da VM.
func met_mat_teto(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("teto", false, args, 1, 1); err != nil {
		return nil, err
	}

	num, err := ptst.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}

	return ptst.Inteiro(math.Ceil(float64(num.(ptst.Decimal)))), nil
}

// _mat_teto cria e define a assinatura do método 'teto' exposto na stdlib do Harpia.
var _mat_teto = ptst.NewMetodoOuPanic(
	"teto",
	met_mat_teto,
	"teto(decimal) -> Retorna o numero arredondado para cima",
)
