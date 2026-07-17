package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/ptst"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func TestModuloIA_Mockado(t *testing.T) {
	// Cria um servidor HTTP mock para simular o Ollama
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var payload struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Responder de forma simulada
		w.Header().Set("Content-Type", "application/json")
		resposta := map[string]interface{}{
			"message": map[string]string{
				"role":    "assistant",
				"content": "Olá, sou o " + payload.Model + "! Recebi sua mensagem: " + payload.Messages[len(payload.Messages)-1].Content,
			},
		}
		json.NewEncoder(w).Encode(resposta)
	}))
	defer mockServer.Close()

	// Configura as variáveis de ambiente temporárias
	os.Setenv("OLLAMA_HOST", mockServer.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "ia" importe Agente

	// Instancia o agente
	var assistente = Agente("HarpiaHelper", "Você é um assistente prestativo", "ollama", "llama3")
	
	// Testa os atributos básicos
	var nome = assistente.nome
	var provedor = assistente.provedor
	var modelo = assistente.modelo

	// Testa a interação/pergunta
	var resposta = assistente.perguntar("como compilar?")
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo ia: %v", err)
	}

	valNome, _ := res.Escopo.ObterValor("nome")
	if string(valNome.(ptst.Texto)) != "HarpiaHelper" {
		t.Errorf("Nome inválido, obteve: %v", valNome)
	}

	valProvedor, _ := res.Escopo.ObterValor("provedor")
	if string(valProvedor.(ptst.Texto)) != "ollama" {
		t.Errorf("Provedor inválido, obteve: %v", valProvedor)
	}

	valModelo, _ := res.Escopo.ObterValor("modelo")
	if string(valModelo.(ptst.Texto)) != "llama3" {
		t.Errorf("Modelo inválido, obteve: %v", valModelo)
	}

	valResposta, _ := res.Escopo.ObterValor("resposta")
	if !strings.Contains(string(valResposta.(ptst.Texto)), "como compilar?") {
		t.Errorf("Resposta simulada inválida, obteve: %v", valResposta)
	}
}

func TestModuloIA_ComunicacaoMultiAgente(t *testing.T) {
	// Mock do servidor de chat
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		json.NewDecoder(r.Body).Decode(&payload)

		w.Header().Set("Content-Type", "application/json")
		resposta := map[string]interface{}{
			"message": map[string]string{
				"role":    "assistant",
				"content": "Agente " + payload.Model + " processou: " + payload.Messages[len(payload.Messages)-1].Content,
			},
		}
		json.NewEncoder(w).Encode(resposta)
	}))
	defer mockServer.Close()

	os.Setenv("OLLAMA_HOST", mockServer.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "ia" importe Agente

	var agente1 = Agente("A1", "Focado em código", "ollama", "agente-codigo")
	var agente2 = Agente("A2", "Focado em revisão", "ollama", "agente-revisao")

	// agente1 pergunta ao agente2 para revisar a mensagem
	var conversa = agente1.comunicar(agente2, "revise esta linha")
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script de comunicação multi-agente: %v", err)
	}

	valConversa, _ := res.Escopo.ObterValor("conversa")
	if !strings.Contains(string(valConversa.(ptst.Texto)), "agente-codigo") {
		t.Errorf("Erro na orquestração de resposta, obteve: %v", valConversa)
	}
}
