package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAPIExecutarCodigoPlayground assevera o funcionamento da API de execução síncrona do playground
func TestAPIExecutarCodigoPlayground(t *testing.T) {
	requestBody, _ := json.Marshal(ExecucaoPlaygroundRequest{
		Codigo: "var a = 42;\nimprimir(a);",
	})

	req, err := http.NewRequest("POST", "/api/executar", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiExecutarCodigoPlayground)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Esperava status 200, obtive %d", rr.Code)
	}

	var response ExecucaoPlaygroundResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Erro ao decodificar JSON de resposta: %v", err)
	}

	// 1. Deve ser executado com sucesso
	if !response.Sucesso {
		t.Errorf("Esperava sucesso na execução do código do playground")
	}

	// 2. Deve coletar e retornar a saída correta de stdout
	if !bytes.Contains([]byte(response.Saida), []byte("42")) {
		t.Errorf("Esperava encontrar '42' na saída capturada de console, recebido: %s", response.Saida)
	}

	// 3. Deve monitorar e registrar a variável 'a' declarada no escopo
	var achouVarA bool
	for _, v := range response.Variaveis {
		if v.Nome == "a" {
			achouVarA = true
			if v.Valor != "42" {
				t.Errorf("Esperava valor de 'a' igual a '42', obtive '%s'", v.Valor)
			}
		}
	}

	if !achouVarA {
		t.Errorf("Variável 'a' declarada no script do playground não foi encontrada no inspetor de escopo")
	}
}
