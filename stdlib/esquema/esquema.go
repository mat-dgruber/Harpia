// Package esquema fornece um validador declarativo de estrutura de dados (estilo Zod ou JSON Schema),
// permitindo criar schemas rigorosos e validar mapas de dados recebidos de APIs ou usuários.
package esquema

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Esquema guarda as regras lógicas e de tipos cadastrados para validação estrutural.
type Esquema struct {
	regras hrp.Mapa
}

// TipoEsquema define e expõe a classe Esquema na VM do Harpia.
var TipoEsquema = hrp.NewTipo("Esquema", "Validador de estrutura de dados Schema")

// Tipo retorna a representação na VM.
func (e *Esquema) Tipo() *hrp.Tipo {
	return TipoEsquema
}

// M__obtem_attributo__ mapeia o método analisar() no tipo Esquema.
func (e *Esquema) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "analisar":
		// analisa um Mapa de dados de entrada contra as regras estruturais e de tipos definidas no Schema.
		// Retorna um par na forma de Lista: [dadosValidados ou Nulo, MensagemDeErro ou Nulo] (sucesso vs falha).
		return hrp.NewMetodoOuPanic("analisar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
				return nil, err
			}

			mapaDados, ok := args[0].(hrp.Mapa)
			if !ok {
				retLista := &hrp.Lista{Itens: []hrp.Objeto{hrp.Nulo, hrp.Texto("Os dados fornecidos para análise devem ser um Mapa")}}
				return retLista, nil
			}

			dadosValidados := hrp.NewMapaVazio()

			for chave, tipoObj := range e.regras {
				val, ok := mapaDados[chave]
				if !ok {
					retLista := &hrp.Lista{Itens: []hrp.Objeto{hrp.Nulo, hrp.Texto(fmt.Sprintf("Campo obrigatório ausente: '%s'", chave))}}
					return retLista, nil
				}

				// Validação básica de correspondência do tipo
				var tipoEsperado string
				if t, ok := tipoObj.(*hrp.Tipo); ok {
					tipoEsperado = t.Nome
				} else {
					tipoEsperado = fmt.Sprintf("%v", tipoObj)
				}
				tipoObtido := val.Tipo().Nome

				if tipoEsperado != tipoObtido {
					retLista := &hrp.Lista{Itens: []hrp.Objeto{hrp.Nulo, hrp.Texto(fmt.Sprintf("Campo '%s' deve ser do tipo %s, obteve %s", chave, tipoEsperado, tipoObtido))}}
					return retLista, nil
				}

				dadosValidados.M__define_item__(hrp.Texto(chave), val)
			}

			// Retorna [dadosValidados, Nulo] significando sucesso sem erros detectados
			retLista := &hrp.Lista{Itens: []hrp.Objeto{dadosValidados, hrp.Nulo}}
			return retLista, nil
		}, "Analisa e valida um Mapa de dados contra as regras de tipagem do Schema."), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Esquema", nome)
}

// met_esquema_criar implementa 'criar(regrasMapa)' para fabricar um validador Esquema.
func met_esquema_criar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("esquema", false, args, 1, 1); err != nil {
		return nil, err
	}
	mapaRegras, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esquema esperava um Mapa contendo as regras de tipo")
	}
	return &Esquema{regras: mapaRegras}, nil
}

func init() {
	// Registra o módulo 'esquema' no sistema central da biblioteca padrão do Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "esquema",
			Arquivo: "stdlib/esquema",
		},
		Constantes: hrp.Mapa{
			"Esquema": TipoEsquema,
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("criar", met_esquema_criar, "Cria um novo validador de Schema de dados baseado em um Mapa de regras."),
		},
	})
}
