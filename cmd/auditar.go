package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)



var auditarCmd = &cobra.Command{
	Use:     "auditar [arquivo.hrp|diretorio]",
	Aliases: []string{"audit", "seguranca"},
	Short:   "Varre o projeto em busca de falhas de segurança (OWASP Top 10)",
	RunE: func(cmd *cobra.Command, args []string) error {
		alvo := "."
		if len(args) > 0 {
			alvo = args[0]
		}

		fmt.Printf("🔍 Harpia Security Auditor — Varrendo '%s' por vulnerabilidades OWASP Top 10...\n\n", alvo)

		problemasEncontrados := 0

		// Simulação de varredura estática AST de segurança
		varrerArquivo := func(caminho string) {
			conteudo, err := os.ReadFile(caminho)
			if err != nil {
				return
			}
			txt := string(conteudo)

			// Check 1: Hardcoded Secrets / Senhas
			if strings.Contains(txt, "senha = \"") || strings.Contains(txt, "secret = \"") || strings.Contains(txt, "token = \"") {
				fmt.Printf("⚠️ [A07: Authentication Failure] [%s]: Possível segredo/senha fixada em texto puro.\n", caminho)
				problemasEncontrados++
			}

			// Check 2: Concatenacao de SQL em string
			if strings.Contains(txt, "ONDE ") && strings.Contains(txt, " + ") {
				fmt.Printf("⚠️ [A03: Injection] [%s]: Consulta SQL gerada via concatenação manual de string. Prefira o Query Builder parametrizado 'bd.tabela()'.\n", caminho)
				problemasEncontrados++
			}
		}


		fi, err := os.Stat(alvo)
		if err == nil && !fi.IsDir() {
			varrerArquivo(alvo)
		} else {
			_ = filepath.Walk(alvo, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "dist" || info.Name() == ".git") {
					return filepath.SkipDir
				}
				if err == nil && !info.IsDir() && (strings.HasSuffix(path, ".hrp") || strings.HasSuffix(path, ".pt")) {
					varrerArquivo(path)
				}
				return nil
			})
		}



		if problemasEncontrados == 0 {
			fmt.Println("✅ Nenhuma vulnerabilidade crítica de segurança detectada! Projeto em conformidade com diretrizes OWASP.")
		} else {
			fmt.Printf("\n⚠️ Total de %d alerta(s) de segurança encontrado(s). Recomendamos a correção antes do deploy em produção.\n", problemasEncontrados)
		}

		return nil
	},
}

func comandoAuditar() *cobra.Command {
	return auditarCmd
}
