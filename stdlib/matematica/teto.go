package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_mat_teto implementa a lógica nativa para a função 'teto()'.
//
// Esta função recebe um número real, valida os argumentos, converte-o para o tipo Decimal,
// realiza o arredondamento para cima (para o menor inteiro maior ou igual) utilizando math.Ceil do Go
// e retorna um tipo Inteiro nativo (hrp.Inteiro) da VM.
func met_mat_teto(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("teto", false, args, 1, 1); err != nil {
		return nil, err
	}

	num, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}

	f := float64(num.(hrp.Decimal))
	if math.IsNaN(f) || math.IsInf(f, 0) || f < math.MinInt64 || f > math.MaxInt64 {
		return nil, hrp.NewErroF(hrp.ValorErro, "valor inválido para teto: %v", f)
	}

	return hrp.Inteiro(math.Ceil(f)), nil
}

// _mat_teto cria e define a assinatura do método 'teto' exposto na stdlib do Harpia.
var _mat_teto = hrp.NewMetodoOuPanic(
	"teto",
	met_mat_teto,
	"teto(decimal) -> Retorna o numero arredondado para cima",
)
