//go:build !js && !wasm

package cmd

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseManifestoJSON(t *testing.T) {
	jsonContent := []byte(`{
		"dependencias": {
			"renderizador": "https://github.com/mat-dgruber/Harpia-render/archive/refs/tags/v1.0.0.zip",
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

	if manifest.Dependencias["renderizador"] != "https://github.com/mat-dgruber/Harpia-render/archive/refs/tags/v1.0.0.zip" {
		t.Errorf("Valor incorreto para 'renderizador'")
	}
}

func TestParseManifestoHarpia(t *testing.T) {
	ptContent := []byte(`
# Manifesto em Harpia
var renderizador = "https://github.com/mat-dgruber/Harpia-render/archive/refs/tags/v1.0.0.zip"
const teste = 'http://exemplo.com/teste.zip'
`)

	manifest, err := parseManifesto(ptContent)
	if err != nil {
		t.Fatalf("Erro ao decodificar manifesto Harpia: %v", err)
	}

	if len(manifest.Dependencias) != 2 {
		t.Errorf("Esperava 2 dependências, obtive %d", len(manifest.Dependencias))
	}

	if manifest.Dependencias["renderizador"] != "https://github.com/mat-dgruber/Harpia-render/archive/refs/tags/v1.0.0.zip" {
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

	f, err := zipWriter.Create("main.hrp")
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
	tempDir, err := os.MkdirTemp("", "Harpia_pacotes_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	_, err = baixarEExtrairPacote("teste_modulo", server.URL, "")
	if err != nil {
		t.Fatalf("Erro ao baixar/extrair pacote: %v", err)
	}

	// Verifica se o arquivo extraído existe na pasta pt_modulos/teste_modulo/
	caminhoArquivo := filepath.Join("pt_modulos", "teste_modulo", "main.hrp")
	if _, err := os.Stat(caminhoArquivo); os.IsNotExist(err) {
		t.Errorf("Arquivo 'main.hrp' não foi extraído do zip do módulo")
	}

	conteudo, err := os.ReadFile(caminhoArquivo)
	if err != nil {
		t.Fatal(err)
	}

	if string(conteudo) != "imprimir('olá do pacote!')" {
		t.Errorf("Conteúdo do arquivo extraído incorreto: %s", string(conteudo))
	}
}

// TestObterUrlDoRegistro assevera a resolução de versões no índice de pacotes.
func TestObterUrlDoRegistro(t *testing.T) {
	regJson := `{
		"pacotes": {
			"banco-dados": {
				"versoes": {
					"1.0.0": {"url": "https://github.com/mat-dgruber/banco/releases/download/v1.0.0/banco.zip", "sha256": "aaa"},
					"1.1.0": {"url": "https://github.com/mat-dgruber/banco/releases/download/v1.1.0/banco.zip", "sha256": "bbb"}
				}
			}
		}
	}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(regJson))
	}))
	defer server.Close()
	t.Setenv("HARPIA_REGISTRO", server.URL)

	versao, urlResolvida, sha, err := obterPacoteDoRegistro("banco-dados", "1.0.0")
	if err != nil || versao != "1.0.0" || sha != "aaa" || !strings.HasSuffix(urlResolvida, "v1.0.0/banco.zip") {
		t.Fatalf("versão exata: %s %s %s (%v)", versao, urlResolvida, sha, err)
	}
	if versao, _, _, err := obterPacoteDoRegistro("banco-dados", ""); err != nil || versao != "1.1.0" {
		t.Fatalf("última versão deveria ser 1.1.0, veio %s (%v)", versao, err)
	}
	if _, _, _, err := obterPacoteDoRegistro("pacote-fantasma", "1.0.0"); err == nil {
		t.Errorf("Esperava erro ao buscar pacote fantasma")
	}
	if _, _, _, err := obterPacoteDoRegistro("banco-dados", "9.9.9"); err == nil {
		t.Errorf("Esperava erro ao buscar versão inexistente")
	}
}

func TestCalcularSHA256(t *testing.T) {
	dados := []byte("")
	esperado := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	obtido := calcularSHA256(dados)
	if obtido != esperado {
		t.Errorf("SHA256 para string vazia incorreto. Esperado: %s, obtido: %s", esperado, obtido)
	}

	dados = []byte("harpia")
	esperado = "8ddd29a5a18b3228c9b3c9a152954669c691f6870e1139dbd3d90037b640b55d"
	obtido = calcularSHA256(dados)
	if obtido != esperado {
		t.Errorf("SHA256 para 'harpia' incorreto. Esperado: %s, obtido: %s", esperado, obtido)
	}
}

