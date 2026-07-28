// Package http implementa o servidor web nativo de alta performance e cliente HTTP do Harpia.
package http

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// regexParametrosRota encontra parâmetros dinâmicos na rota (ex: :id, :cpf, :nome)
var regexParametrosRota = regexp.MustCompile(`:([a-zA-Z0-9_]+)`)

// met_gerar_openapi implementa 'gerar_openapi(servidor)' em nível de script Harpia.
// Analisa a tabela de rotas e metadados no servidor do Harpia e gera a especificação JSON OpenAPI 3.0 completa.
func met_gerar_openapi(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("gerar_openapi", false, args, 1, 1); err != nil {
		return nil, err
	}

	servidor, ok := args[0].(*Servidor)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava um objeto Servidor para gerar_openapi")
	}

	paths := make(map[string]map[string]interface{})
	tagsSet := make(map[string]bool)

	servidor.mu.RLock()
	for metodo, rotas := range servidor.rotas {
		metodoLower := strings.ToLower(metodo)

		for rotaPattern := range rotas {
			if paths[rotaPattern] == nil {
				paths[rotaPattern] = make(map[string]interface{})
			}

			count := servidor.metricas[metodo+" "+rotaPattern]
			meta := servidor.rotasMeta[metodo][rotaPattern]

			op := map[string]interface{}{
				"summary":           fmt.Sprintf("Endpoint %s %s", metodo, rotaPattern),
				"description":       fmt.Sprintf("Processa requisição para %s", rotaPattern),
				"x-harpia-metricas": count,
			}

			// Injeta tags ou detecta automaticamente do path
			var tags []string
			if meta != nil && len(meta.Tags) > 0 {
				tags = meta.Tags
			} else {
				// Fallback limpo de tag baseado no 1º segmento útil do caminho (ex: /rpc/filmes/listar -> filmes)
				partes := strings.Split(strings.Trim(rotaPattern, "/"), "/")
				if len(partes) > 0 && partes[0] != "" {
					tagBase := partes[0]
					if tagBase == "rpc" && len(partes) > 1 {
						tagBase = partes[1]
					}
					tags = []string{tagBase}
				} else {
					tags = []string{"geral"}
				}
			}

			for _, t := range tags {
				tagsSet[t] = true
			}
			op["tags"] = tags

			if meta != nil {
				if meta.Resumo != "" {
					op["summary"] = meta.Resumo
				}
				if meta.Descricao != "" {
					op["description"] = meta.Descricao
				}
				if meta.Deprecated {
					op["deprecated"] = true
					if meta.MensagemDeprecacao != "" {
						op["x-harpia-deprecacao"] = meta.MensagemDeprecacao
					}
				}
			}

			// Processa Parâmetros de Path (:id, :slug, etc.)
			paramMatches := regexParametrosRota.FindAllStringSubmatch(rotaPattern, -1)
			var parameters []map[string]interface{}
			for _, match := range paramMatches {
				pName := match[1]
				parameters = append(parameters, map[string]interface{}{
					"name":        pName,
					"in":          "path",
					"required":    true,
					"description": fmt.Sprintf("Parâmetro de rota :%s", pName),
					"schema": map[string]interface{}{
						"type": "string",
					},
				})
			}
			if len(parameters) > 0 {
				op["parameters"] = parameters
			}

			// Processa Corpo da Requisição (requestBody)
			if meta != nil && meta.Corpo != nil && len(meta.Corpo) > 0 {
				schemaProps := make(map[string]interface{})

				for field, val := range meta.Corpo {
					propType := "string"
					switch v := val.(type) {
					case string:
						switch strings.ToLower(v) {
						case "int", "integer", "numero", "num":
							propType = "integer"
						case "float", "double", "decimal":
							propType = "number"
						case "bool", "boolean", "booleano":
							propType = "boolean"
						case "array", "lista":
							propType = "array"
						case "object", "objeto":
							propType = "object"
						default:
							propType = "string"
						}
					case float64, int, int64:
						propType = "number"
					case bool:
						propType = "boolean"
						case []interface{}:
						propType = "array"
					case map[string]interface{}:
						propType = "object"
					}

					schemaProps[field] = map[string]interface{}{
						"type": propType,
					}
				}

				op["requestBody"] = map[string]interface{}{
					"required":    true,
					"description": "Corpo JSON da requisição",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":       "object",
								"properties": schemaProps,
							},
							"example": meta.Corpo,
						},
					},
				}
			}

			// Processa Respostas HTTP (responses)
			responses := make(map[string]interface{})
			if meta != nil && meta.Respostas != nil && len(meta.Respostas) > 0 {
				for code, desc := range meta.Respostas {
					responses[code] = map[string]interface{}{
						"description": desc,
					}
				}
			} else {
				responses["200"] = map[string]interface{}{
					"description": "Requisição executada com sucesso",
				}
			}
			op["responses"] = responses

			paths[rotaPattern][metodoLower] = op
		}
	}
	servidor.mu.RUnlock()

	var tagsList []map[string]interface{}
	for t := range tagsSet {
		tagsList = append(tagsList, map[string]interface{}{
			"name":        t,
			"description": fmt.Sprintf("Operações relacionadas a %s", t),
		})
	}

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Harpia Servidor API",
			"description": "Especificação OpenAPI 3.0 estruturada e gerada automaticamente pelo Harpia HTTP",
			"version":     "1.0.0",
		},
		"tags":  tagsList,
		"paths": paths,
	}

	bytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao gerar JSON OpenAPI: %v", err)
	}

	return hrp.Texto(bytes), nil
}
