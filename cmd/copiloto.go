package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var modeloCopiloto string
var ollamaURL string

type OllamaGenerateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaGenerateRes struct {
	Response string `json:"response"`
}

func SugerirCopiloto(contexto string) (string, error) {
	prompt := fmt.Sprintf(`Você é o Copiloto da linguagem Harpia (linguagem Full Stack brasileira com sintaxe em português, componentes reativos estilo JSX, OOP simples, tipagem opcional e APIs integradas).
Dado o trecho de código abaixo, complete-o de forma natural e idiomática usando a sintaxe oficial da Harpia.
IMPORTANTE: Retorne APENAS o código de complementação, sem comentários explicativos, sem tags de markdown, sem preâmbulo. Apenas a continuação direta do código.

Contexto de código:
%s`, contexto)

	reqBody := OllamaGenerateReq{
		Model:  modeloCopiloto,
		Prompt: prompt,
		Stream: false,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("Ollama offline em %s: %v", ollamaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro do Ollama (status %d)", resp.StatusCode)
	}

	var resObj OllamaGenerateRes
	if err := json.NewDecoder(resp.Body).Decode(&resObj); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta do Ollama: %v", err)
	}

	return resObj.Response, nil
}

func comandoCopiloto() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copiloto [arquivo]",
		Short: "Sugere complementação de código inteligente via modelo de IA local (Ollama)",
		Long:  "Lê o arquivo ou contexto do código fornecido e aciona o Ollama para autocompletar o código Harpia.",
		Run: func(cmd *cobra.Command, args []string) {
			var contexto string

			if len(args) > 0 {
				bytes, err := os.ReadFile(args[0])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao ler arquivo: %v\n", err)
					os.Exit(1)
				}
				contexto = string(bytes)
			} else {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao ler stdin: %v\n", err)
					os.Exit(1)
				}
				contexto = string(bytes)
			}

			sugestao, err := SugerirCopiloto(contexto)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Copiloto indisponível: %v\n", err)
				os.Exit(0)
			}

			fmt.Print(sugestao)
		},
	}

	cmd.Flags().StringVarP(&modeloCopiloto, "modelo", "m", "llama3", "Modelo do Ollama a ser utilizado")
	cmd.Flags().StringVarP(&ollamaURL, "url", "u", "http://localhost:11434/api/generate", "URL da API do Ollama local")

	return cmd
}
