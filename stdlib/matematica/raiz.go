package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_mat_raiz implementa a lógica nativa para a função 'raiz()'.
//
// Esta função recebe um radicando e opcionalmente um índice de raiz. Se o índice for omitido,
// assume o valor padrão de 2.0 (calculando a raiz quadrada).
func met_mat_raiz(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("raiz", false, args, 1, 2); err != nil {
		return nil, err
	}

	radObj, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	rad := float64(radObj.(hrp.Decimal))

	ind := 2.0
	if len(args) > 1 {
		dec, err := hrp.NewDecimal(args[1])
		if err != nil {
			return nil, err
		}
		ind = float64(dec.(hrp.Decimal))
	}

	if ind == 0 {
		return nil, hrp.NewErroF(hrp.ValorErro, "índice da raiz não pode ser zero")
	}

	if rad < 0 {
		// ponytail: raiz de negativo só é definida no domínio real para índices inteiros ímpares
		isOddInt := math.Trunc(ind) == ind && int64(ind)%2 != 0
		if !isOddInt {
			return nil, hrp.NewErroF(hrp.ValorErro, "raiz de número negativo indefinida para índice não ímpar: %v", ind)
		}
		return hrp.Decimal(-math.Pow(-rad, 1.0/ind)), nil
	}

	return hrp.Decimal(math.Pow(rad, 1.0/ind)), nil
}

// _mat_raiz cria e define a assinatura do método 'raiz' exposto na stdlib do Harpia.
var _mat_raiz = hrp.NewMetodoOuPanic(
	"raiz",
	met_mat_raiz,
	"raiz(radicando, indice?) -> Retorna a raiz de radicando por indice. Se indice não for definido, o padrão é 2 (raiz quadrada do radicando)",
)
