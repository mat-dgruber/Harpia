package esquema

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/ptst"
)

type Esquema struct {
	regras ptst.Mapa
}

var TipoEsquema = ptst.NewTipo("Esquema", "Validador de estrutura de dados Schema")

func (e *Esquema) Tipo() *ptst.Tipo {
	return TipoEsquema
}

func (e *Esquema) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "analisar":
		return ptst.NewMetodoOuPanic("analisar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
				return nil, err
			}

			mapaDados, ok := args[0].(ptst.Mapa)
			if !ok {
				retLista := &ptst.Lista{Itens: []ptst.Objeto{ptst.Nulo, ptst.Texto("Os dados fornecidos para análise devem ser um Mapa")}}
				return retLista, nil
			}

			dadosValidados := ptst.NewMapaVazio()

			for chave, tipoObj := range e.regras {
				val, ok := mapaDados[chave]
				if !ok {
					retLista := &ptst.Lista{Itens: []ptst.Objeto{ptst.Nulo, ptst.Texto(fmt.Sprintf("Campo obrigatório ausente: '%s'", chave))}}
					return retLista, nil
				}

				// Validação básica do tipo
				var tipoEsperado string
				if t, ok := tipoObj.(*ptst.Tipo); ok {
					tipoEsperado = t.Nome
				} else {
					tipoEsperado = fmt.Sprintf("%v", tipoObj)
				}
				tipoObtido := val.Tipo().Nome

				if tipoEsperado != tipoObtido {
					retLista := &ptst.Lista{Itens: []ptst.Objeto{ptst.Nulo, ptst.Texto(fmt.Sprintf("Campo '%s' deve ser do tipo %s, obteve %s", chave, tipoEsperado, tipoObtido))}}
					return retLista, nil
				}

				dadosValidados.M__define_item__(ptst.Texto(chave), val)
			}

			// Retorna [dadosValidados, Nulo] significando sucesso
			retLista := &ptst.Lista{Itens: []ptst.Objeto{dadosValidados, ptst.Nulo}}
			return retLista, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe no Esquema", nome)
}

func met_esquema_criar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("esquema", false, args, 1, 1); err != nil {
		return nil, err
	}
	mapaRegras, ok := args[0].(ptst.Mapa)
	if !ok {
		return nil, ptst.NewErroF(ptst.TipagemErro, "esquema esperava um Mapa contendo as regras de tipo")
	}
	return &Esquema{regras: mapaRegras}, nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "esquema",
			Arquivo: "stdlib/esquema",
		},
		Constantes: ptst.Mapa{
			"Esquema": TipoEsquema,
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("criar", met_esquema_criar, "Cria um novo validador de Schema de dados."),
		},
	})
}
