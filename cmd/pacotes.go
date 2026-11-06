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
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
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
					_, resolved, sha, err := obterPacoteDoRegistro(nome, urlOuVersao)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Erro ao resolver pacote: %v\n", err)
						os.Exit(1)
					}
					urlInstalar, shaEsperado = resolved, sha
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
				if err := salvarLockfile(caminhoLock, lock); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao salvar lockfile: %v\n", err)
					os.Exit(1)
				}
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
					versao, urlPacote, sha, err := obterPacoteDoRegistro(nome, "")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Erro ao resolver pacote: %v\n", err)
						os.Exit(1)
					}
					urlInstalar, shaEsperado, versaoRegistrada = urlPacote, sha, versao
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
				if err := salvarLockfile(caminhoLock, lock); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao salvar lockfile: %v\n", err)
					os.Exit(1)
				}
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
				// lido aqui, e não na goroutine: o laço abaixo escreve em lock.Pacotes ao mesmo tempo
				lockInfo, emLock := lock.Pacotes[nome]
				go func(n, uv string, lockInfo PacoteLockInfo, emLock bool) {
					urlInstalar := uv
					shaEsperado := ""

					if emLock && lockInfo.URL != "" {
						urlInstalar = lockInfo.URL
						shaEsperado = lockInfo.SHA256
					} else if !strings.HasPrefix(uv, "http://") && !strings.HasPrefix(uv, "https://") {
						_, resolved, sha, err := obterPacoteDoRegistro(n, uv)
						shaEsperado = sha
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
				}(nome, urlOuVersao, lockInfo, emLock)
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
				os.Exit(1)
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

// Limites e cliente do gerenciador de pacotes (variáveis para os testes ajustarem).
var (
	clientePacotes                  = &http.Client{Timeout: 60 * time.Second}
	limiteDownloadPacote      int64 = 100 << 20 // 100 MB por arquivo .zip
	limiteDescompressaoPacote int64 = 500 << 20 // 500 MB descomprimidos (zip bomb)
	rxNomePacote                    = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`)
)

// validarNomePacote impede que o nome da dependência (vindo do manifesto ou da linha de comando)
// tire a instalação de pt_modulos: "../../.ssh" escreveria fora do projeto.
func validarNomePacote(nome string) error {
	if !rxNomePacote.MatchString(nome) || strings.Contains(nome, "..") {
		return fmt.Errorf("nome de pacote inválido: %q (use letras, números, '.', '_' e '-')", nome)
	}
	return nil
}

// urlDownloadPermitida exige https (evita MITM), exceto em loopback (espelhos locais e testes).
func urlDownloadPermitida(bruta string) error {
	u, err := url.Parse(bruta)
	if err != nil {
		return fmt.Errorf("URL de pacote inválida: %v", err)
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && hostLocal(u.Host) {
		return nil
	}
	return fmt.Errorf("URL de pacote precisa usar https: %s", bruta)
}

// baixarEExtrairPacote baixa o .zip da dependência, valida o SHA-256 quando informado e extrai
// em ./pt_modulos/<nome>, com proteção contra zip slip, zip bomb e bits setuid.
func baixarEExtrairPacote(nome, urlPacote string, sha256Esperado string) (string, error) {
	if err := validarNomePacote(nome); err != nil {
		return "", err
	}
	if err := urlDownloadPermitida(urlPacote); err != nil {
		return "", err
	}

	fmt.Printf("  ➔ [%s] Baixando dependência de %s...\n", nome, urlPacote)
	resp, err := clientePacotes.Get(urlPacote)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status HTTP inválido: %s", resp.Status)
	}

	corpo, err := io.ReadAll(io.LimitReader(resp.Body, limiteDownloadPacote+1))
	if err != nil {
		return "", err
	}
	if int64(len(corpo)) > limiteDownloadPacote {
		return "", fmt.Errorf("pacote %s excede o limite de download de %d MB", nome, limiteDownloadPacote>>20)
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
	if err := os.MkdirAll(pastaDest, 0755); err != nil {
		return "", err
	}
	pastaLimpa := filepath.Clean(pastaDest)

	restante := limiteDescompressaoPacote
	for _, arquivo := range reader.File {
		// Recusa explícita de '..' e caminho absoluto: é o sanitizador que o CodeQL (go/zipslip)
		// reconhece; a checagem de prefixo abaixo continua como segunda barreira.
		if strings.Contains(arquivo.Name, "..") || filepath.IsAbs(arquivo.Name) {
			return "", fmt.Errorf("caminho de arquivo ilegal detectado no pacote ZIP: %s", arquivo.Name)
		}
		caminhoFisico := filepath.Join(pastaDest, arquivo.Name)

		// impede zip slip (path traversal pelos nomes das entradas)
		caminhoLimpo := filepath.Clean(caminhoFisico)
		if caminhoLimpo != pastaLimpa && !strings.HasPrefix(caminhoLimpo, pastaLimpa+string(filepath.Separator)) {
			return "", fmt.Errorf("caminho de arquivo ilegal detectado no pacote ZIP: %s", arquivo.Name)
		}

		if arquivo.FileInfo().IsDir() {
			if err := os.MkdirAll(caminhoFisico, 0755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(caminhoFisico), 0755); err != nil {
			return "", err
		}

		// nunca preserva setuid/setgid/sticky: só 0644, ou 0755 se o dono podia executar
		modo := os.FileMode(0644)
		if arquivo.Mode()&0100 != 0 {
			modo = 0755
		}

		origem, err := arquivo.Open()
		if err != nil {
			return "", err
		}
		destino, err := os.OpenFile(caminhoFisico, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, modo)
		if err != nil {
			origem.Close()
			return "", err
		}
		// conta o descomprimido de verdade (o tamanho declarado no zip pode mentir)
		n, err := io.Copy(destino, io.LimitReader(origem, restante+1))
		origem.Close()
		destino.Close()
		if err != nil {
			return "", err
		}
		if restante -= n; restante < 0 {
			_ = os.RemoveAll(pastaDest)
			return "", fmt.Errorf("pacote %s excede o limite de descompressão de %d MB (possível zip bomb)", nome, limiteDescompressaoPacote>>20)
		}
		_ = os.Chmod(caminhoFisico, modo) // umask/arquivo existente: garante o modo final
	}

	fmt.Printf("  ✓ [%s] Instalado com sucesso em ./pt_modulos/\n", nome)
	return hashObtido, nil
}

// registroPadrao é o índice oficial, num repositório do próprio projeto (a org GitHub "Harpia"
// pertence a terceiros). HARPIA_REGISTRO troca o índice (espelho ou registro privado).
const registroPadrao = "https://raw.githubusercontent.com/mat-dgruber/harpia-pacotes/main/pacotes.json"

// URL_REGISTRO_CENTRAL mantém compatibilidade: vale quando HARPIA_REGISTRO não está definido.
var URL_REGISTRO_CENTRAL = registroPadrao

func urlDoRegistro() string {
	if v := os.Getenv("HARPIA_REGISTRO"); v != "" {
		return v
	}
	return URL_REGISTRO_CENTRAL
}

// RegistroRemoto descreve o índice de pacotes (pacotes.json).
type RegistroRemoto struct {
	Pacotes map[string]struct {
		Versoes map[string]struct {
			URL    string `json:"url"`
			SHA256 string `json:"sha256"`
		} `json:"versoes"`
	} `json:"pacotes"`
}

// urlPermitidaNoRegistro: no índice oficial, só releases do próprio projeto; um índice
// customizado (HARPIA_REGISTRO) é responsabilidade de quem o configurou.
func urlPermitidaNoRegistro(urlPacote, registro string) bool {
	if registro != registroPadrao {
		return true
	}
	return strings.HasPrefix(urlPacote, "https://github.com/mat-dgruber/")
}

// obterPacoteDoRegistro resolve nome + restrição de versão ("", "ultimo", exata ou prefixo como
// "1" / "1.2") para a versão, a URL e o sha256 do índice. A escolha usa a maior versão semver
// (antes dependia da ordem aleatória de um map).
func obterPacoteDoRegistro(nome, versaoRestricao string) (string, string, string, error) {
	registro := urlDoRegistro()
	resp, err := clientePacotes.Get(registro)
	if err != nil {
		return "", "", "", fmt.Errorf("falha ao se comunicar com o registro de pacotes: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", "", "", fmt.Errorf("registro de pacotes indisponível (%s); instale por URL direta", registro)
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("falha no registro de pacotes (status %s)", resp.Status)
	}

	var reg RegistroRemoto
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&reg); err != nil {
		return "", "", "", fmt.Errorf("registro de pacotes retornou dados inválidos: %v", err)
	}
	pacote, existe := reg.Pacotes[nome]
	if !existe {
		return "", "", "", fmt.Errorf("pacote '%s' não encontrado no registro", nome)
	}

	var escolhida *semver.Version
	var escolhidaTexto string
	for versao := range pacote.Versoes {
		casa := versaoRestricao == "" || versaoRestricao == "latest" || versaoRestricao == "ultimo" ||
			versao == versaoRestricao || strings.HasPrefix(versao, versaoRestricao+".")
		if !casa {
			continue
		}
		v, errV := semver.NewVersion(versao)
		if errV != nil {
			continue
		}
		if escolhida == nil || v.GreaterThan(escolhida) {
			escolhida, escolhidaTexto = v, versao
		}
	}
	if escolhida == nil {
		if versaoRestricao == "" || versaoRestricao == "latest" || versaoRestricao == "ultimo" {
			return "", "", "", fmt.Errorf("pacote '%s' não tem nenhuma versão disponível no registro", nome)
		}
		return "", "", "", fmt.Errorf("versão '%s' do pacote '%s' não está registrada", versaoRestricao, nome)
	}

	dados := pacote.Versoes[escolhidaTexto]
	if dados.SHA256 == "" {
		return "", "", "", fmt.Errorf("o registro não informa sha256 para %s@%s: instalação recusada", nome, escolhidaTexto)
	}
	if !urlPermitidaNoRegistro(dados.URL, registro) {
		return "", "", "", fmt.Errorf("URL de %s@%s fora das releases do projeto: %s", nome, escolhidaTexto, dados.URL)
	}
	return escolhidaTexto, dados.URL, dados.SHA256, nil
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
					versao, urlPacote, sha, err := obterPacoteDoRegistro(nome, "")
					if err != nil {
						cmd.PrintErrf("Erro ao resolver pacote: %v\n", err)
						return
					}
					urlInstalar, shaEsperado, versaoRegistrada = urlPacote, sha, versao
				}
			} else if !strings.HasPrefix(versaoOuUrl, "http://") && !strings.HasPrefix(versaoOuUrl, "https://") {
				cmd.Printf("Buscando pacote '%s' na versão '%s' no registro remoto...\n", nome, versaoOuUrl)
				_, resolved, sha, err := obterPacoteDoRegistro(nome, versaoOuUrl)
				if err != nil {
					cmd.PrintErrf("Erro ao resolver pacote: %v\n", err)
					return
				}
				urlInstalar, shaEsperado = resolved, sha
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
			if filepath.Base(nome) != nome || nome == ".." || nome == "." {
				cmd.Println("Erro: nome de pacote inválido.")
				return
			}

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
