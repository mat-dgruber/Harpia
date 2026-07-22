// Package ia implementa as facilidades de integração com modelos de inteligência artificial generativa.
package ia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Mensagem representa a estrutura padrão para formatação e mapeamento de mensagens de conversação em JSON.
type Mensagem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChamarLLM realiza o roteamento de chamada de IA com base no provedor selecionado (Ollama, Gemini, OpenAI).
// Carrega de forma transparente as chaves de API correspondentes a partir das variáveis de ambiente do sistema.
func ChamarLLM(provedor, modelo string, instrucoes string, historico []Mensagem) (string, error) {
	var mensagens []Mensagem
	if instrucoes != "" {
		mensagens = append(mensagens, Mensagem{Role: "system", Content: instrucoes})
	}
	mensagens = append(mensagens, historico...)

	switch provedor {
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return "", fmt.Errorf("provedor Gemini selecionado, mas a variável GEMINI_API_KEY não está definida")
		}
		return chamarGemini(modelo, apiKey, mensagens)
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return "", fmt.Errorf("provedor OpenAI selecionado, mas a variável OPENAI_API_KEY não está definida")
		}
		return chamarOpenAI(modelo, apiKey, mensagens)
	default: // ollama ou local
		host := os.Getenv("OLLAMA_HOST")
		if host == "" {
			host = "http://localhost:11434"
		}
		return chamarOllama(host, modelo, mensagens)
	}
}

// chamarOllama envia mensagens de chat para o servidor Ollama local via requisição REST POST.
// Caso o Ollama esteja inacessível no localhost, possui um fallback inteligente que redireciona o fluxo
// para as APIs em nuvem do Gemini ou OpenAI, se as chaves correspondentes estiverem definidas, mantendo o sistema funcional.
func chamarOllama(host, modelo string, mensagens []Mensagem) (string, error) {
	url := fmt.Sprintf("%s/api/chat", host)

	payload := map[string]interface{}{
		"model":    modelo,
		"messages": mensagens,
		"stream":   false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		// Fallback amigável se o Ollama local não estiver rodando: tentar nuvens
		if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
			return chamarGemini("gemini-1.5-flash", apiKey, mensagens)
		}
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			return chamarOpenAI("gpt-4o-mini", apiKey, mensagens)
		}
		return "", fmt.Errorf("erro de conexão com Ollama local em %s: %v. Certifique-se de que o Ollama está rodando.", host, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro retornado pelo Ollama (Status %d): %s", resp.StatusCode, string(respBytes))
	}

	var response struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Message.Content, nil
}

// chamarGemini envia requisições para a API oficial do Google Gemini utilizando a versão v1beta.
func chamarGemini(modelo, apiKey string, mensagens []Mensagem) (string, error) {
	if modelo == "" || modelo == "llama3" || modelo == "llama2" {
		modelo = "gemini-1.5-flash"
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelo, apiKey)

	type Part struct {
		Text string `json:"text"`
	}
	type Content struct {
		Role  string `json:"role"` // 'user' ou 'model'
		Parts []Part `json:"parts"`
	}

	var contents []Content
	for _, msg := range mensagens {
		role := "user"
		if msg.Role == "model" || msg.Role == "assistant" {
			role = "model"
		}
		content := msg.Content
		if msg.Role == "system" {
			content = "[Instruções do Sistema: " + content + "]\n"
		}
		contents = append(contents, Content{
			Role:  role,
			Parts: []Part{{Text: content}},
		})
	}

	payload := map[string]interface{}{
		"contents": contents,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro retornado pelo Gemini (Status %d): %s", resp.StatusCode, string(respBytes))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		return response.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("nenhuma resposta gerada pelo Gemini")
}

// chamarOpenAI envia requisições estruturadas para o endpoint Chat Completions da OpenAI.
func chamarOpenAI(modelo, apiKey string, mensagens []Mensagem) (string, error) {
	if modelo == "" || modelo == "llama3" || modelo == "llama2" {
		modelo = "gpt-4o-mini"
	}
	url := "https://api.openai.com/v1/chat/completions"

	payload := map[string]interface{}{
		"model":    modelo,
		"messages": mensagens,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro retornado pelo OpenAI (Status %d): %s", resp.StatusCode, string(respBytes))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("nenhuma resposta gerada pelo OpenAI")
}
