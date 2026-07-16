package cmd

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseManifestoJSON(t *testing.T) {
	jsonContent := []byte(`{
		"dependencias": {
			"renderizador": "https://github.com/natanfeitosa/portuscript-render/archive/refs/tags/v1.0.0.zip",
			"teste": "http://exemplo.com/teste.zip"
		}
	}`)

	manifest, err := parseManifesto(jsonContent)
	if err != nil {
		t.Fatalf("Erro ao decodificar manifesto JSON: %v", err)
	}

	if len(manifest.Dependencias) != 2 {
		t.Errorf("Esperava 2 dependências, obtive %d", len(manifest.Dependencias))
	}

	if manifest.Dependencias["renderizador"] != "https://github.com/natanfeitosa/portuscript-render/archive/refs/tags/v1.0.0.zip" {
		t.Errorf("Valor incorreto para 'renderizador'")
	}
}

func TestParseManifestoPortuscript(t *testing.T) {
	ptContent := []byte(`
# Manifesto em Portuscript
var renderizador = "https://github.com/natanfeitosa/portuscript-render/archive/refs/tags/v1.0.0.zip"
const teste = 'http://exemplo.com/teste.zip'
`)

	manifest, err := parseManifesto(ptContent)
	if err != nil {
		t.Fatalf("Erro ao decodificar manifesto Portuscript: %v", err)
	}

	if len(manifest.Dependencias) != 2 {
		t.Errorf("Esperava 2 dependências, obtive %d", len(manifest.Dependencias))
	}

	if manifest.Dependencias["renderizador"] != "https://github.com/natanfeitosa/portuscript-render/archive/refs/tags/v1.0.0.zip" {
		t.Errorf("Valor incorreto para 'renderizador'")
	}

	if manifest.Dependencias["teste"] != "http://exemplo.com/teste.zip" {
		t.Errorf("Valor incorreto para 'teste'")
	}
}

func TestBaixarEExtrairPacote(t *testing.T) {
	// Cria um arquivo zip em memória
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	f, err := zipWriter.Create("main.ptst")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("imprimir('olá do pacote!')"))
	if err != nil {
		t.Fatal(err)
	}
	zipWriter.Close()

	// Inicia um servidor HTTP mock para entregar o zip
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(buf.Bytes())
	}))
	defer server.Close()

	// Cria pasta temporária e muda o diretório de execução para evitar poluir o disco
	tempDir, err := os.MkdirTemp("", "portuscript_pacotes_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	err = baixarEExtrairPacote("teste_modulo", server.URL)
	if err != nil {
		t.Fatalf("Erro ao baixar/extrair pacote: %v", err)
	}

	// Verifica se o arquivo extraído existe na pasta pt_modulos/teste_modulo/
	caminhoArquivo := filepath.Join("pt_modulos", "teste_modulo", "main.ptst")
	if _, err := os.Stat(caminhoArquivo); os.IsNotExist(err) {
		t.Errorf("Arquivo 'main.ptst' não foi extraído do zip do módulo")
	}

	conteudo, err := os.ReadFile(caminhoArquivo)
	if err != nil {
		t.Fatal(err)
	}

	if string(conteudo) != "imprimir('olá do pacote!')" {
		t.Errorf("Conteúdo do arquivo extraído incorreto: %s", string(conteudo))
	}
}

// TestObterUrlDoRegistro assevera a correta resolução de URLs do registro de pacotes remoto
func TestObterUrlDoRegistro(t *testing.T) {
	// JSON de registro simulado
	regJson := `{
		"pacotes": {
			"banco-dados": {
				"versoes": {
					"1.0.0": {
						"url": "http://exemplo.com/banco-1.0.0.zip"
					},
					"1.1.0": {
						"url": "http://exemplo.com/banco-1.1.0.zip"
					}
				}
			}
		}
	}`

	// Inicia servidor mock para entregar o índice de pacotes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(regJson))
	}))
	defer server.Close()

	// Redireciona temporariamente a URL de registro para o servidor mock
	antigoURL := URL_REGISTRO_CENTRAL
	URL_REGISTRO_CENTRAL = server.URL
	defer func() { URL_REGISTRO_CENTRAL = antigoURL }()

	// 1. Testa resolução de versão exata
	urlResolvida, err := obterUrlDoRegistro("banco-dados", "1.0.0")
	if err != nil {
		t.Fatalf("Erro ao obter URL: %v", err)
	}
	if urlResolvida != "http://exemplo.com/banco-1.0.0.zip" {
		t.Errorf("URL resolvida incorreta para 1.0.0. Obtido: %s", urlResolvida)
	}

	// 2. Testa resolução sem versão informada (deve pegar a última estável disponível no map)
	urlUltimo, err := obterUrlDoRegistro("banco-dados", "")
	if err != nil {
		t.Fatalf("Erro ao obter última versão: %v", err)
	}
	if urlUltimo != "http://exemplo.com/banco-1.1.0.zip" && urlUltimo != "http://exemplo.com/banco-1.0.0.zip" {
		t.Errorf("URL resolvida inesperada para 'ultimo'. Obtido: %s", urlUltimo)
	}

	// 3. Testa erro de pacote inexistente
	_, err = obterUrlDoRegistro("pacote-fantasma", "1.0.0")
	if err == nil {
		t.Errorf("Esperava erro ao buscar pacote fantasma")
	}

	// 4. Testa erro de versão inexistente
	_, err = obterUrlDoRegistro("banco-dados", "9.9.9")
	if err == nil {
		t.Errorf("Esperava erro ao buscar versão inexistente")
	}
}

