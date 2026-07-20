package bd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

const MaxLimiteBuscaVetorial = 10000

type ClienteVetorial struct {
	URL     string
	Colecao string
	HTTPCli *http.Client
}

var TipoClienteVetorial = hrp.TipoObjeto.NewTipo("ClienteVetorial", "Cliente para banco de dados vetorial Qdrant")

func (c *ClienteVetorial) Tipo() *hrp.Tipo {
	return TipoClienteVetorial
}

func (c *ClienteVetorial) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "inserir":
		return hrp.NewMetodoOuPanic("inserir", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("inserir", false, args, 3, 3); err != nil {
				return nil, err
			}
			id, _ := hrp.NewInteiro(args[0])
			vetorObj, ok := args[1].(*hrp.Lista)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "esperava lista para o vetor")
			}
			metaObj, ok := args[2].(hrp.Mapa)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "esperava mapa para os metadados")
			}

			var vetor []float64
			for _, v := range vetorObj.Itens {
				vDec, err := hrp.NewDecimal(v)
				if err != nil {
					return nil, err
				}
				vetor = append(vetor, float64(vDec.(hrp.Decimal)))
			}

			meta := make(map[string]interface{})
			for k, v := range metaObj {
				vTexto, _ := hrp.NewTexto(v)
				meta[k] = string(vTexto.(hrp.Texto))
			}

			ponto := map[string]interface{}{
				"points": []map[string]interface{}{
					{
						"id":      int64(id.(hrp.Inteiro)),
						"vector":  vetor,
						"payload": meta,
					},
				},
			}

			bodyBytes, _ := json.Marshal(ponto)
			reqURL := fmt.Sprintf("%s/collections/%s/points?wait=true", c.URL, c.Colecao)
			req, err := http.NewRequest("PUT", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao criar requisição PUT Qdrant: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := c.HTTPCli.Do(req)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao enviar requisição PUT Qdrant: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return hrp.Falso, nil
			}
			return hrp.Verdadeiro, nil
		}, ""), nil

	case "buscar":
		return hrp.NewMetodoOuPanic("buscar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("buscar", false, args, 2, 2); err != nil {
				return nil, err
			}
			vetorObj, ok := args[0].(*hrp.Lista)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "esperava lista para o vetor")
			}
			limite, _ := hrp.NewInteiro(args[1])

			var vetor []float64
			for _, v := range vetorObj.Itens {
				vDec, err := hrp.NewDecimal(v)
				if err != nil {
					return nil, err
				}
				vetor = append(vetor, float64(vDec.(hrp.Decimal)))
			}

			limVal := int64(limite.(hrp.Inteiro))
			if limVal < 1 || limVal > MaxLimiteBuscaVetorial {
				return nil, hrp.NewErroF(hrp.ValorErro, "limite de busca vetorial inválido (deve ser entre 1 e %d)", MaxLimiteBuscaVetorial)
			}
			query := map[string]interface{}{
				"vector":       vetor,
				"limit":        limVal,
				"with_payload": true,
			}

			bodyBytes, _ := json.Marshal(query)
			reqURL := fmt.Sprintf("%s/collections/%s/points/search", c.URL, c.Colecao)
			req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao criar requisição search Qdrant: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := c.HTTPCli.Do(req)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao enviar requisição search Qdrant: %v", err)
			}
			defer res.Body.Close()

			resultList := &hrp.Lista{Itens: make(hrp.Tupla, 0)}

			if res.StatusCode != http.StatusOK {
				return resultList, nil
			}

			var respBody map[string]interface{}
			json.NewDecoder(res.Body).Decode(&respBody)

			if result, ok := respBody["result"].([]interface{}); ok {
				for _, r := range result {
					if item, ok := r.(map[string]interface{}); ok {
						mapaItem := hrp.NewMapaVazio()
						if idVal, ok := item["id"].(float64); ok {
							mapaItem.M__define_item__(hrp.Texto("id"), hrp.Inteiro(idVal))
						}
						if scoreVal, ok := item["score"].(float64); ok {
							mapaItem.M__define_item__(hrp.Texto("score"), hrp.Decimal(scoreVal))
						}
						if payload, ok := item["payload"].(map[string]interface{}); ok {
							payloadMap := hrp.NewMapaVazio()
							for pk, pv := range payload {
								if pvStr, ok := pv.(string); ok {
									payloadMap.M__define_item__(hrp.Texto(pk), hrp.Texto(pvStr))
								}
							}
							mapaItem.M__define_item__(hrp.Texto("payload"), payloadMap)
						}
						resultList.Adiciona(mapaItem)
					}
				}
			}

			return resultList, nil
		}, ""), nil

	case "deletar":
		return hrp.NewMetodoOuPanic("deletar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("deletar", false, args, 1, 1); err != nil {
				return nil, err
			}
			id, _ := hrp.NewInteiro(args[0])

			query := map[string]interface{}{
				"points": []interface{}{int64(id.(hrp.Inteiro))},
			}

			bodyBytes, _ := json.Marshal(query)
			reqURL := fmt.Sprintf("%s/collections/%s/points/delete?wait=true", c.URL, c.Colecao)
			req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao criar requisição delete Qdrant: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := c.HTTPCli.Do(req)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao enviar requisição delete Qdrant: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return hrp.Falso, nil
			}
			return hrp.Verdadeiro, nil
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ClienteVetorial", nome)
}

func met_conectar_qdrant(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("conectar_qdrant", false, args, 2, 2); err != nil {
		return nil, err
	}
	url, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	colecao, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	return &ClienteVetorial{
		URL:     string(url.(hrp.Texto)),
		Colecao: string(colecao.(hrp.Texto)),
		HTTPCli: &http.Client{Timeout: 5 * time.Second},
	}, nil
}
