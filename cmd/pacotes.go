package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// PacoteManifest representa o arquivo pacote.ptst/json
type PacoteManifest struct {
	Dependencias map[string]string `json:"dependencias"`
}

// comandoInstalar inicializa o subcomando 'portuscript instalar'
func comandoInstalar() *cobra.Command {
	var deArquivo string
	cmdInstalar := &cobra.Command{
		Use:     "instalar [pacote-opcional] [url-opcional]",
		Aliases: []string{"instale", "install"},
		Short:   "Instala dependências do projeto registradas no pacote.ptst ou pacote.json",
		Run: func(cmd *cobra.Command, args []string) {
			// Se o usuário passou argumentos diretamente, instala o pacote específico
			if len(args) == 2 {
				nome := args[0]
				url := args[1]
				fmt.Printf("Instalando pacote '%s' de %s...\n", nome, url)
				err := baixarEExtrairPacote(nome, url)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao instalar %s: %v\n", nome, err)
					os.Exit(1)
				}
				return
			}

			// Caso contrário, lê o manifesto
			manifestoPath := "pacote.ptst"
			if _, err := os.Stat("pacote.json"); err == nil {
				manifestoPath = "pacote.json"
			}
			if deArquivo != "" {
				manifestoPath = deArquivo
			}

			conteudo, err := os.ReadFile(manifestoPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Nenhum arquivo de manifesto de dependências encontrado (%s): %v\n", manifestoPath, err)
				os.Exit(1)
			}

			manifesto, err := parseManifesto(conteudo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao interpretar manifesto %s: %v\n", manifestoPath, err)
				os.Exit(1)
			}

			if len(manifesto.Dependencias) == 0 {
				fmt.Println("Nenhuma dependência registrada no manifesto.")
				return
			}

			fmt.Printf("Instalando %d dependências de %s...\n", len(manifesto.Dependencias), manifestoPath)

			// ponytail: executa as instalações de forma assíncrona
			canalErros := make(chan error, len(manifesto.Dependencias))
			for nome, url := range manifesto.Dependencias {
				go func(n, u string) {
					err := baixarEExtrairPacote(n, u)
					if err != nil {
						canalErros <- fmt.Errorf("falha no pacote %s: %v", n, err)
						return
					}
					canalErros <- nil
				}(nome, url)
			}

			// Aguarda a finalização de todas as goroutines
			erros := 0
			for i := 0; i < len(manifesto.Dependencias); i++ {
				if err := <-canalErros; err != nil {
					fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
					erros++
				}
			}

			if erros > 0 {
				fmt.Fprintf(os.Stderr, "Instalação concluída com %d falhas.\n", erros)
				os.Exit(1)
			}

			fmt.Println("✅ Todas as dependências foram instaladas com sucesso em ./pt_modulos/")
		},
	}
	cmdInstalar.Flags().StringVarP(&deArquivo, "arquivo", "f", "", "Caminho do arquivo de manifesto personalizado")
	return cmdInstalar
}

func parseManifesto(conteudo []byte) (*PacoteManifest, error) {
	var manifest PacoteManifest
	// Tenta fazer o parse do JSON direto
	if err := json.Unmarshal(conteudo, &manifest); err == nil {
		return &manifest, nil
	}

	// Caso falhe, tenta fazer parse linha-a-linha de formato simples key-value
	// Ex: dependencias = { modulo: "url" } ou nome = "url"
	linhas := strings.Split(string(conteudo), "\n")
	manifest.Dependencias = make(map[string]string)
	for _, linha := range linhas {
		linha = strings.TrimSpace(linha)
		if linha == "" || strings.HasPrefix(linha, "#") || strings.HasPrefix(linha, "//") {
			continue
		}
		// Se contiver '=' e não for a declaração dependencias = {
		if strings.Contains(linha, "=") {
			partes := strings.SplitN(linha, "=", 2)
			chave := strings.TrimSpace(partes[0])
			valor := strings.TrimSpace(partes[1])
			// Limpa aspas do valor
			if len(valor) >= 2 && ((valor[0] == '"' && valor[len(valor)-1] == '"') || (valor[0] == '\'' && valor[len(valor)-1] == '\'')) {
				valor = valor[1 : len(valor)-1]
			}
			// Limpa declarações do portuscript como 'var' ou 'const' se houver
			chave = strings.TrimPrefix(chave, "var ")
			chave = strings.TrimPrefix(chave, "const ")
			chave = strings.TrimSpace(chave)
			if chave != "" && valor != "" && !strings.Contains(chave, "{") {
				manifest.Dependencias[chave] = valor
			}
		}
	}
	return &manifest, nil
}

func baixarEExtrairPacote(nome, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status HTTP inválido: %s", resp.Status)
	}

	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(corpo), int64(len(corpo)))
	if err != nil {
		return fmt.Errorf("arquivo baixado não é um zip válido: %v", err)
	}

	pastaDest := filepath.Join("pt_modulos", nome)
	err = os.MkdirAll(pastaDest, 0755)
	if err != nil {
		return err
	}

	for _, arquivo := range reader.File {
		caminhoFisico := filepath.Join(pastaDest, arquivo.Name)
		if arquivo.FileInfo().IsDir() {
			os.MkdirAll(caminhoFisico, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(caminhoFisico), 0755); err != nil {
			return err
		}

		origem, err := arquivo.Open()
		if err != nil {
			return err
		}

		destino, err := os.OpenFile(caminhoFisico, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, arquivo.Mode())
		if err != nil {
			origem.Close()
			return err
		}

		_, err = io.Copy(destino, origem)
		origem.Close()
		destino.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