func TestVerificacaoIntegridadeSHA256(t *testing.T) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	f, err := zipWriter.Create("main.hrp")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("imprimir('teste integrity')"))
	if err != nil {
		t.Fatal(err)
	}
	zipWriter.Close()

	zipBytes := buf.Bytes()
	hashCorreto := calcularSHA256(zipBytes)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(zipBytes)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// 1. Deve falhar com HRP-LOCK-001 se SHA-256 for divergente
	hashRetornado, err := baixarEExtrairPacote("pkg_test", server.URL, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("Esperava erro de integridade HRP-LOCK-001, mas a instalacao teve sucesso")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("HRP-LOCK-001")) {
		t.Errorf("Mensagem de erro esperada HRP-LOCK-001, obtido: %v", err)
	}

	// 2. Deve ter sucesso com hash correto
	hashRetornado, err = baixarEExtrairPacote("pkg_test", server.URL, hashCorreto)
	if err != nil {
		t.Fatalf("Erro inesperado com hash correto: %v", err)
	}
	if hashRetornado != hashCorreto {
		t.Errorf("Hash retornado incorreto. Esperado %s, obtido %s", hashCorreto, hashRetornado)
	}
}

func TestInstalacaoDeterministicaComLockfile(t *testing.T) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	f, err := zipWriter.Create("index.hrp")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("imprimir('locked pkg')"))
	if err != nil {
		t.Fatal(err)
	}
	zipWriter.Close()

	zipBytes := buf.Bytes()
	hashZip := calcularSHA256(zipBytes)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(zipBytes)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Cria pacote.json
	manifestContent := []byte(`{
		"dependencias": {
			"meu-modulo": "1.0.0"
		}
	}`)
	err = os.WriteFile("pacote.json", manifestContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Cria harpia.lock com URL travada no servidor mock
	lockContent := []byte(`{
		"versao_lock": "1.0",
		"pacotes": {
			"meu-modulo": {
				"versao": "1.0.0",
				"url": "` + server.URL + `",
				"sha256": "` + hashZip + `"
			}
		}
	}`)
	err = os.WriteFile("harpia.lock", lockContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := comandoInstalar()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("comandoInstalar falhou: %v", err)
	}

	// Verifica se arquivo foi extraído
	caminhoArquivo := filepath.Join("pt_modulos", "meu-modulo", "index.hrp")
	if _, err := os.Stat(caminhoArquivo); os.IsNotExist(err) {
		t.Errorf("Modulo instalado via lockfile nao foi encontrado em %s", caminhoArquivo)
	}
}

func TestComandoPacoteListarVazio(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cmd := comandoPacote()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"listar"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar pacote listar: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Nenhum pacote instalado")) {
		t.Errorf("saida esperada 'Nenhum pacote instalado', obtido: %s", buf.String())
	}
}

func TestComandoPacoteCicloVida(t *testing.T) {
	// Prepara zip em memória
	var bufZip bytes.Buffer
	zipWriter := zip.NewWriter(&bufZip)
	f, _ := zipWriter.Create("lib.hrp")
	f.Write([]byte("imprimir('lib')"))
	zipWriter.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(bufZip.Bytes())
	}))
	defer server.Close()

	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// 1. Testa adicionar pacote via URL
	cmdAdd := comandoPacote()
	outAdd := new(bytes.Buffer)
	cmdAdd.SetOut(outAdd)
	cmdAdd.SetArgs([]string{"adicionar", "pkg-teste", server.URL})
	if err := cmdAdd.Execute(); err != nil {
		t.Fatalf("erro ao adicionar pacote: %v", err)
	}

	caminhoExt := filepath.Join("pt_modulos", "pkg-teste", "lib.hrp")
	if _, err := os.Stat(caminhoExt); os.IsNotExist(err) {
		t.Fatalf("pacote nao extraido em %s", caminhoExt)
	}

	// 2. Testa listar
	cmdList := comandoPacote()
	outList := new(bytes.Buffer)
	cmdList.SetOut(outList)
	cmdList.SetArgs([]string{"listar"})
	if err := cmdList.Execute(); err != nil {
		t.Fatalf("erro ao listar pacotes: %v", err)
	}
	if !bytes.Contains(outList.Bytes(), []byte("pkg-teste")) {
		t.Errorf("lista nao contem 'pkg-teste': %s", outList.String())
	}

	// 3. Testa remover
	cmdRm := comandoPacote()
	outRm := new(bytes.Buffer)
	cmdRm.SetOut(outRm)
	cmdRm.SetArgs([]string{"remover", "pkg-teste"})
	if err := cmdRm.Execute(); err != nil {
		t.Fatalf("erro ao remover pacote: %v", err)
	}
	if _, err := os.Stat(caminhoExt); !os.IsNotExist(err) {
		t.Errorf("pasta do pacote ainda existe apos remover")
	}

	// 4. Testa limpar
	os.MkdirAll(filepath.Join("pt_modulos", "dummy"), 0755)
	cmdClean := comandoPacote()
	outClean := new(bytes.Buffer)
	cmdClean.SetOut(outClean)
	cmdClean.SetArgs([]string{"limpar"})
	if err := cmdClean.Execute(); err != nil {
		t.Fatalf("erro ao limpar pacotes: %v", err)
	}
	if _, err := os.Stat("pt_modulos"); !os.IsNotExist(err) {
		t.Errorf("pasta pt_modulos ainda existe apos limpar")
	}
}
