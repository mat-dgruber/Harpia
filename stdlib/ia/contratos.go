// Package ia implementa as facilidades de integração com modelos de inteligência artificial generativa.
package ia

import (
	"encoding/json"
	"fmt"

	"github.com/mat-dgruber/Harpia/hrp"
)

// validarCampo valida de forma recursiva os tipos de dados básicos desserializados do payload JSON.
func validarCampo(tipoEsperado string, valor interface{}) bool {
	switch tipoEsperado {
	case "texto":
		_, ok := valor.(string)
		return ok
	case "inteiro":
		if f, ok := valor.(float64); ok {
			return f == float64(int(f))
		}
		return false
	case "decimal":
		_, ok := valor.(float64)
		return ok
	case "booleano":
		_, ok := valor.(bool)
		return ok
	case "lista":
		_, ok := valor.([]interface{})
		return ok
	case "mapa":
		_, ok := valor.(map[string]interface{})
		return ok
	}
	return false
}

// ValidarResposta faz o parsing do payload de resposta do LLM e valida se todos os campos obrigatórios
// descritos no mapa de esquema estrutural do Harpia estão presentes e seguem a tipagem estrita declarada.
func ValidarResposta(esquema hrp.Mapa, resposta string) (bool, error) {
	var dados map[string]interface{}
	err := json.Unmarshal([]byte(resposta), &dados)
	if err != nil {
		return false, fmt.Errorf("formato de resposta inválido (esperado JSON): %v", err)
	}

	for chave, v := range esquema {
		tipoEsp, ok := v.(hrp.Texto)
		if !ok {
			continue
		}

		valorJson, existe := dados[chave]
		if !existe {
			return false, fmt.Errorf("campo obrigatório '%s' ausente na resposta", chave)
		}

		if !validarCampo(string(tipoEsp), valorJson) {
			return false, fmt.Errorf("campo '%s' possui tipo incorreto (esperava '%s')", chave, tipoEsp)
		}
	}

	return true, nil
}

// met_validar_resposta implementa 'validar_resposta(esquemaMapa, respostaJson)' em nível de script Harpia.
// Devolve Verdadeiro ou lança uma exceção estruturada com diagnóstico em caso de incompatibilidade de schema.
func met_validar_resposta(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("validar_resposta", false, args, 2, 2); err != nil {
		return nil, err
	}

	esquema, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperado um objeto Mapa para o esquema")
	}

	resposta, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	valido, errVal := ValidarResposta(esquema, string(resposta.(hrp.Texto)))
	if errVal != nil {
		return hrp.Falso, hrp.NewErroF(hrp.ValorErro, "%v", errVal)
	}

	if valido {
		return hrp.Verdadeiro, nil
	}
	return hrp.Falso, nil
}
