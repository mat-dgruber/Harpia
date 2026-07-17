package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/ptst"
)

// gerar_openapi(servidor) -> gera JSON OpenAPI 3.0 simplificado a partir das rotas do servidor.
func met_gerar_openapi(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("gerar_openapi", false, args, 1, 1); err != nil {
		return nil, err
	}

	servidor, ok := args[0].(*Servidor)
	if !ok {
		return nil, ptst.NewErroF(ptst.TipagemErro, "esperava um objeto Servidor para gerar_openapi")
	}

	paths := make(map[string]map[string]interface{})

	servidor.mu.RLock()
	for metodo, rotas := range servidor.rotas {
		metodoLower := strings.ToLower(metodo)

		for rotaPattern := range rotas {
			if paths[rotaPattern] == nil {
				paths[rotaPattern] = make(map[string]interface{})
			}

			paths[rotaPattern][metodoLower] = map[string]interface{}{
				"summary":     fmt.Sprintf("Rota %s registrada via Harpia HTTP", rotaPattern),
				"description": fmt.Sprintf("Retorna o processamento da rota %s", rotaPattern),
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Sucesso",
					},
				},
			}
		}
	}
	servidor.mu.RUnlock()

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Harpia Servidor API",
			"description": "Especificação OpenAPI gerada automaticamente pelo módulo HTTP do Harpia",
			"version":     "1.0.0",
		},
		"paths": paths,
	}

	bytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "erro ao gerar JSON OpenAPI: %v", err)
	}

	return ptst.Texto(bytes), nil
}
