package embutidos

import "github.com/mat-dgruber/Harpia/hrp"

// emb_tamanho_fn implementa a lógica nativa para a função global 'tamanho()'.
//
// Esta função recebe um único objeto e retorna a quantidade de elementos que ele abriga.
//
// Ela analisa se o argumento implementa a interface de protocolo '__tamanho__' (hrp.I__tamanho__).
// Se implementada (como em Textos, Listas, Tuplas e Dicionários), chama o respectivo método nativo
// 'M__tamanho__()' e retorna este valor inteiro. Caso contrário, lança um erro estruturado de Tipagem (TipagemErro).
func emb_tamanho_fn(mod hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("tamanho", false, args, 1, 1); err != nil {
		return nil, err
	}

	if obj, ok := args[0].(hrp.I__tamanho__); ok {
		return obj.M__tamanho__()
	}

	return nil, hrp.NewErroF(hrp.TipagemErro, "Objeto do tipo '%s' não implementa a interface '__tamanho__'.", args[0].Tipo().Nome)
}

// _emb_tamanho cria e define a assinatura do método 'tamanho' exposto globalmente.
var _emb_tamanho = hrp.NewMetodoOuPanic(
	"tamanho",
	emb_tamanho_fn,
	"tamanho(obj) -> Retorna o tamanho do objeto, mas se o objeto não implementar o método `__tamanho__`, um erro será lançado",
)
