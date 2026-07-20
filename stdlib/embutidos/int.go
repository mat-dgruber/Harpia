package embutidos

import "github.com/mat-dgruber/Harpia/hrp"

// met_emb_int implementa a lógica nativa para a função global 'int()'.
//
// Esta função recebe um objeto (como uma string contendo dígitos ou um número decimal)
// e tenta convertê-lo e representá-lo sob o tipo Inteiro nativo de 64 bits da VM (hrp.Inteiro).
//
// Ela delega o processamento de coerção de tipos para o método construtor 'hrp.NewInteiro'.
func met_emb_int(mod hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("int", false, args, 0, 1); err != nil {
		return nil, err
	}

	return hrp.NewInteiro(args[0])
}

// _emb_int cria e define a assinatura do método 'int' exposto globalmente.
var _emb_int = hrp.NewMetodoOuPanic(
	"int",
	met_emb_int,
	"int(objeto) -> Recebe um objeto e retorna uma representação numérica do tipo inteiro, se possível",
)
