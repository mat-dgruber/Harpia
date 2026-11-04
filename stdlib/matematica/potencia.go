package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_mat_potencia implementa a lógica nativa para a função 'potencia()'.
//
// Esta função recebe uma base e um expoente, valida se a quantidade de parâmetros está correta,
// converte ambos os operandos para Decimal (hrp.Decimal)
// e realiza a exponenciação real (base ^ expoente) por meio de math.Pow do Go.
func met_mat_potencia(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("potencia", false, args, 2, 2); err != nil {
		return nil, err
	}

	base, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}

	expoente, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}

	potencia := math.Pow(float64(base.(hrp.Decimal)), float64(expoente.(hrp.Decimal)))
	return hrp.Decimal(potencia), nil
}

// _mat_potencia cria e define a assinatura do método 'potencia' exposto na stdlib do Harpia.
var _mat_potencia = hrp.NewMetodoOuPanic(
	"potencia",
	met_mat_potencia,
	"potencia(base, expoente) -> Retorna a potencia de base ^ expoente",
)
