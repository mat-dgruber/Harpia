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

// comandoInstalar inicializa o subcomando 'Harpia instalar'
func comandoInstalar() *cobra.Command {
	var deArquivo string
	cmdInstalar := &cobra.Command{
		Use:     "instalar [pacote-opcional] [url-ou-versao-opcional]",
		Aliases: []string{"instale", "install"},
		Short:   "Instala dependências do projeto registradas no pacote.ptst ou pacote.json",
		Run: func(cmd *cobra.Command, args []string) {
			// Se o usuário passou argumentos diretamente, instala o pacote específico
			if len(args) == 2 {
				nome := args[0]
				urlOuVersao := args[1]
				urlInstalar := urlOuVersao

				if !strings.HasPrefix(urlOuVersao, "http://") && !strings.HasPrefix(urlOuVersao, "https://") {
					fmt.Printf("Buscando pacote '%s' na versão '%s' no registro remoto...\n", nome, urlOuVersao)
					resolved, err := obterUrlDoRegistro(nome, urlOuVersao)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Erro ao resolver pacote: %v\n", err)
						os.Exit(1)
					}
					urlInstalar = resolved
				}

				fmt.Printf("Instalando pacote '%s' de %s...\n", nome, urlInstalar)
				err := baixarEExtrairPacote(nome, urlInstalar)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao instalar %s: %v\n", nome, err)
					os.Exit(1)
				}
				return
			}

			if len(args) == 1 {
				nome := args[0]
				fmt.Printf("Buscando pacote '%s' no registro remoto...\n", nome)
				urlPacote, err := obterUrlDoRegistro(nome, "")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao resolver pacote: %v\n", err)
					os.Exit(1)
				}
				err = baixarEExtrairPacote(nome, urlPacote)
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
			for nome, urlOuVersao := range manifesto.Dependencias {
				go func(n, uv string) {
					urlInstalar := uv
					if !strings.HasPrefix(uv, "http://") && !strings.HasPrefix(uv, "https://") {
						resolved, err := obterUrlDoRegistro(n, uv)
						if err != nil {
							canalErros <- fmt.Errorf("falha ao resolver '%s' na versão '%s': %v", n, uv, err)
							return
						}
						urlInstalar = resolved
					}

					err := baixarEExtrairPacote(n, urlInstalar)
					if err != nil {
						canalErros <- fmt.Errorf("falha no pacote %s: %v", n, err)
						return
					}
					canalErros <- nil
				}(nome, urlOuVersao)
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
			// Limpa declarações do Harpia como 'var' ou 'const' se houver
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
	// ponytail: relatórios textuais dinâmicos de progresso para DX amigável
	fmt.Printf("  ➔ [%s] Baixando dependência de %s...\n", nome, url)
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

	fmt.Printf("  ➔ [%s] Extraindo arquivos na pasta local...\n", nome)
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

		// ponytail: impede a vulnerabilidade crítica de Zip Slip (path traversal)
		caminhoLimpo := filepath.Clean(caminhoFisico)
		pastaLimpa := filepath.Clean(pastaDest)
		if !strings.HasPrefix(caminhoLimpo, pastaLimpa) {
			return fmt.Errorf("caminho de arquivo ilegal detectado no pacote ZIP: %s", arquivo.Name)
		}

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

	fmt.Printf("  ✓ [%s] Instalado com sucesso em ./pt_modulos/\n", nome)
	return nil
}

// ponytail: URL oficial contendo os metadados centralizados do índice de pacotes
var URL_REGISTRO_CENTRAL = "https://raw.githubusercontent.com/Harpia/registro/main/pacotes.json"

type RegistroRemoto struct {
	Pacotes map[string]struct {
		Versoes map[string]struct {
			URL string `json:"url"`
		} `json:"versoes"`
	} `json:"pacotes"`
}

func obterUrlDoRegistro(nome, versaoRestricao string) (string, error) {
	resp, err := http.Get(URL_REGISTRO_CENTRAL)
	if err != nil {
		return "", fmt.Errorf("falha ao se comunicar com o registro remoto de pacotes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("falha no registro remoto (status %s)", resp.Status)
	}

	var reg RegistroRemoto
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return "", fmt.Errorf("registro remoto retornou dados inválidos: %v", err)
	}

	pacote, existe := reg.Pacotes[nome]
	if !existe {
		return "", fmt.Errorf("pacote '%s' não encontrado no registro central", nome)
	}

	// Se nenhuma versão foi solicitada ou pediu "ultimo"/"latest", pega a última versão declarada
	if versaoRestricao == "" || versaoRestricao == "latest" || versaoRestricao == "ultimo" {
		var ultimaURL string
		for _, v := range pacote.Versoes {
			ultimaURL = v.URL
		}
		if ultimaURL == "" {
			return "", fmt.Errorf("pacote '%s' não tem nenhuma versão disponível no registro", nome)
		}
		return ultimaURL, nil
	}

	// Busca a versão exata que foi solicitada
	v, existe := pacote.Versoes[versaoRestricao]
	if !existe {
		// ponytail: fallback simples para busca parcial se não achar exata
		for versao, dados := range pacote.Versoes {
			if strings.HasPrefix(versao, versaoRestricao) {
				return dados.URL, nil
			}
		}
		return "", fmt.Errorf("versão '%s' do pacote '%s' não está registrada", versaoRestricao, nome)
	}

	return v.URL, nil
}

