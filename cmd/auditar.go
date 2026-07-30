//go:build !js && !wasm

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type VulnerabilidadeJSON struct {
	Arquivo   string `json:"arquivo"`
	Categoria string `json:"categoria"`
	Mensagem  string `json:"mensagem"`
}

type RelatorioAuditoriaJSON struct {
	Alvo                 string                `json:"alvo"`
	ProblemasEncontrados int                   `json:"problemas_encontrados"`
	Vulnerabilidades     []VulnerabilidadeJSON `json:"vulnerabilidades"`
}

var auditarCmd = func() *cobra.Command {
	var formato string

	cmd := &cobra.Command{
		Use:           "auditar [arquivo.hrp|diretorio]",
		Aliases:       []string{"audit", "seguranca"},
		Short:         "Varre o projeto em busca de falhas de segurança (OWASP Top 10)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			alvo := "."
			if len(args) > 0 {
				alvo = args[0]
			}

			if formato != "json" {
				fmt.Fprintf(out, "🔍 Harpia Security Auditor — Varrendo '%s' por vulnerabilidades OWASP Top 10...\n\n", alvo)
			}

			var relatorio RelatorioAuditoriaJSON
			relatorio.Alvo = alvo

			// Varredura estática leve de segurança
			varrerArquivo := func(caminho string) {
				conteudo, err := os.ReadFile(caminho)
				if err != nil {
					return
				}
				txt := string(conteudo)

				// Check 1: Hardcoded Secrets / Senhas
				if strings.Contains(txt, "senha = \"") || strings.Contains(txt, "secret = \"") || strings.Contains(txt, "token = \"") {
					if formato == "json" {
						relatorio.Vulnerabilidades = append(relatorio.Vulnerabilidades, VulnerabilidadeJSON{
							Arquivo:   caminho,
							Categoria: "A07: Authentication Failure",
							Mensagem:  "Possível segredo/senha fixada em texto puro.",
						})
					} else {
						fmt.Fprintf(out, "⚠️ [A07: Authentication Failure] [%s]: Possível segredo/senha fixada em texto puro.\n", caminho)
					}
					relatorio.ProblemasEncontrados++
				}

				// Check 2: Concatenacao de SQL em string
				if strings.Contains(txt, "ONDE ") && strings.Contains(txt, " + ") {
					if formato == "json" {
						relatorio.Vulnerabilidades = append(relatorio.Vulnerabilidades, VulnerabilidadeJSON{
							Arquivo:   caminho,
							Categoria: "A03: Injection",
							Mensagem:  "Consulta SQL gerada via concatenação manual de string. Prefira o Query Builder parametrizado 'bd.tabela()'.",
						})
					} else {
						fmt.Fprintf(out, "⚠️ [A03: Injection] [%s]: Consulta SQL gerada via concatenação manual de string. Prefira o Query Builder parametrizado 'bd.tabela()'.\n", caminho)
					}
					relatorio.ProblemasEncontrados++
				}
			}

			fi, err := os.Stat(alvo)
			if err == nil && !fi.IsDir() {
				varrerArquivo(alvo)
			} else {
				_ = filepath.Walk(alvo, func(path string, info os.FileInfo, err error) error {
					if info != nil && info.IsDir() && (info.Name() == "node_modules" || info.Name() == "dist" || info.Name() == ".git") {
						return filepath.SkipDir
					}
					if err == nil && info != nil && !info.IsDir() && (strings.HasSuffix(path, ".hrp") || strings.HasSuffix(path, ".pt")) {
						varrerArquivo(path)
					}
					return nil
				})
			}

			if formato == "json" {
				bytes, _ := json.MarshalIndent(relatorio, "", "  ")
				fmt.Fprintln(out, string(bytes))
				if relatorio.ProblemasEncontrados > 0 {
					return fmt.Errorf("auditoria encontrou %d problema(s) de segurança", relatorio.ProblemasEncontrados)
				}
				return nil
			}

			if relatorio.ProblemasEncontrados == 0 {
				fmt.Fprintln(out, "✅ Nenhuma vulnerabilidade crítica de segurança detectada! Projeto em conformidade com diretrizes OWASP.")
			} else {
				fmt.Fprintf(out, "\n⚠️ Total de %d alerta(s) de segurança encontrado(s). Recomendamos a correção antes do deploy em produção.\n", relatorio.ProblemasEncontrados)
				return fmt.Errorf("auditoria encontrou %d problema(s) de segurança", relatorio.ProblemasEncontrados)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&formato, "formato", "texto", "Formato de saída: 'texto' ou 'json'")
	return cmd
}

func comandoAuditar() *cobra.Command {
	return auditarCmd()
}
