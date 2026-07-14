package embutidos

import "github.com/natanfeitosa/portuscript/ptst"

// met_emb_doc implementa a lógica nativa para a função global 'doc()'.
//
// Esta função recebe um único argumento e retorna a documentação (docstring ou texto explicativo)
// associada a ele.
//
// Mecânica de Resolução:
//   - Tenta resolver a função 'imprima' do escopo do módulo atual para exibir o texto resultante;
//   - Verifica se o argumento implementa a interface I_ObtemDoc (que expõe o método ObtemDoc()).
//     Se implementado (como em métodos e funções nativas), chama e exibe essa string;
//   - Caso contrário (como em instâncias de objetos comuns), resolve o tipo base do objeto e obtém
//     a documentação registrada na própria classe de Tipo correspondente como fallback.
func met_emb_doc(mod ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("doc", false, args, 1, 1); err != nil {
		return nil, err
	}

	arg := args[0]
	imp, err := mod.(*ptst.Modulo).Escopo.ObterValor("imprima")
	if err != nil {
		return nil, err
	}

	if obj, ok := arg.(ptst.I_ObtemDoc); ok {
		return ptst.Chamar(imp, ptst.Tupla{ptst.Texto(obj.ObtemDoc())})
	}

	return ptst.Chamar(imp, ptst.Tupla{ptst.Texto(arg.Tipo().ObtemDoc())})
}

// _emb_doc cria e define a assinatura do método 'doc' exposto globalmente.
var _emb_doc = ptst.NewMetodoOuPanic(
	"doc",
	met_emb_doc,
	"doc(objeto) -> Obtem a documentação do objeto",
)