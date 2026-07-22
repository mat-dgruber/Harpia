package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mat-dgruber/Harpia/parser"
	"github.com/spf13/cobra"
)

// swaggerCmd gera a especificação OpenAPI 3.0 a partir de chamadas HTTP Harpia encontradas
// no AST do programa (ex: `obter("/path")`, `postar(...)`, `put(...)`, `deleter(...)`).
//
// Saída canônica: arquivo `swagger.json` no diretório atual, pronto para alimentar
// Swagger UI, Redoc ou geradores de cliente SDK.
var swaggerCmd = &cobra.Command{
	Use:     "docs [arquivo.hrp]",
	Aliases: []string{"swagger", "openapi"},
	Short:   "Gera a especificação Swagger/OpenAPI 3.0 para APIs HTTP do Harpia",
	RunE: func(cmd *cobra.Command, args []string) error {
		entrada := "main.hrp"
		if len(args) > 0 {
			entrada = args[0]
		}

		conteudo, err := os.ReadFile(entrada)
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo '%s': %w", entrada, err)
		}

		p := parser.NewParserFromString(string(conteudo), entrada)
		astNode, err := p.Parse()
		if err != nil {
			return fmt.Errorf("erro ao analisar arquivo '%s': %w", entrada, err)
		}

		doc := extrairOpenAPI(astNode, entrada)
		saidaJson, _ := json.MarshalIndent(doc, "", "  ")

		dest := "swagger.json"
		if err := os.WriteFile(dest, saidaJson, 0644); err != nil {
			return fmt.Errorf("erro ao gravar '%s': %w", dest, err)
		}

		fmt.Printf("✅ Especificação OpenAPI 3.0 gerada com sucesso em '%s'!\n", dest)
		return nil
	},
}

// comandoSwagger retorna o `*cobra.Command` registrado pelo orquestrador da CLI.
func comandoSwagger() *cobra.Command {
	return swaggerCmd
}

// extrairOpenAPI percorre recursivamente a AST, identifica chamadas HTTP Harpia
// (`obter`, `postar`, `put`, `deleter`) e converte cada uma em uma `Operation` OpenAPI.
//
// O resultado é montado em uma especificação OpenAPI 3.0.0 mínima contendo:
//   - paths populados pelas chamadas declaradas no programa.
//   - responses[200] padrão com descrição "Sucesso".
//   - bloco `info` populado dinamicamente a partir do nome do arquivo fonte.
//
// Limitação conhecida: apenas chamadas de função como literais diretos são reconhecidas;
// chamadas via variáveis (`var acao = http.obter`) não são capturadas.
func extrairOpenAPI(ast parser.BaseNode, filename string) map[string]interface{} {
	paths := make(map[string]interface{})

	var varrer func(n parser.BaseNode)
	varrer = func(n parser.BaseNode) {
		if n == nil {
			return
		}
		if call, ok := n.(*parser.ChamadaFuncao); ok {
			if prop, okCall := call.Identificador.(*parser.AcessoMembro); okCall {
				if propIdent, okIdent := prop.Membro.(*parser.Identificador); okIdent {
					metodo := strings.ToLower(propIdent.Nome)
					if metodo == "obter" || metodo == "postar" || metodo == "put" || metodo == "deleter" {
						if len(call.Argumentos) > 0 {
							if txt, okTxt := call.Argumentos[0].(*parser.TextoLiteral); okTxt {
								caminho := txt.Valor
								httpMethod := metodo
								if metodo == "obter" {
									httpMethod = "get"
								} else if metodo == "postar" {
									httpMethod = "post"
								} else if metodo == "deleter" {
									httpMethod = "delete"
								}

								if paths[caminho] == nil {
									paths[caminho] = make(map[string]interface{})
								}
								pMap := paths[caminho].(map[string]interface{})
								pMap[httpMethod] = map[string]interface{}{
									"summary": fmt.Sprintf("Endpoint %s %s", strings.ToUpper(httpMethod), caminho),
									"responses": map[string]interface{}{
										"200": map[string]interface{}{
											"description": "Sucesso",
										},
									},
								}
							}
						}
					}
				}
			}
		}

		switch node := n.(type) {
		case *parser.Programa:
			for _, stmt := range node.Declaracoes {
				varrer(stmt)
			}
		case *parser.Bloco:
			for _, stmt := range node.Declaracoes {
				varrer(stmt)
			}
		case *parser.DeclFuncao:
			varrer(node.Corpo)
		}
	}

	varrer(ast)

	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "API Harpia",
			"version":     "1.0.0",
			"description": fmt.Sprintf("Documentação gerada automaticamente para %s", filepath.Base(filename)),
		},
		"paths": paths,
	}
}
