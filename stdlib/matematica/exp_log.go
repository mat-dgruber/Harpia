package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

func met_mat_exp(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("exp", false, args, 1, 1); err != nil {
		return nil, err
	}
	val, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	return hrp.Decimal(math.Exp(float64(val.(hrp.Decimal)))), nil
}

func met_mat_log(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("log", false, args, 1, 1); err != nil {
		return nil, err
	}
	val, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	f := float64(val.(hrp.Decimal))
	if f <= 0 {
		return nil, hrp.NewErroF(hrp.ValorErro, "Logaritmo indefinido para valores <= 0")
	}
	return hrp.Decimal(math.Log(f)), nil
}

var _mat_exp = hrp.NewMetodoOuPanic("exp", met_mat_exp, "exp(x) -> Retorna e^x")
var _mat_log = hrp.NewMetodoOuPanic("log", met_mat_log, "log(x) -> Retorna o logaritmo natural (ln) de x")
