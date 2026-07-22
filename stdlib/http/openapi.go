// Package http implementa o servidor web nativo de alta performance e cliente HTTP do Harpia.
package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_gerar_openapi implementa 'gerar_openapi(servidor)' em nível de script Harpia.
// Analisa estaticamente a tabela de rotas registradas no servidor do Harpia e gera a especificação JSON OpenAPI 3.0 correspondente.
func met_gerar_openapi(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("gerar_openapi", false, args, 1, 1); err != nil {
		return nil, err
	}

	servidor, ok := args[0].(*Servidor)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava um objeto Servidor para gerar_openapi")
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
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao gerar JSON OpenAPI: %v", err)
	}

	return hrp.Texto(bytes), nil
}
