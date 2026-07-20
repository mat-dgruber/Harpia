package embutidos

import "github.com/mat-dgruber/Harpia/hrp"

// met_emb_instanciaDe implementa a lógica nativa para a função global 'instanciaDe()'.
//
// Esta função recebe um objeto e uma classe (ou uma tupla de classes) e verifica
// se o objeto é herdeiro ou instância direta de algum dos tipos especificados.
//
// Ela delega diretamente o processamento para a função centralizadora 'hrp.InstanciaDe'.
func met_emb_instanciaDe(mod hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("instanciaDe", false, args, 2, 2); err != nil {
		return nil, err
	}

	return hrp.InstanciaDe(args[0], args[1])
}

// _emb_instanciaDe cria e define a assinatura do método 'instanciaDe' exposto globalmente.
var _emb_instanciaDe = hrp.NewMetodoOuPanic(
	"instanciaDe",
	met_emb_instanciaDe,
	"instanciaDe(obj, tipos) -> o parâmetro `tipos` pode ser um tipo ou uma tupla de tipos/classes. Verifica se o obj é instancia de algum dos tipos",
)
