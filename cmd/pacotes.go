//go:build !js && !wasm

package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// calcularSHA256 calcula o hash de integridade SHA-256 dos bytes fornecidos
// e retorna a string hexadecimal resultante de 64 caracteres em minúsculas.
func calcularSHA256(dados []byte) string {
	hash := sha256.Sum256(dados)
	return fmt.Sprintf("%x", hash)
}

// obterCaminhoELockfile busca por arquivos de lockfile existentes no diretório atual
// seguindo a ordem de preferência: harpia.lock, harpia.lock.json, pacote.lock, pacote.lock.json.
// Se nenhum arquivo for encontrado, retorna o caminho padrão "harpia.lock" e uma nova estrutura Lockfile zerada.
func obterCaminhoELockfile() (string, *Lockfile) {
	caminhos := []string{"harpia.lock", "harpia.lock.json", "pacote.lock", "pacote.lock.json"}
	for _, caminho := range caminhos {
		if conteudo, err := os.ReadFile(caminho); err == nil {
			if lock, err := parseLockfile(conteudo); err == nil {
				if lock.Pacotes == nil {
					lock.Pacotes = make(map[string]PacoteLockInfo)
				}
				return caminho, lock
			}
		}
	}
	return "harpia.lock", &Lockfile{
		VersaoLock: "1.0",
		Pacotes:    make(map[string]PacoteLockInfo),
	}
}

// PacoteManifest representa o arquivo de manifesto pacote.hrp ou pacote.json
// mapeando o conjunto de dependências externas declaradas pelo projeto.
type PacoteManifest struct {
	Dependencias map[string]string `json:"dependencias"`
}

// comandoInstalar inicializa e retorna o subcomando Cobra 'instalar' (`harpia instalar`).
// Este comando resolve e instala de forma concorrente e assíncrona dependências locais
// em ./pt_modulos/ buscando metadados no índice de pacotes centralizado.
func comandoInstalar() *cobra.Command {
	var deArquivo string
	cmdInstalar := &cobra.Command{
		Use:     "instalar [pacote-opcional] [url-ou-versao-opcional]",
		Aliases: []string{"instale", "install"},
		Short:   "Instala dependências do projeto registradas no pacote.hrp ou pacote.json",
		Run: func(cmd *cobra.Command, args []string) {
			caminhoLock, lock := obterCaminhoELockfile()

			// Se o usuário passou argumentos diretamente, instala o pacote específico
			if len(args) == 2 {
				nome := args[0]
				urlOuVersao := args[1]
				urlInstalar := urlOuVersao
				shaEsperado := ""

				if lockInfo, emLock := lock.Pacotes[nome]; emLock {
					if lockInfo.Versao == urlOuVersao && lockInfo.URL != "" {
						urlInstalar = lockInfo.URL
						shaEsperado = lockInfo.SHA256
					}
				}

				if !strings.HasPrefix(urlInstalar, "http://") && !strings.HasPrefix(urlInstalar, "https://") {
					fmt.Printf("Buscando pacote '%s' na versão '%s' no registro remoto...\n", nome, urlOuVersao)
					resolved, err := obterUrlDoRegistro(nome, urlOuVersao)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Erro ao resolver pacote: %v\n", err)
						os.Exit(1)
					}
					urlInstalar = resolved
				}

				fmt.Printf("Instalando pacote '%s' de %s...\n", nome, urlInstalar)
				hashCalculado, err := baixarEExtrairPacote(nome, urlInstalar, shaEsperado)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao instalar %s: %v\n", nome, err)
					os.Exit(1)
				}
				salvarDependenciaNoManifesto(nome, urlOuVersao)
				lock.Pacotes[nome] = PacoteLockInfo{
					Versao: urlOuVersao,
					URL:    urlInstalar,
					SHA256: hashCalculado,
				}
				salvarLockfile(caminhoLock, lock)
				return
			}

			if len(args) == 1 {
				nome := args[0]
				urlInstalar := ""
				shaEsperado := ""
				versaoRegistrada := ""

				if lockInfo, emLock := lock.Pacotes[nome]; emLock && lockInfo.URL != "" {
					urlInstalar = lockInfo.URL
					shaEsperado = lockInfo.SHA256
					versaoRegistrada = lockInfo.Versao
				} else {
					fmt.Printf("Buscando pacote '%s' no registro remoto...\n", nome)
					urlPacote, err := obterUrlDoRegistro(nome, "")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Erro ao resolver pacote: %v\n", err)
						os.Exit(1)
					}
					urlInstalar = urlPacote
					versaoRegistrada = urlPacote
				}

				hashCalculado, err := baixarEExtrairPacote(nome, urlInstalar, shaEsperado)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao instalar %s: %v\n", nome, err)
					os.Exit(1)
				}
				salvarDependenciaNoManifesto(nome, versaoRegistrada)
				lock.Pacotes[nome] = PacoteLockInfo{
					Versao: versaoRegistrada,
					URL:    urlInstalar,
					SHA256: hashCalculado,
				}
				salvarLockfile(caminhoLock, lock)
				return
			}

			// Caso contrário, lê o manifesto
			manifestoPath := "pacote.hrp"
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

			type resInstalacao struct {
				nome   string
				versao string
				url    string
				sha256 string
				err    error
			}

			canalResultados := make(chan resInstalacao, len(manifesto.Dependencias))
			for nome, urlOuVersao := range manifesto.Dependencias {
				go func(n, uv string) {
					urlInstalar := uv
					shaEsperado := ""

					if lockInfo, emLock := lock.Pacotes[n]; emLock && lockInfo.URL != "" {
						urlInstalar = lockInfo.URL
						shaEsperado = lockInfo.SHA256
					} else if !strings.HasPrefix(uv, "http://") && !strings.HasPrefix(uv, "https://") {
						resolved, err := obterUrlDoRegistro(n, uv)
						if err != nil {
							canalResultados <- resInstalacao{err: fmt.Errorf("falha ao resolver '%s' na versão '%s': %v", n, uv, err)}
							return
						}
						urlInstalar = resolved
					}

					hashCalculado, err := baixarEExtrairPacote(n, urlInstalar, shaEsperado)
					if err != nil {
						canalResultados <- resInstalacao{err: fmt.Errorf("falha no pacote %s: %v", n, err)}
						return
					}

					canalResultados <- resInstalacao{
						nome:   n,
						versao: uv,
						url:    urlInstalar,
						sha256: hashCalculado,
						err:    nil,
					}
				}(nome, urlOuVersao)
			}

			erros := 0
			for i := 0; i < len(manifesto.Dependencias); i++ {
				res := <-canalResultados
				if res.err != nil {
					fmt.Fprintf(os.Stderr, "Erro: %v\n", res.err)
					erros++
				} else {
					lock.Pacotes[res.nome] = PacoteLockInfo{
						Versao: res.versao,
						URL:    res.url,
						SHA256: res.sha256,
					}
				}
			}

			if erros > 0 {
				fmt.Fprintf(os.Stderr, "Instalação concluída com %d falhas.\n", erros)
				os.Exit(1)
			}

			if err := salvarLockfile(caminhoLock, lock); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao salvar lockfile: %v\n", err)
			}

			fmt.Println("✅ Todas as dependências foram instaladas com sucesso em ./pt_modulos/")
		},
	}
	cmdInstalar.Flags().StringVarP(&deArquivo, "arquivo", "f", "", "Caminho do arquivo de manifesto personalizado")
	return cmdInstalar
}

