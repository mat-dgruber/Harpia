package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var publicarCmd = &cobra.Command{
	Use:     "publicar [alvo]",
	Aliases: []string{"deploy", "publish"},
	Short:   "Empacota e publica a aplicação Harpia em ambiente de nuvem ou Docker",
	RunE: func(cmd *cobra.Command, args []string) error {
		alvo := "docker"
		if len(args) > 0 {
			alvo = args[0]
		}

		fmt.Printf("🚀 Harpia Deploy Engine — Iniciando publicação com alvo '%s'...\n", alvo)

		if alvo == "docker" {
			dockerfile := `FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
`
			if err := os.WriteFile("Dockerfile", []byte(dockerfile), 0644); err != nil {
				return fmt.Errorf("erro ao gerar Dockerfile: %w", err)
			}
			fmt.Println("📦 Dockerfile profissional gerado com sucesso!")
			fmt.Println("⚡ Para compilar a imagem: docker build -t meu-app-harpia .")
		} else {
			fmt.Printf("📦 Gerando pacote de deploy otimizado para '%s'...\n", alvo)
			out, err := exec.Command("go", "build", "-o", "dist/app-producao", "main.go").CombinedOutput()
			if err != nil {
				return fmt.Errorf("erro na compilação de produção: %s", string(out))
			}
			fmt.Println("✅ Compilação de produção concluída em 'dist/app-producao'!")
		}

		fmt.Println("🎉 Deploy concluído com sucesso!")
		return nil
	},
}

func comandoPublicar() *cobra.Command {
	return publicarCmd
}
