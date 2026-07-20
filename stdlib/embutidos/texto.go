package embutidos

import "github.com/mat-dgruber/Harpia/hrp"

// met_emb_texto implementa a lógica nativa para a função global 'texto()'.
//
// Esta função recebe um objeto de qualquer classe (ou nenhum argumento, retornando string vazia)
// e realiza o casting (coerção) estruturado de tipos, devolvendo a representação textual
// correspondente do objeto em formato de Texto nativo da VM (hrp.Texto).
//
// Ela delega o processamento lógico diretamente para a rotina 'hrp.NewTexto'.
func met_emb_texto(mod hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("texto", false, args, 0, 1); err != nil {
		return nil, err
	}

	return hrp.NewTexto(args[0])
}

// _emb_texto cria e define a assinatura do método 'texto' exposto globalmente.
var _emb_texto = hrp.NewMetodoOuPanic(
	"texto",
	met_emb_texto,
	"texto(objeto) -> Recebe um objeto e retorna uma representação no tipo texto, se possível",
)
