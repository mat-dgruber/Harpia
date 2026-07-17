package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCopiloto_Sugerir(t *testing.T) {
	// Cria um servidor HTTP mock para o Ollama
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var reqBody OllamaGenerateReq
		json.NewDecoder(r.Body).Decode(&reqBody)

		resBody := OllamaGenerateRes{
			Response: "var x = 10",
		}
		json.NewEncoder(w).Encode(resBody)
	}))
	defer server.Close()

	// Configura as variáveis globais para apontar para o mock
	modeloCopiloto = "llama3"
	ollamaURL = server.URL

	sugestao, err := SugerirCopiloto("var x =")
	if err != nil {
		t.Fatalf("Erro ao acionar copiloto: %v", err)
	}

	if sugestao != "var x = 10" {
		t.Errorf("Sugestão incorreta, obteve: '%s'", sugestao)
	}
}
