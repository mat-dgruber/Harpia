package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/ptst"
)

// met_mat_potencia implementa a lógica nativa para a função 'potencia()'.
//
// Esta função recebe uma base e um expoente, valida se a quantidade de parâmetros está correta,
// converte ambos os operandos para Decimal (ptst.Decimal)
// e realiza a exponenciação real (base ^ expoente) por meio de math.Pow do Go.
func met_mat_potencia(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("potencia", false, args, 2, 2); err != nil {
		return nil, err
	}

	var base, expoente ptst.Objeto
	expoente = ptst.Decimal(2.0)
	base = args[0]

	if len(args) > 1 {
		expoente = args[1]
	}

	var err error
	if base, err = ptst.NewDecimal(base); err != nil {
		return nil, err
	}

	if expoente, err = ptst.NewDecimal(expoente); err != nil {
		return nil, err
	}

	potencia := math.Pow(float64(base.(ptst.Decimal)), float64(expoente.(ptst.Decimal)))
	return ptst.Decimal(potencia), nil
}

// _mat_potencia cria e define a assinatura do método 'potencia' exposto na stdlib do Harpia.
var _mat_potencia = ptst.NewMetodoOuPanic(
	"potencia",
	met_mat_potencia,
	"potencia(base, expoente) -> Retorna a potencia de base ^ expoente",
)
