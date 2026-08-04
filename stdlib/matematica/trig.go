package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

func met_mat_sen(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("sen", false, args, 1, 1); err != nil {
		return nil, err
	}
	val, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	return hrp.Decimal(math.Sin(float64(val.(hrp.Decimal)))), nil
}

func met_mat_cos(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("cos", false, args, 1, 1); err != nil {
		return nil, err
	}
	val, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	return hrp.Decimal(math.Cos(float64(val.(hrp.Decimal)))), nil
}

func met_mat_tanh(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("tanh", false, args, 1, 1); err != nil {
		return nil, err
	}
	val, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	return hrp.Decimal(math.Tanh(float64(val.(hrp.Decimal)))), nil
}

var _mat_sen = hrp.NewMetodoOuPanic("sen", met_mat_sen, "sen(x) -> Retorna o seno de x (em radianos)")
var _mat_cos = hrp.NewMetodoOuPanic("cos", met_mat_cos, "cos(x) -> Retorna o cosseno de x (em radianos)")
var _mat_tanh = hrp.NewMetodoOuPanic("tanh", met_mat_tanh, "tanh(x) -> Retorna a tangente hiperbolica de x")
