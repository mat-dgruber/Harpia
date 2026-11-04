package matematica

import (
	"math"
	"math/rand"

	"github.com/mat-dgruber/Harpia/hrp"
)

func met_mat_max(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("max", false, args, 2, 2); err != nil {
		return nil, err
	}
	a, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	b, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}
	return hrp.Decimal(math.Max(float64(a.(hrp.Decimal)), float64(b.(hrp.Decimal)))), nil
}

func met_mat_min(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("min", false, args, 2, 2); err != nil {
		return nil, err
	}
	a, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	b, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}
	return hrp.Decimal(math.Min(float64(a.(hrp.Decimal)), float64(b.(hrp.Decimal)))), nil
}

func met_mat_aleatorio(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if len(args) == 0 {
		return hrp.Decimal(rand.Float64()), nil
	}
	if len(args) == 2 {
		minVal, err := hrp.NewDecimal(args[0])
		if err != nil {
			return nil, err
		}
		maxVal, err := hrp.NewDecimal(args[1])
		if err != nil {
			return nil, err
		}
		minF := float64(minVal.(hrp.Decimal))
		maxF := float64(maxVal.(hrp.Decimal))
		return hrp.Decimal(minF + rand.Float64()*(maxF-minF)), nil
	}
	return nil, hrp.NewErroF(hrp.SintaxeErro, "aleatorio aceita 0 ou 2 argumentos: aleatorio() ou aleatorio(min, max)")
}

var _mat_max = hrp.NewMetodoOuPanic("max", met_mat_max, "max(a, b) -> Retorna o maior valor entre a e b")
var _mat_min = hrp.NewMetodoOuPanic("min", met_mat_min, "min(a, b) -> Retorna o menor valor entre a e b")
var _mat_aleatorio = hrp.NewMetodoOuPanic("aleatorio", met_mat_aleatorio, "aleatorio([min, max]) -> Retorna numero decimal aleatorio entre 0 e 1, ou entre min e max")
