package embutidos

import "github.com/mat-dgruber/Harpia/ptst"

// met_emb_int implementa a lógica nativa para a função global 'int()'.
//
// Esta função recebe um objeto (como uma string contendo dígitos ou um número decimal)
// e tenta convertê-lo e representá-lo sob o tipo Inteiro nativo de 64 bits da VM (ptst.Inteiro).
//
// Ela delega o processamento de coerção de tipos para o método construtor 'ptst.NewInteiro'.
func met_emb_int(mod ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("int", false, args, 0, 1); err != nil {
		return nil, err
	}

	return ptst.NewInteiro(args[0])
}

// _emb_int cria e define a assinatura do método 'int' exposto globalmente.
var _emb_int = ptst.NewMetodoOuPanic(
	"int",
	met_emb_int,
	"int(objeto) -> Recebe um objeto e retorna uma representação numérica do tipo inteiro, se possível",
)