// parseManifesto interpreta o conteúdo bruto do manifesto de dependências,
// suportando tanto o formato padrão estruturado JSON quanto uma sintaxe simplificada de linhas chave-valor.
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

// baixarEExtrairPacote baixa síncronamente o arquivo comprimido (.zip) da dependência,
// calcula e valida a integridade via checksum SHA-256 se fornecido, descompacta e extrai seus
// arquivos na pasta local './pt_modulos/', aplicando validação contra Zip Slip.
func baixarEExtrairPacote(nome, url string, sha256Esperado string) (string, error) {
	// ponytail: relatórios textuais dinâmicos de progresso para DX amigável
	fmt.Printf("  ➔ [%s] Baixando dependência de %s...\n", nome, url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status HTTP inválido: %s", resp.Status)
	}

	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	hashObtido := calcularSHA256(corpo)
	if sha256Esperado != "" && sha256Esperado != hashObtido {
		return "", fmt.Errorf("HRP-LOCK-001: Falha na verificação de integridade SHA-256 do pacote %s (esperado: %s, obtido: %s)", nome, sha256Esperado, hashObtido)
	}

	fmt.Printf("  ➔ [%s] Extraindo arquivos na pasta local...\n", nome)
	reader, err := zip.NewReader(bytes.NewReader(corpo), int64(len(corpo)))
	if err != nil {
		return "", fmt.Errorf("arquivo baixado não é um zip válido: %v", err)
	}

	pastaDest := filepath.Join("pt_modulos", nome)
	err = os.MkdirAll(pastaDest, 0755)
	if err != nil {
		return "", err
	}

	for _, arquivo := range reader.File {
		caminhoFisico := filepath.Join(pastaDest, arquivo.Name)

		// ponytail: impede a vulnerabilidade crítica de Zip Slip (path traversal)
		caminhoLimpo := filepath.Clean(caminhoFisico)
		pastaLimpa := filepath.Clean(pastaDest)
		if !strings.HasPrefix(caminhoLimpo, pastaLimpa) {
			return "", fmt.Errorf("caminho de arquivo ilegal detectado no pacote ZIP: %s", arquivo.Name)
		}

		if arquivo.FileInfo().IsDir() {
			os.MkdirAll(caminhoFisico, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(caminhoFisico), 0755); err != nil {
			return "", err
		}

		origem, err := arquivo.Open()
		if err != nil {
			return "", err
		}

		destino, err := os.OpenFile(caminhoFisico, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, arquivo.Mode())
		if err != nil {
			origem.Close()
			return "", err
		}

		_, err = io.Copy(destino, origem)
		origem.Close()
		destino.Close()
		if err != nil {
			return "", err
		}
	}

	fmt.Printf("  ✓ [%s] Instalado com sucesso em ./pt_modulos/\n", nome)
	return hashObtido, nil
}

// ponytail: URL oficial contendo os metadados centralizados do índice de pacotes
var URL_REGISTRO_CENTRAL = "https://raw.githubusercontent.com/Harpia/registro/main/pacotes.json"

// RegistroRemoto descreve a árvore e os metadados das dependências retornadas
// pelo endpoint centralizado de pacotes oficiais do Harpia.
type RegistroRemoto struct {
	Pacotes map[string]struct {
		Versoes map[string]struct {
			URL string `json:"url"`
		} `json:"versoes"`
	} `json:"pacotes"`
}

// obterUrlDoRegistro faz uma requisição HTTP GET ao registro central remoto, localiza
// o pacote solicitado e resolve a URL do zip correspondente à versão pedida ou à mais recente.
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

// salvarDependenciaNoManifesto atualiza ou cria o manifesto pacote.json gravando de
// forma persistente a nova dependência instalada com seu respectivo nome e versão.
func salvarDependenciaNoManifesto(nome, versaoOuUrl string) {
	manifestoPath := "pacote.json"
	if _, err := os.Stat("pacote.hrp"); err == nil {
		manifestoPath = "pacote.hrp"
	}

	var manifest *PacoteManifest
	if conteudo, err := os.ReadFile(manifestoPath); err == nil {
		manifest, _ = parseManifesto(conteudo)
	}
	if manifest == nil {
		manifest = &PacoteManifest{}
	}
	if manifest.Dependencias == nil {
		manifest.Dependencias = make(map[string]string)
	}

	manifest.Dependencias[nome] = versaoOuUrl
	bytes, err := json.MarshalIndent(manifest, "", "  ")
	if err == nil {
		os.WriteFile("pacote.json", bytes, 0644)
		fmt.Printf("  ✓ Dependência '%s' salva em pacote.json\n", nome)
	}
}

// removerDependenciaDoManifesto remove a chave do pacote especificado de pacote.json ou pacote.hrp.
func removerDependenciaDoManifesto(nome string) {
	if _, err := os.Stat("pacote.json"); err == nil {
		if conteudo, err := os.ReadFile("pacote.json"); err == nil {
			var manifest PacoteManifest
			if err := json.Unmarshal(conteudo, &manifest); err == nil && manifest.Dependencias != nil {
				delete(manifest.Dependencias, nome)
				if bytes, err := json.MarshalIndent(manifest, "", "  "); err == nil {
					os.WriteFile("pacote.json", bytes, 0644)
				}
			}
		}
	}
	if _, err := os.Stat("pacote.hrp"); err == nil {
		if conteudo, err := os.ReadFile("pacote.hrp"); err == nil {
			linhas := strings.Split(string(conteudo), "\n")
			var novasLinhas []string
			for _, l := range linhas {
				lTrim := strings.TrimSpace(l)
				if strings.HasPrefix(lTrim, "var "+nome) || strings.HasPrefix(lTrim, "const "+nome) || strings.HasPrefix(lTrim, nome+" =") || strings.HasPrefix(lTrim, nome+"=") {
					continue
				}
				novasLinhas = append(novasLinhas, l)
			}
			os.WriteFile("pacote.hrp", []byte(strings.Join(novasLinhas, "\n")), 0644)
		}
	}
}

// comandoPacote retorna o comando principal Cobra 'pacote' com seus subcomandos de ciclo de vida.
func comandoPacote() *cobra.Command {
	cmdPacote := &cobra.Command{
		Use:     "pacote",
		Aliases: []string{"pacotes", "pkg"},
		Short:   "Gerencia o ciclo de vida dos pacotes e dependências do projeto",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmdAdicionar := &cobra.Command{
		Use:     "adicionar <pacote> [versao-ou-url]",
		Aliases: []string{"add"},
		Short:   "Adiciona e instala uma nova dependência no projeto",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Println("Erro: informe o nome do pacote para adicionar.")
				return
			}
			nome := args[0]
			versaoOuUrl := ""
			if len(args) >= 2 {
				versaoOuUrl = args[1]
			}

			caminhoLock, lock := obterCaminhoELockfile()
			urlInstalar := versaoOuUrl
			shaEsperado := ""
			versaoRegistrada := versaoOuUrl

			if versaoOuUrl == "" {
				if lockInfo, emLock := lock.Pacotes[nome]; emLock && lockInfo.URL != "" {
					urlInstalar = lockInfo.URL
					shaEsperado = lockInfo.SHA256
					versaoRegistrada = lockInfo.Versao
				} else {
					cmd.Printf("Buscando pacote '%s' no registro remoto...\n", nome)
					urlPacote, err := obterUrlDoRegistro(nome, "")
					if err != nil {
						cmd.PrintErrf("Erro ao resolver pacote: %v\n", err)
						return
					}
					urlInstalar = urlPacote
					versaoRegistrada = urlPacote
				}
			} else if !strings.HasPrefix(versaoOuUrl, "http://") && !strings.HasPrefix(versaoOuUrl, "https://") {
				cmd.Printf("Buscando pacote '%s' na versão '%s' no registro remoto...\n", nome, versaoOuUrl)
				resolved, err := obterUrlDoRegistro(nome, versaoOuUrl)
				if err != nil {
					cmd.PrintErrf("Erro ao resolver pacote: %v\n", err)
					return
				}
				urlInstalar = resolved
			}

			hashCalculado, err := baixarEExtrairPacote(nome, urlInstalar, shaEsperado)
			if err != nil {
				cmd.PrintErrf("Erro ao instalar %s: %v\n", nome, err)
				return
			}

			salvarDependenciaNoManifesto(nome, versaoRegistrada)
			lock.Pacotes[nome] = PacoteLockInfo{
				Versao: versaoRegistrada,
				URL:    urlInstalar,
				SHA256: hashCalculado,
			}
			if err := salvarLockfile(caminhoLock, lock); err != nil {
				cmd.PrintErrf("Erro ao salvar lockfile: %v\n", err)
			}
			cmd.Printf("✓ Pacote '%s' adicionado com sucesso.\n", nome)
		},
	}

	cmdRemover := &cobra.Command{
		Use:     "remover <pacote>",
		Aliases: []string{"remove", "rm"},
		Short:   "Remove uma dependência instalada e limpa os registros",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Println("Erro: informe o nome do pacote para remover.")
				return
			}
			nome := args[0]

			pastaModulo := filepath.Join("pt_modulos", nome)
			os.RemoveAll(pastaModulo)

			removerDependenciaDoManifesto(nome)

			caminhoLock, lock := obterCaminhoELockfile()
			delete(lock.Pacotes, nome)
			salvarLockfile(caminhoLock, lock)

			cmd.Printf("✓ Pacote '%s' removido com sucesso.\n", nome)
		},
	}

	cmdListar := &cobra.Command{
		Use:     "listar",
		Aliases: []string{"list", "ls"},
		Short:   "Lista todas as dependências registradas e instaladas no projeto",
		Run: func(cmd *cobra.Command, args []string) {
			_, lock := obterCaminhoELockfile()
			if len(lock.Pacotes) == 0 {
				cmd.Println("Nenhum pacote instalado.")
				return
			}

			cmd.Printf("Pacotes instalados (%d):\n", len(lock.Pacotes))
			for nome, info := range lock.Pacotes {
				versao := info.Versao
				if versao == "" {
					versao = "desconhecida"
				}
				cmd.Printf("  • %s (%s)\n", nome, versao)
			}
		},
	}

	cmdLimpar := &cobra.Command{
		Use:     "limpar",
		Aliases: []string{"clean"},
		Short:   "Remove todos os pacotes baixados na pasta local pt_modulos/",
		Run: func(cmd *cobra.Command, args []string) {
			os.RemoveAll("pt_modulos")
			cmd.Println("✓ Cache e pasta pt_modulos/ limpos com sucesso.")
		},
	}

	cmdPacote.AddCommand(cmdAdicionar, cmdRemover, cmdListar, cmdLimpar)
	return cmdPacote
}

