package bd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mat-dgruber/Harpia/ptst"
)

type ClienteVetorial struct {
	URL     string
	Colecao string
	HTTPCli *http.Client
}

var TipoClienteVetorial = ptst.TipoObjeto.NewTipo("ClienteVetorial", "Cliente para banco de dados vetorial Qdrant")

func (c *ClienteVetorial) Tipo() *ptst.Tipo {
	return TipoClienteVetorial
}

func (c *ClienteVetorial) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "inserir":
		return ptst.NewMetodoOuPanic("inserir", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("inserir", false, args, 3, 3); err != nil {
				return nil, err
			}
			id, _ := ptst.NewInteiro(args[0])
			vetorObj, ok := args[1].(*ptst.Lista)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "esperava lista para o vetor")
			}
			metaObj, ok := args[2].(ptst.Mapa)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "esperava mapa para os metadados")
			}

			var vetor []float64
			for _, v := range vetorObj.Itens {
				vDec, err := ptst.NewDecimal(v)
				if err != nil {
					return nil, err
				}
				vetor = append(vetor, float64(vDec.(ptst.Decimal)))
			}

			meta := make(map[string]interface{})
			for k, v := range metaObj {
				vTexto, _ := ptst.NewTexto(v)
				meta[k] = string(vTexto.(ptst.Texto))
			}

			ponto := map[string]interface{}{
				"points": []map[string]interface{}{
					{
						"id":      int64(id.(ptst.Inteiro)),
						"vector":  vetor,
						"payload": meta,
					},
				},
			}

			bodyBytes, _ := json.Marshal(ponto)
			reqURL := fmt.Sprintf("%s/collections/%s/points?wait=true", c.URL, c.Colecao)
			req, err := http.NewRequest("PUT", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao criar requisição PUT Qdrant: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := c.HTTPCli.Do(req)
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao enviar requisição PUT Qdrant: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return ptst.Falso, nil
			}
			return ptst.Verdadeiro, nil
		}, ""), nil

	case "buscar":
		return ptst.NewMetodoOuPanic("buscar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("buscar", false, args, 2, 2); err != nil {
				return nil, err
			}
			vetorObj, ok := args[0].(*ptst.Lista)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "esperava lista para o vetor")
			}
			limite, _ := ptst.NewInteiro(args[1])

			var vetor []float64
			for _, v := range vetorObj.Itens {
				vDec, err := ptst.NewDecimal(v)
				if err != nil {
					return nil, err
				}
				vetor = append(vetor, float64(vDec.(ptst.Decimal)))
			}

			query := map[string]interface{}{
				"vector":       vetor,
				"limit":        int(limite.(ptst.Inteiro)),
				"with_payload": true,
			}

			bodyBytes, _ := json.Marshal(query)
			reqURL := fmt.Sprintf("%s/collections/%s/points/search", c.URL, c.Colecao)
			req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao criar requisição search Qdrant: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := c.HTTPCli.Do(req)
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao enviar requisição search Qdrant: %v", err)
			}
			defer res.Body.Close()

			resultList := &ptst.Lista{Itens: make(ptst.Tupla, 0)}

			if res.StatusCode != http.StatusOK {
				return resultList, nil
			}

			var respBody map[string]interface{}
			json.NewDecoder(res.Body).Decode(&respBody)

			if result, ok := respBody["result"].([]interface{}); ok {
				for _, r := range result {
					if item, ok := r.(map[string]interface{}); ok {
						mapaItem := ptst.NewMapaVazio()
						if idVal, ok := item["id"].(float64); ok {
							mapaItem.M__define_item__(ptst.Texto("id"), ptst.Inteiro(idVal))
						}
						if scoreVal, ok := item["score"].(float64); ok {
							mapaItem.M__define_item__(ptst.Texto("score"), ptst.Decimal(scoreVal))
						}
						if payload, ok := item["payload"].(map[string]interface{}); ok {
							payloadMap := ptst.NewMapaVazio()
							for pk, pv := range payload {
								if pvStr, ok := pv.(string); ok {
									payloadMap.M__define_item__(ptst.Texto(pk), ptst.Texto(pvStr))
								}
							}
							mapaItem.M__define_item__(ptst.Texto("payload"), payloadMap)
						}
						resultList.Adiciona(mapaItem)
					}
				}
			}

			return resultList, nil
		}, ""), nil

	case "deletar":
		return ptst.NewMetodoOuPanic("deletar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("deletar", false, args, 1, 1); err != nil {
				return nil, err
			}
			id, _ := ptst.NewInteiro(args[0])

			query := map[string]interface{}{
				"points": []interface{}{int64(id.(ptst.Inteiro))},
			}

			bodyBytes, _ := json.Marshal(query)
			reqURL := fmt.Sprintf("%s/collections/%s/points/delete?wait=true", c.URL, c.Colecao)
			req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao criar requisição delete Qdrant: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := c.HTTPCli.Do(req)
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao enviar requisição delete Qdrant: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return ptst.Falso, nil
			}
			return ptst.Verdadeiro, nil
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em ClienteVetorial", nome)
}

func met_conectar_qdrant(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("conectar_qdrant", false, args, 2, 2); err != nil {
		return nil, err
	}
	url, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	colecao, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	return &ClienteVetorial{
		URL:     string(url.(ptst.Texto)),
		Colecao: string(colecao.(ptst.Texto)),
		HTTPCli: &http.Client{Timeout: 5 * time.Second},
	}, nil
}
