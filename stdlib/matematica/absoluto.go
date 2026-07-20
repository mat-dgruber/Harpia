package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_mat_absoluto implementa a lógica nativa para a função 'absoluto()'.
//
// Esta função recebe um único argumento numérico, valida se a quantidade de parâmetros está correta,
// converte o argumento recebido para o formato Decimal nativo da VM (hrp.Decimal)
// e calcula o valor absoluto (magnitude sem sinal) correspondente por meio de math.Abs do Go.
func met_mat_absoluto(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("absoluto", false, args, 1, 1); err != nil {
		return nil, err
	}

	numero, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}

	return hrp.Decimal(math.Abs(float64(numero.(hrp.Decimal)))), nil
}

// _mat_absoluto cria e define a assinatura do método 'absoluto' exposto na stdlib do Harpia.
var _mat_absoluto = hrp.NewMetodoOuPanic(
	"absoluto",
	met_mat_absoluto,
	"absoluto(numero) -> Retorna o valor absoluto de um número, isso é, sem sinal caso houver",
)
