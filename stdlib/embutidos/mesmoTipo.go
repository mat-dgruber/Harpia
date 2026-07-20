package embutidos

import "github.com/mat-dgruber/Harpia/hrp"

// met_emb_mesmoTipo implementa a lógica nativa para a função global 'mesmoTipo()'.
//
// Esta função compara as assinaturas de classe e tipos de base de dois objetos
// e retorna Verdadeiro se ambos forem instanciados a partir do mesmo Tipo (classe),
// ou Falso se forem de tipos divergentes.
//
// Delega o processamento lógico para a rotina centralizada 'hrp.MesmoTipo'.
func met_emb_mesmoTipo(mod hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("mesmoTipo", false, args, 2, 2); err != nil {
		return nil, err
	}

	return hrp.Booleano(hrp.MesmoTipo(args[0], args[1])), nil
}

// _emb_mesmoTipo cria e define a assinatura do método 'mesmoTipo' exposto globalmente.
var _emb_mesmoTipo = hrp.NewMetodoOuPanic(
	"mesmoTipo",
	met_emb_mesmoTipo,
	"mesmoTipo(obj1, obj2) -> Verifica se os dois objetos são do mesmo tipo",
)
