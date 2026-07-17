package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

// isWindows verifica se o sistema operacional atual é Windows.
//
// Esta função é importante para determinar as extensões de arquivos executáveis (.exe)
// e os formatos de compressão de arquivos que são suportados por padrão no sistema
// (ex: .zip para Windows vs .tar.gz para sistemas baseados em Unix).
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// httpClient é o cliente HTTP reutilizável configurado para requisições de rede.
//
// Um timeout explícito de 10 segundos foi configurado para evitar que a CLI fique travada
// indefinidamente caso o usuário esteja em uma rede instável ou lenta, ou quando houver
// problemas de latência nos servidores do GitHub.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// nomeOS retorna o nome padronizado do sistema operacional em formato compatível com
// as tags de build do GoReleaser publicadas no repositório GitHub.
//
// Esta padronização evita problemas de case-sensitivity ou nomes incorretos gerados
// pelo runtime (ex: "Darwin" para macOS, "Linux" para Linux e "Windows" para Windows).
// Caso o SO atual não seja reconhecido, assume "Linux" por questões de compatibilidade.
func nomeOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return "Linux"
	}
}

// nomeArch traduz a arquitetura de processador atual (runtime.GOARCH) para os nomes
// comumente utilizados nos arquivos de release binários pré-compilados do GitHub.
//
// Mapeia especificamente:
//   - "amd64" para "x86_64"
//   - "386" para "i386"
// Para outras arquiteturas (ex: arm64), retorna o próprio valor reportado pelo runtime Go.
func nomeArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "386":
		return "i386"
	default:
		return runtime.GOARCH
	}
}

// jaAtualizado compara a versão local atualmente instalada ('a') com uma versão remota ('b').
//
// Utiliza a especificação de Semantic Versioning (através da biblioteca masterminds/semver)
// para criar uma restrição lógica que verifica se a versão 'a' é menor que a versão 'b'.
//
// Retorna true se a versão 'a' for estritamente menor que 'b', indicando que uma
// atualização está de fato disponível no servidor.
func jaAtualizado(a, b string) bool {
	i, _ := semver.NewConstraint("< " + b)
	n, _ := semver.NewVersion(a)

	return i.Check(n)
}

// urlDaVersao reconstrói dinamicamente o link de download direto do release mais recente
// do Harpia baseado na arquitetura e sistema operacional do cliente atual.
//
// O link aponta para o diretório de downloads do repositório GitHub e anexa as extensões
// correspondentes (.zip para Windows e .tar.gz para as demais plataformas Unix-like).
func urlDaVersao() string {
	url := "https://github.com/mat-dgruber/Harpia/releases/latest/download/"
	url += nomeOS() + "_" + nomeArch()

	if isWindows() {
		return url + ".zip"
	}

	return url + ".tar.gz"
}

// Tag representa a estrutura mínima de dados de uma Tag do Git retornada pela API do GitHub.
// É mapeada diretamente a partir do JSON recebido da API de tags pública do repositório.
type Tag struct {
	// Name é o nome da tag do Git (ex: "v0.1.0").
	Name string `json:"name"`
}

// versaoInstalada executa o binário do Harpia especificado no caminho para descobrir
// a sua versão atual rodando o argumento `-v` ou `--version`.
//
// Retorna a string de versão (limpa de espaços e sem o prefixo do nome do programa) ou um
// erro detalhado caso ocorra falha na execução ou se a versão encontrada for "dev", indicando
// um build de desenvolvimento local que não pode ser atualizado de forma automatizada.
func versaoInstalada(binario string) (string, error) {
	comandoEx, err := exec.Command(binario, "-v").Output()
	if err != nil {
		return "", fmt.Errorf("erro ao obter a versão instalada, provavelmente você ainda não instalou nenhuma versão, veja: <https://github.com/mat-dgruber/Harpia/?tab=readme-ov-file#com-bash>")
	}
	parts := strings.Split(strings.Trim(string(comandoEx), " \t\n"), " ")
	v := parts[len(parts)-1]
	if v == "dev" {
		return v, fmt.Errorf("você tem a versão 'dev' instalada, este comando ainda não é capaz de atualizar nesse cenário")
	}
	return v, nil
}

// ultimaVersao consulta os endpoints públicos da API do GitHub para obter as tags de release do repositório.
//
// Analisa o retorno no formato JSON, extrai a primeira tag disponível (que corresponde ao release mais estável
// e recente devido à ordenação da API) e limpa o caractere de prefixo "v" caso esteja presente,
// retornando a string pura do Semantic Versioning (ex: "0.1.0").
func ultimaVersao() (string, error) {
	response, err := httpClient.Get("https://api.github.com/repos/mat-dgruber/Harpia/tags")
	if err != nil {
		return "", fmt.Errorf("erro ao obter as versões no repositório")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro na resposta do servidor: %s", response.Status)
	}

	var tags []Tag
	if err := json.NewDecoder(response.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("erro ao decodificar a resposta JSON")
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("nenhuma versão encontrada no repositório")
	}

	return strings.TrimPrefix(tags[0].Name, "v"), nil
}

