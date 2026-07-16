package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type DapMessage struct {
	Seq      int    `json:"seq"`
	Type     string `json:"type"`
	Command  string `json:"command,omitempty"`
	Event    string `json:"event,omitempty"`
	Request  string `json:"request_seq,omitempty"`
	Success  bool   `json:"success,omitempty"`
	Body     any    `json:"body,omitempty"`
}

func comandoDepurar() *cobra.Command {
	var porta int

	depurar := &cobra.Command{
		Use:   "depurar",
		Short: "Inicia o servidor de depuração Debug Adapter Protocol (DAP)",
		Run: func(cmd *cobra.Command, args []string) {
			endereco := fmt.Sprintf("127.0.0.1:%d", porta)
			listener, err := net.Listen("tcp", endereco)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao iniciar servidor DAP: %v\n", err)
				os.Exit(1)
			}
			defer listener.Close()

			fmt.Printf("🐞 Servidor DAP Portuscript rodando em: %s\n", endereco)
			fmt.Println("Pronto para conexões de IDEs/Editores (VS Code, etc.). Press ESC/Ctrl+C para parar.")

			for {
				conn, err := listener.Accept()
				if err != nil {
					continue
				}

				go lidarComDap(conn)
			}
		},
	}

	depurar.Flags().IntVarP(&porta, "porta", "p", 4711, "Porta TCP para o servidor DAP.")

	return depurar
}

func lidarComDap(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	// Lida com handshake simples do protocolo DAP
	for scanner.Scan() {
		linha := scanner.Text()
		if strings.HasPrefix(linha, "Content-Length:") {
			// Lê corpo da mensagem
			scanner.Scan() // pula linha em branco
			if scanner.Scan() {
				corpo := scanner.Text()
				var msg DapMessage
				if err := json.Unmarshal([]byte(corpo), &msg); err == nil {
					fmt.Printf("[DAP] Recebido comando: %s\n", msg.Command)
					
					// Resposta de exemplo/handshake de sucesso
					resposta := DapMessage{
						Seq:     msg.Seq + 1,
						Type:    "response",
						Request: fmt.Sprintf("%d", msg.Seq),
						Success: true,
					}
					
					switch msg.Command {
					case "initialize":
						resposta.Body = map[string]any{
							"supportsConfigurationDoneRequest": true,
							"supportsStepBack":                 false,
						}
					}
					
					respBytes, _ := json.Marshal(resposta)
					header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(respBytes))
					_, _ = conn.Write([]byte(header + string(respBytes)))
				}
			}
		}
	}
}
