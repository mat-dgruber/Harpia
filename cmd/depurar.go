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
	Seq     int    `json:"seq"`
	Type    string `json:"type"`
	Command string `json:"command,omitempty"`
	Event   string `json:"event,omitempty"`
	Request string `json:"request_seq,omitempty"`
	Success bool   `json:"success,omitempty"`
	Body    any    `json:"body,omitempty"`
}

// comandoDepurar instancia o servidor de Debug Adapter Protocol (DAP) do Harpia.
//
// O servidor escuta em `127.0.0.1:<porta>` (porta configurável via `--porta`,
// padrão 4711 seguindo convenção de mercado) e atende conexões de qualquer IDE
// ou editor compatível com o protocolo DAP (VS Code, Cursor, neovim, etc.).
//
// Modo de operação:
//   - Loop síncrono de `Accept()` em goroutine que despacha cada cliente para
//     `lidarComDap` em goroutine separada, permitindo múltiplos clientes simultâneos.
//   - Handshake inicial via Content-Length LSP-style: lê cabeçalho, consome
//     linha em branco, e deserializa o corpo JSON para `DapMessage`.
//   - Responde com sucesso padrão às mensagens `initialize`, expondo
//     `supportsConfigurationDoneRequest=true` e `supportsStepBack=false`.
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

			fmt.Printf("🐞 Servidor DAP Harpia rodando em: %s\n", endereco)
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

// lidarComDap processa a conexão TCP de uma única sessão DAP.
//
// Lê uma linha por vez do socket usando `bufio.Scanner` para evitar quebras em
// entradas grandes. Em cada iteração:
//  1. Localiza a linha `Content-Length:` para delimitar o tamanho do corpo JSON;
//  2. Pula a linha em branco subsequente;
//  3. Desserializa o corpo JSON em `DapMessage`;
//
// Ao final, monta a resposta (`Success=true`) com `Seq = request.Seq + 1`,
// serializa em JSON, prefixa com `Content-Length: <tamanho>\r\n\r\n` (conforme
// o protocolo LSP/DAP) e escreve de volta na conexão. A conexão é fechada
// automaticamente via `defer conn.Close()` no retorno de `comandoDepurar`.
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