// downloadEInstalar faz o download do pacote comprimido do Harpia em um arquivo temporário no sistema,
// exibe a barra de progresso no terminal usando a ferramenta nativa curl e, ao final, dispara o processo de
// descompactação do novo binário diretamente no diretório de destino correspondente à instalação.
//
// O arquivo temporário criado no disco é limpo e removido de forma garantida através do recurso defer.
func downloadEInstalar(raizHarpia string) error {
	f, err := os.CreateTemp("", "-ptst")
	if err != nil {
		return fmt.Errorf("erro ao criar um diretorio temporário")
	}
	defer os.Remove(f.Name())

	compactTemp := f.Name()

	fmt.Println("Baixando arquivos necessários")

	curl := exec.Command(
		"curl", "--fail", "--location", "--progress-bar", "--output", compactTemp, urlDaVersao(),
	)
	curl.Stdout = os.Stdout
	curl.Stderr = os.Stderr

	if err := curl.Run(); err != nil {
		return fmt.Errorf("falha ao baixar os arquivos")
	}

	fmt.Println("Instalando a nova versão...")

	return descompactar(compactTemp, raizHarpia)
}

// descompactar extrai o conteúdo do pacote compactado baixado temporariamente para o diretório
// final de execução do Harpia (raizHarpia).
//
// Esta função faz distinção inteligente entre sistemas operacionais:
//   - No Windows: tenta extrair usando o comando "unzip" do sistema ou, se não disponível, o "7z" (7-zip).
//   - Nos demais sistemas (Linux, macOS): executa o utilitário padrão do sistema "tar" com suporte a arquivos comprimidos.
func descompactar(compactTemp, raizHarpia string) error {
	if isWindows() {
		var cmd *exec.Cmd
		if _, err := exec.LookPath("unzip"); err == nil {
			cmd = exec.Command("unzip", "-d", raizHarpia, "-o", compactTemp)
		} else {
			cmd = exec.Command("7z", "x", "-o", raizHarpia, "-y", compactTemp)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("erro ao descompactar: %s", err)
		}
		return nil
	}

	cmd := exec.Command("tar", "-xf", compactTemp, "-C", raizHarpia)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao descompactar com tar: %s", err)
	}
	return nil
}

// atualize centraliza todo o fluxo do subcomando de atualização da CLI.
//
// O processo consiste em:
//  1. Descobrir a pasta home do usuário logado para localizar a instalação padrão em `~/.Harpia/bin/`;
//  2. Identificar qual a versão do Harpia está atualmente instalada localmente;
//  3. Buscar a versão mais recente do interpretador disponibilizada no repositório GitHub;
//  4. Comparar ambas as versões usando Semantic Versioning;
//  5. Se uma atualização estiver disponível, faz o download do pacote correspondente e o instala.
func atualize() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("erro ao tentar montar o caminho da versão atual: %s", err)
	}

	raizHarpia := path.Join(home, ".Harpia/bin/")
	binario := path.Join(raizHarpia, "Harpia")
	if isWindows() {
		binario += ".exe"
	}

	inst, err := versaoInstalada(binario)
	if err != nil {
		return err
	}

	remota, err := ultimaVersao()
	if err != nil {
		return err
	}

	if !jaAtualizado(inst, remota) {
		fmt.Printf("Você já tem a versão mais recente (%s) instalada.\n", inst)
		return nil
	}

	fmt.Printf("Nova versão disponível: %s\n", remota)
	if err := downloadEInstalar(raizHarpia); err != nil {
		return err
	}
	fmt.Println("Nova versão instalada com sucesso!")
	return nil
}

var _ = cobra.Command{}

// comandoAtualize cria e retorna o comando *cobra.Command para atualizar a CLI.
//
// Este comando é montado e registrado na raiz da árvore CLI em cmd.go através
// do método InstalarComandos. Quando acionado no terminal (`Harpia atualize`),
// dispara a lógica em RunE que executa a função de fluxo de atualização (atualize).
func comandoAtualize() *cobra.Command {
	return &cobra.Command{
		Use:   "atualize",
		Short: "Atualiza a CLI do Harpia",
		RunE: func(cmd *cobra.Command, args []string) error {
			return atualize()
		},
	}
}
