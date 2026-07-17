package embutidos

import "github.com/mat-dgruber/Harpia/ptst"

// met_emb_texto implementa a lógica nativa para a função global 'texto()'.
//
// Esta função recebe um objeto de qualquer classe (ou nenhum argumento, retornando string vazia)
// e realiza o casting (coerção) estruturado de tipos, devolvendo a representação textual
// correspondente do objeto em formato de Texto nativo da VM (ptst.Texto).
//
// Ela delega o processamento lógico diretamente para a rotina 'ptst.NewTexto'.
func met_emb_texto(mod ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("texto", false, args, 0, 1); err != nil {
		return nil, err
	}

	return ptst.NewTexto(args[0])
}

// _emb_texto cria e define a assinatura do método 'texto' exposto globalmente.
var _emb_texto = ptst.NewMetodoOuPanic(
	"texto",
	met_emb_texto,
	"texto(objeto) -> Recebe um objeto e retorna uma representação no tipo texto, se possível",
)